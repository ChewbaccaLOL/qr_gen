import argparse
import os
import sys
from dataclasses import dataclass
from typing import Dict, Optional

import segno

from qr_animation import build_float_gif_frames, build_wave_gif_frames
from qr_env import env_bool, env_float, env_int, env_str, load_dotenv
from qr_export import write_pdf, write_png, write_ps
from qr_render import render_catalog_svg, render_svg


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

ANIMATION_VARIANTS = (
    "wave",
    "wave-loop",
    "float",
    "float-tilt-first",
    "float-jagged",
)

DEFAULT_GIF_FPS = 12
DEFAULT_GIF_FRAMES = 40
DEFAULT_GIF_HOLD = 24
DEFAULT_WAVE_AMP = 0.45
DEFAULT_WAVE_PERIOD = 10.0
DEFAULT_FLOAT_JAGGED_SNAP = 0.25
DEFAULT_FLOAT_TILT = 18.0
DEFAULT_FLOAT_ANGLE = 90.0

READABLE_GIF_FPS = 12
READABLE_GIF_FRAMES = 32
READABLE_GIF_HOLD = 32
READABLE_WAVE_AMP = 0.28
READABLE_WAVE_PERIOD = 14.0


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
        choices=ANIMATION_VARIANTS,
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
        choices=ANIMATION_VARIANTS,
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
        help="Number of animation frames (loop segment).",
    )
    parser.add_argument(
        "--gif-hold",
        type=int,
        default=None,
        help="Number of still frames before/after the motion.",
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
        "--float-angle",
        type=float,
        default=None,
        help="Float motion direction in degrees (90 = vertical).",
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




def derive_output_path(svg_path: str, extension: str) -> str:
    extension = extension if extension.startswith(".") else f".{extension}"
    if svg_path.lower().endswith(".svg"):
        return f"{svg_path[:-4]}{extension}"
    return f"{svg_path}{extension}"




def ensure_parent_dir(path: str) -> None:
    directory = os.path.dirname(path)
    if directory:
        os.makedirs(directory, exist_ok=True)


def main() -> None:
    args = parse_args()
    if args.list_variants:
        for name in sorted(VARIANTS.keys()):
            print(name)
        print()
        print("Animations:")
        for name in ANIMATION_VARIANTS:
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
    if args.float_angle is None:
        args.float_angle = env_float("QR_FLOAT_ANGLE")

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
                render_svg=render_svg,
                mode="still",
            )
        elif animation_variant == "wave-loop":
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
                hold=0,
                render_svg=render_svg,
                mode="loop",
            )
        elif animation_variant == "float":
            float_angle = (
                args.float_angle
                if args.float_angle is not None
                else DEFAULT_FLOAT_ANGLE + DEFAULT_FLOAT_TILT
            )
            frames = build_float_gif_frames(
                qr.matrix,
                scale=args.scale,
                border=args.border,
                dark=dark,
                light=light,
                shape=variant.shape,
                radius=radius,
                gradient=gradient,
                float_amp=args.wave_amp,
                float_period=args.wave_period,
                float_angle_deg=float_angle,
                frames=args.gif_frames,
                hold=args.gif_hold,
                render_svg=render_svg,
                mode="still",
                rotate_deg=DEFAULT_FLOAT_TILT,
            )
        elif animation_variant == "float-tilt-first":
            float_angle = args.float_angle if args.float_angle is not None else DEFAULT_FLOAT_ANGLE
            frames = build_float_gif_frames(
                qr.matrix,
                scale=args.scale,
                border=args.border,
                dark=dark,
                light=light,
                shape=variant.shape,
                radius=radius,
                gradient=gradient,
                float_amp=args.wave_amp,
                float_period=args.wave_period,
                float_angle_deg=float_angle,
                frames=args.gif_frames,
                hold=args.gif_hold,
                render_svg=render_svg,
                mode="still",
                rotate_deg=DEFAULT_FLOAT_TILT,
                rotate_mode="before",
            )
        elif animation_variant == "float-jagged":
            float_angle = (
                args.float_angle
                if args.float_angle is not None
                else DEFAULT_FLOAT_ANGLE + DEFAULT_FLOAT_TILT
            )
            frames = build_float_gif_frames(
                qr.matrix,
                scale=args.scale,
                border=args.border,
                dark=dark,
                light=light,
                shape=variant.shape,
                radius=radius,
                gradient=gradient,
                float_amp=args.wave_amp,
                float_period=args.wave_period,
                float_angle_deg=float_angle,
                frames=args.gif_frames,
                hold=args.gif_hold,
                render_svg=render_svg,
                mode="still",
                snap=DEFAULT_FLOAT_JAGGED_SNAP,
                rotate_deg=DEFAULT_FLOAT_TILT,
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
