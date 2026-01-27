import io
import math
from typing import Callable, Dict, List, Optional

from qr_export import require_pillow, svg_to_png_bytes


SvgRenderer = Callable[
    [
        List[List[int]],
        int,
        int,
        str,
        Optional[str],
        str,
        float,
        Optional[Dict[str, str]],
        Optional[List[float]],
        int,
    ],
    str,
]


def compute_wave_offsets(size: int, amplitude_px: float, period: float, phase: float) -> List[float]:
    if amplitude_px == 0:
        return [0.0 for _ in range(size)]
    return [
        amplitude_px * math.sin((2 * math.pi * (x / period)) + phase)
        for x in range(size)
    ]


def build_wave_gif_frames(
    matrix: List[List[int]],
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
    render_svg: SvgRenderer,
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
            scale,
            border,
            dark,
            light,
            shape,
            radius,
            gradient,
            offsets,
            extra_pad_y,
        )
        png_bytes = svg_to_png_bytes(svg_frame, scale=1.0)
        image = Image.open(io.BytesIO(png_bytes))
        image.load()
        images.append(image)
    return images
