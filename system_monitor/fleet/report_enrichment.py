from __future__ import annotations

import time
from typing import Any

from ..disk_discovery import _mount_to_sensor_id
from ..protocol.models import AgentReport, MetricPoint, SensorMeta

_DEFAULT_SENSORS: dict[str, SensorMeta] = {
    "cpu_percent": SensorMeta(
        id="cpu_percent",
        name="Загрузка процессора",
        type="system.cpu_percent",
        unit="%",
    ),
    "ram_used": SensorMeta(
        id="ram_used",
        name="Использование памяти",
        type="system.memory_percent",
        unit="%",
    ),
}


def _disk_partitions(system: dict[str, Any]) -> list[dict[str, Any]]:
    for key in ("disks", "partitions", "storage"):
        value = system.get(key)
        if isinstance(value, list) and value:
            return value
    return []


def reading_from_system(system: dict[str, Any], sensor_id: str, ts: float | None = None) -> dict[str, Any] | None:
    """Build a current reading for a sensor from a stored system snapshot."""
    if not system:
        return None
    now = ts if ts is not None else time.time()

    if sensor_id == "cpu_percent":
        value = (system.get("cpu") or {}).get("percent")
        if value is None:
            return None
        return {"value": float(value), "status": "ok", "ts": now, "unit": "%"}

    if sensor_id == "ram_used":
        value = (system.get("memory") or {}).get("percent")
        if value is None:
            return None
        return {"value": float(value), "status": "ok", "ts": now, "unit": "%"}

    if sensor_id.startswith("disk_auto_"):
        for part in _disk_partitions(system):
            mountpoint = str(part.get("mountpoint") or "")
            if _mount_to_sensor_id(mountpoint) != sensor_id:
                continue
            value = part.get("percent")
            if value is None:
                return None
            return {"value": float(value), "status": "ok", "ts": now, "unit": "%"}

    return None


def sensor_metas_from_system(system: dict[str, Any]) -> list[dict[str, Any]]:
    """Build sensor metadata list from a stored system snapshot."""
    metas = [meta.model_dump(mode="json") for meta in _DEFAULT_SENSORS.values()]
    known = {meta["id"] for meta in metas}
    for part in _disk_partitions(system):
        mountpoint = str(part.get("mountpoint") or "")
        sensor_id = _mount_to_sensor_id(mountpoint)
        if sensor_id in known:
            continue
        known.add(sensor_id)
        metas.append(
            SensorMeta(
                id=sensor_id,
                name=f"Диск {mountpoint}",
                type="system.disk_usage",
                unit="%",
            ).model_dump(mode="json")
        )
    return metas


def enrich_agent_report(report: AgentReport) -> AgentReport:
    """Ensure core metrics are present, using the system snapshot as source of truth."""
    system = report.system or {}
    now = time.time()
    metrics_map = {point.sensor_id: point for point in report.metrics}

    def set_metric(sensor_id: str, value: Any) -> None:
        if value is None:
            metrics_map.pop(sensor_id, None)
            return
        metrics_map[sensor_id] = MetricPoint(
            sensor_id=sensor_id,
            ts=now,
            value=float(value),
            status="ok",
        )

    cpu = system.get("cpu") or {}
    set_metric("cpu_percent", cpu.get("percent"))

    memory = system.get("memory") or {}
    set_metric("ram_used", memory.get("percent"))

    for part in _disk_partitions(system):
        mountpoint = str(part.get("mountpoint") or "")
        sensor_id = _mount_to_sensor_id(mountpoint)
        set_metric(sensor_id, part.get("percent"))

    sensors = list(report.sensors)
    known_ids = {sensor.id for sensor in sensors}
    for sensor_id in metrics_map:
        if sensor_id in known_ids:
            continue
        if sensor_id in _DEFAULT_SENSORS:
            sensors.append(_DEFAULT_SENSORS[sensor_id])
            continue
        if sensor_id.startswith("disk_auto_"):
            mount = next(
                (
                    str(part.get("mountpoint") or "")
                    for part in _disk_partitions(system)
                    if _mount_to_sensor_id(str(part.get("mountpoint") or "")) == sensor_id
                ),
                sensor_id.removeprefix("disk_auto_"),
            )
            sensors.append(
                SensorMeta(
                    id=sensor_id,
                    name=f"Диск {mount}",
                    type="system.disk_usage",
                    unit="%",
                )
            )

    if not sensors and metrics_map:
        sensors = [SensorMeta.model_validate(item) for item in sensor_metas_from_system(system)]

    return report.model_copy(
        update={
            "metrics": list(metrics_map.values()),
            "sensors": sensors,
        }
    )
