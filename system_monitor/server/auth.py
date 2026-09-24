from __future__ import annotations

import base64
import hashlib
import hmac
import os
import re
import secrets
import time

from starlette.datastructures import Headers
from starlette.requests import Request
from starlette.responses import JSONResponse, RedirectResponse, Response
from starlette.types import ASGIApp, Receive, Scope, Send

SESSION_COOKIE = "sm_hub_session"
SESSION_TTL_SEC = 7 * 24 * 3600
_AGENT_PATH_RE = re.compile(r"^/api/agents/[^/]+/(metrics|heartbeat|config)$")


def hub_credentials() -> tuple[str, str] | None:
    name = os.environ.get("HUB_NAME", "").strip()
    key = os.environ.get("HUB_KEY", "").strip()
    if name and key:
        return name, key
    return None


def hub_auth_enabled() -> bool:
    return hub_credentials() is not None


def hub_name() -> str:
    creds = hub_credentials()
    return creds[0] if creds else ""


def verify_credentials(name: str, key: str) -> bool:
    creds = hub_credentials()
    if creds is None:
        return True
    return secrets.compare_digest(name, creds[0]) and secrets.compare_digest(key, creds[1])


def _session_secret() -> bytes:
    creds = hub_credentials()
    material = creds[1] if creds else "system-monitor-dev"
    return hashlib.sha256(material.encode("utf-8")).digest()


def create_session_token() -> str:
    expires = int(time.time()) + SESSION_TTL_SEC
    nonce = secrets.token_urlsafe(16)
    payload = f"{expires}.{nonce}"
    signature = hmac.new(_session_secret(), payload.encode("utf-8"), hashlib.sha256).hexdigest()
    raw = f"{payload}.{signature}".encode("utf-8")
    return base64.urlsafe_b64encode(raw).decode("ascii").rstrip("=")


def verify_session_token(token: str) -> bool:
    if not token:
        return False
    try:
        padding = "=" * (-len(token) % 4)
        raw = base64.urlsafe_b64decode(token + padding).decode("utf-8")
        expires_s, nonce, signature = raw.rsplit(".", 2)
        payload = f"{expires_s}.{nonce}"
        expected = hmac.new(_session_secret(), payload.encode("utf-8"), hashlib.sha256).hexdigest()
        if not secrets.compare_digest(signature, expected):
            return False
        return int(expires_s) > time.time()
    except Exception:
        return False


def parse_basic_auth(authorization: str | None) -> tuple[str, str] | None:
    if not authorization or not authorization.lower().startswith("basic "):
        return None
    try:
        decoded = base64.b64decode(authorization[6:].strip()).decode("utf-8")
    except Exception:
        return None
    username, sep, password = decoded.partition(":")
    if not sep:
        return None
    return username, password


def cookies_from_headers(headers: Headers) -> dict[str, str]:
    cookies: dict[str, str] = {}
    raw = headers.get("cookie")
    if not raw:
        return cookies
    for part in raw.split(";"):
        part = part.strip()
        if "=" in part:
            key, value = part.split("=", 1)
            cookies[key.strip()] = value.strip()
    return cookies


def is_authenticated(headers: Headers, cookies: dict[str, str]) -> bool:
    if not hub_auth_enabled():
        return True
    token = cookies.get(SESSION_COOKIE)
    if token and verify_session_token(token):
        return True
    basic = parse_basic_auth(headers.get("authorization"))
    if basic and verify_credentials(basic[0], basic[1]):
        return True
    return False


def is_public_path(path: str, method: str) -> bool:
    if path.startswith("/static/"):
        return True
    if path == "/login":
        return True
    if path == "/api/auth/login" and method.upper() == "POST":
        return True
    if path == "/api/auth/status" and method.upper() == "GET":
        return True
    if method.upper() == "POST" and _AGENT_PATH_RE.match(path):
        return True
    return False


def session_cookie_header(token: str, secure: bool = False) -> str:
    parts = [
        f"{SESSION_COOKIE}={token}",
        "Path=/",
        "HttpOnly",
        "SameSite=Lax",
        f"Max-Age={SESSION_TTL_SEC}",
    ]
    if secure:
        parts.append("Secure")
    return "; ".join(parts)


def clear_session_cookie_header(secure: bool = False) -> str:
    parts = [f"{SESSION_COOKIE}=", "Path=/", "HttpOnly", "SameSite=Lax", "Max-Age=0"]
    if secure:
        parts.append("Secure")
    return "; ".join(parts)


def request_is_secure(request: Request) -> bool:
    forwarded = request.headers.get("x-forwarded-proto", "")
    if forwarded:
        return forwarded.split(",")[0].strip().lower() == "https"
    return request.url.scheme == "https"


class HubAuthMiddleware:
    def __init__(self, app: ASGIApp, mode: str) -> None:
        self.app = app
        self.mode = mode

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        if self.mode != "hub" or not hub_auth_enabled():
            await self.app(scope, receive, send)
            return

        path = scope.get("path", "")
        method = scope.get("method", "GET")
        if is_public_path(path, method):
            await self.app(scope, receive, send)
            return

        headers = Headers(scope=scope)
        if is_authenticated(headers, cookies_from_headers(headers)):
            await self.app(scope, receive, send)
            return

        if path.startswith("/api/"):
            response: Response = JSONResponse({"detail": "Требуется авторизация"}, status_code=401)
        else:
            response = RedirectResponse(url="/login", status_code=302)
        await response(scope, receive, send)
