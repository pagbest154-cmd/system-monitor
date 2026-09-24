from __future__ import annotations

import argparse
import signal
import sys
from pathlib import Path

from ..paths import AGENT_CONFIG
from .runner import build_runner


def _configure_stdio() -> None:
    if sys.platform != "win32":
        return
    for stream in (sys.stdout, sys.stderr):
        reconfigure = getattr(stream, "reconfigure", None)
        if reconfigure is not None:
            try:
                reconfigure(encoding="utf-8", errors="replace")
            except (OSError, ValueError):
                pass


def _attach_parent_console() -> None:
    """Allow --check-update to print when the frozen exe has no console."""
    if sys.platform != "win32" or not getattr(sys, "frozen", False):
        return
    import ctypes

    if ctypes.windll.kernel32.AttachConsole(-1):  # ATTACH_PARENT_PROCESS
        sys.stdout = open("CONOUT$", "w", encoding="utf-8", errors="replace")  # noqa: SIM115
        sys.stderr = open("CONOUT$", "w", encoding="utf-8", errors="replace")  # noqa: SIM115
        _configure_stdio()


def main() -> None:
    _configure_stdio()
    parser = argparse.ArgumentParser(description="system-monitor-agent")
    parser.add_argument("--config", type=Path, default=AGENT_CONFIG, help="Путь к agent.yaml")
    parser.add_argument("--tray", action="store_true", help="Иконка в системном трее (Windows)")
    parser.add_argument("--settings", action="store_true", help="Окно настроек (Windows)")
    parser.add_argument("--check-update", action="store_true", help="Проверить наличие обновлений")
    args = parser.parse_args()

    if args.check_update:
        from .updates import check_for_updates, format_update_message

        _attach_parent_console()
        result = check_for_updates(force=True)
        print(format_update_message(result))
        if result.error:
            sys.exit(1)
        if result.update_available:
            sys.exit(2)
        sys.exit(0)

    if args.tray or args.settings:
        if sys.platform != "win32":
            print("Режим трея и настроек доступен только на Windows")
            sys.exit(1)

    if args.tray:
        from .tray import run_tray

        run_tray(args.config)
        return

    if args.settings:
        from .settings_ui import run_settings

        run_settings(args.config)
        return

    runner = build_runner(args.config)
    print(f"system-monitor-agent запущен: {runner.agent_id} -> {runner.config.hub_url}")

    def _stop(_signum: int, _frame: object) -> None:
        runner.stop()

    signal.signal(signal.SIGINT, _stop)
    signal.signal(signal.SIGTERM, _stop)

    runner.start()
    try:
        if runner._thread:
            runner._thread.join()
    except KeyboardInterrupt:
        runner.stop()


if __name__ == "__main__":
    try:
        main()
    except Exception:
        import traceback

        traceback.print_exc()
        sys.exit(1)
