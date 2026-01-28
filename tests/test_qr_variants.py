import json

import pytest

import qr_variants


def test_load_variants_config_default_path():
    config = qr_variants.load_variants_config()
    assert "classic" in config.variants
    assert "wave" in config.animation_variants
    assert config.defaults.gif.fps > 0
    assert config.defaults.readable_gif.wave_period > 0


def test_load_variants_config_missing_fields(tmp_path):
    bad_path = tmp_path / "variants.json"
    bad_path.write_text(json.dumps({"variants": []}), encoding="utf-8")
    with pytest.raises(ValueError):
        qr_variants.load_variants_config(str(bad_path))
