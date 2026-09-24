from __future__ import annotations

import os
import subprocess
import sys
import threading
import time
from datetime import datetime
from pathlib import Path

import pystray
from PIL import Image

from ..paths import AGENT_CONFIG, CONFIG_DIR
from .branding import render_icon
from .service_control import get_service_state, restart_service
from .status import AgentStatus, read_agent_status
from .updates import check_for_updates, format_update_message, get_installed_version, open_update_page


def _make_icon(color_key: str) -> Image.Image:
    return render_icon(64, color_key)


def _format_ts(ts: float | None) -> str:
    if ts is None:
        return "—"
    return datetime.fromtimestamp(ts).strftime("%d.%m.%Y %H:%M:%S")


def _status_icon_key(status: AgentStatus | None, service_state: str) -> str:
    if service_state != "running":
        return "error"
    if status is None:
        return "idle"
    if status.connected and not status.is_stale():
        return "ok"
    if status.last_error:
        return "error"
    return "idle"


def _status_tooltip(status: AgentStatus | None, service_state: str) -> str:
    lines = ["system-monitor agent"]
    if service_state == "missing":
        lines.append("Служба не установлена")
        return "\n".join(lines)
    if service_state == "stopped":
        lines.append("Служба остановлена")
        return "\n".join(lines)

    if status is None:
        lines.append("Ожидание данных от службы")
        return "\n".join(lines)

    lines.append(f"Статус: {status.status_label()}")
    if status.hub_url:
        lines.append(f"Hub: {status.hub_url}")
    if status.agent_id:
        lines.append(f"Agent ID: {status.agent_id}")
    if status.last_success_ts:
        lines.append(f"Последняя отправка: {_format_ts(status.last_success_ts)}")
    if status.last_error:
        lines.append(f"Ошибка: {status.last_error}")
    if status.update_available and status.latest_version:
        lines.append(f"Доступно обновление: {status.latest_version}")
    return "\n".join(lines)


class TrayApp:
    def __init__(self, config_path: Path | None = None) -> None:
        self.config_path = config_path
        self._stop = threading.Event()
        self._update_available = False
        self._latest_version = ""
        self._icon = pystray.Icon(
            "system-monitor-agent",
            _make_icon("idle"),
            "system-monitor agent",
            menu=self._build_menu(),
        )

    def _build_menu(self) -> pystray.Menu:
        return pystray.Menu(
            pystray.MenuItem(lambda icon, item: self._menu_status_text(), None, enabled=False),
            pystray.Menu.SEPARATOR,
            pystray.MenuItem("Настройки", self._open_settings),
            pystray.MenuItem("Перезапустить службу", self._restart_service),
            pystray.MenuItem("Открыть лог", self._open_log),
            pystray.MenuItem("Открыть папку конфигурации", self._open_config_dir),
            pystray.Menu.SEPARATOR,
            pystray.MenuItem(lambda icon, item: self._menu_version_text(), None, enabled=False),
            pystray.MenuItem("Проверить обновления", self._check_updates),
            pystray.MenuItem(
                lambda icon, item: self._menu_update_text(),
                self._open_update,
                visible=lambda item: self._update_available,
            ),
            pystray.Menu.SEPARATOR,
            pystray.MenuItem("Выход", self._quit),
        )

    def _menu_status_text(self) -> str:
        status = read_agent_status()
        service_state = get_service_state()
        if service_state != "running":
            return "Служба: не запущена"
        if status is None:
            return "Статус: ожидание"
        if status.connected and not status.is_stale():
            return f"Статус: подключён ({status.agent_id})"
        if status.last_error:
            return f"Статус: ошибка"
        return "Статус: ожидание"

    def _open_settings(self, icon: pystray.Icon, item: pystray.MenuItem) -> None:
        _launch_settings_window(self.config_path)

    def _restart_service(self, icon: pystray.Icon, item: pystray.MenuItem) -> None:
        ok, detail = restart_service()
        icon.title = detail if ok else f"Ошибка: {detail}"

    def _open_log(self, icon: pystray.Icon, item: pystray.MenuItem) -> None:
        log_path = CONFIG_DIR / "agent.err.log"
        if not log_path.exists():
            log_path = CONFIG_DIR / "agent.log"
        if log_path.exists():
            os.startfile(log_path)  # type: ignore[attr-defined]
        else:
            icon.title = "Лог не найден"

    def _open_config_dir(self, icon: pystray.Icon, item: pystray.MenuItem) -> None:
        CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        os.startfile(CONFIG_DIR)  # type: ignore[attr-defined]

    def _menu_version_text(self) -> str:
        return f"Версия: {get_installed_version()}"

    def _menu_update_text(self) -> str:
        if self._latest_version:
            return f"Скачать {self._latest_version}"
        return "Скачать обновление"

    def _check_updates(self, icon: pystray.Icon, item: pystray.MenuItem) -> None:
        def _run() -> None:
            result = check_for_updates(force=True)
            self._update_available = result.update_available
            self._latest_version = result.latest_version
            icon.title = format_update_message(result)
            icon.update_menu()

        threading.Thread(target=_run, daemon=True).start()

    def _open_update(self, icon: pystray.Icon, item: pystray.MenuItem) -> None:
        result = check_for_updates(force=True)
        if result.update_available:
            open_update_page(result)

    def _quit(self, icon: pystray.Icon, item: pystray.MenuItem) -> None:
        self._stop.set()
        icon.stop()

    def _refresh_loop(self) -> None:
        while not self._stop.is_set():
            status = read_agent_status()
            service_state = get_service_state()
            color_key = _status_icon_key(status, service_state)
            if status is not None:
                self._update_available = status.update_available
                self._latest_version = status.latest_version or ""
            self._icon.icon = _make_icon(color_key)
            self._icon.title = _status_tooltip(status, service_state)
            self._icon.update_menu()
            time.sleep(3)

    def run(self) -> None:
        refresh_thread = threading.Thread(target=self._refresh_loop, daemon=True)
        refresh_thread.start()
        self._icon.run()


def _launch_settings_window(config_path: Path | None = None) -> None:
    if getattr(sys, "frozen", False):
        args = [sys.executable, "--settings"]
    else:
        args = [sys.executable, "-m", "system_monitor.agent", "--settings"]
    if config_path:
        args.extend(["--config", str(config_path)])
    subprocess.Popen(args)


def run_tray(config_path: Path | None = None) -> None:
    TrayApp(config_path or AGENT_CONFIG).run()
