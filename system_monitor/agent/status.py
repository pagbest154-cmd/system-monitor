from __future__ import annotations

import json
import time
from dataclasses import asdict, dataclass
from pathlib import Path

from ..paths import AGENT_STATUS_FILE


@dataclass
class AgentStatus:
    connected: bool = False
    agent_id: str = ""
    hub_url: str = ""
    hostname: str = ""
    last_success_ts: float | None = None
    last_error: str | None = None
    last_error_ts: float | None = None
    metrics_count: int = 0

    def status_label(self) -> str:
        if self.connected:
            return "Подключён"
        if self.last_error:
            return "Ошибка"
        return "Ожидание"

    def is_stale(self, max_age_sec: float = 30.0) -> bool:
        if self.last_success_ts is None:
            return True
        return (time.time() - self.last_success_ts) > max_age_sec


def write_agent_status(status: AgentStatus, path: Path | None = None) -> None:
    path = path or AGENT_STATUS_FILE
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        json.dumps(asdict(status), ensure_ascii=False, indent=2),
        encoding="utf-8",
    )


def read_agent_status(path: Path | None = None) -> AgentStatus | None:
    path = path or AGENT_STATUS_FILE
    if not path.exists():
        return None
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
        return AgentStatus(**data)
    except (json.JSONDecodeError, TypeError, ValueError):
        return None
