from __future__ import annotations

import os
from pathlib import Path

_PKG_DIR = Path(__file__).resolve().parent
_DEV_ROOT = _PKG_DIR.parent


def _use_dev_layout() -> bool:
    return (_DEV_ROOT / "config").is_dir() and (_DEV_ROOT / "web").is_dir()


if _use_dev_layout():
    ROOT_DIR = _DEV_ROOT
    CONFIG_DIR = ROOT_DIR / "config"
    WEB_DIR = ROOT_DIR / "web"
    DATA_DIR = ROOT_DIR / "data"
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
HUB_CONFIG = CONFIG_DIR / "hub.yaml"
DB_PATH = DATA_DIR / "metrics.db"

# Bundled defaults shipped with the agent deb
_PKG_CONFIG_DIR = _PKG_DIR.parent / "config" if _use_dev_layout() else ROOT_DIR / "config"
