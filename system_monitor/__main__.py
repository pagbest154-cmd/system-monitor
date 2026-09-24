from __future__ import annotations

import argparse

import uvicorn

from .server import create_app


def main() -> None:
    parser = argparse.ArgumentParser(description="system-monitor — система мониторинга")
    parser.add_argument(
        "--mode",
        choices=("standalone", "hub"),
        default="standalone",
        help="standalone — локальный мониторинг; hub — центральный дашборд",
    )
    parser.add_argument("--host", default="127.0.0.1", help="Адрес сервера")
    parser.add_argument("--port", type=int, default=8080, help="Порт сервера")
    parser.add_argument(
        "--proxy-headers",
        action="store_true",
        help="Доверять X-Forwarded-* (за reverse proxy / доменом)",
    )
    args = parser.parse_args()

    app = create_app(mode=args.mode)
    print(f"system-monitor ({args.mode}) запущен: http://{args.host}:{args.port}")
    uvicorn.run(
        app,
        host=args.host,
        port=args.port,
        log_level="info",
        proxy_headers=args.proxy_headers or args.mode == "hub",
        forwarded_allow_ips="*",
    )


if __name__ == "__main__":
    main()
