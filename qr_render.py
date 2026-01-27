import math
from typing import Dict, Iterable, List, Optional, Tuple


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
    matrix: List[List[int]],
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
    matrix: List[List[int]],
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
        parts.append(f'<rect width="100%" height="100%" fill="{light}"/>')

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
    matrix: List[List[int]],
    scale: int,
    border: int,
    variants: Iterable,
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
