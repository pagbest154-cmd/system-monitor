from __future__ import annotations

from typing import Callable

from ..config_loader import SensorConfig
from .base import Sensor
from .gpio import Dht22Sensor
from .remote import HttpJsonSensor, MqttSensor
from .system import (
    CpuPercentSensor,
    DiskUsageSensor,
    MemoryPercentSensor,
    NetworkBytesSensor,
)
from .temperature import GpuTemperatureSensor, TemperatureSensor

SensorFactory = Callable[[SensorConfig], Sensor]

_REGISTRY: dict[str, SensorFactory] = {
    "system.cpu_percent": CpuPercentSensor,
    "system.memory_percent": MemoryPercentSensor,
    "system.disk_usage": DiskUsageSensor,
    "system.network_bytes": NetworkBytesSensor,
    "system.temperature": TemperatureSensor,
    "system.gpu_temperature": GpuTemperatureSensor,
    "gpio.dht22": Dht22Sensor,
    "remote.http_json": HttpJsonSensor,
    "remote.mqtt": MqttSensor,
}


def get_registered_types() -> list[str]:
    return sorted(_REGISTRY.keys())


def create_sensor(config: SensorConfig) -> Sensor:
    factory = _REGISTRY.get(config.type)
    if factory is None:
        raise ValueError(f"Неизвестный тип датчика: {config.type}")
    return factory(config)
