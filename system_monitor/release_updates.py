from __future__ import annotations

import json
import re
import time
from collections.abc import Callable
from dataclasses import asdict, dataclass
from pathlib import Path

import httpx

GITHUB_REPO = "pagbest154-cmd/system-monitor"
RELEASES_LATEST_PAGE = f"https://github.com/{GITHUB_REPO}/releases/latest"
DEFAULT_CHECK_INTERVAL_SEC = 24 * 60 * 60


@dataclass
class ReleaseCheckResult:
    current_version: str
    latest_version: str
    update_available: bool
    release_url: str
    download_url: str | None
    checked_at: float
    error: str | None = None


def normalize_version(version: str) -> str:
    value = version.strip().lstrip("v")
    if "-" in value:
        value = value.split("-", 1)[0]
    return value


def version_key(version: str) -> tuple[int, ...]:
    normalized = normalize_version(version)
    parts: list[int] = []
    for piece in normalized.split("."):
        try:
            parts.append(int(piece))
        except ValueError:
            parts.append(0)
    return tuple(parts)


def is_newer_version(latest: str, current: str) -> bool:
    return version_key(latest) > version_key(current)


def release_download_url(tag: str, latest_version: str, asset_name: str) -> str:
    tag_name = tag if tag.startswith("v") else f"v{latest_version}"
    return f"https://github.com/{GITHUB_REPO}/releases/download/{tag_name}/{asset_name}"


def _read_cache(cache_path: Path) -> ReleaseCheckResult | None:
    if not cache_path.exists():
        return None
    try:
        data = json.loads(cache_path.read_text(encoding="utf-8"))
        return ReleaseCheckResult(**data)
    except (json.JSONDecodeError, TypeError, ValueError):
        return None


def _write_cache(cache_path: Path, result: ReleaseCheckResult) -> None:
    cache_path.parent.mkdir(parents=True, exist_ok=True)
    cache_path.write_text(
        json.dumps(asdict(result), ensure_ascii=False, indent=2),
        encoding="utf-8",
    )


def _parse_release_tag(final_url: str, body: str) -> str | None:
    match = re.search(r"/releases/tag/(v?[^/?#]+)", final_url)
    if match:
        return match.group(1)
    match = re.search(r'href="[^"]*/releases/tag/(v?[^"/?#]+)"', body)
    if match:
        return match.group(1)
    return None


def fetch_latest_release(user_agent: str = "system-monitor") -> tuple[str, str, str]:
    headers = {"User-Agent": user_agent}
    with httpx.Client(timeout=15.0, follow_redirects=True) as client:
        response = client.get(RELEASES_LATEST_PAGE, headers=headers)
        response.raise_for_status()

    final_url = str(response.url)
    tag = _parse_release_tag(final_url, response.text)
    if not tag:
        raise ValueError("Не удалось определить версию из GitHub Releases")
    latest_version = normalize_version(tag)
    if not latest_version:
        raise ValueError("В релизе не указана версия")
    return latest_version, final_url, tag


def check_release_updates(
    current_version: str,
    *,
    cache_path: Path | None = None,
    force: bool = False,
    user_agent: str = "system-monitor",
    check_interval_sec: int = DEFAULT_CHECK_INTERVAL_SEC,
    download_asset_name: Callable[[str], str] | None = None,
) -> ReleaseCheckResult:
    current = normalize_version(current_version)
    now = time.time()
    cached = _read_cache(cache_path) if cache_path is not None else None

    if not force and cached is not None and (now - cached.checked_at) < check_interval_sec:
        cached.current_version = current
        cached.update_available = is_newer_version(cached.latest_version, current)
        cached.error = None
        return cached

    try:
        latest_version, release_url, tag = fetch_latest_release(user_agent=user_agent)
        download_url = (
            release_download_url(tag, latest_version, download_asset_name(latest_version))
            if download_asset_name is not None
            else None
        )
        result = ReleaseCheckResult(
            current_version=current,
            latest_version=latest_version,
            update_available=is_newer_version(latest_version, current),
            release_url=release_url,
            download_url=download_url,
            checked_at=now,
        )
        if cache_path is not None:
            _write_cache(cache_path, result)
        return result
    except Exception as exc:
        if cached is not None and not force:
            cached.current_version = current
            cached.update_available = is_newer_version(cached.latest_version, current)
            cached.error = None
            return cached
        return ReleaseCheckResult(
            current_version=current,
            latest_version=current,
            update_available=False,
            release_url=RELEASES_LATEST_PAGE,
            download_url=None,
            checked_at=now,
            error=str(exc),
        )
