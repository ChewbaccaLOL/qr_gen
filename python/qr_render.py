import math
from typing import Dict, Iterable, List, Optional, Tuple, Union


ColumnOffset = Union[float, Tuple[float, float]]


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


def column_offset_value(
    column_offsets: Optional[List[ColumnOffset]], index: int
) -> Tuple[float, float]:
    if not column_offsets:
        return (0.0, 0.0)
    if index >= len(column_offsets):
        return (0.0, 0.0)
    value = column_offsets[index]
    if value is None:
        return (0.0, 0.0)
    if isinstance(value, (tuple, list)):
        if len(value) >= 2:
            return (float(value[0]), float(value[1]))
        if len(value) == 1:
            return (0.0, float(value[0]))
        return (0.0, 0.0)
    return (0.0, float(value))


def module_element(
    shape: str, px: float, py: float, scale: int, radius: float, fill: str
) -> str:
    if shape == "square":
        return f'<rect x="{px}" y="{py}" width="{scale}" height="{scale}" fill="{fill}"/>'
    if shape == "rounded":
        corner = max(0.0, min(radius, 0.5)) * scale
        return (
            f'<rect x="{px}" y="{py}" width="{scale}" height="{scale}" '
            f'rx="{corner}" ry="{corner}" fill="{fill}"/>'
        )
    if shape == "dot":
        r = scale * 0.45
        offset_center = scale / 2
        cx = px + offset_center
        cy = py + offset_center
        return f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{fill}"/>'
    raise ValueError(f"Unknown shape: {shape}")


def render_modules(
    matrix: List[List[int]],
    scale: int,
    border: int,
    fill: str,
    shape: str,
    radius: float,
    offset: Tuple[int, int] = (0, 0),
    column_offsets: Optional[List[ColumnOffset]] = None,
) -> Iterable[str]:
    offset_x, offset_y = offset
    for y, row in enumerate(matrix):
        for x, cell in enumerate(row):
            if cell:
                column_offset_x, column_offset_y = column_offset_value(column_offsets, x)
                px = (x + border) * scale + offset_x + column_offset_x
                py = (y + border) * scale + offset_y + column_offset_y
                yield module_element(shape, px, py, scale, radius, fill)


def render_modules_for_column(
    matrix: List[List[int]],
    scale: int,
    border: int,
    fill: str,
    shape: str,
    radius: float,
    column_index: int,
    offset: Tuple[int, int] = (0, 0),
) -> Iterable[str]:
    offset_x, offset_y = offset
    for y, row in enumerate(matrix):
        if row[column_index]:
            px = (column_index + border) * scale + offset_x
            py = (y + border) * scale + offset_y
            yield module_element(shape, px, py, scale, radius, fill)


def render_svg(
    matrix: List[List[int]],
    scale: int,
    border: int,
    dark: str,
    light: Optional[str],
    shape: str,
    radius: float,
    gradient: Optional[Dict[str, str]],
    column_offsets: Optional[List[ColumnOffset]] = None,
    extra_pad_y: int = 0,
    extra_pad_x: int = 0,
    rotate_deg: float = 0.0,
    rotate_mode: str = "after",
    rotate_modules: bool = True,
) -> str:
    size = len(matrix)
    total_modules = size + border * 2
    dimension = total_modules * scale
    content_width = dimension + extra_pad_x * 2
    content_height = dimension + extra_pad_y * 2
    width = content_width
    height = content_height
    shape_rendering = "crispEdges" if shape == "square" else "geometricPrecision"
    gradient_id = gradient.get("id", "fg") if gradient else None
    fill = module_fill(dark, gradient_id)
    if rotate_deg:
        angle = math.radians(rotate_deg)
        cos_a = abs(math.cos(angle))
        sin_a = abs(math.sin(angle))
        width = content_width * cos_a + content_height * sin_a
        height = content_width * sin_a + content_height * cos_a

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

    translate_x = (width - content_width) / 2
    translate_y = (height - content_height) / 2
    parts.append(f'<g transform="translate({translate_x} {translate_y})">')
    if rotate_mode not in ("after", "before"):
        raise ValueError(f"Unknown rotate mode: {rotate_mode}")

    center_x = content_width / 2
    center_y = content_height / 2
    base_offset = (extra_pad_x, extra_pad_y)

    if rotate_deg and not rotate_modules:
        angle = math.radians(rotate_deg)
        cos_a = math.cos(angle)
        sin_a = math.sin(angle)
        base_offset = (extra_pad_x, extra_pad_y)
        for y, row in enumerate(matrix):
            for x, cell in enumerate(row):
                if not cell:
                    continue
                column_offset_x, column_offset_y = column_offset_value(column_offsets, x)
                px = (x + border) * scale + base_offset[0]
                py = (y + border) * scale + base_offset[1]
                if rotate_mode == "before":
                    px += column_offset_x
                    py += column_offset_y
                dx = px - center_x
                dy = py - center_y
                rx = dx * cos_a - dy * sin_a
                ry = dx * sin_a + dy * cos_a
                px = rx + center_x
                py = ry + center_y
                if rotate_mode == "after":
                    px += column_offset_x
                    py += column_offset_y
                parts.append(module_element(shape, px, py, scale, radius, fill))
    elif rotate_mode == "before" and column_offsets:
        for column_index in range(size):
            column_offset_x, column_offset_y = column_offset_value(
                column_offsets, column_index
            )
            if column_offset_x or column_offset_y:
                parts.append(
                    f'<g transform="translate({column_offset_x} {column_offset_y})">'
                )
            if rotate_deg:
                parts.append(f'<g transform="rotate({rotate_deg} {center_x} {center_y})">')
            parts.extend(
                render_modules_for_column(
                    matrix,
                    scale=scale,
                    border=border,
                    fill=fill,
                    shape=shape,
                    radius=radius,
                    column_index=column_index,
                    offset=base_offset,
                )
            )
            if rotate_deg:
                parts.append("</g>")
            if column_offset_x or column_offset_y:
                parts.append("</g>")
    else:
        if rotate_deg:
            parts.append(f'<g transform="rotate({rotate_deg} {center_x} {center_y})">')
        parts.extend(
            render_modules(
                matrix,
                scale=scale,
                border=border,
                fill=fill,
                shape=shape,
                radius=radius,
                offset=base_offset,
                column_offsets=column_offsets,
            )
        )
        if rotate_deg:
            parts.append("</g>")
    parts.append("</g>")

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
