from __future__ import annotations

import asyncio
from typing import Any

from fastapi import APIRouter, Header, HTTPException, Query, Request, Response, WebSocket, WebSocketDisconnect
from pydantic import BaseModel

from ..collector.runner import Collector
from ..config_loader import (
    AgentsFile,
    DashboardFile,
    HubFile,
    HubSettings,
    SensorsFile,
    hub_trusted_hosts,
    load_agents_config,
    load_dashboard_config,
    load_hub_config,
    load_sensors_config,
    normalize_domain,
    prepare_agents_for_save,
    resolve_public_url,
    save_agents_config,
    save_dashboard_config,
    save_hub_config,
    save_sensors_config,
)
from ..protocol.models import AgentConfigResponse, AgentReport, prefixed_sensor_id, strip_agent_prefix
from ..storage import MetricStore
from ..system_info import get_system_info
from ..fleet.report_enrichment import sensor_metas_from_system
from ..fleet.service import build_agent_config_response, ingest_agent_report, verify_agent_token
from .auth import (
    clear_session_cookie_header,
    create_session_token,
    hub_auth_enabled,
    hub_name,
    is_authenticated,
    request_is_secure,
    session_cookie_header,
    verify_credentials,
)
from .state import FleetState, LiveHub

router = APIRouter()


class AppState:
    mode: str = "standalone"
    collector: Collector | None = None
    store: MetricStore
    hub: LiveHub
    fleet: FleetState
    retention_days: int = 31


def setup_state(
    *,
    mode: str,
    store: MetricStore,
    hub: LiveHub,
    fleet: FleetState,
    collector: Collector | None = None,
    retention_days: int = 31,
) -> None:
    router.app_state = AppState()
    router.app_state.mode = mode
    router.app_state.collector = collector
    router.app_state.store = store
    router.app_state.hub = hub
    router.app_state.fleet = fleet
    router.app_state.retention_days = retention_days

    if collector is not None:

        def on_update(data: dict[str, Any]) -> None:
            try:
                loop = asyncio.get_running_loop()
            except RuntimeError:
                return
            loop.create_task(hub.broadcast({"type": "update", "data": data}))

        collector.on_update = on_update


def _state() -> AppState:
    return router.app_state


def _require_token(agent_id: str, authorization: str | None) -> None:
    token = None
    if authorization and authorization.lower().startswith("bearer "):
        token = authorization[7:].strip()
    if not verify_agent_token(agent_id, token):
        raise HTTPException(status_code=401, detail="Неверный токен агента")


@router.get("/api/mode")
def get_mode() -> dict[str, str]:
    return {"mode": _state().mode}


@router.get("/api/sensors")
def list_sensors(agent: str | None = Query(default=None)) -> dict[str, Any]:
    state = _state()
    if agent:
        record = state.store.get_agent(agent)
        if record is None:
            raise HTTPException(status_code=404, detail="Агент не найден")
        latest = state.fleet.get_agent_snapshot(agent)
        metas = record.get("sensors") or []
        if not metas:
            system = record.get("system") or {}
            if system:
                metas = sensor_metas_from_system(system)
            else:
                metas = [
                    {"id": sensor_id, "name": sensor_id, "type": "unknown", "unit": "%"}
                    for sensor_id in state.store.list_agent_sensor_ids(agent)
                ]
        items = []
        for meta in metas:
            sensor_id = meta["id"]
            full_id = prefixed_sensor_id(agent, sensor_id)
            current = latest.get(full_id)
            if current is None:
                current = state.store.get_latest(full_id)
                if current is not None:
                    latest[full_id] = {**current, "sensor_id": full_id}
            items.append(
                {
                    **meta,
                    "id": sensor_id,
                    "full_id": full_id,
                    "supported": True,
                    "available": True,
                    "current": current,
                }
            )
        if latest:
            state.fleet.update_agent(agent, latest)
        return {"settings": {"retention_days": state.retention_days}, "sensors": items, "agent_id": agent}

    collector = state.collector
    if collector is None:
        return {"settings": {"retention_days": state.retention_days}, "sensors": []}

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


@router.get("/api/metrics/{sensor_id:path}")
def get_metrics(
    sensor_id: str,
    period: str = "1h",
    agent: str | None = Query(default=None),
) -> dict[str, Any]:
    store = _state().store
    if agent:
        sensor_id = prefixed_sensor_id(agent, sensor_id)
    return {"sensor_id": sensor_id, "period": period, "points": store.get_history(sensor_id, period)}


@router.get("/api/system")
def system_info(agent: str | None = Query(default=None)) -> dict[str, Any]:
    if agent:
        record = _state().store.get_agent(agent)
        if record is None or not record.get("system"):
            raise HTTPException(status_code=404, detail="Данные системы агента недоступны")
        return record["system"]
    if _state().mode == "hub":
        raise HTTPException(status_code=400, detail="Укажите agent=<id> в hub-режиме")
    return get_system_info()


@router.get("/api/agents")
def list_agents() -> dict[str, Any]:
    agents = _state().store.list_agents()
    return {"agents": agents}


@router.get("/api/agents/{agent_id}")
def get_agent(agent_id: str) -> dict[str, Any]:
    record = _state().store.get_agent(agent_id)
    if record is None:
        raise HTTPException(status_code=404, detail="Агент не найден")
    return record


@router.get("/api/agents/{agent_id}/system")
def agent_system(agent_id: str) -> dict[str, Any]:
    record = _state().store.get_agent(agent_id)
    if record is None or not record.get("system"):
        raise HTTPException(status_code=404, detail="Данные системы агента недоступны")
    return record["system"]


@router.post("/api/agents/{agent_id}/metrics")
def push_metrics(
    agent_id: str,
    body: AgentReport,
    authorization: str | None = Header(default=None),
) -> dict[str, Any]:
    if body.agent_id != agent_id:
        raise HTTPException(status_code=400, detail="agent_id в URL и теле не совпадают")
    _require_token(agent_id, authorization)
    state = _state()
    from ..fleet.report_enrichment import enrich_agent_report

    report = enrich_agent_report(body)
    result = ingest_agent_report(
        report,
        store=state.store,
        live_hub=state.hub,
        retention_days=state.retention_days,
    )
    latest = {
        prefixed_sensor_id(agent_id, point.sensor_id): {
            "sensor_id": prefixed_sensor_id(agent_id, point.sensor_id),
            "value": point.value,
            "status": point.status,
            "ts": point.ts,
        }
        for point in report.metrics
    }
    state.fleet.update_agent(agent_id, latest)
    return result


@router.post("/api/agents/{agent_id}/heartbeat")
def agent_heartbeat(
    agent_id: str,
    authorization: str | None = Header(default=None),
) -> dict[str, str]:
    _require_token(agent_id, authorization)
    _state().store.touch_agent(agent_id)
    return {"status": "ok"}


@router.post("/api/agents/{agent_id}/config")
def sync_config(
    agent_id: str,
    config_version: int = 0,
    authorization: str | None = Header(default=None),
) -> AgentConfigResponse:
    _require_token(agent_id, authorization)
    return build_agent_config_response(agent_id, config_version)


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


class AgentsUpdate(BaseModel):
    agents: list[dict[str, Any]]


class HubUpdate(BaseModel):
    hub: dict[str, Any]


class LoginRequest(BaseModel):
    name: str
    key: str


@router.get("/api/auth/status")
def auth_status(request: Request) -> dict[str, Any]:
    enabled = _state().mode == "hub" and hub_auth_enabled()
    return {
        "auth_required": enabled,
        "hub_name": hub_name() if enabled else "",
        "authenticated": is_authenticated(request.headers, request.cookies),
    }


@router.post("/api/auth/login")
def login(body: LoginRequest, request: Request) -> Response:
    if _state().mode != "hub" or not hub_auth_enabled():
        raise HTTPException(status_code=400, detail="Авторизация не настроена")
    if not verify_credentials(body.name.strip(), body.key):
        raise HTTPException(status_code=401, detail="Неверное имя хаба или ключ")
    token = create_session_token()
    response = Response(content='{"status":"ok"}', media_type="application/json")
    response.headers["set-cookie"] = session_cookie_header(token, secure=request_is_secure(request))
    return response


@router.post("/api/auth/logout")
def logout(request: Request) -> Response:
    response = Response(content='{"status":"ok"}', media_type="application/json")
    response.headers["set-cookie"] = clear_session_cookie_header(secure=request_is_secure(request))
    return response


@router.get("/api/hub/info")
def hub_info() -> dict[str, Any]:
    if _state().mode != "hub":
        return {"mode": _state().mode, "public_url": "", "domain": ""}
    settings = load_hub_config().hub
    return {
        "mode": "hub",
        "domain": settings.domain,
        "public_url": resolve_public_url(settings),
        "use_https": settings.use_https,
    }


@router.get("/api/config/hub")
def get_hub_config() -> dict[str, Any]:
    config = load_hub_config()
    data = config.model_dump(mode="json")
    data["hub"]["public_url_resolved"] = resolve_public_url(config.hub)
    return data


@router.put("/api/config/hub")
def update_hub_config(body: HubUpdate) -> dict[str, Any]:
    if _state().mode != "hub":
        raise HTTPException(status_code=400, detail="Доступно только в hub-режиме")
    current = load_hub_config()
    data = current.model_dump(mode="json")
    data["hub"].update(body.hub or {})
    settings = HubSettings.model_validate(data["hub"])
    settings = settings.model_copy(update={"domain": normalize_domain(settings.domain)})
    if settings.domain:
        settings = settings.model_copy(
            update={"trusted_hosts": [settings.domain, f"*.{settings.domain}"]}
        )
    config = HubFile(hub=settings)
    save_hub_config(config)
    return {
        "status": "ok",
        "hub": {
            **config.model_dump(mode="json")["hub"],
            "public_url_resolved": resolve_public_url(settings),
        },
    }


@router.put("/api/config/sensors")
def update_sensors_config(body: SensorsUpdate) -> dict[str, str]:
    if _state().mode == "hub":
        raise HTTPException(status_code=400, detail="Датчики настраиваются на агентах")
    current = load_sensors_config()
    data = current.model_dump(mode="json")
    if body.settings is not None:
        data["settings"] = body.settings
    if body.sensors is not None:
        data["sensors"] = body.sensors
    config = SensorsFile.model_validate(data)
    save_sensors_config(config)
    collector = _state().collector
    if collector is not None:
        collector.reload_config()
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


@router.get("/api/config/agents")
def get_agents_config() -> dict[str, Any]:
    return load_agents_config().model_dump(mode="json")


@router.put("/api/config/agents")
def update_agents_config(body: AgentsUpdate) -> dict[str, Any]:
    incoming = AgentsFile.model_validate({"agents": body.agents})
    config = prepare_agents_for_save(incoming)
    save_agents_config(config)
    return {"status": "ok", "agents": config.model_dump(mode="json")["agents"]}


@router.get("/api/sensor-types")
def sensor_types() -> dict[str, list[str]]:
    from ..collector.registry import get_registered_types

    return {"types": get_registered_types()}


@router.websocket("/ws/live")
async def live_ws(websocket: WebSocket, agent: str | None = Query(default=None)) -> None:
    state = _state()
    if state.mode == "hub" and hub_auth_enabled():
        if not is_authenticated(websocket.headers, websocket.cookies):
            await websocket.close(code=1008, reason="Unauthorized")
            return
    await state.hub.connect(websocket)
    try:
        if agent:
            raw = state.fleet.get_agent_snapshot(agent)
            snapshot = {
                strip_agent_prefix(agent, key): value for key, value in raw.items()
            }
        elif state.collector is not None:
            snapshot = state.collector.get_latest_snapshot()
        elif state.mode == "hub":
            snapshot = state.fleet.get_combined_snapshot()
        else:
            snapshot = {}
        await websocket.send_json({"type": "snapshot", "data": snapshot, "agent_id": agent})
        while True:
            await websocket.receive_text()
    except WebSocketDisconnect:
        state.hub.disconnect(websocket)
