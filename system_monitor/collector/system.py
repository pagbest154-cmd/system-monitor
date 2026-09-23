from __future__ import annotations

import time

import psutil

from ..config_loader import default_disk_path
from .base import Sensor, SensorReading

_net_counters: dict[str, tuple[int, float]] = {}


class CpuPercentSensor(Sensor):
    def read(self) -> SensorReading:
        value = float(psutil.cpu_percent(interval=None))
        return SensorReading(
            sensor_id=self.id,
            value=round(value, 2),
            status=self.evaluate_status(value),
        )


class MemoryPercentSensor(Sensor):
    def read(self) -> SensorReading:
        value = float(psutil.virtual_memory().percent)
        return SensorReading(
            sensor_id=self.id,
            value=round(value, 2),
            status=self.evaluate_status(value),
        )


class DiskUsageSensor(Sensor):
    def read(self) -> SensorReading:
        path = self.config.params.get("path") or default_disk_path()
        try:
            usage = psutil.disk_usage(path)
        except OSError as exc:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="error",
                error=str(exc),
            )
        value = float(usage.percent)
        return SensorReading(
            sensor_id=self.id,
            value=round(value, 2),
            status=self.evaluate_status(value),
        )


class NetworkBytesSensor(Sensor):
    def read(self) -> SensorReading:
        direction = self.config.params.get("direction", "recv")
        counters = psutil.net_io_counters()
        now = time.time()
        current = counters.bytes_recv if direction == "recv" else counters.bytes_sent

        prev = _net_counters.get(self.id)
        if prev is None:
            _net_counters[self.id] = (current, now)
            return SensorReading(sensor_id=self.id, value=0.0, status="ok")

        prev_bytes, prev_time = prev
        elapsed = max(now - prev_time, 0.001)
        delta_mb = (current - prev_bytes) / (1024 * 1024) / elapsed
        _net_counters[self.id] = (current, now)

        value = max(float(delta_mb), 0.0)
        return SensorReading(
            sensor_id=self.id,
            value=round(value, 3),
            status=self.evaluate_status(value),
        )
