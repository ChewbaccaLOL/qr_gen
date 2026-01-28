import json
import os
import sys
from typing import Dict, List, Optional, Tuple

import segno

from qr_export import prepare_cairo
from qr_generator import Variant, VARIANTS
from qr_render import render_svg


MAX_PREVIEW_MATRIX = 90
MAX_PREVIEW_DARK_MODULES = 9000
MAX_PREVIEW_INPUT_LEN = 512
PREVIEW_SKIPPED_MESSAGE = "Preview skipped for large data"
PREVIEW_TOO_LONG_MESSAGE = "Preview skipped for long input"
PRESET_FILE = "qr_presets.json"
DEFAULT_GUI_DATA = "https://example.com"

QT_AVAILABLE = False
QT_IMPORT_ERROR = None
try:  # pragma: no cover - optional runtime dependency
    from PySide6 import QtCore, QtGui, QtWidgets, QtSvg

    QT_AVAILABLE = True
except Exception as exc:  # pragma: no cover - optional runtime dependency
    QT_IMPORT_ERROR = exc
    QtCore = QtGui = QtWidgets = QtSvg = None


def _variant_from_preset(entry: Dict[str, str]) -> Optional[Variant]:
    name = str(entry.get("name", "")).strip()
    if not name:
        return None
    shape = str(entry.get("shape", "square")).strip() or "square"
    if shape not in ("square", "rounded", "dot"):
        shape = "square"
    dark = str(entry.get("dark", "#000000")).strip() or "#000000"
    light_value = entry.get("light")
    light = str(light_value).strip() if light_value is not None else None
    transparent = bool(entry.get("transparent", False))
    if transparent:
        light = None
    radius = entry.get("radius", 0.0)
    try:
        radius = float(radius)
    except (TypeError, ValueError):
        radius = 0.0
    gradient = entry.get("gradient")
    if isinstance(gradient, dict):
        if not gradient.get("from") or not gradient.get("to"):
            gradient = None
    else:
        gradient = None
    return Variant(name=name, shape=shape, dark=dark, light=light, radius=radius, gradient=gradient)


def _build_gradient(
    enabled: bool, color_from: Optional[str], color_to: Optional[str]
) -> Optional[Dict[str, str]]:
    if not enabled:
        return None
    color_from = (color_from or "").strip()
    color_to = (color_to or "").strip()
    if not color_from or not color_to:
        return None
    return {"id": "fg", "from": color_from, "to": color_to}


def _should_reload_variant(
    current: Optional[str], new: Optional[str], custom_dirty: bool
) -> bool:
    if not new:
        return False
    if current != new:
        return True
    return not custom_dirty


def load_presets(path: str = PRESET_FILE) -> List[Dict[str, str]]:
    try:
        with open(path, "r", encoding="utf-8") as handle:
            data = json.load(handle)
    except FileNotFoundError:
        return []
    except json.JSONDecodeError:
        return []
    if not isinstance(data, list):
        return []
    return [item for item in data if isinstance(item, dict)]


def save_presets(entries: List[Dict[str, str]], path: str = PRESET_FILE) -> bool:
    try:
        with open(path, "w", encoding="utf-8") as handle:
            json.dump(entries, handle, indent=2)
    except OSError:
        return False
    return True


def build_variant_catalog(preset_entries: List[Dict[str, str]]) -> Tuple[Dict[str, Variant], List[str]]:
    variant_map = dict(VARIANTS)
    variant_order = sorted(VARIANTS.keys())
    for entry in preset_entries:
        variant = _variant_from_preset(entry)
        if not variant:
            continue
        if variant.name in variant_map:
            continue
        variant_map[variant.name] = variant
        variant_order.append(variant.name)
    return variant_map, variant_order


def _count_dark_modules(matrix: List[List[int]]) -> int:
    return sum(1 for row in matrix for cell in row if cell)


def _preview_too_large(size: int, dark_modules: Optional[int] = None) -> bool:
    if size > MAX_PREVIEW_MATRIX:
        return True
    if dark_modules is None:
        return False
    return dark_modules > MAX_PREVIEW_DARK_MODULES


def _preview_too_long(data: str) -> bool:
    return len(data) > MAX_PREVIEW_INPUT_LEN


def _compute_preview_scale(matrix_size: int, target_px: int, border: int) -> float:
    total_modules = max(1, matrix_size + border * 2)
    return max(1.0, float(target_px) / total_modules)


def _default_output_path(variant_name: str, fmt: str) -> str:
    extension = "svg" if fmt.lower() == "svg" else fmt.lower()
    return os.path.join("out", f"qr-{variant_name}.{extension}")


def _strip_known_extension(name: str, extensions: Tuple[str, ...]) -> str:
    base = os.path.basename(name.strip())
    if not base:
        return ""
    lower = base.lower()
    for ext in extensions:
        suffix = f".{ext}"
        if lower.endswith(suffix):
            return base[: -len(suffix)]
    return base


def _clamp_zoom(value: float, minimum: float, maximum: float) -> float:
    return max(minimum, min(float(value), maximum))


def _import_cairosvg():
    try:
        prepare_cairo()
        import cairosvg

        return cairosvg, None
    except Exception:
        return None, (
            "PNG/PDF/PS export requires cairosvg + cairo (pip install cairosvg). "
            "On Windows, the Cairo DLLs must be bundled or installed."
        )


if QT_AVAILABLE:  # pragma: no cover - GUI code exercised manually

    class PreviewZoomDialog(QtWidgets.QDialog):
        def __init__(
            self,
            parent: QtWidgets.QWidget,
            *,
            svg_text: str,
            background_color: Optional[str],
            render_pixmap,
        ) -> None:
            super().__init__(parent)
            self.svg_text = svg_text
            self.background_color = background_color
            self.render_pixmap = render_pixmap
            self.zoom = 1.0
            self.min_zoom = 0.2
            self.max_zoom = 6.0
            self.base_fit_px: Optional[int] = None

            self.setWindowTitle("QR Preview")
            self.resize(940, 940)

            layout = QtWidgets.QVBoxLayout(self)
            layout.setContentsMargins(16, 16, 16, 16)
            layout.setSpacing(10)

            header = QtWidgets.QWidget()
            header_layout = QtWidgets.QHBoxLayout(header)
            header_layout.setContentsMargins(0, 0, 0, 0)
            header_layout.setSpacing(10)

            title = QtWidgets.QLabel("Preview")
            title.setStyleSheet("font-weight: 600;")
            hint = QtWidgets.QLabel("Scroll to zoom. Click anywhere to close.")
            hint.setStyleSheet("color: #666666;")
            close_btn = QtWidgets.QPushButton("X")
            close_btn.setFixedWidth(32)
            close_btn.clicked.connect(self.close)

            header_layout.addWidget(title)
            header_layout.addStretch(1)
            header_layout.addWidget(hint)
            header_layout.addWidget(close_btn)
            layout.addWidget(header)

            self.image_label = QtWidgets.QLabel()
            self.image_label.setAlignment(QtCore.Qt.AlignCenter)

            self.scroll = QtWidgets.QScrollArea()
            self.scroll.setWidget(self.image_label)
            self.scroll.setWidgetResizable(False)
            self.scroll.setAlignment(QtCore.Qt.AlignCenter)
            layout.addWidget(self.scroll, 1)

            for widget in (header, title, hint):
                widget.installEventFilter(self)
            self.scroll.viewport().installEventFilter(self)
            self.image_label.installEventFilter(self)

            self._render()

        def _device_pixel_ratio(self, widget: QtWidgets.QWidget) -> float:
            try:
                return float(widget.devicePixelRatioF())
            except AttributeError:
                return 1.0

        def _update_base_fit(self, force: bool = False) -> None:
            if self.base_fit_px is not None and not force and abs(self.zoom - 1.0) > 0.01:
                return
            viewport = self.scroll.viewport().size()
            base = min(viewport.width(), viewport.height())
            self.base_fit_px = max(240, int(base))

        def _render(self) -> None:
            if not self.svg_text:
                return
            if self.base_fit_px is None:
                self._update_base_fit(force=True)
            fit_px = self.base_fit_px or 240
            target_px = max(64, int(round(fit_px * self.zoom)))
            dpr = self._device_pixel_ratio(self.scroll.viewport())
            pixmap = self.render_pixmap(
                self.svg_text,
                target_px,
                self.background_color,
                dpr,
            )
            self.image_label.setPixmap(pixmap)
            self.image_label.adjustSize()

        def _apply_zoom(self, factor: float) -> None:
            self.zoom = _clamp_zoom(self.zoom * factor, self.min_zoom, self.max_zoom)
            self._render()

        def eventFilter(self, obj, event) -> bool:
            if event.type() == QtCore.QEvent.Wheel:
                delta = event.angleDelta().y()
                if delta:
                    self._apply_zoom(1.15 if delta > 0 else 1 / 1.15)
                    return True
            if event.type() == QtCore.QEvent.MouseButtonPress:
                self.close()
                return True
            return super().eventFilter(obj, event)

        def mousePressEvent(self, event: QtGui.QMouseEvent) -> None:
            self.close()
            event.accept()

        def resizeEvent(self, event: QtGui.QResizeEvent) -> None:
            super().resizeEvent(event)
            self._update_base_fit(force=True)
            self._render()


    class QrGuiApp(QtWidgets.QMainWindow):
        def __init__(self) -> None:
            super().__init__()
            self.setWindowTitle("QR Generator")
            self.setMinimumSize(1100, 900)
            self.resize(1140, 980)

            self.preview_icon_size = 110
            self.detail_preview_size = 240
            self.preview_cache: Dict[Tuple[str, int], QtGui.QPixmap] = {}
            self._initial_show = False

            self.preset_entries = load_presets()
            self.variant_map, self.variant_order = build_variant_catalog(self.preset_entries)
            self.selected_variant = self.variant_order[0]
            self.suspend_custom_update = False
            self.custom_dirty = False
            self.suspend_output_name_update = False
            self.output_name_dirty = False
            self.output_name_default = f"qr-{self.selected_variant}"

            self._build_ui()
            self._populate_variants()
            self._load_variant_into_custom_fields(self.selected_variant)
            self.selected_label.setText(self.selected_variant)
            self.data_edit.setPlainText(DEFAULT_GUI_DATA)
            self.schedule_preview_update()

        def showEvent(self, event: QtGui.QShowEvent) -> None:
            super().showEvent(event)
            if not self._initial_show:
                self._initial_show = True
                QtCore.QTimer.singleShot(0, self.refresh_previews)

        def _build_ui(self) -> None:
            root = QtWidgets.QWidget()
            layout = QtWidgets.QHBoxLayout(root)
            layout.setContentsMargins(18, 18, 18, 18)
            layout.setSpacing(18)

            self.variant_list = QtWidgets.QListWidget()
            self.variant_list.setViewMode(QtWidgets.QListView.IconMode)
            self.variant_list.setResizeMode(QtWidgets.QListView.Adjust)
            self.variant_list.setIconSize(QtCore.QSize(self.preview_icon_size, self.preview_icon_size))
            self.variant_list.setGridSize(QtCore.QSize(self.preview_icon_size + 30, self.preview_icon_size + 46))
            self.variant_list.setMovement(QtWidgets.QListView.Static)
            self.variant_list.setUniformItemSizes(True)
            self.variant_list.setSpacing(12)
            self.variant_list.itemSelectionChanged.connect(self._on_variant_selected)
            layout.addWidget(self.variant_list, 2)

            right = QtWidgets.QWidget()
            right_layout = QtWidgets.QVBoxLayout(right)
            right_layout.setSpacing(14)

            right_layout.addWidget(self._build_input_panel())
            right_layout.addWidget(self._build_selected_panel())
            right_layout.addWidget(self._build_customize_panel())
            right_layout.addWidget(self._build_preset_panel())
            right_layout.addWidget(self._build_export_panel())

            right_layout.addStretch(1)

            right_scroll = QtWidgets.QScrollArea()
            right_scroll.setWidgetResizable(True)
            right_scroll.setFrameShape(QtWidgets.QFrame.NoFrame)
            right_scroll.setWidget(right)
            layout.addWidget(right_scroll, 3)
            self.setCentralWidget(root)
            self.setStatusBar(QtWidgets.QStatusBar())

            self.preview_timer = QtCore.QTimer(self)
            self.preview_timer.setSingleShot(True)
            self.preview_timer.timeout.connect(self.refresh_previews)

        def _build_input_panel(self) -> QtWidgets.QGroupBox:
            panel = QtWidgets.QGroupBox("Input")
            layout = QtWidgets.QGridLayout(panel)
            layout.setColumnStretch(1, 1)

            layout.addWidget(QtWidgets.QLabel("Data"), 0, 0, QtCore.Qt.AlignTop)
            self.data_edit = QtWidgets.QPlainTextEdit()
            self.data_edit.setPlaceholderText("Text or URL to encode")
            self.data_edit.setFixedHeight(70)
            self.data_edit.textChanged.connect(self.schedule_preview_update)
            layout.addWidget(self.data_edit, 0, 1, 1, 2)

            self.live_preview_check = QtWidgets.QCheckBox("Live preview")
            self.live_preview_check.setChecked(True)
            self.live_preview_check.stateChanged.connect(self.schedule_preview_update)
            layout.addWidget(self.live_preview_check, 1, 1)

            layout.addWidget(QtWidgets.QLabel("Error"), 2, 0)
            self.error_combo = QtWidgets.QComboBox()
            self.error_combo.addItems(["l", "m", "q", "h"])
            self.error_combo.setCurrentText("m")
            self.error_combo.currentIndexChanged.connect(self.schedule_preview_update)
            layout.addWidget(self.error_combo, 2, 1)

            layout.addWidget(QtWidgets.QLabel("Scale"), 3, 0)
            self.scale_spin = QtWidgets.QSpinBox()
            self.scale_spin.setRange(1, 50)
            self.scale_spin.setValue(10)
            self.scale_spin.valueChanged.connect(self.schedule_preview_update)
            layout.addWidget(self.scale_spin, 3, 1)

            layout.addWidget(QtWidgets.QLabel("Border"), 4, 0)
            self.border_spin = QtWidgets.QSpinBox()
            self.border_spin.setRange(0, 20)
            self.border_spin.setValue(4)
            self.border_spin.valueChanged.connect(self.schedule_preview_update)
            layout.addWidget(self.border_spin, 4, 1)

            return panel

        def _build_selected_panel(self) -> QtWidgets.QGroupBox:
            panel = QtWidgets.QGroupBox("Selected style")
            layout = QtWidgets.QVBoxLayout(panel)

            self.selected_label = QtWidgets.QLabel("")
            layout.addWidget(self.selected_label)

            self.preview_label = QtWidgets.QLabel()
            self.preview_label.setFixedSize(self.detail_preview_size, self.detail_preview_size)
            self.preview_label.setAlignment(QtCore.Qt.AlignCenter)
            self.preview_label.setFrameStyle(QtWidgets.QFrame.Panel | QtWidgets.QFrame.Sunken)
            self.preview_label.setCursor(QtGui.QCursor(QtCore.Qt.PointingHandCursor))
            self.preview_label.mousePressEvent = self._on_preview_clicked
            layout.addWidget(self.preview_label, alignment=QtCore.Qt.AlignCenter)

            self.copy_png_button = QtWidgets.QPushButton("Copy PNG")
            self.copy_png_button.clicked.connect(self.copy_selected_png)
            layout.addWidget(self.copy_png_button, alignment=QtCore.Qt.AlignLeft)

            return panel

        def _build_customize_panel(self) -> QtWidgets.QGroupBox:
            panel = QtWidgets.QGroupBox("Customize")
            layout = QtWidgets.QGridLayout(panel)
            layout.setColumnStretch(1, 1)

            layout.addWidget(QtWidgets.QLabel("Shape"), 0, 0)
            self.shape_combo = QtWidgets.QComboBox()
            self.shape_combo.addItems(["square", "rounded", "dot"])
            self.shape_combo.currentIndexChanged.connect(self.on_custom_change)
            layout.addWidget(self.shape_combo, 0, 1)

            layout.addWidget(QtWidgets.QLabel("Dark"), 1, 0)
            self.dark_edit = QtWidgets.QLineEdit()
            self.dark_edit.textChanged.connect(self.on_custom_change)
            layout.addWidget(self.dark_edit, 1, 1)
            dark_btn = QtWidgets.QPushButton("Pick")
            dark_btn.clicked.connect(lambda: self.pick_color(self.dark_edit))
            layout.addWidget(dark_btn, 1, 2)

            layout.addWidget(QtWidgets.QLabel("Light"), 2, 0)
            self.light_edit = QtWidgets.QLineEdit()
            self.light_edit.textChanged.connect(self.on_custom_change)
            layout.addWidget(self.light_edit, 2, 1)
            light_btn = QtWidgets.QPushButton("Pick")
            light_btn.clicked.connect(lambda: self.pick_color(self.light_edit))
            layout.addWidget(light_btn, 2, 2)

            self.transparent_check = QtWidgets.QCheckBox("Transparent")
            self.transparent_check.stateChanged.connect(self.on_custom_change)
            layout.addWidget(self.transparent_check, 3, 1)

            layout.addWidget(QtWidgets.QLabel("Radius"), 4, 0)
            self.radius_spin = QtWidgets.QDoubleSpinBox()
            self.radius_spin.setRange(0.0, 0.5)
            self.radius_spin.setSingleStep(0.02)
            self.radius_spin.valueChanged.connect(self.on_custom_change)
            layout.addWidget(self.radius_spin, 4, 1)

            self.gradient_check = QtWidgets.QCheckBox("Gradient")
            self.gradient_check.stateChanged.connect(self.on_custom_change)
            layout.addWidget(self.gradient_check, 5, 1)

            layout.addWidget(QtWidgets.QLabel("From"), 6, 0)
            self.gradient_from_edit = QtWidgets.QLineEdit()
            self.gradient_from_edit.textChanged.connect(self.on_custom_change)
            layout.addWidget(self.gradient_from_edit, 6, 1)
            gradient_from_btn = QtWidgets.QPushButton("Pick")
            gradient_from_btn.clicked.connect(lambda: self.pick_color(self.gradient_from_edit))
            layout.addWidget(gradient_from_btn, 6, 2)

            layout.addWidget(QtWidgets.QLabel("To"), 7, 0)
            self.gradient_to_edit = QtWidgets.QLineEdit()
            self.gradient_to_edit.textChanged.connect(self.on_custom_change)
            layout.addWidget(self.gradient_to_edit, 7, 1)
            gradient_to_btn = QtWidgets.QPushButton("Pick")
            gradient_to_btn.clicked.connect(lambda: self.pick_color(self.gradient_to_edit))
            layout.addWidget(gradient_to_btn, 7, 2)

            reset_btn = QtWidgets.QPushButton("Reset to default")
            reset_btn.clicked.connect(self.reset_custom_settings)
            layout.addWidget(reset_btn, 8, 1, 1, 2, QtCore.Qt.AlignLeft)

            return panel

        def _build_preset_panel(self) -> QtWidgets.QGroupBox:
            panel = QtWidgets.QGroupBox("Presets")
            layout = QtWidgets.QGridLayout(panel)
            layout.setColumnStretch(1, 1)

            layout.addWidget(QtWidgets.QLabel("Name"), 0, 0)
            self.preset_name_edit = QtWidgets.QLineEdit()
            layout.addWidget(self.preset_name_edit, 0, 1)
            save_btn = QtWidgets.QPushButton("Save preset")
            save_btn.clicked.connect(self.save_preset)
            layout.addWidget(save_btn, 0, 2)

            return panel

        def _build_export_panel(self) -> QtWidgets.QGroupBox:
            panel = QtWidgets.QGroupBox("Export")
            layout = QtWidgets.QGridLayout(panel)
            layout.setColumnStretch(1, 1)

            layout.addWidget(QtWidgets.QLabel("Format"), 0, 0)
            self.format_combo = QtWidgets.QComboBox()
            self.format_combo.addItems(["SVG", "PNG", "PDF", "PS"])
            self.format_combo.currentIndexChanged.connect(self._on_format_change)
            layout.addWidget(self.format_combo, 0, 1)

            layout.addWidget(QtWidgets.QLabel("PNG scale"), 1, 0)
            self.png_scale_spin = QtWidgets.QDoubleSpinBox()
            self.png_scale_spin.setRange(0.5, 10.0)
            self.png_scale_spin.setSingleStep(0.5)
            self.png_scale_spin.setValue(3.0)
            layout.addWidget(self.png_scale_spin, 1, 1)

            layout.addWidget(QtWidgets.QLabel("Location"), 2, 0)
            output_row = QtWidgets.QWidget()
            output_layout = QtWidgets.QGridLayout(output_row)
            output_layout.setContentsMargins(0, 0, 0, 0)
            output_layout.setColumnStretch(0, 1)
            output_layout.setColumnStretch(1, 0)

            self.output_dir_edit = QtWidgets.QLineEdit("out")
            self.output_dir_edit.setPlaceholderText("Folder")
            output_layout.addWidget(self.output_dir_edit, 0, 0)
            browse_btn = QtWidgets.QPushButton("Browse")
            browse_btn.clicked.connect(self.browse_output)
            output_layout.addWidget(browse_btn, 0, 1)

            name_row = QtWidgets.QWidget()
            name_layout = QtWidgets.QHBoxLayout(name_row)
            name_layout.setContentsMargins(0, 0, 0, 0)
            self.output_name_edit = QtWidgets.QLineEdit(self.output_name_default)
            self.output_name_edit.setPlaceholderText("File name")
            self.output_name_edit.textChanged.connect(self._on_output_name_changed)
            name_layout.addWidget(self.output_name_edit, 1)
            self.output_ext_label = QtWidgets.QLabel(".svg")
            name_layout.addWidget(self.output_ext_label)

            layout.addWidget(output_row, 2, 1, 1, 2)

            export_btn = QtWidgets.QPushButton("Export")
            export_btn.clicked.connect(self.export_output)
            layout.addWidget(QtWidgets.QLabel("File name"), 3, 0)
            layout.addWidget(name_row, 3, 1, 1, 2)
            layout.addWidget(export_btn, 4, 1, 1, 2, QtCore.Qt.AlignLeft)

            self._on_format_change()
            return panel

        def _populate_variants(self) -> None:
            self.variant_list.clear()
            self.variant_items = {}
            for name in self.variant_order:
                item = QtWidgets.QListWidgetItem(name)
                item.setData(QtCore.Qt.UserRole, name)
                self.variant_list.addItem(item)
                self.variant_items[name] = item

            if self.variant_items:
                first_item = self.variant_items[self.selected_variant]
                self.variant_list.setCurrentItem(first_item)

        def _load_variant_into_custom_fields(self, name: str) -> None:
            variant = self.variant_map.get(name)
            if not variant:
                return
            self.suspend_custom_update = True
            self.shape_combo.setCurrentText(variant.shape)
            self.dark_edit.setText(variant.dark)
            if variant.light is None:
                self.light_edit.setText("")
                self.transparent_check.setChecked(True)
            else:
                self.light_edit.setText(variant.light)
                self.transparent_check.setChecked(False)
            self.radius_spin.setValue(float(variant.radius))
            if variant.gradient:
                self.gradient_check.setChecked(True)
                self.gradient_from_edit.setText(variant.gradient.get("from", ""))
                self.gradient_to_edit.setText(variant.gradient.get("to", ""))
            else:
                self.gradient_check.setChecked(False)
                self.gradient_from_edit.setText("")
                self.gradient_to_edit.setText("")
            self.suspend_custom_update = False
            self.custom_dirty = False

        def _on_variant_selected(self) -> None:
            current = self.variant_list.currentItem()
            if not current:
                return
            name = current.data(QtCore.Qt.UserRole)
            if not name:
                return
            if not _should_reload_variant(self.selected_variant, name, self.custom_dirty):
                self.selected_label.setText(name)
                return
            self.selected_variant = name
            self.selected_label.setText(name)
            self._load_variant_into_custom_fields(name)
            self._set_output_name_default()
            self.schedule_preview_update()

        def _on_format_change(self) -> None:
            fmt = self.format_combo.currentText().lower()
            show_png = fmt == "png"
            self.png_scale_spin.setVisible(show_png)
            extension = self._output_extension(fmt)
            self.output_ext_label.setText(f".{extension}")

        def _output_extension(self, fmt: str) -> str:
            return "svg" if fmt.lower() == "svg" else fmt.lower()

        def _set_output_name_default(self) -> None:
            default_name = f"qr-{self.selected_variant}"
            self.output_name_default = default_name
            if self.output_name_dirty and self.output_name_edit.text().strip():
                return
            self.suspend_output_name_update = True
            self.output_name_edit.setText(default_name)
            self.suspend_output_name_update = False
            self.output_name_dirty = False

        def _on_output_name_changed(self) -> None:
            if self.suspend_output_name_update:
                return
            self.output_name_dirty = True

        def _current_output_path(self) -> str:
            fmt = self.format_combo.currentText().lower()
            extension = self._output_extension(fmt)
            output_dir = self.output_dir_edit.text().strip() or "out"
            filename = self.output_name_edit.text().strip()
            if not filename:
                filename = self.output_name_default
            filename = _strip_known_extension(
                filename, ("svg", "png", "pdf", "ps")
            )
            if not filename:
                filename = self.output_name_default
            return os.path.join(output_dir, f"{filename}.{extension}")

        def pick_color(self, target: QtWidgets.QLineEdit) -> None:
            current = target.text().strip()
            initial = QtGui.QColor(current) if current else QtGui.QColor("#000000")
            if not initial.isValid():
                initial = QtGui.QColor("#000000")
            dialog = QtWidgets.QColorDialog(initial, self)
            dialog.setOption(QtWidgets.QColorDialog.ShowAlphaChannel, False)
            dialog.setCurrentColor(initial)
            if dialog.exec() == QtWidgets.QDialog.Accepted:
                picked = dialog.selectedColor()
                if picked.isValid():
                    target.setText(picked.name(QtGui.QColor.HexRgb))

        def on_custom_change(self) -> None:
            if self.suspend_custom_update:
                return
            if self.shape_combo.currentText() == "rounded" and self.radius_spin.value() <= 0:
                self.suspend_custom_update = True
                self.radius_spin.setValue(0.28)
                self.suspend_custom_update = False
            self.custom_dirty = True
            self.schedule_preview_update()

        def reset_custom_settings(self) -> None:
            self._load_variant_into_custom_fields(self.selected_variant)
            self.schedule_preview_update()

        def build_custom_variant(self, base: Variant) -> Variant:
            shape = self.shape_combo.currentText() or base.shape
            dark = self.dark_edit.text().strip() or base.dark
            if self.transparent_check.isChecked():
                light = None
            else:
                light = self.light_edit.text().strip() or base.light or "#ffffff"
            try:
                radius = float(self.radius_spin.value())
            except (TypeError, ValueError):
                radius = base.radius
            gradient = _build_gradient(
                self.gradient_check.isChecked(),
                self.gradient_from_edit.text(),
                self.gradient_to_edit.text(),
            )
            return Variant(
                name=base.name,
                shape=shape,
                dark=dark,
                light=light,
                radius=radius,
                gradient=gradient,
            )

        def schedule_preview_update(self) -> None:
            self.preview_timer.start(120)

        def refresh_previews(self) -> None:
            self.preview_cache.clear()
            data = self.data_edit.toPlainText().strip()
            if not data:
                self._set_preview_message("Enter text")
                self._clear_variant_icons()
                return
            if _preview_too_long(data):
                self._set_preview_message(PREVIEW_TOO_LONG_MESSAGE)
                self._clear_variant_icons()
                return
            try:
                qr = segno.make(data, error=self.error_combo.currentText())
            except Exception as exc:
                self._set_preview_message(f"Invalid data: {exc}")
                self._clear_variant_icons()
                return

            size = len(qr.matrix)
            dark_modules = _count_dark_modules(qr.matrix)
            if _preview_too_large(size, dark_modules):
                self._set_preview_message(PREVIEW_SKIPPED_MESSAGE)
                self._clear_variant_icons()
                return

            if not self.live_preview_check.isChecked():
                self._set_preview_message("Preview paused")
                self._clear_variant_icons()
                return

            self._render_variant_icons(qr.matrix)
            self._render_selected_preview(qr.matrix)

        def _set_preview_message(self, message: str) -> None:
            self.preview_label.setText(message)
            self.preview_label.setPixmap(QtGui.QPixmap())

        def _clear_variant_icons(self) -> None:
            for item in self.variant_items.values():
                item.setIcon(QtGui.QIcon())

        def _render_variant_icons(self, matrix: List[List[int]]) -> None:
            target_px = self.preview_icon_size
            scale = _compute_preview_scale(len(matrix), target_px, self.border_spin.value())
            dpr = self._device_pixel_ratio(self.variant_list.viewport())
            for name in self.variant_order:
                variant = self.variant_map.get(name)
                if not variant:
                    continue
                svg = render_svg(
                    matrix,
                    scale=scale,
                    border=self.border_spin.value(),
                    dark=variant.dark,
                    light=variant.light,
                    shape=variant.shape,
                    radius=variant.radius,
                    gradient=variant.gradient,
                )
                background = variant.light if variant.light is not None else None
                pixmap = self._render_svg_to_pixmap(
                    svg,
                    target_px,
                    background_color=background,
                    device_pixel_ratio=dpr,
                )
                self.variant_items[name].setIcon(QtGui.QIcon(pixmap))
            self.variant_list.doItemsLayout()
            self.variant_list.viewport().update()

        def _render_selected_preview(self, matrix: List[List[int]]) -> None:
            base_variant = self.variant_map.get(self.selected_variant)
            if not base_variant:
                return
            variant = self.build_custom_variant(base_variant)
            target_px = min(self.preview_label.width(), self.preview_label.height())
            if target_px <= 0:
                target_px = self.detail_preview_size
            scale = _compute_preview_scale(len(matrix), target_px, self.border_spin.value())
            svg = render_svg(
                matrix,
                scale=scale,
                border=self.border_spin.value(),
                dark=variant.dark,
                light=variant.light,
                shape=variant.shape,
                radius=variant.radius,
                gradient=variant.gradient,
            )
            background = variant.light if variant.light is not None else None
            pixmap = self._render_svg_to_pixmap(
                svg,
                target_px,
                background_color=background,
                device_pixel_ratio=self._device_pixel_ratio(self.preview_label),
            )
            self.preview_label.setPixmap(pixmap)
            self.preview_label.setText("")

        def _on_preview_clicked(self, _event: QtGui.QMouseEvent) -> None:
            svg_text = self.get_variant_svg(self.selected_variant, use_custom=True)
            if not svg_text:
                return
            base_variant = self.variant_map.get(self.selected_variant)
            if not base_variant:
                return
            variant = self.build_custom_variant(base_variant)
            dialog = PreviewZoomDialog(
                self,
                svg_text=svg_text,
                background_color=variant.light,
                render_pixmap=self._render_svg_to_pixmap,
            )
            dialog.exec()

        def _render_svg_to_pixmap(
            self,
            svg_text: str,
            size: int,
            background_color: Optional[str],
            device_pixel_ratio: float,
        ) -> QtGui.QPixmap:
            dpr = max(1.0, float(device_pixel_ratio or 1.0))
            key = (svg_text, size, round(dpr, 2), background_color or "transparent")
            cached = self.preview_cache.get(key)
            if cached is not None:
                return cached
            renderer = QtSvg.QSvgRenderer(bytearray(svg_text.encode("utf-8")))
            pixel_size = max(1, int(round(size * dpr)))
            image = QtGui.QImage(pixel_size, pixel_size, QtGui.QImage.Format_ARGB32)
            if background_color is None:
                self._paint_checkerboard(image, pixel_size)
            else:
                image.fill(QtGui.QColor(background_color))
            painter = QtGui.QPainter(image)
            renderer.render(painter, QtCore.QRectF(0, 0, pixel_size, pixel_size))
            painter.end()
            pixmap = QtGui.QPixmap.fromImage(image)
            pixmap.setDevicePixelRatio(dpr)
            self.preview_cache[key] = pixmap
            return pixmap

        def _paint_checkerboard(self, image: QtGui.QImage, size: int) -> None:
            image.fill(QtCore.Qt.white)
            painter = QtGui.QPainter(image)
            tile = 8
            color_a = QtGui.QColor("#e6e6e6")
            color_b = QtGui.QColor("#cfcfcf")
            for y in range(0, size, tile):
                for x in range(0, size, tile):
                    painter.fillRect(
                        QtCore.QRect(x, y, tile, tile),
                        color_a if (x // tile + y // tile) % 2 == 0 else color_b,
                    )
            painter.end()

        def _device_pixel_ratio(self, widget: QtWidgets.QWidget) -> float:
            try:
                return float(widget.devicePixelRatioF())
            except AttributeError:
                return 1.0

        def copy_selected_svg(self) -> None:
            svg_text = self.get_variant_svg(self.selected_variant, use_custom=True)
            if not svg_text:
                return
            QtWidgets.QApplication.clipboard().setText(svg_text)
            self.show_status(f"Copied {self.selected_variant} SVG to clipboard")

        def copy_selected_png(self) -> None:
            svg_text = self.get_variant_svg(self.selected_variant, use_custom=True)
            if not svg_text:
                return
            renderer = QtSvg.QSvgRenderer(bytearray(svg_text.encode("utf-8")))
            size = renderer.defaultSize()
            if not size.isValid():
                size = QtCore.QSize(self.detail_preview_size, self.detail_preview_size)
            image = QtGui.QImage(size, QtGui.QImage.Format_ARGB32)
            image.fill(QtCore.Qt.transparent)
            painter = QtGui.QPainter(image)
            renderer.render(painter)
            painter.end()
            QtWidgets.QApplication.clipboard().setImage(image)
            self.show_status(f"Copied {self.selected_variant} PNG to clipboard")

        def save_preset(self) -> None:
            name = self.preset_name_edit.text().strip()
            if not name:
                self.show_status("Preset name is required")
                return
            if name in self.variant_map:
                self.show_status("Preset name already exists")
                return
            base_variant = self.variant_map.get(self.selected_variant)
            if not base_variant:
                return
            variant = self.build_custom_variant(base_variant)
            variant = Variant(
                name=name,
                shape=variant.shape,
                dark=variant.dark,
                light=variant.light,
                radius=variant.radius,
                gradient=variant.gradient,
            )
            entry = {
                "name": variant.name,
                "shape": variant.shape,
                "dark": variant.dark,
                "light": variant.light,
                "transparent": variant.light is None,
                "radius": variant.radius,
                "gradient": variant.gradient,
            }
            self.preset_entries.append(entry)
            if not save_presets(self.preset_entries):
                self.show_status("Failed to save presets")
                return
            self.variant_map[variant.name] = variant
            self.variant_order.append(variant.name)
            self.preset_name_edit.setText("")
            self._populate_variants()
            self.selected_variant = variant.name
            self._load_variant_into_custom_fields(variant.name)
            self.schedule_preview_update()

        def get_variant_svg(self, name: str, use_custom: bool = False) -> Optional[str]:
            data = self.data_edit.toPlainText().strip()
            if not data:
                self.show_status("Enter data first")
                return None
            try:
                qr = segno.make(data, error=self.error_combo.currentText())
            except Exception as exc:
                self.show_status(f"Invalid data: {exc}")
                return None
            variant = self.variant_map.get(name)
            if not variant:
                return None
            if use_custom and name == self.selected_variant:
                variant = self.build_custom_variant(variant)
            return render_svg(
                qr.matrix,
                scale=self.scale_spin.value(),
                border=self.border_spin.value(),
                dark=variant.dark,
                light=variant.light,
                shape=variant.shape,
                radius=variant.radius,
                gradient=variant.gradient,
            )

        def browse_output(self) -> None:
            initial = self.output_dir_edit.text().strip() or "out"
            selected = QtWidgets.QFileDialog.getExistingDirectory(self, "Select output folder", initial)
            if selected:
                self.output_dir_edit.setText(selected)

        def export_output(self) -> None:
            fmt = self.format_combo.currentText().lower()
            output_path = self._current_output_path()
            svg_text = self.get_variant_svg(self.selected_variant, use_custom=True)
            if not svg_text:
                return
            os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)

            if fmt == "svg":
                try:
                    with open(output_path, "w", encoding="utf-8") as handle:
                        handle.write(svg_text)
                except OSError as exc:
                    self.show_status(f"Failed to save SVG: {exc}")
                    return
                self.show_status(f"Saved {output_path}")
                return

            cairosvg, message = _import_cairosvg()
            if not cairosvg:
                self.show_status(message)
                QtWidgets.QMessageBox.warning(self, "Missing dependency", message)
                return

            try:
                if fmt == "png":
                    cairosvg.svg2png(
                        bytestring=svg_text.encode("utf-8"),
                        write_to=output_path,
                        scale=self.png_scale_spin.value(),
                    )
                elif fmt == "pdf":
                    cairosvg.svg2pdf(bytestring=svg_text.encode("utf-8"), write_to=output_path)
                elif fmt == "ps":
                    cairosvg.svg2ps(bytestring=svg_text.encode("utf-8"), write_to=output_path)
                else:
                    self.show_status(f"Unsupported format: {fmt}")
                    return
            except Exception as exc:
                self.show_status(f"Export failed: {exc}")
                return
            self.show_status(f"Saved {output_path}")

        def show_status(self, message: str) -> None:
            status = self.statusBar()
            if status is None:
                return
            status.showMessage(message, 2200)



def main() -> None:
    if not QT_AVAILABLE:
        message = (
            "PySide6 is required for the Qt GUI. "
            "Install it with: pip install pyside6\n"
        )
        if QT_IMPORT_ERROR:
            message += f"\nImport error: {QT_IMPORT_ERROR}"
        print(message, file=sys.stderr)
        return
    app = QtWidgets.QApplication(sys.argv)
    window = QrGuiApp()
    window.show()
    sys.exit(app.exec())


if __name__ == "__main__":
    main()
