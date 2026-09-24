from __future__ import annotations

import asyncio
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles
from starlette.middleware.trustedhost import TrustedHostMiddleware

from ..collector.runner import Collector
from ..config_loader import hub_trusted_hosts, load_hub_config, load_sensors_config
from ..paths import DB_PATH, WEB_DIR
from ..storage import MetricStore
from .auth import HubAuthMiddleware
from .routes import LiveHub, router, setup_state
from .state import FleetState


def create_app(mode: str = "standalone") -> FastAPI:
    store = MetricStore(DB_PATH)
    hub = LiveHub()
    fleet = FleetState()
    collector: Collector | None = None
    retention_days = 31

    if mode in ("standalone",):
        sensors = load_sensors_config()
        retention_days = sensors.settings.retention_days
        collector = Collector(store)

    setup_state(
        mode=mode,
        store=store,
        hub=hub,
        fleet=fleet,
        collector=collector,
        retention_days=retention_days,
    )

    @asynccontextmanager
    async def lifespan(app: FastAPI):
        hub.bind_loop(asyncio.get_running_loop())
        if collector is not None:
            collector.start()
        yield
        if collector is not None:
            collector.stop()

    app = FastAPI(
        title="system-monitor",
        description="Лёгкая система мониторинга",
        lifespan=lifespan,
    )

    if mode == "hub":
        trusted = hub_trusted_hosts(load_hub_config().hub)
        if trusted != ["*"]:
            app.add_middleware(TrustedHostMiddleware, allowed_hosts=trusted)
        app.add_middleware(HubAuthMiddleware, mode=mode)

    app.include_router(router)

    if WEB_DIR.exists():
        app.mount("/static", StaticFiles(directory=WEB_DIR), name="static")

        @app.get("/")
        def index() -> FileResponse:
            return FileResponse(WEB_DIR / "index.html")

        @app.get("/settings")
        def settings_page() -> FileResponse:
            return FileResponse(WEB_DIR / "settings.html")

        @app.get("/hosts")
        def hosts_page() -> FileResponse:
            return FileResponse(WEB_DIR / "hosts.html")

        @app.get("/login")
        def login_page() -> FileResponse:
            return FileResponse(WEB_DIR / "login.html")

    return app
