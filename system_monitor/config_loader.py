from __future__ import annotations

import platform
import sys
from pathlib import Path
from typing import Any

import yaml
from pydantic import BaseModel, Field

from .paths import DASHBOARD_CONFIG, SENSORS_CONFIG


def current_platform() -> str:
    name = sys.platform
    if name.startswith("win"):
        return "windows"
    if name.startswith("linux"):
        return "linux"
    if name.startswith("darwin"):
        return "macos"
    return name


class SettingsConfig(BaseModel):
    retention_days: int = 7
    default_interval_sec: int = 5


class SensorConfig(BaseModel):
    id: str
    name: str
    type: str
    enabled: bool = True
    interval_sec: int | None = None
    unit: str = ""
    params: dict[str, Any] = Field(default_factory=dict)
    platforms: list[str] | None = None
    warn_above: float | None = None
    critical_above: float | None = None

    def is_supported_on_platform(self) -> bool:
        if not self.platforms:
            return True
        return current_platform() in self.platforms


class SensorsFile(BaseModel):
    settings: SettingsConfig = Field(default_factory=SettingsConfig)
    sensors: list[SensorConfig] = Field(default_factory=list)


class DashboardMeta(BaseModel):
    title: str = "Мониторинг системы"
    refresh_sec: int = 3


class PanelConfig(BaseModel):
    id: str
    title: str
    type: str
    sensors: list[str] = Field(default_factory=list)
    period: str = "1h"
    col: int = 1
    row: int = 1
    span: int = 1


class DashboardFile(BaseModel):
    dashboard: DashboardMeta = Field(default_factory=DashboardMeta)
    panels: list[PanelConfig] = Field(default_factory=list)


def _load_yaml(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}
    with path.open("r", encoding="utf-8") as fh:
        data = yaml.safe_load(fh) or {}
    if not isinstance(data, dict):
        raise ValueError(f"Конфиг {path} должен быть объектом YAML")
    return data


def _save_yaml(path: Path, data: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as fh:
        yaml.safe_dump(
            data,
            fh,
            allow_unicode=True,
            default_flow_style=False,
            sort_keys=False,
        )


def load_sensors_config(path: Path | None = None) -> SensorsFile:
    path = path or SENSORS_CONFIG
    return SensorsFile.model_validate(_load_yaml(path))


def save_sensors_config(config: SensorsFile, path: Path | None = None) -> None:
    path = path or SENSORS_CONFIG
    _save_yaml(path, config.model_dump(mode="json"))


def load_dashboard_config(path: Path | None = None) -> DashboardFile:
    path = path or DASHBOARD_CONFIG
    return DashboardFile.model_validate(_load_yaml(path))


def save_dashboard_config(config: DashboardFile, path: Path | None = None) -> None:
    path = path or DASHBOARD_CONFIG
    _save_yaml(path, config.model_dump(mode="json"))


def default_disk_path() -> str:
    if platform.system() == "Windows":
        return "C:\\"
    return "/"
