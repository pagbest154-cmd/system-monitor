from __future__ import annotations

import time

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


def enrich_agent_report(report: AgentReport) -> AgentReport:
    """Add core metrics from system snapshot when the collector did not send them."""
    existing_ids = {point.sensor_id for point in report.metrics}
    extra_metrics: list[MetricPoint] = []
    now = time.time()
    system = report.system or {}

    cpu = system.get("cpu") or {}
    if "cpu_percent" not in existing_ids and cpu.get("percent") is not None:
        extra_metrics.append(
            MetricPoint(
                sensor_id="cpu_percent",
                ts=now,
                value=float(cpu["percent"]),
                status="ok",
            )
        )

    memory = system.get("memory") or {}
    if "ram_used" not in existing_ids and memory.get("percent") is not None:
        extra_metrics.append(
            MetricPoint(
                sensor_id="ram_used",
                ts=now,
                value=float(memory["percent"]),
                status="ok",
            )
        )

    for part in system.get("storage") or []:
        mountpoint = str(part.get("mountpoint") or "")
        sensor_id = _mount_to_sensor_id(mountpoint)
        if sensor_id in existing_ids:
            continue
        percent = part.get("percent")
        if percent is None:
            continue
        extra_metrics.append(
            MetricPoint(
                sensor_id=sensor_id,
                ts=now,
                value=float(percent),
                status="ok",
            )
        )

    if not extra_metrics:
        return report

    sensors = list(report.sensors)
    known_ids = {sensor.id for sensor in sensors}
    for point in extra_metrics:
        if point.sensor_id in known_ids:
            continue
        if point.sensor_id in _DEFAULT_SENSORS:
            sensors.append(_DEFAULT_SENSORS[point.sensor_id])
            continue
        if point.sensor_id.startswith("disk_auto_"):
            mount = next(
                (
                    str(part.get("mountpoint") or "")
                    for part in system.get("storage") or []
                    if _mount_to_sensor_id(str(part.get("mountpoint") or "")) == point.sensor_id
                ),
                point.sensor_id.removeprefix("disk_auto_"),
            )
            sensors.append(
                SensorMeta(
                    id=point.sensor_id,
                    name=f"Диск {mount}",
                    type="system.disk_usage",
                    unit="%",
                )
            )

    return report.model_copy(update={"metrics": [*report.metrics, *extra_metrics], "sensors": sensors})
