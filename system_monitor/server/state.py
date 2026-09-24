from __future__ import annotations

import asyncio
import threading
from typing import Any


class LiveHub:
    def __init__(self) -> None:
        self.connections: list[Any] = []
        self._loop: asyncio.AbstractEventLoop | None = None

    def bind_loop(self, loop: asyncio.AbstractEventLoop) -> None:
        self._loop = loop

    async def connect(self, websocket: Any) -> None:
        await websocket.accept()
        self.connections.append(websocket)

    def disconnect(self, websocket: Any) -> None:
        if websocket in self.connections:
            self.connections.remove(websocket)

    async def broadcast(self, payload: dict[str, Any]) -> None:
        dead: list[Any] = []
        for ws in self.connections:
            try:
                await ws.send_json(payload)
            except Exception:
                dead.append(ws)
        for ws in dead:
            self.disconnect(ws)

    def schedule_broadcast(self, payload: dict[str, Any]) -> None:
        if self._loop is None or not self.connections:
            return
        asyncio.run_coroutine_threadsafe(self.broadcast(payload), self._loop)


class FleetState:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._latest_by_agent: dict[str, dict[str, dict[str, Any]]] = {}

    def update_agent(self, agent_id: str, latest: dict[str, dict[str, Any]]) -> None:
        with self._lock:
            current = self._latest_by_agent.setdefault(agent_id, {})
            current.update(latest)

    def get_agent_snapshot(self, agent_id: str) -> dict[str, dict[str, Any]]:
        with self._lock:
            return dict(self._latest_by_agent.get(agent_id, {}))

    def get_combined_snapshot(self) -> dict[str, dict[str, Any]]:
        with self._lock:
            combined: dict[str, dict[str, Any]] = {}
            for agent_latest in self._latest_by_agent.values():
                combined.update(agent_latest)
            return combined
