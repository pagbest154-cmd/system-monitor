from __future__ import annotations

import socket
from typing import Any

from ..config_loader import find_agent_entry
from ..protocol.models import AgentConfigResponse, AgentReport, SensorOverride, prefixed_sensor_id


def default_agent_id(configured: str) -> str:
    if configured.strip():
        return configured.strip()
    return socket.gethostname()


def verify_agent_token(agent_id: str, token: str | None) -> bool:
    if not token:
        return False
    entry = find_agent_entry(agent_id)
    if entry is None:
        return False
    return entry.token == token


def build_agent_config_response(agent_id: str, config_version: int) -> AgentConfigResponse:
    entry = find_agent_entry(agent_id)
    if entry is None:
        return AgentConfigResponse(version=config_version, overrides=[])
    overrides = [
        SensorOverride(
            sensor_id=item.sensor_id,
            enabled=item.enabled,
            interval_sec=item.interval_sec,
            warn_above=item.warn_above,
            critical_above=item.critical_above,
            params=item.params,
        )
        for item in entry.overrides
    ]
    version = config_version if not overrides else max(config_version, 1)
    return AgentConfigResponse(version=version, overrides=overrides)


def ingest_agent_report(
    report: AgentReport,
    *,
    store: Any,
    live_hub: Any | None = None,
    retention_days: int = 31,
) -> dict[str, Any]:
    entry = find_agent_entry(report.agent_id)
    display_name = entry.name if entry and entry.name else report.agent_id

    rows: list[tuple[str, float, float | None, str]] = []
    for point in report.metrics:
        full_id = prefixed_sensor_id(report.agent_id, point.sensor_id)
        rows.append((full_id, point.ts, point.value, point.status))

    store.insert_batch(rows)
    sensors_payload = [item.model_dump(mode="json") for item in report.sensors]
    store.upsert_agent(
        report.agent_id,
        name=display_name,
        hostname=report.hostname,
        status="online",
        system=report.system,
        sensors=sensors_payload,
    )

    if live_hub is not None:
        latest = {
            prefixed_sensor_id(report.agent_id, point.sensor_id): {
                "sensor_id": prefixed_sensor_id(report.agent_id, point.sensor_id),
                "value": point.value,
                "status": point.status,
                "ts": point.ts,
            }
            for point in report.metrics
        }
        if latest:
            live_hub.schedule_broadcast(
                {"type": "update", "data": latest, "agent_id": report.agent_id}
            )

    return {"status": "ok", "metrics": len(rows)}
