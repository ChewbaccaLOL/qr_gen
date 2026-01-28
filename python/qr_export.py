import os
import sys


def prepare_cairo() -> None:
    if sys.platform != "win32":
        return
    candidates = []
    base = getattr(sys, "_MEIPASS", None)
    if base and os.path.isdir(base):
        candidates.append(base)
        cairo_dir = os.path.join(base, "cairo")
        if os.path.isdir(cairo_dir):
            candidates.append(cairo_dir)
    exe_dir = os.path.dirname(sys.executable)
    if exe_dir and os.path.isdir(exe_dir):
        candidates.append(exe_dir)
        cairo_dir = os.path.join(exe_dir, "cairo")
        if os.path.isdir(cairo_dir):
            candidates.append(cairo_dir)
    if not candidates:
        return
    for folder in candidates:
        try:
            os.add_dll_directory(folder)
        except (AttributeError, OSError):
            continue
    os.environ["PATH"] = os.pathsep.join(candidates + [os.environ.get("PATH", "")])


def require_cairosvg():
    try:
        prepare_cairo()
        import cairosvg
    except Exception:
        print(
            "error: PNG/PDF/PS export requires cairosvg + cairo (pip install cairosvg). "
            "On Windows, the Cairo DLLs must be bundled or installed.",
            file=sys.stderr,
        )
        sys.exit(2)
    return cairosvg


def require_pillow():
    try:
        from PIL import Image
    except ImportError:
        print(
            "error: GIF export requires Pillow (pip install pillow)",
            file=sys.stderr,
        )
        sys.exit(2)
    return Image


def write_png(svg_text: str, output_path: str, scale: float) -> None:
    cairosvg = require_cairosvg()
    cairosvg.svg2png(
        bytestring=svg_text.encode("utf-8"),
        write_to=output_path,
        scale=scale,
    )


def write_pdf(svg_text: str, output_path: str) -> None:
    cairosvg = require_cairosvg()
    cairosvg.svg2pdf(bytestring=svg_text.encode("utf-8"), write_to=output_path)


def write_ps(svg_text: str, output_path: str) -> None:
    cairosvg = require_cairosvg()
    cairosvg.svg2ps(bytestring=svg_text.encode("utf-8"), write_to=output_path)


def svg_to_png_bytes(svg_text: str, scale: float = 1.0) -> bytes:
    cairosvg = require_cairosvg()
    return cairosvg.svg2png(bytestring=svg_text.encode("utf-8"), scale=scale)
