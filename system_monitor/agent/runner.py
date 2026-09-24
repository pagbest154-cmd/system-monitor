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
    load_agent_token,
    merge_agent_config,
)
from ..protocol.models import AgentReport, MetricPoint, SensorMeta
from ..fleet.service import default_agent_id
from ..system_info import get_system_info
from .status import AgentStatus, write_agent_status
from .transport import AgentTransport, create_transport
from .updates import UPDATE_CHECK_INTERVAL_SEC, check_for_updates


class AgentRunner:
    def __init__(self, config: AgentFileConfig, transport: AgentTransport) -> None:
        self.config = config
        self.transport = transport
        self.agent_id = default_agent_id(config.agent_id)
        self._config_version = 0
        self._stop = threading.Event()
        self._thread: threading.Thread | None = None
        self._collector: Collector | None = None
        self._last_status = AgentStatus(
            agent_id=self.agent_id,
            hub_url=config.hub_url,
            hostname=socket.gethostname(),
        )
        self._last_update_check = 0.0

    def _write_status(self, **updates: object) -> None:
        for key, value in updates.items():
            setattr(self._last_status, key, value)
        write_agent_status(self._last_status)

    def _maybe_check_updates(self, now: float) -> None:
        if now - self._last_update_check < UPDATE_CHECK_INTERVAL_SEC:
            return
        self._last_update_check = now
        try:
            result = check_for_updates()
            if result.update_available:
                print(
                    "system-monitor-agent: доступно обновление "
                    f"{result.latest_version} (установлена {result.current_version})"
                )
            elif result.error:
                print(f"system-monitor-agent: проверка обновлений: {result.error}")
            self._write_status(
                update_available=result.update_available,
                latest_version=result.latest_version,
                release_url=result.release_url,
                update_checked_at=result.checked_at,
            )
        except Exception as exc:
            print(f"system-monitor-agent: ошибка проверки обновлений: {exc}")

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
        return len(metrics)

    def _run(self) -> None:
        self._ensure_collector()
        self._write_status(connected=False)
        last_config_sync = 0.0
        self._maybe_check_updates(time.time())
        while not self._stop.is_set():
            now = time.time()
            self._maybe_check_updates(now)
            if now - last_config_sync > 60:
                try:
                    self.reload_collector()
                    last_config_sync = now
                except Exception as exc:
                    print(f"system-monitor-agent: ошибка SyncConfig: {exc}")
                    self._write_status(
                        connected=False,
                        last_error=str(exc),
                        last_error_ts=now,
                    )
            try:
                metrics_count = self._push_once()
                self._write_status(
                    connected=True,
                    last_success_ts=time.time(),
                    last_error=None,
                    last_error_ts=None,
                    metrics_count=metrics_count,
                )
            except Exception as exc:
                print(f"system-monitor-agent: ошибка отправки метрик: {exc}")
                self._write_status(
                    connected=False,
                    last_error=str(exc),
                    last_error_ts=time.time(),
                )
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
    token = load_agent_token(config)
    transport = create_transport(config.transport, config.hub_url, token)
    return AgentRunner(config, transport)
