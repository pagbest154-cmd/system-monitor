from .base import AgentTransport
from .grpc import GrpcTransport
from .http import HttpTransport

__all__ = ["AgentTransport", "GrpcTransport", "HttpTransport"]


def create_transport(transport: str, hub_url: str, token: str) -> AgentTransport:
    if transport == "grpc":
        return GrpcTransport(hub_url, token)
    return HttpTransport(hub_url, token)
