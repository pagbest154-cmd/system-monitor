from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Any

from ..config_loader import SensorConfig


@dataclass
class SensorReading:
    sensor_id: str
    value: float | None
    status: str = "ok"
    error: str | None = None

    def to_dict(self) -> dict[str, Any]:
        return {
            "sensor_id": self.sensor_id,
            "value": self.value,
            "status": self.status,
            "error": self.error,
        }


class Sensor(ABC):
    def __init__(self, config: SensorConfig) -> None:
        self.config = config

    @property
    def id(self) -> str:
        return self.config.id

    @abstractmethod
    def read(self) -> SensorReading:
        raise NotImplementedError

    def evaluate_status(self, value: float | None) -> str:
        if value is None:
            return "unknown"
        critical = self.config.critical_above
        warn = self.config.warn_above
        if critical is not None and value >= critical:
            return "critical"
        if warn is not None and value >= warn:
            return "warning"
        return "ok"
