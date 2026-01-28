import json

import qr_gui
from qr_generator import Variant


def test_default_gui_data_value():
    assert qr_gui.DEFAULT_GUI_DATA.startswith("https://")


def test_variant_from_preset_builds_variant():
    entry = {
        "name": "custom",
        "shape": "rounded",
        "dark": "#111111",
        "light": "#ffffff",
        "radius": 0.25,
        "gradient": {"from": "#000000", "to": "#ffffff"},
    }
    variant = qr_gui._variant_from_preset(entry)
    assert isinstance(variant, Variant)
    assert variant.name == "custom"
    assert variant.shape == "rounded"
    assert variant.dark == "#111111"
    assert variant.light == "#ffffff"
    assert variant.radius == 0.25
    assert variant.gradient == {"from": "#000000", "to": "#ffffff"}


def test_variant_from_preset_defaults_shape():
    entry = {"name": "bad", "shape": "hex", "dark": "#000"}
    variant = qr_gui._variant_from_preset(entry)
    assert variant is not None
    assert variant.shape == "square"


def test_build_variant_catalog_includes_presets():
    presets = [{"name": "extra", "shape": "dot", "dark": "#123"}]
    variant_map, variant_order = qr_gui.build_variant_catalog(presets)
    assert "extra" in variant_map
    assert "extra" in variant_order


def test_compute_preview_scale():
    scale = qr_gui._compute_preview_scale(matrix_size=21, target_px=100, border=4)
    assert abs(scale - (100 / 29)) < 0.01


def test_default_output_path_svg():
    path = qr_gui._default_output_path("classic", "svg")
    assert path.endswith("out/qr-classic.svg")


def test_save_and_load_presets(tmp_path):
    entries = [{"name": "one", "shape": "square", "dark": "#000"}]
    file_path = tmp_path / "presets.json"
    assert qr_gui.save_presets(entries, path=str(file_path))
    loaded = qr_gui.load_presets(path=str(file_path))
    assert loaded == entries


def test_load_presets_rejects_non_list(tmp_path):
    file_path = tmp_path / "presets.json"
    file_path.write_text(json.dumps({"name": "bad"}), encoding="utf-8")
    loaded = qr_gui.load_presets(path=str(file_path))
    assert loaded == []


def test_build_gradient_requires_enabled_and_colors():
    assert qr_gui._build_gradient(False, "#111111", "#222222") is None
    assert qr_gui._build_gradient(True, "", "#222222") is None
    assert qr_gui._build_gradient(True, "#111111", "  ") is None
    assert qr_gui._build_gradient(True, "#111111", "#222222") == {
        "id": "fg",
        "from": "#111111",
        "to": "#222222",
    }


def test_should_reload_variant_respects_custom_dirty():
    assert qr_gui._should_reload_variant("classic", "sunset", False) is True
    assert qr_gui._should_reload_variant("classic", "classic", True) is False
    assert qr_gui._should_reload_variant("classic", "classic", False) is True
    assert qr_gui._should_reload_variant("classic", None, False) is False


def test_strip_known_extension():
    extensions = ("svg", "png", "pdf", "ps")
    assert qr_gui._strip_known_extension("qr.svg", extensions) == "qr"
    assert qr_gui._strip_known_extension("qr.PNG", extensions) == "qr"
    assert qr_gui._strip_known_extension("my.qr.svg", extensions) == "my.qr"
    assert qr_gui._strip_known_extension("my.qr", extensions) == "my.qr"
    assert qr_gui._strip_known_extension("", extensions) == ""


def test_clamp_zoom():
    assert qr_gui._clamp_zoom(1.0, 0.5, 2.0) == 1.0
    assert qr_gui._clamp_zoom(0.1, 0.5, 2.0) == 0.5
    assert qr_gui._clamp_zoom(3.0, 0.5, 2.0) == 2.0
