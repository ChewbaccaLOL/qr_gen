import pytest

import qr_animation


MATRIX = [
    [1, 0, 1],
    [0, 1, 0],
    [1, 0, 1],
]


def fake_render_svg(*args, **kwargs):
    return "<svg></svg>"


class DummyImage:
    def load(self):
        return None


class DummyImageModule:
    @staticmethod
    def open(_buffer):
        return DummyImage()


def setup_pillow(monkeypatch):
    monkeypatch.setattr(qr_animation, "require_pillow", lambda: DummyImageModule())
    monkeypatch.setattr(qr_animation, "svg_to_png_bytes", lambda *args, **kwargs: b"png")


def test_wave_offsets_and_helpers():
    zeros = qr_animation.compute_wave_offsets(3, 0.0, 10.0, 0.0)
    assert zeros == [0.0, 0.0, 0.0]
    vals = qr_animation.compute_wave_offsets(2, 2.0, 4.0, 0.0)
    assert len(vals) == 2
    assert qr_animation.smoothstep(-1.0) == 0.0
    assert qr_animation.smoothstep(2.0) == 1.0
    assert qr_animation.wave_ramp_multiplier(0, 1, 2) == 1.0
    assert qr_animation.wave_ramp_multiplier(0, 10, 4) < 1.0
    assert qr_animation.wave_ramp_multiplier(9, 10, 4) < 1.0
    assert qr_animation.quantize_offset(2.4, 0.0) == 2.4
    assert qr_animation.quantize_offset(2.4, 1.0) == 2.0


def test_build_wave_gif_frames(monkeypatch):
    setup_pillow(monkeypatch)
    frames = qr_animation.build_wave_gif_frames(
        MATRIX,
        scale=10,
        border=2,
        dark="#000",
        light="#fff",
        shape="square",
        radius=0.0,
        gradient=None,
        wave_amp=0.3,
        wave_period=8.0,
        frames=3,
        hold=2,
        render_svg=fake_render_svg,
        mode="still",
    )
    assert len(frames) == 7

    with pytest.raises(ValueError):
        qr_animation.build_wave_gif_frames(
            MATRIX,
            scale=10,
            border=2,
            dark="#000",
            light="#fff",
            shape="square",
            radius=0.0,
            gradient=None,
            wave_amp=0.3,
            wave_period=8.0,
            frames=3,
            hold=2,
            render_svg=fake_render_svg,
            mode="bad",
        )


def test_build_wave_gif_frames_loop(monkeypatch):
    setup_pillow(monkeypatch)
    frames = qr_animation.build_wave_gif_frames(
        MATRIX,
        scale=8,
        border=1,
        dark="#000",
        light=None,
        shape="square",
        radius=0.0,
        gradient=None,
        wave_amp=0.4,
        wave_period=6.0,
        frames=4,
        hold=1,
        render_svg=fake_render_svg,
        mode="loop",
    )
    assert len(frames) == 4
    assert qr_animation.build_wave_gif_frames(
        MATRIX,
        scale=8,
        border=1,
        dark="#000",
        light=None,
        shape="square",
        radius=0.0,
        gradient=None,
        wave_amp=0.4,
        wave_period=6.0,
        frames=0,
        hold=1,
        render_svg=fake_render_svg,
        mode="still",
    ) == []


def test_build_float_gif_frames(monkeypatch):
    setup_pillow(monkeypatch)
    frames = qr_animation.build_float_gif_frames(
        MATRIX,
        scale=10,
        border=2,
        dark="#000",
        light="#fff",
        shape="square",
        radius=0.0,
        gradient=None,
        float_amp=0.3,
        float_period=8.0,
        float_angle_deg=90.0,
        frames=3,
        hold=1,
        render_svg=fake_render_svg,
        mode="still",
        snap=0.1,
        rotate_deg=0.0,
        rotate_mode="after",
    )
    assert len(frames) == 5

    with pytest.raises(ValueError):
        qr_animation.build_float_gif_frames(
            MATRIX,
            scale=10,
            border=2,
            dark="#000",
            light="#fff",
            shape="square",
            radius=0.0,
            gradient=None,
            float_amp=0.3,
            float_period=0.0,
            float_angle_deg=90.0,
            frames=3,
            hold=1,
            render_svg=fake_render_svg,
            mode="still",
        )

    with pytest.raises(ValueError):
        qr_animation.build_float_gif_frames(
            MATRIX,
            scale=10,
            border=2,
            dark="#000",
            light="#fff",
            shape="square",
            radius=0.0,
            gradient=None,
            float_amp=0.3,
            float_period=8.0,
            float_angle_deg=90.0,
            frames=3,
            hold=1,
            render_svg=fake_render_svg,
            mode="bad",
        )

    with pytest.raises(ValueError):
        qr_animation.build_float_gif_frames(
            MATRIX,
            scale=10,
            border=2,
            dark="#000",
            light="#fff",
            shape="square",
            radius=0.0,
            gradient=None,
            float_amp=0.3,
            float_period=8.0,
            float_angle_deg=90.0,
            frames=3,
            hold=1,
            render_svg=fake_render_svg,
            mode="still",
            rotate_mode="bad",
        )


def test_build_float_gif_frames_loop(monkeypatch):
    setup_pillow(monkeypatch)
    frames = qr_animation.build_float_gif_frames(
        MATRIX,
        scale=10,
        border=2,
        dark="#000",
        light="#fff",
        shape="square",
        radius=0.0,
        gradient=None,
        float_amp=0.3,
        float_period=8.0,
        float_angle_deg=90.0,
        frames=4,
        hold=1,
        render_svg=fake_render_svg,
        mode="loop",
        snap=0.1,
        rotate_deg=12.0,
        rotate_mode="after",
    )
    assert len(frames) == 4
    assert qr_animation.build_float_gif_frames(
        MATRIX,
        scale=10,
        border=2,
        dark="#000",
        light="#fff",
        shape="square",
        radius=0.0,
        gradient=None,
        float_amp=0.3,
        float_period=8.0,
        float_angle_deg=90.0,
        frames=0,
        hold=1,
        render_svg=fake_render_svg,
    ) == []
