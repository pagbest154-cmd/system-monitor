from __future__ import annotations

import socket
import threading
import time
from pathlib import Path

from ..collector.runner import Collector
from ..config_loader import (
    AgentFileConfig,
    SensorOverrideConfig,
    load_agent_config,
    load_agent_sensors_config,
    merge_agent_config,
)
from ..protocol.models import AgentReport, MetricPoint, SensorMeta
from ..fleet.service import default_agent_id
from ..system_info import get_system_info
from .transport import AgentTransport, create_transport


class AgentRunner:
    def __init__(self, config: AgentFileConfig, transport: AgentTransport) -> None:
        self.config = config
        self.transport = transport
        self.agent_id = default_agent_id(config.agent_id)
        self._config_version = 0
        self._stop = threading.Event()
        self._thread: threading.Thread | None = None
        self._collector: Collector | None = None

    def reload_collector(self) -> None:
        base = load_agent_sensors_config()
        remote = self.transport.sync_config(self.agent_id, self._config_version)
        self._config_version = remote.version
        overrides = [
            SensorOverrideConfig.model_validate(item.model_dump())
            for item in remote.overrides
        ]
        merged = merge_agent_config(base, overrides)
        if self._collector is None:
            self._collector = Collector(store=None, config=merged)
        else:
            self._collector.reload_config(merged)

    def _ensure_collector(self) -> Collector:
        if self._collector is None:
            self._collector = Collector(store=None, config=load_agent_sensors_config())
            self._collector.start()
        return self._collector

    def _push_once(self) -> None:
        collector = self._ensure_collector()
        snapshot = collector.get_latest_snapshot()
        metrics = [
            MetricPoint(
                sensor_id=str(item.get("sensor_id", sensor_id)),
                ts=float(item.get("ts", time.time())),
                value=item.get("value"),
                status=str(item.get("status", "unknown")),
            )
            for sensor_id, item in snapshot.items()
        ]
        sensors = [SensorMeta.model_validate(item) for item in collector.build_sensor_meta()]
        report = AgentReport(
            agent_id=self.agent_id,
            hostname=socket.gethostname(),
            metrics=metrics,
            system=get_system_info(),
            config_version=self._config_version,
            sensors=sensors,
        )
        self.transport.push_report(report)

    def _run(self) -> None:
        self._ensure_collector()
        last_config_sync = 0.0
        while not self._stop.is_set():
            now = time.time()
            if now - last_config_sync > 60:
                try:
                    self.reload_collector()
                    last_config_sync = now
                except Exception as exc:
                    print(f"system-monitor-agent: ошибка SyncConfig: {exc}")
            try:
                self._push_once()
            except Exception as exc:
                print(f"system-monitor-agent: ошибка отправки метрик: {exc}")
            time.sleep(max(1, self.config.interval_sec))

    def start(self) -> None:
        if self._thread and self._thread.is_alive():
            return
        self._stop.clear()
        self._thread = threading.Thread(target=self._run, name="system-monitor-agent", daemon=False)
        self._thread.start()

    def stop(self) -> None:
        self._stop.set()
        if self._collector is not None:
            self._collector.stop()
        if self._thread:
            self._thread.join(timeout=3)
        self.transport.close()


def build_runner(config_path: Path | None = None) -> AgentRunner:
    config = load_agent_config(config_path)
    token = config.token
    if not token and config.token_file:
        token_path = Path(config.token_file)
        if token_path.exists():
            token = token_path.read_text(encoding="utf-8").strip()
    transport = create_transport(config.transport, config.hub_url, token)
    return AgentRunner(config, transport)
