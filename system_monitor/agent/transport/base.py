from __future__ import annotations

from abc import ABC, abstractmethod

from ...protocol.models import AgentConfigResponse, AgentReport


class AgentTransport(ABC):
    @abstractmethod
    def push_report(self, report: AgentReport) -> None:
        raise NotImplementedError

    @abstractmethod
    def sync_config(self, agent_id: str, config_version: int) -> AgentConfigResponse:
        raise NotImplementedError

    def close(self) -> None:
        return None
