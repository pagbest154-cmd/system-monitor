from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


def prefixed_sensor_id(agent_id: str, sensor_id: str) -> str:
    return f"{agent_id}:{sensor_id}"


def strip_agent_prefix(agent_id: str, full_id: str) -> str:
    prefix = f"{agent_id}:"
    if full_id.startswith(prefix):
        return full_id[len(prefix) :]
    return full_id


class MetricPoint(BaseModel):
    sensor_id: str
    ts: float
    value: float | None = None
    status: str = "unknown"


class SensorMeta(BaseModel):
    id: str
    name: str
    type: str
    unit: str = ""
    enabled: bool = True
    interval_sec: int | None = None
    params: dict[str, Any] = Field(default_factory=dict)
    warn_above: float | None = None
    critical_above: float | None = None


class AgentReport(BaseModel):
    agent_id: str
    hostname: str = ""
    metrics: list[MetricPoint] = Field(default_factory=list)
    system: dict[str, Any] | None = None
    config_version: int = 0
    sensors: list[SensorMeta] = Field(default_factory=list)


class SensorOverride(BaseModel):
    sensor_id: str
    enabled: bool | None = None
    interval_sec: int | None = None
    warn_above: float | None = None
    critical_above: float | None = None
    params: dict[str, Any] | None = None


class AgentConfigResponse(BaseModel):
    version: int = 0
    overrides: list[SensorOverride] = Field(default_factory=list)
