from system_monitor.release_updates import is_newer_version, normalize_version, version_key


def test_normalize_version_strips_v_prefix() -> None:
    assert normalize_version("v1.2.3") == "1.2.3"


def test_is_newer_version() -> None:
    assert is_newer_version("1.0.1", "1.0.0")
    assert not is_newer_version("1.0.0", "1.0.1")
    assert not is_newer_version("1.0.0", "1.0.0")


def test_version_key_handles_short_versions() -> None:
    assert version_key("1.2") < version_key("1.2.1")
