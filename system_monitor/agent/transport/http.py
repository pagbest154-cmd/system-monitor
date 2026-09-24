from __future__ import annotations

import httpx

from ...protocol.models import AgentConfigResponse, AgentReport
from .base import AgentTransport


class HttpTransport(AgentTransport):
    def __init__(self, hub_url: str, token: str, timeout: float = 10.0) -> None:
        self.hub_url = hub_url.rstrip("/")
        self.token = token
        self.timeout = timeout
        self._client = httpx.Client(timeout=timeout)

    def _headers(self) -> dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers

    def push_report(self, report: AgentReport) -> None:
        url = f"{self.hub_url}/api/agents/{report.agent_id}/metrics"
        response = self._client.post(url, json=report.model_dump(mode="json"), headers=self._headers())
        response.raise_for_status()

    def sync_config(self, agent_id: str, config_version: int) -> AgentConfigResponse:
        url = f"{self.hub_url}/api/agents/{agent_id}/config"
        response = self._client.post(
            url,
            params={"config_version": str(config_version)},
            headers=self._headers(),
        )
        response.raise_for_status()
        return AgentConfigResponse.model_validate(response.json())

    def close(self) -> None:
        self._client.close()
