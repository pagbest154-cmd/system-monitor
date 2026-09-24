from __future__ import annotations

from system_monitor.fleet.report_enrichment import (
    enrich_agent_report,
    reading_from_system,
    sensor_metas_from_system,
)
from system_monitor.protocol.models import AgentReport, MetricPoint


def test_enrich_uses_disks_not_storage() -> None:
    report = AgentReport(
        agent_id="homepc",
        metrics=[],
        sensors=[],
        system={
            "cpu": {"percent": 24.5},
            "memory": {"percent": 71.7},
            "disks": [
                {"mountpoint": "C:\\", "percent": 55.0},
                {"mountpoint": "D:\\", "percent": 12.0},
            ],
        },
    )

    enriched = enrich_agent_report(report)
    sensor_ids = {point.sensor_id for point in enriched.metrics}

    assert "cpu_percent" in sensor_ids
    assert "ram_used" in sensor_ids
    assert "disk_auto_c" in sensor_ids
    assert "disk_auto_d" in sensor_ids
    assert enriched.metrics[0].value == 24.5


def test_sensor_metas_from_system_includes_defaults_and_disks() -> None:
    metas = sensor_metas_from_system(
        {
            "cpu": {"percent": 1},
            "disks": [{"mountpoint": "E:\\", "percent": 80}],
        }
    )
    ids = {item["id"] for item in metas}
    assert "cpu_percent" in ids
    assert "ram_used" in ids
    assert "disk_auto_e" in ids


def test_reading_from_system_uses_disks_key() -> None:
    reading = reading_from_system(
        {
            "cpu": {"percent": 12.5},
            "memory": {"percent": 64.0},
            "disks": [{"mountpoint": "C:\\", "percent": 40.0}],
        },
        "disk_auto_c",
    )
    assert reading is not None
    assert reading["value"] == 40.0


def test_enrich_overrides_null_collector_metric() -> None:
    report = AgentReport(
        agent_id="homepc",
        metrics=[MetricPoint(sensor_id="cpu_percent", ts=1, value=None, status="unknown")],
        sensors=[],
        system={"cpu": {"percent": 33.3}, "memory": {"percent": 50.0}},
    )
    enriched = enrich_agent_report(report)
    cpu = next(point for point in enriched.metrics if point.sensor_id == "cpu_percent")
    assert cpu.value == 33.3
    assert cpu.status == "ok"
