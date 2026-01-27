import sys


def require_cairosvg():
    try:
        import cairosvg
    except ImportError:
        print(
            "error: PNG/PDF/PS export requires cairosvg (pip install cairosvg)",
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
