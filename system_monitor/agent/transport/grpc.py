from __future__ import annotations

from ...protocol.models import AgentConfigResponse, AgentReport
from .base import AgentTransport
from .http import HttpTransport


class GrpcTransport(AgentTransport):
    """gRPC-транспорт: пока делегирует HTTP (fallback). Точка расширения для grpcio."""

    def __init__(self, hub_url: str, token: str, timeout: float = 10.0) -> None:
        self._http = HttpTransport(hub_url, token, timeout=timeout)
        self._grpc_enabled = False
        try:
            import grpc  # noqa: F401

            self._grpc_enabled = True
        except ImportError:
            self._grpc_enabled = False

    def push_report(self, report: AgentReport) -> None:
        self._http.push_report(report)

    def sync_config(self, agent_id: str, config_version: int) -> AgentConfigResponse:
        return self._http.sync_config(agent_id, config_version)

    def close(self) -> None:
        self._http.close()

    @property
    def grpc_available(self) -> bool:
        return self._grpc_enabled
