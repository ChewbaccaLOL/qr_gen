import builtins
import sys
import types

import pytest

import qr_export


def test_require_cairosvg_missing(monkeypatch, capsys):
    real_import = builtins.__import__

    def fake_import(name, globals=None, locals=None, fromlist=(), level=0):
        if name == "cairosvg":
            raise ImportError("blocked")
        return real_import(name, globals, locals, fromlist, level)

    monkeypatch.setattr(builtins, "__import__", fake_import)
    with pytest.raises(SystemExit) as exc:
        qr_export.require_cairosvg()
    assert exc.value.code == 2
    assert "cairosvg" in capsys.readouterr().err


def test_require_cairosvg_present(monkeypatch):
    dummy = types.SimpleNamespace(svg2png=lambda **kwargs: b"ok")
    monkeypatch.setitem(sys.modules, "cairosvg", dummy)
    assert qr_export.require_cairosvg() is dummy


def test_require_pillow_missing(monkeypatch, capsys):
    real_import = builtins.__import__

    def fake_import(name, globals=None, locals=None, fromlist=(), level=0):
        if name == "PIL":
            raise ImportError("blocked")
        return real_import(name, globals, locals, fromlist, level)

    monkeypatch.setattr(builtins, "__import__", fake_import)
    with pytest.raises(SystemExit) as exc:
        qr_export.require_pillow()
    assert exc.value.code == 2
    assert "Pillow" in capsys.readouterr().err


def test_require_pillow_present(monkeypatch):
    class DummyImage:
        pass

    dummy = types.SimpleNamespace(Image=DummyImage)
    monkeypatch.setitem(sys.modules, "PIL", dummy)
    assert qr_export.require_pillow() is DummyImage


def test_svg_to_png_bytes(monkeypatch):
    class DummyCairo:
        def svg2png(self, **kwargs):
            return b"png-bytes"

    monkeypatch.setattr(qr_export, "require_cairosvg", lambda: DummyCairo())
    assert qr_export.svg_to_png_bytes("<svg></svg>") == b"png-bytes"


def test_write_png_pdf_ps(monkeypatch):
    calls = []

    class DummyCairo:
        def svg2png(self, **kwargs):
            calls.append(("png", kwargs))

        def svg2pdf(self, **kwargs):
            calls.append(("pdf", kwargs))

        def svg2ps(self, **kwargs):
            calls.append(("ps", kwargs))

    monkeypatch.setattr(qr_export, "require_cairosvg", lambda: DummyCairo())
    qr_export.write_png("<svg></svg>", "out.png", 2.0)
    qr_export.write_pdf("<svg></svg>", "out.pdf")
    qr_export.write_ps("<svg></svg>", "out.ps")

    assert calls[0][0] == "png"
    assert calls[1][0] == "pdf"
    assert calls[2][0] == "ps"
