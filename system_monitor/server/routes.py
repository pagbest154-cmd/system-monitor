from __future__ import annotations

import asyncio
from typing import Any

from fastapi import APIRouter, WebSocket, WebSocketDisconnect
from pydantic import BaseModel

from ..collector.runner import Collector
from ..config_loader import (
    DashboardFile,
    SensorsFile,
    load_dashboard_config,
    load_sensors_config,
    save_dashboard_config,
    save_sensors_config,
)
from ..storage import MetricStore
from ..system_info import get_system_info

router = APIRouter()


class LiveHub:
    def __init__(self) -> None:
        self.connections: list[WebSocket] = []

    async def connect(self, websocket: WebSocket) -> None:
        await websocket.accept()
        self.connections.append(websocket)

    def disconnect(self, websocket: WebSocket) -> None:
        if websocket in self.connections:
            self.connections.remove(websocket)

    async def broadcast(self, payload: dict[str, Any]) -> None:
        dead: list[WebSocket] = []
        for ws in self.connections:
            try:
                await ws.send_json(payload)
            except Exception:
                dead.append(ws)
        for ws in dead:
            self.disconnect(ws)


def setup_state(collector: Collector, store: MetricStore, hub: LiveHub) -> None:
    router.collector = collector
    router.store = store
    router.hub = hub

    def on_update(data: dict[str, Any]) -> None:
        try:
            loop = asyncio.get_running_loop()
        except RuntimeError:
            return
        loop.create_task(hub.broadcast({"type": "update", "data": data}))

    collector.on_update = on_update


@router.get("/api/sensors")
def list_sensors() -> dict[str, Any]:
    collector: Collector = router.collector
    configs = collector.get_sensor_configs()
    latest = collector.get_latest_snapshot()
    active_ids = collector.get_active_sensor_ids()
    items = []
    for cfg in configs:
        current = latest.get(cfg.id)
        supported = cfg.is_supported_on_platform()
        items.append(
            {
                "id": cfg.id,
                "name": cfg.name,
                "type": cfg.type,
                "enabled": cfg.enabled,
                "interval_sec": cfg.interval_sec,
                "unit": cfg.unit,
                "params": cfg.params,
                "platforms": cfg.platforms,
                "warn_above": cfg.warn_above,
                "critical_above": cfg.critical_above,
                "supported": supported,
                "available": supported and cfg.id in active_ids,
                "current": current,
            }
        )
    settings = collector.get_settings()
    return {"settings": settings, "sensors": items}


@router.get("/api/metrics/{sensor_id}")
def get_metrics(sensor_id: str, period: str = "1h") -> dict[str, Any]:
    store: MetricStore = router.store
    return {"sensor_id": sensor_id, "period": period, "points": store.get_history(sensor_id, period)}


@router.get("/api/system")
def system_info() -> dict[str, Any]:
    return get_system_info()


@router.get("/api/dashboard")
def get_dashboard() -> dict[str, Any]:
    config = load_dashboard_config()
    return config.model_dump(mode="json")


class SensorsUpdate(BaseModel):
    settings: dict[str, Any] | None = None
    sensors: list[dict[str, Any]]


class DashboardUpdate(BaseModel):
    dashboard: dict[str, Any] | None = None
    panels: list[dict[str, Any]]


@router.put("/api/config/sensors")
def update_sensors_config(body: SensorsUpdate) -> dict[str, str]:
    current = load_sensors_config()
    data = current.model_dump(mode="json")
    if body.settings is not None:
        data["settings"] = body.settings
    if body.sensors is not None:
        data["sensors"] = body.sensors
    config = SensorsFile.model_validate(data)
    save_sensors_config(config)
    router.collector.reload_config()
    return {"status": "ok"}


@router.put("/api/config/dashboard")
def update_dashboard_config(body: DashboardUpdate) -> dict[str, str]:
    current = load_dashboard_config()
    data = current.model_dump(mode="json")
    if body.dashboard is not None:
        data["dashboard"] = body.dashboard
    if body.panels is not None:
        data["panels"] = body.panels
    config = DashboardFile.model_validate(data)
    save_dashboard_config(config)
    return {"status": "ok"}


@router.get("/api/sensor-types")
def sensor_types() -> dict[str, list[str]]:
    from ..collector.registry import get_registered_types

    return {"types": get_registered_types()}


@router.websocket("/ws/live")
async def live_ws(websocket: WebSocket) -> None:
    hub: LiveHub = router.hub
    collector: Collector = router.collector
    await hub.connect(websocket)
    try:
        await websocket.send_json(
            {"type": "snapshot", "data": collector.get_latest_snapshot()}
        )
        while True:
            await websocket.receive_text()
    except WebSocketDisconnect:
        hub.disconnect(websocket)
