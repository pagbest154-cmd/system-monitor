from __future__ import annotations

import json
import re
import sys
import time
import webbrowser
from dataclasses import asdict, dataclass
from pathlib import Path

import httpx

from .. import __version__
from ..paths import AGENT_UPDATE_CACHE, CONFIG_DIR

GITHUB_REPO = "pagbest154-cmd/system-monitor"
RELEASES_LATEST_PAGE = f"https://github.com/{GITHUB_REPO}/releases/latest"
RELEASES_PAGE_URL = RELEASES_LATEST_PAGE
APT_SOURCE_FILE = Path("/etc/apt/sources.list.d/system-monitor.list")
UPDATE_CHECK_INTERVAL_SEC = 24 * 60 * 60


@dataclass
class UpdateCheckResult:
    current_version: str
    latest_version: str
    update_available: bool
    release_url: str
    download_url: str | None
    install_hint: str
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


def get_installed_version() -> str:
    if sys.platform.startswith("linux"):
        try:
            import subprocess

            result = subprocess.run(
                ["dpkg-query", "-W", "-f=${Version}", "system-monitor-agent"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if result.returncode == 0:
                installed = result.stdout.strip()
                if installed and installed != "none":
                    return normalize_version(installed)
        except (FileNotFoundError, OSError, subprocess.SubprocessError):
            pass
    return normalize_version(__version__)


def _apt_repo_configured() -> bool:
    return APT_SOURCE_FILE.exists()


def _asset_name(latest_version: str) -> str:
    if sys.platform == "win32":
        return f"system-monitor-agent_{latest_version}_setup.exe"
    return f"system-monitor-agent_{latest_version}-1_amd64.deb"


def _release_download_url(tag: str, latest_version: str) -> str:
    tag_name = tag if tag.startswith("v") else f"v{latest_version}"
    asset_name = _asset_name(latest_version)
    return f"https://github.com/{GITHUB_REPO}/releases/download/{tag_name}/{asset_name}"


def _install_hint(latest_version: str, download_url: str | None) -> str:
    if sys.platform == "win32":
        if download_url:
            return f"Скачайте и запустите установщик:\n{download_url}"
        return f"Скачайте установщик с GitHub Releases:\n{RELEASES_PAGE_URL}"

    if _apt_repo_configured():
        return (
            "Обновление через APT-репозиторий:\n"
            "  sudo apt update\n"
            "  sudo apt install --only-upgrade system-monitor-agent"
        )

    deb_name = _asset_name(latest_version)
    if download_url:
        return (
            f"Скачайте пакет и установите:\n"
            f"  wget {download_url}\n"
            f"  sudo apt install ./{deb_name}"
        )
    return f"Скачайте {deb_name} с GitHub Releases:\n{RELEASES_PAGE_URL}"


def _read_cache() -> UpdateCheckResult | None:
    if not AGENT_UPDATE_CACHE.exists():
        return None
    try:
        data = json.loads(AGENT_UPDATE_CACHE.read_text(encoding="utf-8"))
        return UpdateCheckResult(**data)
    except (json.JSONDecodeError, TypeError, ValueError):
        return None


def _write_cache(result: UpdateCheckResult) -> None:
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    AGENT_UPDATE_CACHE.write_text(
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


def _fetch_latest_release() -> tuple[str, str, str]:
    headers = {"User-Agent": "system-monitor-agent"}
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
    release_url = final_url
    download_url = _release_download_url(tag, latest_version)
    return latest_version, release_url, download_url


def check_for_updates(force: bool = False) -> UpdateCheckResult:
    current_version = get_installed_version()
    now = time.time()
    cached = _read_cache()

    if not force:
        if cached is not None and (now - cached.checked_at) < UPDATE_CHECK_INTERVAL_SEC:
            cached.current_version = current_version
            cached.update_available = is_newer_version(cached.latest_version, current_version)
            cached.error = None
            return cached

    try:
        latest_version, release_url, download_url = _fetch_latest_release()
        result = UpdateCheckResult(
            current_version=current_version,
            latest_version=latest_version,
            update_available=is_newer_version(latest_version, current_version),
            release_url=release_url,
            download_url=download_url,
            install_hint=_install_hint(latest_version, download_url),
            checked_at=now,
        )
        _write_cache(result)
        return result
    except Exception as exc:
        if cached is not None and not force:
            cached.current_version = current_version
            cached.update_available = is_newer_version(cached.latest_version, current_version)
            cached.error = None
            return cached
        return UpdateCheckResult(
            current_version=current_version,
            latest_version=current_version,
            update_available=False,
            release_url=RELEASES_PAGE_URL,
            download_url=None,
            install_hint="",
            checked_at=now,
            error=str(exc),
        )


def format_update_message(result: UpdateCheckResult) -> str:
    if result.error:
        if "rate limit" in result.error.lower():
            return (
                "Слишком много запросов к GitHub.\n"
                "Попробуйте позже или откройте страницу релизов вручную."
            )
        return f"Не удалось проверить обновления: {result.error}"
    if result.update_available:
        return (
            f"Доступна новая версия {result.latest_version} "
            f"(установлена {result.current_version}).\n\n"
            f"{result.install_hint}"
        )
    return f"Установлена актуальная версия {result.current_version}."


def open_update_page(result: UpdateCheckResult) -> None:
    url = result.download_url or result.release_url or RELEASES_PAGE_URL
    webbrowser.open(url)
