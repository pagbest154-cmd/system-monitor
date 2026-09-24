from __future__ import annotations

import threading
import time
from typing import Any, Callable

import psutil

from ..config_loader import SensorConfig, SensorsFile, load_sensors_config
from ..disk_discovery import discover_disk_sensors
from ..storage import MetricStore
from .base import Sensor, SensorReading
from .registry import create_sensor


class Collector:
    def __init__(
        self,
        store: MetricStore | None = None,
        on_update: Callable[[dict[str, Any]], None] | None = None,
        config: SensorsFile | None = None,
        config_loader: Callable[[], SensorsFile] | None = None,
    ) -> None:
        self.store = store
        self.on_update = on_update
        self._config_loader = config_loader or load_sensors_config
        self._stop = threading.Event()
        self._thread: threading.Thread | None = None
        self._config = config or self._config_loader()
        self._sensors: dict[str, Sensor] = {}
        self._last_poll: dict[str, float] = {}
        self._latest: dict[str, dict[str, Any]] = {}
        self._lock = threading.Lock()
        self.reload_config()

    def reload_config(self, config: SensorsFile | None = None) -> None:
        if config is not None:
            self._config = config
        else:
            self._config = self._config_loader()
        sensors: dict[str, Sensor] = {}
        configured_paths = {
            cfg.params.get("path")
            for cfg in self._config.sensors
            if cfg.type == "system.disk_usage" and cfg.params.get("path")
        }

        for item in self._config.sensors:
            if not item.enabled or not item.is_supported_on_platform():
                continue
            try:
                sensors[item.id] = create_sensor(item)
            except ValueError:
                continue

        if self._config.settings.auto_discover_disks:
            for item in discover_disk_sensors():
                mount = item.params.get("path")
                if mount in configured_paths:
                    continue
                try:
                    sensors[item.id] = create_sensor(item)
                except ValueError:
                    continue

        with self._lock:
            self._sensors = sensors

    def start(self) -> None:
        if self._thread and self._thread.is_alive():
            return
        psutil.cpu_percent(interval=None)
        self._stop.clear()
        self._thread = threading.Thread(target=self._run, name="system-monitor-collector", daemon=True)
        self._thread.start()

    def stop(self) -> None:
        self._stop.set()
        if self._thread:
            self._thread.join(timeout=2)

    def _interval_for(self, config: SensorConfig) -> int:
        return config.interval_sec or self._config.settings.default_interval_sec

    def _run(self) -> None:
        last_cleanup = 0.0
        while not self._stop.is_set():
            now = time.time()
            updates: dict[str, dict[str, Any]] = {}

            with self._lock:
                sensors = dict(self._sensors)

            for sensor_id, sensor in sensors.items():
                config = sensor.config
                interval = self._interval_for(config)
                last = self._last_poll.get(sensor_id, 0.0)
                if now - last < interval:
                    continue

                reading = self._safe_read(sensor)
                if self.store is not None:
                    self.store.insert(sensor_id, reading.value, reading.status, ts=now)
                payload = {
                    **reading.to_dict(),
                    "name": config.name,
                    "unit": config.unit,
                    "type": config.type,
                    "ts": now,
                }
                with self._lock:
                    self._latest[sensor_id] = payload
                updates[sensor_id] = payload
                self._last_poll[sensor_id] = now

            if updates and self.on_update:
                self.on_update(updates)

            if self.store is not None:
                retention = self._config.settings.retention_days
                if now - last_cleanup > 3600:
                    self.store.cleanup(retention)
                    last_cleanup = now

            time.sleep(0.5)

    def _safe_read(self, sensor: Sensor) -> SensorReading:
        try:
            return sensor.read()
        except Exception as exc:
            return SensorReading(
                sensor_id=sensor.id,
                value=None,
                status="error",
                error=str(exc),
            )

    def get_latest_snapshot(self) -> dict[str, dict[str, Any]]:
        with self._lock:
            return dict(self._latest)

    def get_sensor_configs(self) -> list[SensorConfig]:
        configs = list(self._config.sensors)
        configured_paths = {
            cfg.params.get("path")
            for cfg in configs
            if cfg.type == "system.disk_usage" and cfg.params.get("path")
        }
        for item in discover_disk_sensors():
            if item.params.get("path") in configured_paths:
                continue
            configs.append(item)
        return configs

    def get_settings(self) -> dict[str, Any]:
        return self._config.settings.model_dump(mode="json")

    def get_active_sensor_ids(self) -> set[str]:
        with self._lock:
            return set(self._sensors.keys())

    def build_sensor_meta(self) -> list[dict[str, Any]]:
        active_ids = self.get_active_sensor_ids()
        items = []
        for cfg in self.get_sensor_configs():
            if cfg.id not in active_ids:
                continue
            items.append(
                {
                    "id": cfg.id,
                    "name": cfg.name,
                    "type": cfg.type,
                    "unit": cfg.unit,
                    "enabled": cfg.enabled,
                    "interval_sec": cfg.interval_sec,
                    "params": cfg.params,
                    "warn_above": cfg.warn_above,
                    "critical_above": cfg.critical_above,
                }
            )
        return items
