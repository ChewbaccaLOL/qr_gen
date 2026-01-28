import json
import os
from dataclasses import dataclass
from typing import Dict, Optional, Tuple


DEFAULT_VARIANTS_PATH = os.path.join(os.path.dirname(__file__), "variants.json")


@dataclass(frozen=True)
class Variant:
    name: str
    shape: str
    dark: str
    light: Optional[str]
    radius: float = 0.0
    gradient: Optional[Dict[str, str]] = None


@dataclass(frozen=True)
class GifDefaults:
    fps: int
    frames: int
    hold: int
    wave_amp: float
    wave_period: float


@dataclass(frozen=True)
class AnimationDefaults:
    gif: GifDefaults
    readable_gif: GifDefaults
    float_jagged_snap: float
    float_tilt: float
    float_angle: float


@dataclass(frozen=True)
class VariantsConfig:
    variants: Dict[str, Variant]
    animation_variants: Tuple[str, ...]
    defaults: AnimationDefaults


def _coerce_variant(raw: Dict[str, object]) -> Variant:
    name = raw.get("name")
    shape = raw.get("shape")
    dark = raw.get("dark")
    if not isinstance(name, str) or not name:
        raise ValueError("Variant name is required")
    if not isinstance(shape, str) or not shape:
        raise ValueError(f"Variant '{name}' is missing a shape")
    if not isinstance(dark, str) or not dark:
        raise ValueError(f"Variant '{name}' is missing a dark color")
    light = raw.get("light")
    radius = raw.get("radius", 0.0)
    gradient = raw.get("gradient")
    if light is not None and not isinstance(light, str):
        raise ValueError(f"Variant '{name}' has an invalid light color")
    if gradient is not None and not isinstance(gradient, dict):
        raise ValueError(f"Variant '{name}' has an invalid gradient")
    return Variant(
        name=name,
        shape=shape,
        dark=dark,
        light=light,
        radius=float(radius),
        gradient=gradient,
    )


def _coerce_gif_defaults(raw: Dict[str, object], label: str) -> GifDefaults:
    for key in ("gif_fps", "gif_frames", "gif_hold", "wave_amp", "wave_period"):
        if key not in raw:
            raise ValueError(f"Defaults '{label}' missing '{key}'")
    return GifDefaults(
        fps=int(raw["gif_fps"]),
        frames=int(raw["gif_frames"]),
        hold=int(raw["gif_hold"]),
        wave_amp=float(raw["wave_amp"]),
        wave_period=float(raw["wave_period"]),
    )


def load_variants_config(path: Optional[str] = None) -> VariantsConfig:
    config_path = path or DEFAULT_VARIANTS_PATH
    with open(config_path, "r", encoding="utf-8") as handle:
        raw = json.load(handle)

    raw_variants = raw.get("variants")
    if not isinstance(raw_variants, list) or not raw_variants:
        raise ValueError("Variants config must include a non-empty 'variants' list")

    variants: Dict[str, Variant] = {}
    for item in raw_variants:
        if not isinstance(item, dict):
            raise ValueError("Each variant entry must be an object")
        variant = _coerce_variant(item)
        if variant.name in variants:
            raise ValueError(f"Duplicate variant '{variant.name}'")
        variants[variant.name] = variant

    raw_animation_variants = raw.get("animation_variants")
    if not isinstance(raw_animation_variants, list) or not raw_animation_variants:
        raise ValueError("Variants config must include 'animation_variants'")
    animation_variants = tuple(str(item) for item in raw_animation_variants)

    defaults_raw = raw.get("defaults")
    if not isinstance(defaults_raw, dict):
        raise ValueError("Variants config must include 'defaults'")

    readable_raw = defaults_raw.get("readable_gif")
    if not isinstance(readable_raw, dict):
        raise ValueError("Variants config must include 'defaults.readable_gif'")

    for key in ("float_jagged_snap", "float_tilt", "float_angle"):
        if key not in defaults_raw:
            raise ValueError(f"Defaults missing '{key}'")

    defaults = AnimationDefaults(
        gif=_coerce_gif_defaults(defaults_raw, "defaults"),
        readable_gif=_coerce_gif_defaults(readable_raw, "readable_gif"),
        float_jagged_snap=float(defaults_raw["float_jagged_snap"]),
        float_tilt=float(defaults_raw["float_tilt"]),
        float_angle=float(defaults_raw["float_angle"]),
    )

    return VariantsConfig(
        variants=variants,
        animation_variants=animation_variants,
        defaults=defaults,
    )
