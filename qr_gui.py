import base64
import json
import os
import tkinter as tk
from tkinter import colorchooser, filedialog, messagebox, ttk
from typing import Dict, List, Optional

import segno

from qr_animation import build_float_gif_frames, build_wave_gif_frames
from qr_generator import (
    ANIMATION_VARIANTS,
    DEFAULT_FLOAT_ANGLE,
    DEFAULT_FLOAT_JAGGED_SNAP,
    DEFAULT_FLOAT_TILT,
    DEFAULT_GIF_FPS,
    DEFAULT_GIF_FRAMES,
    DEFAULT_GIF_HOLD,
    DEFAULT_WAVE_AMP,
    DEFAULT_WAVE_PERIOD,
    Variant,
    VARIANTS,
)
from qr_export import write_pdf, write_png, write_ps
from qr_render import render_svg


UI_SCALE = 1.25


def ui(value: float) -> int:
    return int(round(value * UI_SCALE))


BASE_PREVIEW_SIZE = ui(148)
DETAIL_PREVIEW_SIZE = ui(220)
CHECKER_SIZE = max(6, ui(10))
CARD_COLUMNS = 3
PRESET_FILE = "qr_presets.json"
THEMES = {
    "system": {
        "bg": "#f3f2ef",
        "surface": "#ffffff",
        "surface_border": "#d6d6d6",
        "text": "#1a1a1a",
        "muted": "#6a6a6a",
        "accent": "#1c6fe8",
        "preview_bg": "#f0f0f0",
        "checker_a": "#e6e6e6",
        "checker_b": "#cfcfcf",
    },
    "light": {
        "bg": "#f3f2ef",
        "surface": "#ffffff",
        "surface_border": "#d6d6d6",
        "text": "#1a1a1a",
        "muted": "#6a6a6a",
        "accent": "#1c6fe8",
        "preview_bg": "#f0f0f0",
        "checker_a": "#e6e6e6",
        "checker_b": "#cfcfcf",
    },
    "dark": {
        "bg": "#1f2124",
        "surface": "#2a2d31",
        "surface_border": "#3a3d42",
        "text": "#f2f2f2",
        "muted": "#b5b5b5",
        "accent": "#7bb5ff",
        "preview_bg": "#24272b",
        "checker_a": "#3a3d42",
        "checker_b": "#2f3237",
    },
}


def try_svg_to_png_bytes(svg_text: str, scale: float = 1.0) -> bytes:
    try:
        import cairosvg
    except ImportError as exc:
        raise RuntimeError("cairosvg is required for previews and PNG export") from exc
    return cairosvg.svg2png(bytestring=svg_text.encode("utf-8"), scale=scale)


class ScrollableFrame(ttk.Frame):
    def __init__(self, parent: tk.Widget, zoom_callback=None):
        super().__init__(parent)
        self.zoom_callback = zoom_callback
        self.canvas = tk.Canvas(self, highlightthickness=0)
        self.scrollbar = ttk.Scrollbar(self, orient="vertical", command=self.canvas.yview)
        self.inner = ttk.Frame(self.canvas)
        self.inner_id = self.canvas.create_window((0, 0), window=self.inner, anchor="nw")
        self.canvas.configure(yscrollcommand=self.scrollbar.set)

        self.canvas.grid(row=0, column=0, sticky="nsew")
        self.scrollbar.grid(row=0, column=1, sticky="ns")
        self.grid_rowconfigure(0, weight=1)
        self.grid_columnconfigure(0, weight=1)

        self.inner.bind("<Configure>", self._on_inner_configure)
        self.canvas.bind("<Configure>", self._on_canvas_configure)
        self.canvas.bind("<MouseWheel>", self._on_mousewheel)
        self.canvas.bind("<Control-MouseWheel>", self._on_control_mousewheel)
        self.canvas.bind("<Button-4>", self._on_button4)
        self.canvas.bind("<Button-5>", self._on_button5)
        self.canvas.bind("<Control-Button-4>", self._on_control_button4)
        self.canvas.bind("<Control-Button-5>", self._on_control_button5)

    def _on_inner_configure(self, _event: tk.Event) -> None:
        self.canvas.configure(scrollregion=self.canvas.bbox("all"))

    def _on_canvas_configure(self, event: tk.Event) -> None:
        self.canvas.itemconfigure(self.inner_id, width=event.width)

    def _on_mousewheel(self, event: tk.Event) -> None:
        if event.delta:
            self.canvas.yview_scroll(int(-1 * (event.delta / 120)), "units")

    def _on_control_mousewheel(self, event: tk.Event) -> None:
        if self.zoom_callback and event.delta:
            self.zoom_callback(1 if event.delta > 0 else -1)

    def _on_button4(self, _event: tk.Event) -> None:
        self.canvas.yview_scroll(-1, "units")

    def _on_button5(self, _event: tk.Event) -> None:
        self.canvas.yview_scroll(1, "units")

    def _on_control_button4(self, _event: tk.Event) -> None:
        if self.zoom_callback:
            self.zoom_callback(1)

    def _on_control_button5(self, _event: tk.Event) -> None:
        if self.zoom_callback:
            self.zoom_callback(-1)

    def apply_theme(self, colors: Dict[str, str]) -> None:
        self.canvas.configure(bg=colors["bg"])


class VariantCard:
    def __init__(self, app: "QrGuiApp", parent: tk.Widget, name: str, variant: Variant):
        self.app = app
        self.name = name
        self.variant = variant
        self.photo: Optional[tk.PhotoImage] = None
        self.cached_svg: Optional[str] = None
        self._hover_count = 0
        self.preview_size = app.preview_size

        self.frame = tk.Frame(
            parent,
            highlightthickness=2,
            highlightbackground=app.colors["surface_border"],
            background=app.colors["surface"],
        )
        self.canvas = tk.Canvas(
            self.frame,
            width=self.preview_size,
            height=self.preview_size,
            highlightthickness=0,
            background=app.colors["preview_bg"],
        )
        self.canvas.grid(row=0, column=0, padx=10, pady=(10, 6))
        self.label = ttk.Label(self.frame, text=name)
        self.label.grid(row=1, column=0, padx=10, pady=(0, 10))

        icon_size = ui(22)
        self.copy_icon = tk.Canvas(
            self.frame,
            width=icon_size,
            height=icon_size,
            highlightthickness=0,
            background=app.colors["surface"],
        )
        self._draw_copy_icon()
        self.copy_icon.place_forget()

        for widget in (self.frame, self.canvas, self.label, self.copy_icon):
            widget.bind("<Enter>", self._hover_enter, add="+")
            widget.bind("<Leave>", self._hover_leave, add="+")
            widget.bind("<Button-1>", self._select)
        self.copy_icon.bind("<Button-1>", lambda _event: self.copy())

    def _select(self, _event: tk.Event) -> None:
        self.app.select_variant(self.name)

    def _hover_enter(self, _event: tk.Event) -> None:
        self._hover_count += 1
        self._update_hover()

    def _hover_leave(self, _event: tk.Event) -> None:
        self._hover_count = max(0, self._hover_count - 1)
        self._update_hover()

    def _update_hover(self) -> None:
        if self._hover_count > 0:
            self.copy_icon.place(in_=self.canvas, relx=1.0, x=-ui(6), y=ui(6), anchor="ne")
        else:
            self.copy_icon.place_forget()

    def copy(self) -> None:
        svg_text = self.app.get_variant_svg(self.name)
        if not svg_text:
            return
        self.app.root.clipboard_clear()
        self.app.root.clipboard_append(svg_text)
        self.app.show_status(f"Copied {self.name} SVG to clipboard")

    def set_selected(self, selected: bool) -> None:
        border = self.app.colors["accent"] if selected else self.app.colors["surface_border"]
        self.frame.configure(highlightbackground=border)

    def update_preview(self, svg_text: Optional[str], placeholder: str = "Preview paused") -> None:
        self.cached_svg = svg_text
        self.canvas.delete("all")
        transparent = self.variant.light is None
        if transparent:
            self._draw_checkerboard()
        else:
            self.canvas.configure(background=self.app.colors["preview_bg"])
        if not svg_text:
            self.canvas.create_text(
                self.preview_size / 2,
                self.preview_size / 2,
                text=placeholder,
                fill=self.app.colors["muted"],
            )
            if transparent:
                self._draw_transparent_badge()
            return
        try:
            png_bytes = try_svg_to_png_bytes(svg_text, scale=1.0)
        except RuntimeError as exc:
            self.canvas.create_text(
                self.preview_size / 2,
                self.preview_size / 2,
                text="Install cairosvg",
                fill=self.app.colors["muted"],
            )
            self.app.show_status(str(exc))
            return
        self.photo = tk.PhotoImage(data=base64.b64encode(png_bytes))
        image_x = self.preview_size / 2
        image_y = self.preview_size / 2
        self.canvas.create_image(image_x, image_y, anchor="center", image=self.photo)
        if transparent:
            self._draw_transparent_badge()

    def _draw_checkerboard(self) -> None:
        self.canvas.configure(background=self.app.colors["preview_bg"])
        size = CHECKER_SIZE
        colors = (self.app.colors["checker_a"], self.app.colors["checker_b"])
        for y in range(0, self.preview_size, size):
            for x in range(0, self.preview_size, size):
                color = colors[(x // size + y // size) % 2]
                self.canvas.create_rectangle(
                    x,
                    y,
                    x + size,
                    y + size,
                    outline=color,
                    fill=color,
                )

    def _draw_transparent_badge(self) -> None:
        badge_bg = self.app.colors["surface"]
        badge_border = self.app.colors["surface_border"]
        badge_text = self.app.colors["muted"]
        x0, y0 = ui(6), ui(6)
        x1, y1 = ui(80), ui(24)
        self.canvas.create_rectangle(x0, y0, x1, y1, fill=badge_bg, outline=badge_border)
        self.canvas.create_text(
            (x0 + x1) / 2,
            (y0 + y1) / 2,
            text="Transparent",
            fill=badge_text,
            font=("Arial", max(8, ui(8))),
        )

    def _draw_copy_icon(self) -> None:
        self.copy_icon.delete("all")
        bg = self.app.colors["surface"]
        stroke = self.app.colors["text"]
        self.copy_icon.configure(background=bg)
        x0, y0 = ui(6), ui(4)
        x1, y1 = ui(16), ui(14)
        x2, y2 = ui(8), ui(6)
        x3, y3 = ui(18), ui(16)
        self.copy_icon.create_rectangle(x0, y0, x1, y1, outline=stroke, fill=bg)
        self.copy_icon.create_rectangle(x2, y2, x3, y3, outline=stroke, fill=bg)

    def apply_theme(self, colors: Dict[str, str]) -> None:
        self.frame.configure(background=colors["surface"], highlightbackground=colors["surface_border"])
        self.canvas.configure(background=colors["preview_bg"])
        self._draw_copy_icon()

    def set_preview_size(self, size: int) -> None:
        if size == self.preview_size:
            return
        self.preview_size = size
        self.canvas.configure(width=size, height=size)


class QrGuiApp:
    def __init__(self, root: tk.Tk) -> None:
        self.root = root
        self.root.title("QR Generator")
        self.root.tk.call("tk", "scaling", UI_SCALE)
        self.root.geometry(f"{ui(1120)}x{ui(860)}")

        self.colors = THEMES["system"].copy()
        self.preview_job: Optional[str] = None
        self.suspend_custom_trace = False

        self.preview_zoom = 1.0
        self.preview_size = BASE_PREVIEW_SIZE
        self.card_columns = CARD_COLUMNS

        self.data_var = tk.StringVar()
        self.live_preview_var = tk.BooleanVar(value=True)
        self.theme_var = tk.StringVar(value="System")
        self.format_var = tk.StringVar(value="SVG")
        self.status_var = tk.StringVar(value="")
        self.scale_var = tk.IntVar(value=10)
        self.border_var = tk.IntVar(value=4)
        self.error_var = tk.StringVar(value="m")
        self.png_scale_var = tk.DoubleVar(value=3.0)
        self.gif_variant_var = tk.StringVar(value="wave")
        self.gif_fps_var = tk.IntVar(value=DEFAULT_GIF_FPS)
        self.gif_frames_var = tk.IntVar(value=DEFAULT_GIF_FRAMES)
        self.gif_hold_var = tk.IntVar(value=DEFAULT_GIF_HOLD)
        self.wave_amp_var = tk.DoubleVar(value=DEFAULT_WAVE_AMP)
        self.wave_period_var = tk.DoubleVar(value=DEFAULT_WAVE_PERIOD)
        self.float_angle_var = tk.DoubleVar(value=DEFAULT_FLOAT_ANGLE)
        self.output_path_var = tk.StringVar()

        self.custom_shape_var = tk.StringVar()
        self.custom_dark_var = tk.StringVar()
        self.custom_light_var = tk.StringVar()
        self.custom_radius_var = tk.DoubleVar(value=0.0)
        self.custom_transparent_var = tk.BooleanVar(value=False)
        self.custom_gradient_enabled_var = tk.BooleanVar(value=False)
        self.custom_gradient_from_var = tk.StringVar()
        self.custom_gradient_to_var = tk.StringVar()
        self.preset_name_var = tk.StringVar()

        self.cards: Dict[str, VariantCard] = {}
        self.variant_map: Dict[str, Variant] = {}
        self.variant_order: List[str] = []
        self.preset_entries: List[Dict[str, str]] = []
        self.detail_photo: Optional[tk.PhotoImage] = None
        self.enlarged_window: Optional[tk.Toplevel] = None
        self.enlarged_photo: Optional[tk.PhotoImage] = None
        self.enlarged_png_bytes: Optional[bytes] = None
        self.selected_variant = sorted(VARIANTS.keys())[0]

        self._load_presets()
        self._build_variant_catalog()
        self._build_ui()
        self.apply_theme("System")
        self.select_variant(self.selected_variant)
        self.schedule_preview_update()

    def _build_ui(self) -> None:
        self.root.grid_rowconfigure(1, weight=1)
        self.root.grid_columnconfigure(0, weight=1)

        top = ttk.Frame(self.root)
        top.grid(row=0, column=0, sticky="ew", padx=16, pady=12)
        top.grid_columnconfigure(1, weight=1)

        ttk.Label(top, text="QR Data").grid(row=0, column=0, sticky="w")
        entry = ttk.Entry(top, textvariable=self.data_var)
        entry.grid(row=0, column=1, sticky="ew", padx=12)
        entry.focus()

        preview_toggle = ttk.Checkbutton(
            top,
            text="Live preview",
            variable=self.live_preview_var,
            command=self.on_preview_toggle,
        )
        preview_toggle.grid(row=0, column=2, sticky="e", padx=(0, 12))

        theme_box = ttk.Combobox(
            top,
            textvariable=self.theme_var,
            values=["System", "Light", "Dark"],
            state="readonly",
            width=10,
        )
        theme_box.grid(row=0, column=3, sticky="e")
        theme_box.bind("<<ComboboxSelected>>", self.on_theme_change)

        content = ttk.Frame(self.root)
        content.grid(row=1, column=0, sticky="nsew", padx=16)
        content.grid_columnconfigure(0, weight=1)
        content.grid_rowconfigure(0, weight=1)

        self.scroll = ScrollableFrame(content, zoom_callback=self.on_zoom)
        self.scroll.grid(row=0, column=0, sticky="nsew")
        self.scroll.canvas.bind("<Configure>", lambda _event: self.update_grid_layout())

        self.detail_panel = ttk.Frame(content)
        self.detail_panel.grid(row=0, column=1, sticky="ns", padx=(16, 0))

        self._build_cards()
        self._build_detail_panel()

        export = ttk.LabelFrame(self.root, text="Export")
        export.grid(row=2, column=0, sticky="ew", padx=16, pady=12)
        export.grid_columnconfigure(1, weight=1)

        ttk.Label(export, text="Format").grid(row=0, column=0, sticky="w", padx=(12, 6), pady=(8, 4))
        format_box = ttk.Combobox(
            export,
            textvariable=self.format_var,
            values=["SVG", "PNG", "PDF", "PS", "GIF"],
            state="readonly",
            width=8,
        )
        format_box.grid(row=0, column=1, sticky="w", pady=(8, 4))
        format_box.bind("<<ComboboxSelected>>", self.on_format_change)

        ttk.Label(export, text="Output").grid(row=0, column=2, sticky="w", padx=(16, 6), pady=(8, 4))
        output_entry = ttk.Entry(export, textvariable=self.output_path_var)
        output_entry.grid(row=0, column=3, sticky="ew", pady=(8, 4))
        ttk.Button(export, text="Browse", command=self.browse_output).grid(
            row=0, column=4, sticky="w", padx=6, pady=(8, 4)
        )
        ttk.Button(export, text="Export", command=self.export).grid(
            row=0, column=5, sticky="e", padx=(6, 12), pady=(8, 4)
        )

        settings = ttk.Frame(export)
        settings.grid(row=1, column=0, columnspan=6, sticky="ew", padx=12, pady=(4, 10))
        settings.grid_columnconfigure(2, weight=1)

        qr_settings = ttk.LabelFrame(settings, text="QR settings")
        qr_settings.grid(row=0, column=0, sticky="nw", padx=(0, 12), pady=4)

        ttk.Label(qr_settings, text="Scale").grid(row=0, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(qr_settings, from_=1, to=40, width=6, textvariable=self.scale_var).grid(
            row=0, column=1, sticky="w", padx=8, pady=4
        )
        ttk.Label(qr_settings, text="Border").grid(row=1, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(qr_settings, from_=0, to=20, width=6, textvariable=self.border_var).grid(
            row=1, column=1, sticky="w", padx=8, pady=4
        )
        ttk.Label(qr_settings, text="Error").grid(row=2, column=0, sticky="w", padx=8, pady=4)
        ttk.Combobox(
            qr_settings,
            textvariable=self.error_var,
            values=["l", "m", "q", "h"],
            state="readonly",
            width=4,
        ).grid(row=2, column=1, sticky="w", padx=8, pady=4)

        self.format_settings = ttk.LabelFrame(settings, text="Format settings")
        self.format_settings.grid(row=0, column=1, sticky="nw", pady=4)

        self.png_settings = ttk.Frame(self.format_settings)
        ttk.Label(self.png_settings, text="PNG scale").grid(row=0, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(
            self.png_settings,
            from_=0.5,
            to=10.0,
            increment=0.5,
            width=6,
            textvariable=self.png_scale_var,
        ).grid(row=0, column=1, sticky="w", padx=8, pady=4)

        self.gif_settings = ttk.Frame(self.format_settings)
        ttk.Label(self.gif_settings, text="Variant").grid(row=0, column=0, sticky="w", padx=8, pady=4)
        ttk.Combobox(
            self.gif_settings,
            textvariable=self.gif_variant_var,
            values=list(ANIMATION_VARIANTS),
            state="readonly",
            width=14,
        ).grid(row=0, column=1, sticky="w", padx=8, pady=4)
        ttk.Label(self.gif_settings, text="FPS").grid(row=1, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(self.gif_settings, from_=1, to=60, width=6, textvariable=self.gif_fps_var).grid(
            row=1, column=1, sticky="w", padx=8, pady=4
        )
        ttk.Label(self.gif_settings, text="Frames").grid(row=2, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(self.gif_settings, from_=2, to=120, width=6, textvariable=self.gif_frames_var).grid(
            row=2, column=1, sticky="w", padx=8, pady=4
        )
        ttk.Label(self.gif_settings, text="Hold").grid(row=3, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(self.gif_settings, from_=0, to=120, width=6, textvariable=self.gif_hold_var).grid(
            row=3, column=1, sticky="w", padx=8, pady=4
        )
        ttk.Label(self.gif_settings, text="Wave amp").grid(row=4, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(
            self.gif_settings,
            from_=0.0,
            to=2.0,
            increment=0.05,
            width=6,
            textvariable=self.wave_amp_var,
        ).grid(row=4, column=1, sticky="w", padx=8, pady=4)
        ttk.Label(self.gif_settings, text="Wave period").grid(row=5, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(
            self.gif_settings,
            from_=2.0,
            to=40.0,
            increment=0.5,
            width=6,
            textvariable=self.wave_period_var,
        ).grid(row=5, column=1, sticky="w", padx=8, pady=4)
        ttk.Label(self.gif_settings, text="Float angle").grid(row=6, column=0, sticky="w", padx=8, pady=4)
        ttk.Spinbox(
            self.gif_settings,
            from_=-180,
            to=180,
            increment=5,
            width=6,
            textvariable=self.float_angle_var,
        ).grid(row=6, column=1, sticky="w", padx=8, pady=4)

        self.format_settings.grid_columnconfigure(0, weight=1)
        self.on_format_change()

        status = ttk.Label(self.root, textvariable=self.status_var)
        status.grid(row=3, column=0, sticky="e", padx=18, pady=(0, 10))

        self.data_var.trace_add("write", lambda *_: self.on_data_change())
        self.scale_var.trace_add("write", lambda *_: self.on_data_change())
        self.border_var.trace_add("write", lambda *_: self.on_data_change())
        self.error_var.trace_add("write", lambda *_: self.on_data_change())

    def _build_cards(self) -> None:
        names = self.variant_order
        for index, name in enumerate(names):
            variant = self.variant_map[name]
            card = VariantCard(self, self.scroll.inner, name, variant)
            self.cards[name] = card
        self.update_grid_layout(force=True)

    def _rebuild_cards(self) -> None:
        for card in self.cards.values():
            card.frame.destroy()
        self.cards = {}
        self._build_cards()
        self.select_variant(self.selected_variant)
        self.schedule_preview_update()

    def update_grid_layout(self, force: bool = False) -> None:
        canvas_width = self.scroll.canvas.winfo_width()
        if canvas_width <= 1:
            return
        card_width = self.preview_size + ui(40)
        columns = max(1, int(canvas_width // max(1, card_width)))
        if columns != self.card_columns or force:
            self.card_columns = columns
            for index, name in enumerate(self.variant_order):
                card = self.cards.get(name)
                if not card:
                    continue
                row = index // columns
                col = index % columns
                card.frame.grid(row=row, column=col, padx=ui(8), pady=ui(8), sticky="n")

    def _build_detail_panel(self) -> None:
        preview_frame = ttk.LabelFrame(self.detail_panel, text="Selected style")
        preview_frame.grid(row=0, column=0, sticky="ew")
        preview_frame.grid_columnconfigure(0, weight=1)

        self.selected_label = ttk.Label(preview_frame, text="")
        self.selected_label.grid(row=0, column=0, sticky="w", padx=12, pady=(10, 4))

        self.detail_canvas = tk.Canvas(
            preview_frame,
            width=DETAIL_PREVIEW_SIZE,
            height=DETAIL_PREVIEW_SIZE,
            highlightthickness=0,
            background=self.colors["preview_bg"],
        )
        self.detail_canvas.grid(row=1, column=0, padx=12, pady=(0, 8))
        self.detail_canvas.bind("<Button-1>", self.open_enlarged_preview)

        preview_actions = ttk.Frame(preview_frame)
        preview_actions.grid(row=2, column=0, sticky="ew", padx=12, pady=(0, 10))
        ttk.Button(preview_actions, text="Copy SVG", command=self.copy_selected_svg).grid(
            row=0, column=0, sticky="w"
        )

        customize_frame = ttk.LabelFrame(self.detail_panel, text="Customize")
        customize_frame.grid(row=1, column=0, sticky="ew", pady=(12, 0))
        customize_frame.grid_columnconfigure(1, weight=1)

        ttk.Label(customize_frame, text="Shape").grid(row=0, column=0, sticky="w", padx=12, pady=6)
        ttk.Combobox(
            customize_frame,
            textvariable=self.custom_shape_var,
            values=["square", "rounded", "dot"],
            state="readonly",
            width=10,
        ).grid(row=0, column=1, sticky="w", padx=12, pady=6)

        ttk.Label(customize_frame, text="Dark").grid(row=1, column=0, sticky="w", padx=12, pady=6)
        dark_row = ttk.Frame(customize_frame)
        dark_row.grid(row=1, column=1, sticky="w", padx=12, pady=6)
        ttk.Entry(dark_row, textvariable=self.custom_dark_var, width=16).grid(row=0, column=0, sticky="w")
        ttk.Button(dark_row, text="Pick", command=lambda: self.pick_color(self.custom_dark_var)).grid(
            row=0, column=1, sticky="w", padx=(6, 0)
        )

        ttk.Label(customize_frame, text="Light").grid(row=2, column=0, sticky="w", padx=12, pady=6)
        light_row = ttk.Frame(customize_frame)
        light_row.grid(row=2, column=1, sticky="w", padx=12, pady=6)
        self.custom_light_entry = ttk.Entry(light_row, textvariable=self.custom_light_var, width=16)
        self.custom_light_entry.grid(row=0, column=0, sticky="w")
        self.custom_light_pick = ttk.Button(
            light_row, text="Pick", command=lambda: self.pick_color(self.custom_light_var)
        )
        self.custom_light_pick.grid(row=0, column=1, sticky="w", padx=(6, 0))

        ttk.Checkbutton(
            customize_frame,
            text="Transparent background",
            variable=self.custom_transparent_var,
            command=self.update_custom_field_states,
        ).grid(row=3, column=0, columnspan=2, sticky="w", padx=12, pady=6)

        ttk.Label(customize_frame, text="Radius").grid(row=4, column=0, sticky="w", padx=12, pady=6)
        ttk.Spinbox(
            customize_frame,
            from_=0.0,
            to=0.5,
            increment=0.02,
            width=6,
            textvariable=self.custom_radius_var,
        ).grid(row=4, column=1, sticky="w", padx=12, pady=6)

        ttk.Checkbutton(
            customize_frame,
            text="Use gradient",
            variable=self.custom_gradient_enabled_var,
            command=self.update_custom_field_states,
        ).grid(row=5, column=0, columnspan=2, sticky="w", padx=12, pady=(6, 2))

        gradient_row = ttk.Frame(customize_frame)
        gradient_row.grid(row=6, column=0, columnspan=2, sticky="w", padx=12, pady=(0, 6))
        ttk.Label(gradient_row, text="From").grid(row=0, column=0, sticky="w")
        self.gradient_from_entry = ttk.Entry(gradient_row, textvariable=self.custom_gradient_from_var, width=10)
        self.gradient_from_entry.grid(row=0, column=1, sticky="w", padx=(6, 6))
        self.gradient_from_pick = ttk.Button(
            gradient_row, text="Pick", command=lambda: self.pick_color(self.custom_gradient_from_var)
        )
        self.gradient_from_pick.grid(row=0, column=2, sticky="w", padx=(0, 12))
        ttk.Label(gradient_row, text="To").grid(row=0, column=3, sticky="w")
        self.gradient_to_entry = ttk.Entry(gradient_row, textvariable=self.custom_gradient_to_var, width=10)
        self.gradient_to_entry.grid(row=0, column=4, sticky="w", padx=(6, 6))
        self.gradient_to_pick = ttk.Button(
            gradient_row, text="Pick", command=lambda: self.pick_color(self.custom_gradient_to_var)
        )
        self.gradient_to_pick.grid(row=0, column=5, sticky="w")

        ttk.Button(customize_frame, text="Reset to default", command=self.reset_custom_settings).grid(
            row=7, column=0, columnspan=2, sticky="w", padx=12, pady=(8, 10)
        )

        preset_frame = ttk.LabelFrame(self.detail_panel, text="Presets")
        preset_frame.grid(row=2, column=0, sticky="ew", pady=(12, 0))
        preset_frame.grid_columnconfigure(1, weight=1)
        ttk.Label(preset_frame, text="Name").grid(row=0, column=0, sticky="w", padx=12, pady=8)
        ttk.Entry(preset_frame, textvariable=self.preset_name_var, width=18).grid(
            row=0, column=1, sticky="w", padx=12, pady=8
        )
        ttk.Button(preset_frame, text="Save preset", command=self.save_preset).grid(
            row=1, column=0, columnspan=2, sticky="w", padx=12, pady=(0, 10)
        )

        self.custom_shape_var.trace_add("write", lambda *_: self.on_custom_change())
        self.custom_dark_var.trace_add("write", lambda *_: self.on_custom_change())
        self.custom_light_var.trace_add("write", lambda *_: self.on_custom_change())
        self.custom_radius_var.trace_add("write", lambda *_: self.on_custom_change())
        self.custom_transparent_var.trace_add("write", lambda *_: self.on_custom_change())
        self.custom_gradient_enabled_var.trace_add("write", lambda *_: self.on_custom_change())
        self.custom_gradient_from_var.trace_add("write", lambda *_: self.on_custom_change())
        self.custom_gradient_to_var.trace_add("write", lambda *_: self.on_custom_change())

    def _load_presets(self) -> None:
        self.preset_entries = []
        if not os.path.exists(PRESET_FILE):
            return
        try:
            with open(PRESET_FILE, "r", encoding="utf-8") as handle:
                data = json.load(handle)
        except (OSError, json.JSONDecodeError):
            return
        if not isinstance(data, list):
            return
        for item in data:
            if isinstance(item, dict):
                self.preset_entries.append(item)

    def _save_presets(self) -> None:
        try:
            with open(PRESET_FILE, "w", encoding="utf-8") as handle:
                json.dump(self.preset_entries, handle, indent=2)
        except OSError:
            self.show_status("Failed to save presets")

    def _build_variant_catalog(self) -> None:
        self.variant_map = dict(VARIANTS)
        self.variant_order = sorted(VARIANTS.keys())
        for entry in self.preset_entries:
            variant = self._variant_from_preset(entry)
            if not variant:
                continue
            if variant.name in self.variant_map:
                continue
            self.variant_map[variant.name] = variant
            self.variant_order.append(variant.name)

    def _variant_from_preset(self, entry: Dict[str, str]) -> Optional[Variant]:
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

    def apply_theme(self, mode: str) -> None:
        key = mode.lower()
        if key not in THEMES:
            key = "system"
        self.colors = THEMES[key].copy()
        style = ttk.Style()

        self.root.configure(background=self.colors["bg"])
        style.configure("TFrame", background=self.colors["bg"])
        style.configure("TLabel", background=self.colors["bg"], foreground=self.colors["text"])
        style.configure("TLabelFrame", background=self.colors["bg"], foreground=self.colors["text"])
        style.configure("TLabelFrame.Label", background=self.colors["bg"], foreground=self.colors["text"])
        style.configure("TEntry", fieldbackground=self.colors["surface"], foreground=self.colors["text"])
        style.configure("TButton", background=self.colors["surface"], foreground=self.colors["text"])
        style.configure("TCombobox", fieldbackground=self.colors["surface"], foreground=self.colors["text"])
        style.configure("TCheckbutton", background=self.colors["bg"], foreground=self.colors["text"])

        self.scroll.apply_theme(self.colors)
        for card in self.cards.values():
            card.apply_theme(self.colors)
            card.set_selected(card.name == self.selected_variant)
        if hasattr(self, "detail_canvas"):
            self.detail_canvas.configure(background=self.colors["preview_bg"])
            self.update_selected_preview()
        self.schedule_preview_update()

    def on_zoom(self, direction: float) -> None:
        step = 0.1 if direction > 0 else -0.1
        new_zoom = max(0.6, min(2.0, round(self.preview_zoom + step, 2)))
        if new_zoom == self.preview_zoom:
            return
        self.preview_zoom = new_zoom
        self.preview_size = max(ui(96), int(round(BASE_PREVIEW_SIZE * self.preview_zoom)))
        for card in self.cards.values():
            card.set_preview_size(self.preview_size)
        self.update_grid_layout(force=True)
        self.schedule_preview_update()

    def select_variant(self, name: str) -> None:
        self.selected_variant = name
        for card_name, card in self.cards.items():
            card.set_selected(card_name == name)
        self.selected_label.configure(text=name)
        self.load_variant_into_custom_fields(name)
        self.update_selected_preview()

    def on_preview_toggle(self) -> None:
        if self.live_preview_var.get():
            self.schedule_preview_update()
            self.update_selected_preview()
        else:
            for card in self.cards.values():
                card.update_preview(None)
            self.update_selected_preview()
            self.show_status("Preview paused")

    def on_theme_change(self, _event: tk.Event) -> None:
        self.apply_theme(self.theme_var.get())

    def on_format_change(self, _event: Optional[tk.Event] = None) -> None:
        for child in self.format_settings.winfo_children():
            child.grid_forget()
        fmt = self.format_var.get().upper()
        if fmt == "PNG":
            self.png_settings.grid(row=0, column=0, sticky="w")
        elif fmt == "GIF":
            self.gif_settings.grid(row=0, column=0, sticky="w")
        else:
            placeholder = ttk.Label(self.format_settings, text="No extra settings")
            placeholder.grid(row=0, column=0, sticky="w", padx=8, pady=8)

    def on_data_change(self) -> None:
        if not self.live_preview_var.get():
            return
        self.schedule_preview_update()

    def schedule_preview_update(self) -> None:
        if not self.live_preview_var.get():
            return
        if self.preview_job:
            self.root.after_cancel(self.preview_job)
        self.preview_job = self.root.after(150, self.update_previews)

    def update_previews(self) -> None:
        self.preview_job = None
        data = self.data_var.get().strip()
        if not data:
            for card in self.cards.values():
                card.update_preview(None, placeholder="Enter text")
            self.update_selected_preview()
            return
        try:
            qr = segno.make(data, error=self.error_var.get())
        except Exception as exc:
            self.show_status(f"Invalid data: {exc}")
            return
        preview_scale = self._compute_preview_scale(len(qr.matrix))
        border = self.border_var.get()
        for name, card in self.cards.items():
            variant = self.variant_map[name]
            svg = render_svg(
                qr.matrix,
                scale=preview_scale,
                border=border,
                dark=variant.dark,
                light=variant.light,
                shape=variant.shape,
                radius=variant.radius,
                gradient=variant.gradient,
            )
            card.update_preview(svg)
        self.update_selected_preview()

    def _compute_preview_scale(self, size: int) -> int:
        border = self.border_var.get()
        total_modules = max(1, size + border * 2)
        return max(2, self.preview_size // total_modules)

    def _compute_detail_scale(self, size: int) -> int:
        border = self.border_var.get()
        total_modules = max(1, size + border * 2)
        return max(2, DETAIL_PREVIEW_SIZE // total_modules)

    def load_variant_into_custom_fields(self, name: str) -> None:
        variant = self.variant_map.get(name)
        if not variant:
            return
        self.suspend_custom_trace = True
        self.custom_shape_var.set(variant.shape)
        self.custom_dark_var.set(variant.dark)
        if variant.light is None:
            self.custom_transparent_var.set(True)
            self.custom_light_var.set("")
        else:
            self.custom_transparent_var.set(False)
            self.custom_light_var.set(variant.light)
        self.custom_radius_var.set(variant.radius)
        if variant.gradient:
            self.custom_gradient_enabled_var.set(True)
            self.custom_gradient_from_var.set(variant.gradient.get("from", ""))
            self.custom_gradient_to_var.set(variant.gradient.get("to", ""))
        else:
            self.custom_gradient_enabled_var.set(False)
            self.custom_gradient_from_var.set("")
            self.custom_gradient_to_var.set("")
        self.suspend_custom_trace = False
        self.update_custom_field_states()

    def update_custom_field_states(self) -> None:
        light_state = "disabled" if self.custom_transparent_var.get() else "normal"
        self.custom_light_entry.configure(state=light_state)
        self.custom_light_pick.configure(state=light_state)
        gradient_state = "normal" if self.custom_gradient_enabled_var.get() else "disabled"
        self.gradient_from_entry.configure(state=gradient_state)
        self.gradient_to_entry.configure(state=gradient_state)
        self.gradient_from_pick.configure(state=gradient_state)
        self.gradient_to_pick.configure(state=gradient_state)

    def pick_color(self, target: tk.Variable) -> None:
        current = target.get() if hasattr(target, "get") else ""
        color = colorchooser.askcolor(color=current if current else None, parent=self.root)
        if color and color[1]:
            target.set(color[1])

    def on_custom_change(self) -> None:
        if self.suspend_custom_trace:
            return
        if self.custom_shape_var.get() == "rounded":
            try:
                current_radius = float(self.custom_radius_var.get())
            except (tk.TclError, TypeError, ValueError):
                current_radius = 0.0
            if current_radius <= 0:
                self.suspend_custom_trace = True
                self.custom_radius_var.set(0.28)
                self.suspend_custom_trace = False
        self.update_selected_preview()

    def build_custom_variant(self, base: Variant) -> Variant:
        shape = self.custom_shape_var.get() or base.shape
        dark = self.custom_dark_var.get().strip() or base.dark
        if self.custom_transparent_var.get():
            light = None
        else:
            light = self.custom_light_var.get().strip() or base.light or "#ffffff"
        try:
            radius_value = self.custom_radius_var.get()
        except tk.TclError:
            radius_value = base.radius
        try:
            radius = float(radius_value)
        except (TypeError, ValueError):
            radius = base.radius
        gradient = None
        if self.custom_gradient_enabled_var.get():
            color_from = self.custom_gradient_from_var.get().strip()
            color_to = self.custom_gradient_to_var.get().strip()
            if color_from and color_to:
                gradient = {"id": "fg", "from": color_from, "to": color_to}
        return Variant(
            name=base.name,
            shape=shape,
            dark=dark,
            light=light,
            radius=radius,
            gradient=gradient,
        )

    def update_selected_preview(self) -> None:
        if not hasattr(self, "detail_canvas"):
            return
        self.detail_canvas.delete("all")
        self.detail_canvas.configure(background=self.colors["preview_bg"])
        if not self.live_preview_var.get():
            self.detail_canvas.create_text(
                DETAIL_PREVIEW_SIZE / 2,
                DETAIL_PREVIEW_SIZE / 2,
                text="Preview paused",
                fill=self.colors["muted"],
            )
            return
        data = self.data_var.get().strip()
        if not data:
            self.detail_canvas.create_text(
                DETAIL_PREVIEW_SIZE / 2,
                DETAIL_PREVIEW_SIZE / 2,
                text="Enter text",
                fill=self.colors["muted"],
            )
            return
        try:
            qr = segno.make(data, error=self.error_var.get())
        except Exception as exc:
            self.show_status(f"Invalid data: {exc}")
            return
        base_variant = self.variant_map.get(self.selected_variant)
        if not base_variant:
            return
        variant = self.build_custom_variant(base_variant)
        if variant.light is None:
            self._draw_detail_checkerboard()
        preview_scale = self._compute_detail_scale(len(qr.matrix))
        svg = render_svg(
            qr.matrix,
            scale=preview_scale,
            border=self.border_var.get(),
            dark=variant.dark,
            light=variant.light,
            shape=variant.shape,
            radius=variant.radius,
            gradient=variant.gradient,
        )
        try:
            png_bytes = try_svg_to_png_bytes(svg, scale=1.0)
        except RuntimeError as exc:
            self.detail_canvas.create_text(
                DETAIL_PREVIEW_SIZE / 2,
                DETAIL_PREVIEW_SIZE / 2,
                text="Install cairosvg",
                fill=self.colors["muted"],
            )
            self.show_status(str(exc))
            return
        self.detail_photo = tk.PhotoImage(data=base64.b64encode(png_bytes))
        self.detail_canvas.create_image(
            DETAIL_PREVIEW_SIZE / 2,
            DETAIL_PREVIEW_SIZE / 2,
            anchor="center",
            image=self.detail_photo,
        )
        if variant.light is None:
            self._draw_detail_transparent_badge()

    def _draw_detail_checkerboard(self) -> None:
        self._draw_checkerboard(self.detail_canvas, DETAIL_PREVIEW_SIZE)

    def _draw_detail_transparent_badge(self) -> None:
        badge_bg = self.colors["surface"]
        badge_border = self.colors["surface_border"]
        badge_text = self.colors["muted"]
        x0, y0 = ui(8), ui(8)
        x1, y1 = ui(92), ui(28)
        self.detail_canvas.create_rectangle(x0, y0, x1, y1, fill=badge_bg, outline=badge_border)
        self.detail_canvas.create_text(
            (x0 + x1) / 2,
            (y0 + y1) / 2,
            text="Transparent",
            fill=badge_text,
            font=("Arial", max(9, ui(9))),
        )

    def _draw_checkerboard(self, canvas: tk.Canvas, size: int) -> None:
        tile = CHECKER_SIZE
        colors = (self.colors["checker_a"], self.colors["checker_b"])
        for y in range(0, size, tile):
            for x in range(0, size, tile):
                color = colors[(x // tile + y // tile) % 2]
                canvas.create_rectangle(
                    x,
                    y,
                    x + tile,
                    y + tile,
                    outline=color,
                    fill=color,
                )

    def reset_custom_settings(self) -> None:
        self.load_variant_into_custom_fields(self.selected_variant)
        self.update_selected_preview()

    def open_enlarged_preview(self, _event: Optional[tk.Event] = None) -> None:
        if self.enlarged_window and self.enlarged_window.winfo_exists():
            self.close_enlarged_preview()
            return
        data = self.data_var.get().strip()
        if not data:
            self.show_status("Enter data first")
            return
        try:
            qr = segno.make(data, error=self.error_var.get())
        except Exception as exc:
            self.show_status(f"Invalid data: {exc}")
            return
        base_variant = self.variant_map.get(self.selected_variant)
        if not base_variant:
            return
        variant = self.build_custom_variant(base_variant)
        total_modules = max(1, len(qr.matrix) + self.border_var.get() * 2)
        self.root.update_idletasks()
        root_w = max(ui(600), self.root.winfo_width())
        root_h = max(ui(500), self.root.winfo_height())
        target = int(min(root_w, root_h) * 0.7)
        scale = max(2, target // total_modules)
        svg = render_svg(
            qr.matrix,
            scale=scale,
            border=self.border_var.get(),
            dark=variant.dark,
            light=variant.light,
            shape=variant.shape,
            radius=variant.radius,
            gradient=variant.gradient,
        )
        try:
            png_bytes = try_svg_to_png_bytes(svg, scale=1.0)
        except RuntimeError as exc:
            self.show_status(str(exc))
            return

        win = tk.Toplevel(self.root)
        self.enlarged_window = win
        self.enlarged_png_bytes = png_bytes
        win.transient(self.root)
        win.configure(background=self.colors["bg"])
        win.attributes("-topmost", True)
        x = self.root.winfo_rootx()
        y = self.root.winfo_rooty()
        win.geometry(f"{root_w}x{root_h}+{x}+{y}")

        def close_event(_event=None):
            self.close_enlarged_preview()

        win.bind("<Button-1>", close_event)
        win.bind("<Escape>", close_event)

        container = tk.Frame(win, background=self.colors["bg"])
        container.place(relx=0.5, rely=0.5, anchor="center")
        container.bind("<Button-1>", lambda _event: "break")

        self.enlarged_photo = tk.PhotoImage(data=base64.b64encode(png_bytes))
        canvas = tk.Canvas(
            container,
            width=self.enlarged_photo.width(),
            height=self.enlarged_photo.height(),
            highlightthickness=0,
            background=self.colors["preview_bg"],
        )
        canvas.grid(row=0, column=0, padx=12, pady=(12, 8))
        canvas.bind("<Button-1>", lambda _event: "break")
        if variant.light is None:
            self._draw_checkerboard(canvas, self.enlarged_photo.width())
        canvas.create_image(
            self.enlarged_photo.width() / 2,
            self.enlarged_photo.height() / 2,
            anchor="center",
            image=self.enlarged_photo,
        )

        hint = ttk.Label(container, text="Click outside the image to close")
        hint.grid(row=1, column=0, pady=(0, 6))
        hint.bind("<Button-1>", lambda _event: "break")

        copy_btn = ttk.Button(container, text="Copy PNG", command=self.copy_enlarged_png)
        copy_btn.grid(row=2, column=0, pady=(0, 12))
        copy_btn.bind("<Button-1>", lambda _event: "break")

    def close_enlarged_preview(self) -> None:
        if self.enlarged_window and self.enlarged_window.winfo_exists():
            self.enlarged_window.destroy()
        self.enlarged_window = None
        self.enlarged_photo = None
        self.enlarged_png_bytes = None

    def copy_enlarged_png(self) -> None:
        if not self.enlarged_png_bytes:
            return
        try:
            self.root.clipboard_clear()
            self.root.clipboard_append(self.enlarged_png_bytes, type="image/png")
            self.show_status("Copied PNG to clipboard")
        except tk.TclError:
            encoded = base64.b64encode(self.enlarged_png_bytes).decode("ascii")
            self.root.clipboard_clear()
            self.root.clipboard_append(encoded)
            self.show_status("Copied PNG bytes as base64")

    def copy_selected_svg(self) -> None:
        svg_text = self.get_variant_svg(self.selected_variant, use_custom=True)
        if not svg_text:
            return
        self.root.clipboard_clear()
        self.root.clipboard_append(svg_text)
        self.show_status(f"Copied {self.selected_variant} SVG to clipboard")

    def save_preset(self) -> None:
        name = self.preset_name_var.get().strip()
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
        self._save_presets()
        self.variant_map[variant.name] = variant
        self.variant_order.append(variant.name)
        self.preset_name_var.set("")
        self._rebuild_cards()
        self.select_variant(variant.name)

    def get_variant_svg(self, name: str, use_custom: bool = False) -> Optional[str]:
        data = self.data_var.get().strip()
        if not data:
            self.show_status("Enter data first")
            return None
        try:
            qr = segno.make(data, error=self.error_var.get())
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
            scale=self.scale_var.get(),
            border=self.border_var.get(),
            dark=variant.dark,
            light=variant.light,
            shape=variant.shape,
            radius=variant.radius,
            gradient=variant.gradient,
        )

    def browse_output(self) -> None:
        fmt = self.format_var.get().lower()
        extension = "svg" if fmt == "svg" else fmt
        initial = os.path.join("out", f"qr-{self.selected_variant}.{extension}")
        path = filedialog.asksaveasfilename(
            defaultextension=f".{extension}",
            initialfile=os.path.basename(initial),
        )
        if path:
            self.output_path_var.set(path)

    def export(self) -> None:
        fmt = self.format_var.get().lower()
        ext = "svg" if fmt == "svg" else fmt
        output_path = self.output_path_var.get().strip()
        if not output_path:
            output_path = os.path.join("out", f"qr-{self.selected_variant}.{ext}")
        svg_text = self.get_variant_svg(self.selected_variant, use_custom=True)
        if not svg_text:
            return
        os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)

        if fmt == "svg":
            with open(output_path, "w", encoding="utf-8") as handle:
                handle.write(svg_text)
            self.show_status(f"Saved {output_path}")
            return
        if fmt == "png":
            try:
                write_png(svg_text, output_path, self.png_scale_var.get())
            except SystemExit:
                messagebox.showerror("Missing dependency", "PNG export requires cairosvg")
                return
            self.show_status(f"Saved {output_path}")
            return
        if fmt == "pdf":
            try:
                write_pdf(svg_text, output_path)
            except SystemExit:
                messagebox.showerror("Missing dependency", "PDF export requires cairosvg")
                return
            self.show_status(f"Saved {output_path}")
            return
        if fmt == "ps":
            try:
                write_ps(svg_text, output_path)
            except SystemExit:
                messagebox.showerror("Missing dependency", "PS export requires cairosvg")
                return
            self.show_status(f"Saved {output_path}")
            return
        if fmt == "gif":
            self.export_gif(output_path)
            return

    def export_gif(self, output_path: str) -> None:
        data = self.data_var.get().strip()
        if not data:
            self.show_status("Enter data first")
            return
        try:
            qr = segno.make(data, error=self.error_var.get())
        except Exception as exc:
            self.show_status(f"Invalid data: {exc}")
            return
        base_variant = self.variant_map.get(self.selected_variant)
        if not base_variant:
            return
        variant = self.build_custom_variant(base_variant)
        animation_variant = self.gif_variant_var.get()
        scale = self.scale_var.get()
        border = self.border_var.get()
        kwargs = dict(
            matrix=qr.matrix,
            scale=scale,
            border=border,
            dark=variant.dark,
            light=variant.light,
            shape=variant.shape,
            radius=variant.radius,
            gradient=variant.gradient,
            frames=self.gif_frames_var.get(),
            hold=self.gif_hold_var.get(),
            render_svg=render_svg,
        )
        try:
            if animation_variant == "wave":
                frames = build_wave_gif_frames(
                    wave_amp=self.wave_amp_var.get(),
                    wave_period=self.wave_period_var.get(),
                    mode="still",
                    **kwargs,
                )
            elif animation_variant == "wave-loop":
                frames = build_wave_gif_frames(
                    wave_amp=self.wave_amp_var.get(),
                    wave_period=self.wave_period_var.get(),
                    mode="loop",
                    hold=0,
                    **kwargs,
                )
            elif animation_variant == "float":
                frames = build_float_gif_frames(
                    float_amp=self.wave_amp_var.get(),
                    float_period=self.wave_period_var.get(),
                    float_angle_deg=self.float_angle_var.get() + DEFAULT_FLOAT_TILT,
                    mode="still",
                    rotate_deg=DEFAULT_FLOAT_TILT,
                    **kwargs,
                )
            elif animation_variant == "float-tilt-first":
                frames = build_float_gif_frames(
                    float_amp=self.wave_amp_var.get(),
                    float_period=self.wave_period_var.get(),
                    float_angle_deg=self.float_angle_var.get(),
                    mode="still",
                    rotate_deg=DEFAULT_FLOAT_TILT,
                    rotate_mode="before",
                    **kwargs,
                )
            elif animation_variant == "float-jagged":
                frames = build_float_gif_frames(
                    float_amp=self.wave_amp_var.get(),
                    float_period=self.wave_period_var.get(),
                    float_angle_deg=self.float_angle_var.get() + DEFAULT_FLOAT_TILT,
                    mode="still",
                    snap=DEFAULT_FLOAT_JAGGED_SNAP,
                    rotate_deg=DEFAULT_FLOAT_TILT,
                    **kwargs,
                )
            else:
                self.show_status(f"Unknown GIF variant: {animation_variant}")
                return
        except SystemExit:
            messagebox.showerror("Missing dependency", "GIF export requires cairosvg and Pillow")
            return
        if not frames:
            self.show_status("No GIF frames generated")
            return
        duration_ms = int(1000 / max(1, self.gif_fps_var.get()))
        frames[0].save(
            output_path,
            save_all=True,
            append_images=frames[1:],
            duration=duration_ms,
            loop=0,
            disposal=2,
            optimize=False,
        )
        self.show_status(f"Saved {output_path}")

    def show_status(self, message: str) -> None:
        self.status_var.set(message)
        self.root.after(2200, lambda: self.status_var.set(""))


def main() -> None:
    root = tk.Tk()
    app = QrGuiApp(root)
    root.mainloop()


if __name__ == "__main__":
    main()
