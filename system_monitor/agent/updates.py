from __future__ import annotations

import sys
import webbrowser
from dataclasses import dataclass
from pathlib import Path

from .. import __version__
from ..paths import AGENT_UPDATE_CACHE
from ..release_updates import (
    RELEASES_LATEST_PAGE,
    ReleaseCheckResult,
    check_release_updates,
    normalize_version,
)

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


def _install_hint(latest_version: str, download_url: str | None) -> str:
    if sys.platform == "win32":
        if download_url:
            return f"Скачайте и запустите установщик:\n{download_url}"
        return f"Скачайте установщик с GitHub Releases:\n{RELEASES_LATEST_PAGE}"

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
    return f"Скачайте {deb_name} с GitHub Releases:\n{RELEASES_LATEST_PAGE}"


def _to_agent_result(release: ReleaseCheckResult) -> UpdateCheckResult:
    return UpdateCheckResult(
        current_version=release.current_version,
        latest_version=release.latest_version,
        update_available=release.update_available,
        release_url=release.release_url,
        download_url=release.download_url,
        install_hint=_install_hint(release.latest_version, release.download_url),
        checked_at=release.checked_at,
        error=release.error,
    )


def check_for_updates(force: bool = False) -> UpdateCheckResult:
    release = check_release_updates(
        get_installed_version(),
        cache_path=AGENT_UPDATE_CACHE,
        force=force,
        user_agent="system-monitor-agent",
        check_interval_sec=UPDATE_CHECK_INTERVAL_SEC,
        download_asset_name=_asset_name,
    )
    return _to_agent_result(release)


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
    url = result.download_url or result.release_url or RELEASES_LATEST_PAGE
    webbrowser.open(url)
