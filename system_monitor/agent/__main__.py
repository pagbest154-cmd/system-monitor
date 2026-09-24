from __future__ import annotations

import argparse
import signal
from pathlib import Path

from ..paths import AGENT_CONFIG
from .runner import build_runner


def main() -> None:
    parser = argparse.ArgumentParser(description="system-monitor-agent")
    parser.add_argument("--config", type=Path, default=AGENT_CONFIG, help="Путь к agent.yaml")
    args = parser.parse_args()

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
