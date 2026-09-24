from __future__ import annotations

import argparse
import signal
import sys
from pathlib import Path

from ..paths import AGENT_CONFIG
from .runner import build_runner


def main() -> None:
    parser = argparse.ArgumentParser(description="system-monitor-agent")
    parser.add_argument("--config", type=Path, default=AGENT_CONFIG, help="Путь к agent.yaml")
    parser.add_argument("--tray", action="store_true", help="Иконка в системном трее (Windows)")
    parser.add_argument("--settings", action="store_true", help="Окно настроек (Windows)")
    args = parser.parse_args()

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
    print(f"system-monitor-agent запущен: {runner.agent_id} → {runner.config.hub_url}")

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
    main()
