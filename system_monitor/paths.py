from __future__ import annotations

import os
import platform
import sys
from pathlib import Path

_PKG_DIR = Path(__file__).resolve().parent
_DEV_ROOT = _PKG_DIR.parent


def _is_frozen() -> bool:
    return getattr(sys, "frozen", False)


def _use_dev_layout() -> bool:
    return (_DEV_ROOT / "config").is_dir() and (_DEV_ROOT / "web").is_dir()


def _windows_program_files() -> Path:
    return Path(os.environ.get("ProgramFiles", r"C:\Program Files")) / "system-monitor-agent"


def _windows_program_data() -> Path:
    return Path(os.environ.get("ProgramData", r"C:\ProgramData")) / "system-monitor"


if _use_dev_layout():
    ROOT_DIR = _DEV_ROOT
    CONFIG_DIR = ROOT_DIR / "config"
    WEB_DIR = ROOT_DIR / "web"
    DATA_DIR = ROOT_DIR / "data"
elif _is_frozen():
    ROOT_DIR = Path(sys.executable).resolve().parent
    if platform.system() == "Windows":
        CONFIG_DIR = Path(
            os.environ.get("SYSTEM_MONITOR_CONFIG_DIR", str(_windows_program_data()))
        )
        DATA_DIR = Path(
            os.environ.get("SYSTEM_MONITOR_DATA_DIR", str(_windows_program_data() / "data"))
        )
    else:
        CONFIG_DIR = Path(os.environ.get("SYSTEM_MONITOR_CONFIG_DIR", "/etc/system-monitor"))
        DATA_DIR = Path(os.environ.get("SYSTEM_MONITOR_DATA_DIR", "/var/lib/system-monitor"))
    WEB_DIR = ROOT_DIR / "web"
elif platform.system() == "Windows":
    ROOT_DIR = Path(os.environ.get("SYSTEM_MONITOR_ROOT", str(_windows_program_files())))
    CONFIG_DIR = Path(os.environ.get("SYSTEM_MONITOR_CONFIG_DIR", str(_windows_program_data())))
    WEB_DIR = ROOT_DIR / "web"
    DATA_DIR = Path(os.environ.get("SYSTEM_MONITOR_DATA_DIR", str(_windows_program_data() / "data")))
else:
    ROOT_DIR = Path(os.environ.get("SYSTEM_MONITOR_ROOT", "/usr/share/system-monitor"))
    CONFIG_DIR = Path(os.environ.get("SYSTEM_MONITOR_CONFIG_DIR", "/etc/system-monitor"))
    WEB_DIR = ROOT_DIR / "web"
    DATA_DIR = Path(os.environ.get("SYSTEM_MONITOR_DATA_DIR", "/var/lib/system-monitor"))

SENSORS_CONFIG = CONFIG_DIR / "sensors.yaml"
DASHBOARD_CONFIG = CONFIG_DIR / "dashboard.yaml"
AGENTS_CONFIG = CONFIG_DIR / "agents.yaml"
AGENT_CONFIG = CONFIG_DIR / "agent.yaml"
AGENT_SENSORS_CONFIG = CONFIG_DIR / "agent_sensors.yaml"
AGENT_TOKEN_FILE = CONFIG_DIR / "agent.token"
AGENT_STATUS_FILE = CONFIG_DIR / "agent.status.json"
AGENT_UPDATE_CACHE = CONFIG_DIR / "agent.update.json"
HUB_CONFIG = CONFIG_DIR / "hub.yaml"
DB_PATH = DATA_DIR / "metrics.db"

if _use_dev_layout():
    _PKG_CONFIG_DIR = _DEV_ROOT / "config"
elif _is_frozen():
    _bundle_base = Path(getattr(sys, "_MEIPASS", ROOT_DIR))
    _PKG_CONFIG_DIR = _bundle_base / "config"
else:
    _PKG_CONFIG_DIR = ROOT_DIR / "config"
