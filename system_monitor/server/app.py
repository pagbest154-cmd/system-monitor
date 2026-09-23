from __future__ import annotations

from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

from ..collector.runner import Collector
from ..paths import DB_PATH, WEB_DIR
from ..storage import MetricStore
from .routes import LiveHub, router, setup_state


def create_app() -> FastAPI:
    store = MetricStore(DB_PATH)
    hub = LiveHub()
    collector = Collector(store)
    setup_state(collector, store, hub)

    @asynccontextmanager
    async def lifespan(_app: FastAPI):
        collector.start()
        yield
        collector.stop()

    app = FastAPI(
        title="system-monitor",
        description="Лёгкая система мониторинга",
        lifespan=lifespan,
    )

    app.include_router(router)

    if WEB_DIR.exists():
        app.mount("/static", StaticFiles(directory=WEB_DIR), name="static")

        @app.get("/")
        def index() -> FileResponse:
            return FileResponse(WEB_DIR / "index.html")

        @app.get("/settings")
        def settings_page() -> FileResponse:
            return FileResponse(WEB_DIR / "settings.html")

    return app
