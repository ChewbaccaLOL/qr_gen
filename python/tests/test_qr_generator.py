import argparse
import io
import os
import runpy
import sys

import pytest

import qr_generator


class DummyQR:
    def __init__(self, matrix):
        self.matrix = matrix


class DummyImage:
    def __init__(self):
        self.saved = None

    def save(self, *args, **kwargs):
        self.saved = {"args": args, "kwargs": kwargs}


def build_args(tmp_path, **overrides):
    base = argparse.Namespace(
        data="hello",
        output=str(tmp_path / "qr.svg"),
        png=False,
        png_output=None,
        png_scale=3.0,
        gif=False,
        animation=False,
        animation_format=None,
        animation_variant=None,
        gif_output=None,
        gif_variant=None,
        gif_fps=None,
        gif_frames=None,
        gif_hold=None,
        wave_amp=None,
        wave_period=None,
        float_angle=None,
        readable_gif=False,
        pdf=False,
        pdf_output=None,
        ps=False,
        ps_output=None,
        variant="classic",
        scale=10,
        border=4,
        error="m",
        dark=None,
        light=None,
        no_background=False,
        radius=None,
        list_variants=False,
        catalog=False,
        catalog_columns=3,
        catalog_background="#ffffff",
        catalog_label_size=0,
    )
    for key, value in overrides.items():
        setattr(base, key, value)
    return base


def test_derive_output_path():
    assert qr_generator.derive_output_path("out/qr.svg", "png") == "out/qr.png"
    assert qr_generator.derive_output_path("out/qr", ".gif") == "out/qr.gif"


def test_ensure_parent_dir(tmp_path):
    target = tmp_path / "nested" / "qr.svg"
    qr_generator.ensure_parent_dir(str(target))
    assert (tmp_path / "nested").is_dir()


def test_read_data_from_args():
    args = argparse.Namespace(data="hello")
    assert qr_generator.read_data(args) == "hello"


def test_read_data_from_stdin(monkeypatch):
    class DummyStdin(io.StringIO):
        def isatty(self):
            return False

    monkeypatch.setattr(sys, "stdin", DummyStdin("piped"))
    args = argparse.Namespace(data=None)
    assert qr_generator.read_data(args) == "piped"


def test_read_data_from_env(monkeypatch):
    class DummyStdin(io.StringIO):
        def isatty(self):
            return True

    monkeypatch.setattr(sys, "stdin", DummyStdin(""))
    monkeypatch.setenv("QR_DATA", "env-value")
    args = argparse.Namespace(data=None)
    assert qr_generator.read_data(args) == "env-value"


def test_read_data_missing_exits(monkeypatch, capsys):
    class DummyStdin(io.StringIO):
        def isatty(self):
            return True

    monkeypatch.setattr(sys, "stdin", DummyStdin(""))
    monkeypatch.delenv("QR_DATA", raising=False)
    args = argparse.Namespace(data=None)
    with pytest.raises(SystemExit) as exc:
        qr_generator.read_data(args)
    assert exc.value.code == 2
    assert "data is required" in capsys.readouterr().err


def test_parse_args_env_defaults(monkeypatch):
    monkeypatch.setenv("QR_PNG", "true")
    monkeypatch.setenv("QR_SCALE", "7")
    monkeypatch.setenv("QR_VARIANT", "rounded")
    monkeypatch.setenv("QR_ERROR", "h")
    monkeypatch.setenv("QR_NO_BACKGROUND", "true")
    monkeypatch.setenv("QR_PNG_SCALE", "2.5")
    monkeypatch.setattr(sys, "argv", ["prog", "hello"])

    args = qr_generator.parse_args()
    assert args.png is True
    assert args.scale == 7
    assert args.variant == "rounded"
    assert args.error == "h"
    assert args.no_background is True
    assert args.png_scale == 2.5


def test_main_list_variants(monkeypatch, capsys):
    monkeypatch.setattr(qr_generator, "parse_args", lambda: argparse.Namespace(list_variants=True))
    qr_generator.main()
    output = capsys.readouterr().out
    assert "classic" in output
    assert "Animations:" in output


def test_main_catalog_animation_conflict(monkeypatch, capsys):
    args = argparse.Namespace(
        list_variants=False,
        data="hello",
        animation=True,
        gif=False,
        catalog=True,
    )
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    with pytest.raises(SystemExit) as exc:
        qr_generator.main()
    assert exc.value.code == 2
    assert "--catalog" in capsys.readouterr().err


def test_main_png_scale_validation(monkeypatch, tmp_path, capsys):
    args = build_args(tmp_path, png=True, png_scale=0)
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    with pytest.raises(SystemExit) as exc:
        qr_generator.main()
    assert exc.value.code == 2
    assert "--png-scale" in capsys.readouterr().err


def test_main_catalog_default_output(monkeypatch, tmp_path):
    args = build_args(tmp_path, catalog=True, output=None)
    monkeypatch.delenv("QR_OUTPUT", raising=False)
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_catalog_svg", lambda *a, **k: "<svg></svg>")

    if os.path.exists("out/catalog.svg"):
        os.remove("out/catalog.svg")

    qr_generator.main()

    assert os.path.exists("out/catalog.svg")


def test_main_gif_format_mismatch(monkeypatch, tmp_path, capsys):
    args = build_args(tmp_path, gif=True, animation_format="mp4")
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    with pytest.raises(SystemExit) as exc:
        qr_generator.main()
    assert exc.value.code == 2
    assert "--gif" in capsys.readouterr().err


def test_main_export_outputs(monkeypatch, tmp_path):
    calls = []
    args = build_args(tmp_path, png=True, pdf=True, ps=True, png_scale=2.0)
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")

    monkeypatch.setattr(qr_generator, "write_png", lambda *a: calls.append("png"))
    monkeypatch.setattr(qr_generator, "write_pdf", lambda *a: calls.append("pdf"))
    monkeypatch.setattr(qr_generator, "write_ps", lambda *a: calls.append("ps"))

    qr_generator.main()

    assert calls == ["png", "pdf", "ps"]


@pytest.mark.parametrize(
    "override, fragment",
    [
        ({"gif_fps": 0}, "--gif-fps"),
        ({"gif_frames": 0}, "--gif-frames"),
        ({"gif_hold": -1}, "--gif-hold"),
        ({"wave_amp": -0.1}, "--wave-amp"),
        ({"wave_period": 0}, "--wave-period"),
    ],
)
def test_main_gif_validation_errors(monkeypatch, tmp_path, capsys, override, fragment):
    args = build_args(tmp_path, animation_variant="wave", animation=True, **override)
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    monkeypatch.setattr(qr_generator, "build_wave_gif_frames", lambda *a, **k: [DummyImage()])
    with pytest.raises(SystemExit) as exc:
        qr_generator.main()
    assert exc.value.code == 2
    assert fragment in capsys.readouterr().err


def test_main_animation_format_error(monkeypatch, tmp_path, capsys):
    args = build_args(tmp_path, animation=True, animation_format="mp4")
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    with pytest.raises(SystemExit) as exc:
        qr_generator.main()
    assert exc.value.code == 2
    assert "format" in capsys.readouterr().err


def test_main_wave_loop_variant(monkeypatch, tmp_path):
    args = build_args(tmp_path, animation=True, animation_variant="wave-loop")
    captured = {}
    dummy_image = DummyImage()

    def fake_wave_frames(*_, **kwargs):
        captured.update(kwargs)
        return [dummy_image]

    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    monkeypatch.setattr(qr_generator, "build_wave_gif_frames", fake_wave_frames)

    qr_generator.main()

    assert captured["mode"] == "loop"
    assert captured["hold"] == 0
    assert dummy_image.saved is not None


@pytest.mark.parametrize(
    "variant, key, expected",
    [
        ("float", "float_angle_deg", qr_generator.DEFAULT_FLOAT_ANGLE + qr_generator.DEFAULT_FLOAT_TILT),
        ("float-tilt-first", "rotate_mode", "before"),
        ("float-jagged", "snap", qr_generator.DEFAULT_FLOAT_JAGGED_SNAP),
    ],
)
def test_main_float_variants(monkeypatch, tmp_path, variant, key, expected):
    args = build_args(tmp_path, animation=True, animation_variant=variant)
    captured = {}
    dummy_image = DummyImage()

    def fake_float_frames(*_, **kwargs):
        captured.update(kwargs)
        return [dummy_image]

    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    monkeypatch.setattr(qr_generator, "build_float_gif_frames", fake_float_frames)

    qr_generator.main()

    assert captured.get(key) == expected
    assert dummy_image.saved is not None


def test_main_unknown_animation_variant(monkeypatch, tmp_path, capsys):
    args = build_args(tmp_path, animation=True, animation_variant="nope")
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    with pytest.raises(SystemExit) as exc:
        qr_generator.main()
    assert exc.value.code == 2
    assert "variant" in capsys.readouterr().err


def test_main_empty_animation_frames(monkeypatch, tmp_path, capsys):
    args = build_args(tmp_path, animation=True, animation_variant="wave")
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    monkeypatch.setattr(qr_generator, "build_wave_gif_frames", lambda *a, **k: [])
    with pytest.raises(SystemExit) as exc:
        qr_generator.main()
    assert exc.value.code == 2
    assert "no GIF frames" in capsys.readouterr().err


def test_module_main_execution(tmp_path, monkeypatch):
    output = tmp_path / "cli.svg"
    monkeypatch.setattr(sys, "argv", ["qr_generator.py", "hello", "--output", str(output)])
    runpy.run_module("qr_generator", run_name="__main__")
    assert output.exists()


def test_main_readable_gif_defaults(monkeypatch, tmp_path):
    args = build_args(
        tmp_path,
        animation=True,
        readable_gif=True,
        animation_variant="wave",
    )

    captured = {}
    dummy_image = DummyImage()

    def fake_wave_frames(*_, **kwargs):
        captured.update(kwargs)
        return [dummy_image, DummyImage()]

    monkeypatch.delenv("QR_OUTPUT", raising=False)
    monkeypatch.setattr(qr_generator, "parse_args", lambda: args)
    monkeypatch.setattr(qr_generator.segno, "make", lambda data, error: DummyQR([[1]]))
    monkeypatch.setattr(qr_generator, "render_svg", lambda *a, **k: "<svg></svg>")
    monkeypatch.setattr(qr_generator, "build_wave_gif_frames", fake_wave_frames)

    qr_generator.main()

    assert captured["wave_amp"] == qr_generator.READABLE_WAVE_AMP
    assert captured["wave_period"] == qr_generator.READABLE_WAVE_PERIOD
    assert captured["frames"] == qr_generator.READABLE_GIF_FRAMES
    assert captured["hold"] == qr_generator.READABLE_GIF_HOLD
    assert dummy_image.saved is not None
    assert dummy_image.saved["kwargs"]["duration"] == int(1000 / qr_generator.READABLE_GIF_FPS)
