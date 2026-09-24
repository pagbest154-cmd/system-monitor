from __future__ import annotations

"""gRPC-сервис hub (фаза 2): точка расширения. HTTP ingest — основной путь."""

def start_grpc_server(_host: str = "0.0.0.0", _port: int = 9080) -> None:
    try:
        import grpc  # noqa: F401
    except ImportError:
        return
    # Полная реализация — после codegen из system_monitor/proto/agent.proto
