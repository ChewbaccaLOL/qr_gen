import io
import math
from typing import Callable, Dict, List, Optional, Tuple, Union

from qr_export import require_pillow, svg_to_png_bytes


ColumnOffset = Union[float, Tuple[float, float]]
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
        Optional[List[ColumnOffset]],
        int,
        int,
        float,
        str,
    ],
    str,
]


def compute_wave_offsets(
    size: int, amplitude_px: float, period: float, phase: float
) -> List[float]:
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


def quantize_offset(offset: float, snap_px: float) -> float:
    if snap_px <= 0:
        return offset
    return round(offset / snap_px) * snap_px


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
            column_offsets=offsets,
            extra_pad_y=extra_pad_y,
            extra_pad_x=0,
            rotate_deg=0.0,
            rotate_mode="after",
        )
        png_bytes = svg_to_png_bytes(svg_frame, scale=1.0)
        image = Image.open(io.BytesIO(png_bytes))
        image.load()
        images.append(image)
    return images


def build_float_gif_frames(
    matrix: List[List[int]],
    scale: int,
    border: int,
    dark: str,
    light: Optional[str],
    shape: str,
    radius: float,
    gradient: Optional[Dict[str, str]],
    float_amp: float,
    float_period: float,
    float_angle_deg: float,
    frames: int,
    hold: int,
    render_svg: SvgRenderer,
    mode: str = "still",
    snap: float = 0.0,
    rotate_deg: float = 0.0,
    rotate_mode: str = "after",
) -> List["Image.Image"]:
    Image = require_pillow()
    size = len(matrix)
    amplitude_px = float_amp * scale
    if frames <= 0:
        return []
    looped = mode == "loop"
    if mode not in ("still", "loop"):
        raise ValueError(f"Unknown float mode: {mode}")
    if float_period <= 0:
        raise ValueError("Float period must be greater than 0")
    if rotate_mode not in ("after", "before"):
        raise ValueError(f"Unknown rotate mode: {rotate_mode}")
    phase_denominator = max(frames - 1, 1) if looped else frames
    phase_step = (2 * math.pi) / phase_denominator
    still_offsets = [0.0 for _ in range(size)]
    snap_px = snap * scale
    cloth_amp = amplitude_px * 0.4
    angle_rad = math.radians(float_angle_deg)
    if rotate_mode == "after" and rotate_deg:
        angle_rad = math.radians(float_angle_deg - rotate_deg)
    cos_a = math.cos(angle_rad)
    sin_a = math.sin(angle_rad)
    max_offset = abs(amplitude_px) + abs(cloth_amp)
    max_offset_x = abs(max_offset * cos_a) if max_offset else 0.0
    max_offset_y = abs(max_offset * sin_a) if max_offset else 0.0
    extra_pad_x = int(math.ceil(max_offset_x)) + 1 if max_offset_x else 0
    extra_pad_y = int(math.ceil(max_offset_y)) + 1 if max_offset_y else 0

    images: List["Image.Image"] = []
    total_frames = frames if looped else hold + frames + hold
    ramp_frames = 0
    if not looped:
        ramp_frames = max(2, int(round(frames * 0.2)))
    for frame_index in range(total_frames):
        if looped:
            phase = phase_step * frame_index
            base_offset = amplitude_px * math.sin(phase)
            phase_mod = phase * 0.7
            offsets = [
                base_offset + cloth_amp * math.sin((2 * math.pi * (x / float_period)) + phase_mod)
                for x in range(size)
            ]
        else:
            if frame_index < hold or frame_index >= hold + frames:
                offsets = still_offsets
                svg_frame = render_svg(
                    matrix,
                    scale,
                    border,
                    dark,
                    light,
                    shape,
                    radius,
                    gradient,
                    column_offsets=offsets,
                    extra_pad_y=extra_pad_y,
                    extra_pad_x=extra_pad_x,
                    rotate_deg=rotate_deg,
                    rotate_mode=rotate_mode,
                )
                png_bytes = svg_to_png_bytes(svg_frame, scale=1.0)
                image = Image.open(io.BytesIO(png_bytes))
                image.load()
                images.append(image)
                continue
            float_index = frame_index - hold
            phase = phase_step * float_index
            ramp = wave_ramp_multiplier(float_index, frames, ramp_frames)
            base_offset = amplitude_px * math.sin(phase) * ramp
            phase_mod = phase * 0.7
            offsets = [
                base_offset
                + cloth_amp * ramp * math.sin((2 * math.pi * (x / float_period)) + phase_mod)
                for x in range(size)
            ]
        offsets = [quantize_offset(offset, snap_px) for offset in offsets]
        offsets = [(offset * cos_a, offset * sin_a) for offset in offsets]
        svg_frame = render_svg(
            matrix,
            scale,
            border,
            dark,
            light,
            shape,
            radius,
            gradient,
            column_offsets=offsets,
            extra_pad_y=extra_pad_y,
            extra_pad_x=extra_pad_x,
            rotate_deg=rotate_deg,
            rotate_mode=rotate_mode,
        )
        png_bytes = svg_to_png_bytes(svg_frame, scale=1.0)
        image = Image.open(io.BytesIO(png_bytes))
        image.load()
        images.append(image)
    return images
