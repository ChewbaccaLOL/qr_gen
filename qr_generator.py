import argparse
import io
import math
import os
import sys
from dataclasses import dataclass
from typing import Dict, Iterable, List, Optional, Tuple

import segno


@dataclass(frozen=True)
class Variant:
    name: str
    shape: str
    dark: str
    light: Optional[str]
    radius: float = 0.0
    gradient: Optional[Dict[str, str]] = None


VARIANTS: Dict[str, Variant] = {
    "classic": Variant("classic", "square", "#000000", "#ffffff"),
    "square": Variant("square", "square", "#000000", "#ffffff"),
    "rounded": Variant("rounded", "rounded", "#111111", "#ffffff", radius=0.28),
    "dot": Variant("dot", "dot", "#111111", "#ffffff"),
    "clear": Variant("clear", "square", "#111111", None),
    "clear-rounded": Variant("clear-rounded", "rounded", "#111111", None, radius=0.28),
    "clear-dot": Variant("clear-dot", "dot", "#111111", None),
    "inverted": Variant("inverted", "square", "#ffffff", "#000000"),
    "midnight": Variant("midnight", "square", "#e6f1ff", "#0b1020"),
    "sunset": Variant(
        "sunset",
        "rounded",
        "#2a0a44",
        "#fff2e3",
        radius=0.3,
        gradient={"id": "fg", "from": "#ff7a59", "to": "#7a2cff"},
    ),
    "neon": Variant(
        "neon",
        "dot",
        "#00f0ff",
        "#06060a",
        gradient={"id": "fg", "from": "#00f0ff", "to": "#6bff2e"},
    ),
}

DEFAULT_GIF_FPS = 12
DEFAULT_GIF_FRAMES = 20
DEFAULT_GIF_HOLD = 12
DEFAULT_WAVE_AMP = 0.45
DEFAULT_WAVE_PERIOD = 10.0

READABLE_GIF_FPS = 12
READABLE_GIF_FRAMES = 16
READABLE_GIF_HOLD = 16
READABLE_WAVE_AMP = 0.28
READABLE_WAVE_PERIOD = 14.0

def load_dotenv(path: str = ".env") -> None:
    if not os.path.isfile(path):
        return
    try:
        with open(path, "r", encoding="utf-8") as handle:
            for raw_line in handle:
                line = raw_line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, value = line.split("=", 1)
                key = key.strip()
                value = value.strip()
                if not key:
                    continue
                if len(value) >= 2 and value[0] == value[-1] and value[0] in ("\"", "'"):
                    value = value[1:-1]
                if key not in os.environ:
                    os.environ[key] = value
    except OSError:
        return


def env_str(name: str, default: Optional[str] = None) -> Optional[str]:
    return os.getenv(name, default)


def env_int(name: str, default: Optional[int] = None) -> Optional[int]:
    value = os.getenv(name)
    if value is None:
        return default
    try:
        return int(value)
    except ValueError:
        return default


def env_float(name: str, default: Optional[float] = None) -> Optional[float]:
    value = os.getenv(name)
    if value is None:
        return default
    try:
        return float(value)
    except ValueError:
        return default


def env_bool(name: str, default: bool = False) -> bool:
    value = os.getenv(name)
    if value is None:
        return default
    value = value.strip().lower()
    if value in ("1", "true", "yes", "y", "on", "t"):
        return True
    if value in ("0", "false", "no", "n", "off", "f"):
        return False
    return default


def parse_args() -> argparse.Namespace:
    load_dotenv()
    parser = argparse.ArgumentParser(
        description="Generate stylized QR codes as SVG (optional PNG/PDF/PS)."
    )
    parser.add_argument(
        "data",
        nargs="?",
        help="Text or URL to encode. If omitted, read from stdin.",
    )
    parser.add_argument(
        "-o",
        "--output",
        default=None,
        help="Output SVG file (default: out/qr.svg or QR_OUTPUT).",
    )
    parser.add_argument(
        "--png",
        action="store_true",
        default=env_bool("QR_PNG", False),
        help="Also write a PNG file alongside the SVG (requires cairosvg).",
    )
    parser.add_argument(
        "--png-output",
        default=None,
        help="PNG output file (default: derived from --output).",
    )
    parser.add_argument(
        "--png-scale",
        type=float,
        default=env_float("QR_PNG_SCALE", 3.0),
        help="Scale factor for PNG export (default: 3.0).",
    )
    parser.add_argument(
        "--gif",
        action="store_true",
        default=env_bool("QR_GIF", False),
        help="Alias for --animation --animation-format gif.",
    )
    parser.add_argument(
        "--animation",
        action="store_true",
        default=env_bool("QR_ANIMATION", False),
        help="Also write an animated output (default format: gif).",
    )
    parser.add_argument(
        "--animation-format",
        default=env_str("QR_ANIMATION_FORMAT"),
        choices=["gif"],
        help="Animation format (default: gif).",
    )
    parser.add_argument(
        "--animation-variant",
        default=env_str("QR_ANIMATION_VARIANT"),
        choices=["wave"],
        help="Animation variant (default: wave).",
    )
    parser.add_argument(
        "--gif-output",
        default=None,
        help="GIF output file (default: derived from --output).",
    )
    parser.add_argument(
        "--gif-variant",
        default=None,
        choices=["wave"],
        help="GIF animation variant (default: wave).",
    )
    parser.add_argument(
        "--gif-fps",
        type=int,
        default=None,
        help="Frames per second for GIF animation.",
    )
    parser.add_argument(
        "--gif-frames",
        type=int,
        default=None,
        help="Number of wave animation frames (loop segment).",
    )
    parser.add_argument(
        "--gif-hold",
        type=int,
        default=None,
        help="Number of still frames before/after the wave.",
    )
    parser.add_argument(
        "--wave-amp",
        type=float,
        default=None,
        help="Wave amplitude in module units (default: expressive).",
    )
    parser.add_argument(
        "--wave-period",
        type=float,
        default=None,
        help="Wave period in columns (modules).",
    )
    parser.add_argument(
        "--readable-gif",
        action="store_true",
        default=env_bool("QR_READABLE_GIF", False),
        help="Use scan-safer defaults for GIF wave settings.",
    )
    parser.add_argument(
        "--pdf",
        action="store_true",
        default=env_bool("QR_PDF", False),
        help="Also write a PDF file alongside the SVG (requires cairosvg).",
    )
    parser.add_argument(
        "--pdf-output",
        default=None,
        help="PDF output file (default: derived from --output).",
    )
    parser.add_argument(
        "--ps",
        action="store_true",
        default=env_bool("QR_PS", False),
        help="Also write a PostScript file alongside the SVG (requires cairosvg).",
    )
    parser.add_argument(
        "--ps-output",
        default=None,
        help="PostScript output file (default: derived from --output).",
    )
    parser.add_argument(
        "-v",
        "--variant",
        default=env_str("QR_VARIANT", "classic"),
        choices=sorted(VARIANTS.keys()),
        help="Visual style variant.",
    )
    parser.add_argument(
        "--scale",
        type=int,
        default=env_int("QR_SCALE", 10),
        help="Module size in pixels (default: 10)",
    )
    parser.add_argument(
        "--border",
        type=int,
        default=env_int("QR_BORDER", 4),
        help="Quiet zone in modules (default: 4)",
    )
    parser.add_argument(
        "--error",
        default=env_str("QR_ERROR", "m"),
        choices=["l", "m", "q", "h"],
        help="Error correction level (default: m)",
    )
    parser.add_argument(
        "--dark",
        default=env_str("QR_DARK"),
        help="Override foreground color (hex or CSS color).",
    )
    parser.add_argument(
        "--light",
        default=env_str("QR_LIGHT"),
        help="Override background color (hex or CSS color).",
    )
    parser.add_argument(
        "--no-background",
        action="store_true",
        default=env_bool("QR_NO_BACKGROUND", False),
        help="Make the background transparent.",
    )
    parser.add_argument(
        "--radius",
        type=float,
        default=env_float("QR_RADIUS"),
        help="Corner radius for rounded modules (0-0.5).",
    )
    parser.add_argument(
        "--list-variants",
        action="store_true",
        help="List available variants and exit.",
    )
    parser.add_argument(
        "--catalog",
        action="store_true",
        default=env_bool("QR_CATALOG", False),
        help="Generate a catalog grid containing all variants.",
    )
    parser.add_argument(
        "--catalog-columns",
        type=int,
        default=env_int("QR_CATALOG_COLUMNS", 3),
        help="Number of columns in the catalog grid (default: 3).",
    )
    parser.add_argument(
        "--catalog-background",
        default=env_str("QR_CATALOG_BACKGROUND", "#ffffff"),
        help="Background color for the catalog canvas (default: #ffffff).",
    )
    parser.add_argument(
        "--catalog-label-size",
        type=int,
        default=env_int("QR_CATALOG_LABEL_SIZE", 0),
        help="Label font size for catalog (0 = auto).",
    )
    return parser.parse_args()


def read_data(args: argparse.Namespace) -> str:
    if args.data:
        return args.data
    if not sys.stdin.isatty():
        value = sys.stdin.read().strip()
        if value:
            return value
    env_value = env_str("QR_DATA")
    if env_value:
        return env_value
    print("error: data is required (pass text or pipe via stdin)", file=sys.stderr)
    sys.exit(2)


def svg_gradient_def(
    gradient: Optional[Dict[str, str]], gradient_id: Optional[str]
) -> str:
    if not gradient or not gradient_id:
        return ""
    color_from = gradient.get("from", "#000000")
    color_to = gradient.get("to", "#ffffff")
    return (
        f"<linearGradient id=\"{gradient_id}\" x1=\"0%\" y1=\"0%\" x2=\"100%\" y2=\"100%\">"
        f"<stop offset=\"0%\" stop-color=\"{color_from}\"/>"
        f"<stop offset=\"100%\" stop-color=\"{color_to}\"/>"
        "</linearGradient>"
    )


def module_fill(dark: str, gradient_id: Optional[str]) -> str:
    if gradient_id:
        return f"url(#{gradient_id})"
    return dark


def escape_xml(text: str) -> str:
    return (
        text.replace("&", "&amp;")
        .replace("<", "&lt;")
        .replace(">", "&gt;")
        .replace('"', "&quot;")
    )


def render_modules(
    matrix,
    scale: int,
    border: int,
    fill: str,
    shape: str,
    radius: float,
    offset: Tuple[int, int] = (0, 0),
    column_offsets: Optional[List[float]] = None,
) -> Iterable[str]:
    offset_x, offset_y = offset
    if shape == "square":
        for y, row in enumerate(matrix):
            for x, cell in enumerate(row):
                if cell:
                    column_offset = column_offsets[x] if column_offsets else 0.0
                    px = (x + border) * scale + offset_x
                    py = (y + border) * scale + offset_y + column_offset
                    yield (
                        f'<rect x="{px}" y="{py}" width="{scale}" height="{scale}" fill="{fill}"/>'
                    )
    elif shape == "rounded":
        corner = max(0.0, min(radius, 0.5)) * scale
        for y, row in enumerate(matrix):
            for x, cell in enumerate(row):
                if cell:
                    column_offset = column_offsets[x] if column_offsets else 0.0
                    px = (x + border) * scale + offset_x
                    py = (y + border) * scale + offset_y + column_offset
                    yield (
                        f'<rect x="{px}" y="{py}" width="{scale}" height="{scale}" '
                        f'rx="{corner}" ry="{corner}" fill="{fill}"/>'
                    )
    elif shape == "dot":
        r = scale * 0.45
        offset_center = scale / 2
        for y, row in enumerate(matrix):
            for x, cell in enumerate(row):
                if cell:
                    column_offset = column_offsets[x] if column_offsets else 0.0
                    cx = (x + border) * scale + offset_center + offset_x
                    cy = (y + border) * scale + offset_center + offset_y + column_offset
                    yield f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{fill}"/>'
    else:
        raise ValueError(f"Unknown shape: {shape}")


def render_svg(
    matrix,
    scale: int,
    border: int,
    dark: str,
    light: Optional[str],
    shape: str,
    radius: float,
    gradient: Optional[Dict[str, str]],
    column_offsets: Optional[List[float]] = None,
    extra_pad_y: int = 0,
) -> str:
    size = len(matrix)
    total_modules = size + border * 2
    dimension = total_modules * scale
    width = dimension
    height = dimension + extra_pad_y * 2
    shape_rendering = "crispEdges" if shape == "square" else "geometricPrecision"
    gradient_id = gradient.get("id", "fg") if gradient else None
    fill = module_fill(dark, gradient_id)

    parts = [
        '<?xml version="1.0" encoding="UTF-8"?>',
        (
            f'<svg xmlns="http://www.w3.org/2000/svg" '
            f'width="{width}" height="{height}" '
            f'viewBox="0 0 {width} {height}" '
            f'shape-rendering="{shape_rendering}">'
        ),
    ]
    gradient_def = svg_gradient_def(gradient, gradient_id)
    if gradient_def:
        parts.append(f"<defs>{gradient_def}</defs>")
    if light is not None:
        parts.append(
            f'<rect width="100%" height="100%" fill="{light}"/>'
        )

    parts.extend(
        render_modules(
            matrix,
            scale=scale,
            border=border,
            fill=fill,
            shape=shape,
            radius=radius,
            offset=(0, extra_pad_y),
            column_offsets=column_offsets,
        )
    )

    parts.append("</svg>")
    return "\n".join(parts)


def render_catalog_svg(
    matrix,
    scale: int,
    border: int,
    variants: Iterable[Variant],
    columns: int,
    background: str,
    label_size: int,
) -> str:
    variants = list(variants)
    if not variants:
        raise ValueError("No variants available for catalog.")

    size = len(matrix)
    tile_dim = (size + border * 2) * scale
    padding = max(8, int(scale * 1.2))
    if label_size <= 0:
        label_size = max(10, int(scale * 1.4))
    label_height = int(label_size * 1.6)
    tile_total_height = tile_dim + label_height + padding
    columns = max(1, columns)
    rows = int(math.ceil(len(variants) / columns))
    width = columns * tile_dim + (columns + 1) * padding
    height = rows * tile_total_height + padding

    parts = [
        '<?xml version="1.0" encoding="UTF-8"?>',
        (
            f'<svg xmlns="http://www.w3.org/2000/svg" '
            f'width="{width}" height="{height}" '
            f'viewBox="0 0 {width} {height}" '
            'shape-rendering="geometricPrecision">'
        ),
    ]

    gradient_defs = []
    for variant in variants:
        if variant.gradient:
            gradient_id = f"fg-{variant.name}"
            gradient_defs.append(svg_gradient_def(variant.gradient, gradient_id))
    if gradient_defs:
        parts.append(f"<defs>{''.join(gradient_defs)}</defs>")

    parts.append(f'<rect width="100%" height="100%" fill="{background}"/>')

    for index, variant in enumerate(variants):
        col = index % columns
        row = index // columns
        origin_x = padding + col * (tile_dim + padding)
        origin_y = padding + row * tile_total_height
        tile_bg = variant.light if variant.light is not None else background
        parts.append(
            f'<rect x="{origin_x}" y="{origin_y}" '
            f'width="{tile_dim}" height="{tile_dim}" fill="{tile_bg}"/>'
        )

        gradient_id = f"fg-{variant.name}" if variant.gradient else None
        fill = module_fill(variant.dark, gradient_id)
        parts.extend(
            render_modules(
                matrix,
                scale=scale,
                border=border,
                fill=fill,
                shape=variant.shape,
                radius=variant.radius,
                offset=(origin_x, origin_y),
            )
        )

        label_x = origin_x + tile_dim / 2
        label_y = origin_y + tile_dim + label_height * 0.75
        parts.append(
            f'<text x="{label_x}" y="{label_y}" '
            f'font-size="{label_size}" '
            'font-family="Helvetica, Arial, sans-serif" '
            'fill="#1a1a1a" text-anchor="middle">'
            f"{escape_xml(variant.name)}"
            "</text>"
        )

    parts.append("</svg>")
    return "\n".join(parts)


def derive_output_path(svg_path: str, extension: str) -> str:
    extension = extension if extension.startswith(".") else f".{extension}"
    if svg_path.lower().endswith(".svg"):
        return f"{svg_path[:-4]}{extension}"
    return f"{svg_path}{extension}"


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


def compute_wave_offsets(size: int, amplitude_px: float, period: float, phase: float) -> List[float]:
    if amplitude_px == 0:
        return [0.0 for _ in range(size)]
    return [
        amplitude_px * math.sin((2 * math.pi * (x / period)) + phase)
        for x in range(size)
    ]


def build_wave_gif_frames(
    matrix,
    scale: int,
    border: int,
    dark: str,
    light: Optional[str],
    shape: str,
    radius: float,
    gradient: Optional[Dict[str, str]],
    wave_amp: float,
    wave_period: float,
    frames: int,
    hold: int,
) -> List["Image.Image"]:
    Image = require_pillow()
    size = len(matrix)
    amplitude_px = wave_amp * scale
    extra_pad_y = int(math.ceil(abs(amplitude_px))) + 1 if amplitude_px else 0
    phase_step = (2 * math.pi) / frames if frames > 0 else 0.0
    still_offsets = [0.0 for _ in range(size)]

    images: List["Image.Image"] = []
    total_frames = hold + frames + hold
    for frame_index in range(total_frames):
        if frame_index < hold or frame_index >= hold + frames:
            offsets = still_offsets
        else:
            phase = phase_step * (frame_index - hold)
            offsets = compute_wave_offsets(size, amplitude_px, wave_period, phase)
        svg_frame = render_svg(
            matrix,
            scale=scale,
            border=border,
            dark=dark,
            light=light,
            shape=shape,
            radius=radius,
            gradient=gradient,
            column_offsets=offsets,
            extra_pad_y=extra_pad_y,
        )
        png_bytes = svg_to_png_bytes(svg_frame, scale=1.0)
        image = Image.open(io.BytesIO(png_bytes))
        image.load()
        images.append(image)
    return images


def ensure_parent_dir(path: str) -> None:
    directory = os.path.dirname(path)
    if directory:
        os.makedirs(directory, exist_ok=True)


def main() -> None:
    args = parse_args()
    if args.list_variants:
        for name in sorted(VARIANTS.keys()):
            print(name)
        return

    data = read_data(args)
    animation_enabled = args.animation or args.gif
    if args.catalog and animation_enabled:
        print("error: animation output is not supported with --catalog", file=sys.stderr)
        sys.exit(2)
    had_output_flag = args.output is not None
    env_output = os.getenv("QR_OUTPUT")
    default_output = env_output or "out/qr.svg"
    args.output = args.output or default_output
    if args.catalog and not had_output_flag and env_output is None:
        args.output = "out/catalog.svg"

    if args.png_output is None:
        args.png_output = env_str("QR_PNG_OUTPUT")
    if args.pdf_output is None:
        args.pdf_output = env_str("QR_PDF_OUTPUT")
    if args.ps_output is None:
        args.ps_output = env_str("QR_PS_OUTPUT")
    if args.gif_output is None:
        args.gif_output = env_str("QR_GIF_OUTPUT")
    if args.gif_variant is None:
        args.gif_variant = env_str("QR_GIF_VARIANT")
    if args.gif_fps is None:
        args.gif_fps = env_int("QR_GIF_FPS")
    if args.gif_frames is None:
        args.gif_frames = env_int("QR_GIF_FRAMES")
    if args.gif_hold is None:
        args.gif_hold = env_int("QR_GIF_HOLD")
    if args.wave_amp is None:
        args.wave_amp = env_float("QR_WAVE_AMP")
    if args.wave_period is None:
        args.wave_period = env_float("QR_WAVE_PERIOD")

    animation_format = args.animation_format or "gif"
    if args.gif and animation_format != "gif":
        print("error: --gif can only be used with GIF output", file=sys.stderr)
        sys.exit(2)
    animation_variant = args.animation_variant or args.gif_variant or "wave"

    if args.readable_gif:
        gif_defaults = {
            "fps": READABLE_GIF_FPS,
            "frames": READABLE_GIF_FRAMES,
            "hold": READABLE_GIF_HOLD,
            "amp": READABLE_WAVE_AMP,
            "period": READABLE_WAVE_PERIOD,
        }
    else:
        gif_defaults = {
            "fps": DEFAULT_GIF_FPS,
            "frames": DEFAULT_GIF_FRAMES,
            "hold": DEFAULT_GIF_HOLD,
            "amp": DEFAULT_WAVE_AMP,
            "period": DEFAULT_WAVE_PERIOD,
        }

    if args.gif_fps is None:
        args.gif_fps = gif_defaults["fps"]
    if args.gif_frames is None:
        args.gif_frames = gif_defaults["frames"]
    if args.gif_hold is None:
        args.gif_hold = gif_defaults["hold"]
    if args.wave_amp is None:
        args.wave_amp = gif_defaults["amp"]
    if args.wave_period is None:
        args.wave_period = gif_defaults["period"]

    qr = segno.make(data, error=args.error)

    if args.catalog:
        svg = render_catalog_svg(
            qr.matrix,
            scale=args.scale,
            border=args.border,
            variants=[VARIANTS[name] for name in sorted(VARIANTS.keys())],
            columns=args.catalog_columns,
            background=args.catalog_background,
            label_size=args.catalog_label_size,
        )
    else:
        variant = VARIANTS[args.variant]
        dark = args.dark or variant.dark
        light = None if args.no_background else (args.light or variant.light)
        radius = args.radius if args.radius is not None else variant.radius
        gradient = None if args.dark else variant.gradient
        svg = render_svg(
            qr.matrix,
            scale=args.scale,
            border=args.border,
            dark=dark,
            light=light,
            shape=variant.shape,
            radius=radius,
            gradient=gradient,
        )

    ensure_parent_dir(args.output)
    with open(args.output, "w", encoding="utf-8") as handle:
        handle.write(svg)
    print(f"Saved {args.output}")

    if args.png:
        if args.png_scale <= 0:
            print("error: --png-scale must be greater than 0", file=sys.stderr)
            sys.exit(2)
        png_path = args.png_output or derive_output_path(args.output, ".png")
        ensure_parent_dir(png_path)
        write_png(svg, png_path, args.png_scale)
        print(f"Saved {png_path}")
    if args.pdf:
        pdf_path = args.pdf_output or derive_output_path(args.output, ".pdf")
        ensure_parent_dir(pdf_path)
        write_pdf(svg, pdf_path)
        print(f"Saved {pdf_path}")
    if args.ps:
        ps_path = args.ps_output or derive_output_path(args.output, ".ps")
        ensure_parent_dir(ps_path)
        write_ps(svg, ps_path)
        print(f"Saved {ps_path}")
    if animation_enabled:
        if animation_format != "gif":
            print(
                f"error: animation format '{animation_format}' is not supported yet",
                file=sys.stderr,
            )
            sys.exit(2)
        if args.gif_fps <= 0:
            print("error: --gif-fps must be greater than 0", file=sys.stderr)
            sys.exit(2)
        if args.gif_frames <= 0:
            print("error: --gif-frames must be greater than 0", file=sys.stderr)
            sys.exit(2)
        if args.gif_hold < 0:
            print("error: --gif-hold must be 0 or greater", file=sys.stderr)
            sys.exit(2)
        if args.wave_amp < 0:
            print("error: --wave-amp must be 0 or greater", file=sys.stderr)
            sys.exit(2)
        if args.wave_period <= 0:
            print("error: --wave-period must be greater than 0", file=sys.stderr)
            sys.exit(2)
        gif_path = args.gif_output or derive_output_path(args.output, ".gif")
        ensure_parent_dir(gif_path)
        variant = VARIANTS[args.variant]
        dark = args.dark or variant.dark
        light = None if args.no_background else (args.light or variant.light)
        radius = args.radius if args.radius is not None else variant.radius
        gradient = None if args.dark else variant.gradient
        if animation_variant == "wave":
            frames = build_wave_gif_frames(
                qr.matrix,
                scale=args.scale,
                border=args.border,
                dark=dark,
                light=light,
                shape=variant.shape,
                radius=radius,
                gradient=gradient,
                wave_amp=args.wave_amp,
                wave_period=args.wave_period,
                frames=args.gif_frames,
                hold=args.gif_hold,
            )
        else:
            print(
                f"error: animation variant '{animation_variant}' is not supported yet",
                file=sys.stderr,
            )
            sys.exit(2)
        if not frames:
            print("error: no GIF frames generated", file=sys.stderr)
            sys.exit(2)
        duration_ms = int(1000 / args.gif_fps)
        frames[0].save(
            gif_path,
            save_all=True,
            append_images=frames[1:],
            duration=duration_ms,
            loop=0,
            disposal=2,
            optimize=False,
        )
        print(f"Saved {gif_path}")


if __name__ == "__main__":
    main()
