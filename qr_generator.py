import argparse
import math
import os
import sys
from dataclasses import dataclass
from typing import Dict, Iterable, Optional, Tuple

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
        description="Generate stylized QR codes as SVG (optional PNG)."
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
) -> Iterable[str]:
    offset_x, offset_y = offset
    if shape == "square":
        for y, row in enumerate(matrix):
            for x, cell in enumerate(row):
                if cell:
                    px = (x + border) * scale + offset_x
                    py = (y + border) * scale + offset_y
                    yield (
                        f'<rect x="{px}" y="{py}" width="{scale}" height="{scale}" fill="{fill}"/>'
                    )
    elif shape == "rounded":
        corner = max(0.0, min(radius, 0.5)) * scale
        for y, row in enumerate(matrix):
            for x, cell in enumerate(row):
                if cell:
                    px = (x + border) * scale + offset_x
                    py = (y + border) * scale + offset_y
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
                    cx = (x + border) * scale + offset_center + offset_x
                    cy = (y + border) * scale + offset_center + offset_y
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
) -> str:
    size = len(matrix)
    total_modules = size + border * 2
    dimension = total_modules * scale
    shape_rendering = "crispEdges" if shape == "square" else "geometricPrecision"
    gradient_id = gradient.get("id", "fg") if gradient else None
    fill = module_fill(dark, gradient_id)

    parts = [
        '<?xml version="1.0" encoding="UTF-8"?>',
        (
            f'<svg xmlns="http://www.w3.org/2000/svg" '
            f'width="{dimension}" height="{dimension}" '
            f'viewBox="0 0 {dimension} {dimension}" '
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


def derive_png_path(svg_path: str) -> str:
    if svg_path.lower().endswith(".svg"):
        return f"{svg_path[:-4]}.png"
    return f"{svg_path}.png"


def write_png(svg_text: str, output_path: str) -> None:
    try:
        import cairosvg
    except ImportError:
        print(
            "error: PNG output requires cairosvg (pip install cairosvg)",
            file=sys.stderr,
        )
        sys.exit(2)
    cairosvg.svg2png(bytestring=svg_text.encode("utf-8"), write_to=output_path)


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
    had_output_flag = args.output is not None
    env_output = os.getenv("QR_OUTPUT")
    default_output = env_output or "out/qr.svg"
    args.output = args.output or default_output
    if args.catalog and not had_output_flag and env_output is None:
        args.output = "out/catalog.svg"

    if args.png_output is None:
        args.png_output = env_str("QR_PNG_OUTPUT")

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
        png_path = args.png_output or derive_png_path(args.output)
        ensure_parent_dir(png_path)
        write_png(svg, png_path)
        print(f"Saved {png_path}")


if __name__ == "__main__":
    main()
