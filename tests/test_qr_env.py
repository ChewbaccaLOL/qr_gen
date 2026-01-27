import os

from qr_env import env_bool, env_float, env_int, env_str, load_dotenv


def test_load_dotenv_sets_values(tmp_path, monkeypatch):
    env_path = tmp_path / ".env"
    env_path.write_text(
        """
# comment
QR_DATA=hello
QUOTED=\"world\"
EMPTY=
 =should_skip
BADLINE
""".lstrip()
    )
    monkeypatch.delenv("QR_DATA", raising=False)
    monkeypatch.setenv("QUOTED", "keep")

    load_dotenv(str(env_path))

    assert os.getenv("QR_DATA") == "hello"
    assert os.getenv("QUOTED") == "keep"
    assert os.getenv("EMPTY") == ""


def test_load_dotenv_handles_oserror(monkeypatch):
    def fake_open(*_args, **_kwargs):
        raise OSError("boom")

    monkeypatch.setattr("os.path.isfile", lambda _path: True)
    monkeypatch.setattr("builtins.open", fake_open)
    load_dotenv(".env")


def test_env_helpers(monkeypatch):
    monkeypatch.setenv("INT_VAL", "12")
    monkeypatch.setenv("INT_BAD", "nope")
    monkeypatch.setenv("FLOAT_VAL", "2.5")
    monkeypatch.setenv("FLOAT_BAD", "none")
    monkeypatch.setenv("BOOL_TRUE", "yes")
    monkeypatch.setenv("BOOL_FALSE", "off")
    monkeypatch.setenv("BOOL_BAD", "maybe")

    assert env_str("INT_VAL") == "12"
    assert env_int("INT_VAL") == 12
    assert env_int("INT_BAD", 7) == 7
    assert env_float("FLOAT_VAL") == 2.5
    assert env_float("FLOAT_BAD", 1.25) == 1.25
    assert env_bool("BOOL_TRUE") is True
    assert env_bool("BOOL_FALSE") is False
    assert env_bool("BOOL_BAD", True) is True
