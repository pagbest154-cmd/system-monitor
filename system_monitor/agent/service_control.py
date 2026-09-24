from __future__ import annotations

import subprocess
import sys
from pathlib import Path

SERVICE_NAME = "system-monitor-agent"


def _nssm_path() -> Path | None:
    if getattr(sys, "frozen", False):
        candidate = Path(sys.executable).resolve().parent / "nssm" / "nssm.exe"
        if candidate.exists():
            return candidate
    packaging_nssm = (
        Path(__file__).resolve().parents[2] / "packaging" / "windows" / "third_party" / "nssm" / "win64" / "nssm.exe"
    )
    if packaging_nssm.exists():
        return packaging_nssm
    return None


def get_service_state() -> str:
    result = subprocess.run(
        ["sc", "query", SERVICE_NAME],
        capture_output=True,
        text=True,
        timeout=10,
        creationflags=subprocess.CREATE_NO_WINDOW if sys.platform == "win32" else 0,
    )
    output = result.stdout or ""
    if "RUNNING" in output:
        return "running"
    if "STOPPED" in output:
        return "stopped"
    if "does not exist" in output.lower() or result.returncode != 0:
        return "missing"
    return "unknown"


def restart_service() -> tuple[bool, str]:
    nssm = _nssm_path()
    flags = subprocess.CREATE_NO_WINDOW if sys.platform == "win32" else 0
    if nssm is not None:
        result = subprocess.run(
            [str(nssm), "restart", SERVICE_NAME],
            capture_output=True,
            text=True,
            timeout=30,
            creationflags=flags,
        )
        if result.returncode == 0:
            return True, "Служба перезапущена"
        detail = (result.stderr or result.stdout or "").strip()
        return False, detail or f"Код выхода {result.returncode}"

    for command in (
        ["net", "stop", SERVICE_NAME],
        ["net", "start", SERVICE_NAME],
    ):
        result = subprocess.run(
            command,
            capture_output=True,
            text=True,
            timeout=30,
            creationflags=flags,
        )
        if result.returncode != 0:
            detail = (result.stderr or result.stdout or "").strip()
            return False, detail or f"Не удалось выполнить {' '.join(command)}"
    return True, "Служба перезапущена"
