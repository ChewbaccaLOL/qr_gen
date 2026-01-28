import pytest

try:
    import tkinter as tk
except Exception:  # pragma: no cover - optional runtime dependency
    tk = None

from legacy import qr_gui_tk as qr_gui
from legacy.qr_gui_tk import (
    MAX_PREVIEW_INPUT_LEN,
    MAX_PREVIEW_MATRIX,
    QrGuiApp,
    ScrollableFrame,
    ui,
)


class DummyStyle:
    def __init__(self, themes):
        self._themes = list(themes)
        self.used = None

    def theme_names(self):
        return list(self._themes)

    def theme_use(self, name=None):
        if name is None:
            return self.used or (self._themes[0] if self._themes else "")
        self.used = name
        return name


def _make_root():
    if tk is None:
        pytest.skip("tkinter not available")
    try:
        root = tk.Tk()
    except tk.TclError as exc:
        pytest.skip(f"Tk not available: {exc}")
    root.withdraw()
    return root


def test_scrollable_frame_event_area_filters():
    root = _make_root()
    scroll = ScrollableFrame(root)
    scroll.pack()
    inside = tk.Label(scroll.inner, text="Inside")
    inside.pack()
    outside = tk.Label(root, text="Outside")
    outside.pack()

    assert scroll._event_in_scroll_area(inside) is True
    assert scroll._event_in_scroll_area(outside) is False

    root.destroy()


def test_update_grid_layout_regrids_unmanaged_cards():
    root = _make_root()
    app = QrGuiApp(root)
    root.update_idletasks()

    for card in app.cards.values():
        card.frame.grid_forget()

    canvas_width = app.scroll.canvas.winfo_width()
    card_width = app.preview_size + ui(40)
    columns = max(1, int(canvas_width // max(1, card_width)))
    app.card_columns = columns

    app.update_grid_layout(force=False)

    assert any(card.frame.winfo_manager() == "grid" for card in app.cards.values())

    root.destroy()


def test_disable_preview_renderer_sets_flag():
    root = _make_root()
    app = QrGuiApp(root)
    app.disable_preview_renderer("boom")
    assert app.preview_renderer_ok is False
    root.destroy()


def test_update_previews_skips_large_data(monkeypatch):
    root = _make_root()
    app = QrGuiApp(root)
    app.data_var.set("x")

    class DummyQR:
        def __init__(self, size):
            self.matrix = [[1] * size for _ in range(size)]

    monkeypatch.setattr(qr_gui.segno, "make", lambda _data, error: DummyQR(MAX_PREVIEW_MATRIX + 1))
    app.update_previews()
    assert app.preview_skip_size == MAX_PREVIEW_MATRIX + 1
    root.destroy()


def test_update_previews_skips_long_input(monkeypatch):
    root = _make_root()
    app = QrGuiApp(root)
    app.data_var.set("x" * (MAX_PREVIEW_INPUT_LEN + 1))

    def fail_make(_data, error):
        raise AssertionError("segno.make should not be called for long input")

    monkeypatch.setattr(qr_gui.segno, "make", fail_make)
    app.update_previews()
    assert app.preview_skip_token == ("len", MAX_PREVIEW_INPUT_LEN + 1)
    root.destroy()


def test_compute_tk_scaling_uses_dpi_for_windows():
    scale = qr_gui._compute_tk_scaling(144.0, is_windows=True)
    assert scale == 2.0


def test_select_ttk_theme_prefers_native_on_windows():
    style = DummyStyle(["clam", "vista", "alt"])
    theme = qr_gui._select_ttk_theme(style, prefer_native=True, os_name="nt", platform="win32")
    assert theme == "vista"
    assert style.used == "vista"


def test_select_ttk_theme_falls_back_to_clam():
    style = DummyStyle(["clam", "alt"])
    theme = qr_gui._select_ttk_theme(style, prefer_native=False, os_name="nt", platform="win32")
    assert theme == "clam"
