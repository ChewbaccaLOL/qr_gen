from dataclasses import dataclass
from typing import Dict, Optional

import pytest

from qr_render import (
    column_offset_value,
    escape_xml,
    module_element,
    module_fill,
    render_modules_for_column,
    render_catalog_svg,
    render_svg,
    svg_gradient_def,
)


MATRIX = [
    [1, 0],
    [0, 1],
]


@dataclass
class DummyVariant:
    name: str
    shape: str
    dark: str
    light: Optional[str]
    radius: float = 0.0
    gradient: Optional[Dict[str, str]] = None


def test_svg_gradient_def_and_fill():
    assert svg_gradient_def(None, None) == ""
    gradient = {"id": "fg", "from": "#111", "to": "#222"}
    result = svg_gradient_def(gradient, "fg")
    assert "linearGradient" in result
    assert "#111" in result
    assert "#222" in result
    assert module_fill("#000", None) == "#000"
    assert module_fill("#000", "fg") == "url(#fg)"


def test_escape_xml():
    assert escape_xml("a&b<c>\"") == "a&amp;b&lt;c&gt;&quot;"


def test_column_offset_value():
    assert column_offset_value(None, 0) == (0.0, 0.0)
    assert column_offset_value([], 0) == (0.0, 0.0)
    assert column_offset_value([None], 0) == (0.0, 0.0)
    assert column_offset_value([()], 0) == (0.0, 0.0)
    assert column_offset_value([1.25], 0) == (0.0, 1.25)
    assert column_offset_value([(1.0, 2.0)], 0) == (1.0, 2.0)
    assert column_offset_value([(3.5,)], 0) == (0.0, 3.5)
    assert column_offset_value([1.0], 3) == (0.0, 0.0)


def test_module_element_shapes():
    rect = module_element("square", 0, 0, 10, 0.0, "#000")
    assert rect.startswith("<rect")
    rounded = module_element("rounded", 0, 0, 10, 0.2, "#000")
    assert "rx" in rounded and "ry" in rounded
    dot = module_element("dot", 0, 0, 10, 0.0, "#000")
    assert dot.startswith("<circle")
    with pytest.raises(ValueError):
        module_element("triangle", 0, 0, 10, 0.0, "#000")


def test_render_svg_basic():
    svg = render_svg(
        MATRIX,
        scale=10,
        border=1,
        dark="#000",
        light="#fff",
        shape="square",
        radius=0.0,
        gradient=None,
    )
    assert svg.startswith("<?xml")
    assert "shape-rendering=\"crispEdges\"" in svg
    assert "<rect width=\"100%\"" in svg


def test_render_svg_rotation_and_gradient():
    svg = render_svg(
        MATRIX,
        scale=8,
        border=1,
        dark="#000",
        light=None,
        shape="rounded",
        radius=0.2,
        gradient={"id": "grad", "from": "#111", "to": "#222"},
        rotate_deg=12,
        rotate_mode="after",
    )
    assert "<defs>" in svg
    assert "url(#grad)" in svg
    assert "rotate(12" in svg

    with pytest.raises(ValueError):
        render_svg(
            MATRIX,
            scale=8,
            border=1,
            dark="#000",
            light=None,
            shape="rounded",
            radius=0.2,
            gradient=None,
            rotate_mode="nope",
        )


def test_render_modules_for_column():
    parts = list(
        render_modules_for_column(
            MATRIX,
            scale=5,
            border=1,
            fill="#000",
            shape="square",
            radius=0.0,
            column_index=0,
        )
    )
    assert len(parts) == 1
    assert parts[0].startswith("<rect")


def test_render_svg_rotate_before_with_offsets():
    svg = render_svg(
        MATRIX,
        scale=10,
        border=1,
        dark="#000",
        light="#fff",
        shape="rounded",
        radius=0.1,
        gradient=None,
        column_offsets=[(1.0, 2.0), (0.0, 0.0)],
        rotate_deg=15,
        rotate_mode="before",
    )
    assert "translate(1.0 2.0)" in svg
    assert "rotate(15" in svg


def test_render_catalog_svg():
    variants = [
        DummyVariant("classic", "square", "#000", "#fff"),
        DummyVariant(
            "sunset",
            "rounded",
            "#111",
            "#eee",
            radius=0.2,
            gradient={"from": "#111", "to": "#222"},
        ),
    ]
    svg = render_catalog_svg(
        MATRIX,
        scale=6,
        border=1,
        variants=variants,
        columns=0,
        background="#fff",
        label_size=0,
    )
    assert "<text" in svg
    assert "classic" in svg
    assert "sunset" in svg
    assert "linearGradient" in svg

    with pytest.raises(ValueError):
        render_catalog_svg(
            MATRIX,
            scale=6,
            border=1,
            variants=[],
            columns=1,
            background="#fff",
            label_size=10,
        )
