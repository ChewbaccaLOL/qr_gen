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


def smoothstep(t: float) -> float:
    t = max(0.0, min(1.0, t))
    return t * t * (3 - 2 * t)


def wave_ramp_multiplier(frame_index: int, frames: int, ramp_frames: int) -> float:
    if ramp_frames <= 1 or frames <= 1:
        return 1.0
    ramp_frames = min(ramp_frames, frames // 2)
    if ramp_frames <= 1:
        return 1.0
    if frame_index < ramp_frames:
        t = frame_index / (ramp_frames - 1)
        return smoothstep(t)
    if frame_index >= frames - ramp_frames:
        t = (frames - 1 - frame_index) / (ramp_frames - 1)
        return smoothstep(t)
    return 1.0


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
    mode: str = "still",
) -> List["Image.Image"]:
    Image = require_pillow()
    size = len(matrix)
    amplitude_px = wave_amp * scale
    extra_pad_y = int(math.ceil(abs(amplitude_px))) + 1 if amplitude_px else 0
    if frames <= 0:
        return []
    looped = mode == "loop"
    if mode not in ("still", "loop"):
        raise ValueError(f"Unknown wave mode: {mode}")
    phase_denominator = max(frames - 1, 1) if looped else frames
    phase_step = (2 * math.pi) / phase_denominator
    still_offsets = [0.0 for _ in range(size)]

    images: List["Image.Image"] = []
    total_frames = frames if looped else hold + frames + hold
    ramp_frames = 0
    if not looped:
        ramp_frames = max(2, int(round(frames * 0.2)))
    for frame_index in range(total_frames):
        if looped:
            phase = phase_step * frame_index
            offsets = compute_wave_offsets(size, amplitude_px, wave_period, phase)
        else:
            if frame_index < hold or frame_index >= hold + frames:
                offsets = still_offsets
            else:
                wave_index = frame_index - hold
                phase = phase_step * wave_index
                ramp = wave_ramp_multiplier(wave_index, frames, ramp_frames)
                offsets = compute_wave_offsets(
                    size,
                    amplitude_px * ramp,
                    wave_period,
                    phase,
                )
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
