from __future__ import annotations

import platform
import secrets
import sys
from pathlib import Path
from typing import Any

import yaml
from pydantic import BaseModel, Field

from .paths import (
    AGENT_CONFIG,
    AGENT_SENSORS_CONFIG,
    AGENTS_CONFIG,
    DASHBOARD_CONFIG,
    HUB_CONFIG,
    SENSORS_CONFIG,
)


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


class SensorOverrideConfig(BaseModel):
    sensor_id: str
    enabled: bool | None = None
    interval_sec: int | None = None
    warn_above: float | None = None
    critical_above: float | None = None
    params: dict[str, Any] | None = None


class AgentEntry(BaseModel):
    id: str
    name: str = ""
    token: str = ""
    overrides: list[SensorOverrideConfig] = Field(default_factory=list)


class AgentsFile(BaseModel):
    agents: list[AgentEntry] = Field(default_factory=list)


class HubSettings(BaseModel):
    domain: str = ""
    public_url: str = ""
    use_https: bool = True
    trusted_hosts: list[str] = Field(default_factory=lambda: ["*"])


class HubFile(BaseModel):
    hub: HubSettings = Field(default_factory=HubSettings)


class AgentFileConfig(BaseModel):
    hub_url: str = "http://127.0.0.1:8080"
    agent_id: str = ""
    token: str = ""
    token_file: str = ""
    interval_sec: int = 5
    transport: str = "http"


def load_agents_config(path: Path | None = None) -> AgentsFile:
    path = path or AGENTS_CONFIG
    return AgentsFile.model_validate(_load_yaml(path))


def save_agents_config(config: AgentsFile, path: Path | None = None) -> None:
    path = path or AGENTS_CONFIG
    _save_yaml(path, config.model_dump(mode="json"))


def normalize_domain(value: str) -> str:
    value = value.strip()
    if not value:
        return ""
    for prefix in ("https://", "http://"):
        if value.startswith(prefix):
            value = value[len(prefix) :]
    return value.split("/")[0].strip()


def resolve_public_url(settings: HubSettings) -> str:
    if settings.public_url.strip():
        return settings.public_url.strip().rstrip("/")
    domain = normalize_domain(settings.domain)
    if not domain:
        return ""
    scheme = "https" if settings.use_https else "http"
    return f"{scheme}://{domain}"


def load_hub_config(path: Path | None = None) -> HubFile:
    path = path or HUB_CONFIG
    return HubFile.model_validate(_load_yaml(path))


def save_hub_config(config: HubFile, path: Path | None = None) -> None:
    path = path or HUB_CONFIG
    _save_yaml(path, config.model_dump(mode="json"))


def hub_trusted_hosts(settings: HubSettings) -> list[str]:
    hosts = [item.strip() for item in settings.trusted_hosts if item.strip()]
    domain = normalize_domain(settings.domain)
    if domain:
        hosts.extend([domain, f"*.{domain}"])
    hosts = list(dict.fromkeys(hosts))
    return hosts or ["*"]


def load_agent_config(path: Path | None = None) -> AgentFileConfig:
    path = path or AGENT_CONFIG
    return AgentFileConfig.model_validate(_load_yaml(path))


def save_agent_config(config: AgentFileConfig, path: Path | None = None) -> None:
    path = path or AGENT_CONFIG
    _save_yaml(path, config.model_dump(mode="json"))


def load_agent_sensors_config(path: Path | None = None) -> SensorsFile:
    path = path or AGENT_SENSORS_CONFIG
    if not path.exists():
        return load_sensors_config(SENSORS_CONFIG)
    return SensorsFile.model_validate(_load_yaml(path))


def merge_agent_config(base: SensorsFile, overrides: list[SensorOverrideConfig]) -> SensorsFile:
    if not overrides:
        return base

    override_map = {item.sensor_id: item for item in overrides}
    merged_sensors: list[SensorConfig] = []
    for sensor in base.sensors:
        override = override_map.get(sensor.id)
        if override is None:
            merged_sensors.append(sensor)
            continue
        data = sensor.model_dump(mode="json")
        if override.enabled is not None:
            data["enabled"] = override.enabled
        if override.interval_sec is not None:
            data["interval_sec"] = override.interval_sec
        if override.warn_above is not None:
            data["warn_above"] = override.warn_above
        if override.critical_above is not None:
            data["critical_above"] = override.critical_above
        if override.params is not None:
            data["params"] = {**data.get("params", {}), **override.params}
        merged_sensors.append(SensorConfig.model_validate(data))
    return SensorsFile(settings=base.settings, sensors=merged_sensors)


def find_agent_entry(agent_id: str, config: AgentsFile | None = None) -> AgentEntry | None:
    config = config or load_agents_config()
    for entry in config.agents:
        if entry.id == agent_id:
            return entry
    return None


_PLACEHOLDER_TOKENS = frozenset({"", "change-me", "changeme"})


def generate_agent_token() -> str:
    return secrets.token_urlsafe(32)


def _needs_new_token(token: str) -> bool:
    return token.strip().lower() in _PLACEHOLDER_TOKENS


def prepare_agents_for_save(incoming: AgentsFile) -> AgentsFile:
    """Сохраняет существующие токены; для новых и пустых — автогенерация."""
    previous = load_agents_config()
    previous_map = {entry.id: entry for entry in previous.agents}
    prepared: list[AgentEntry] = []

    for agent in incoming.agents:
        token = agent.token.strip()
        if _needs_new_token(token):
            prev = previous_map.get(agent.id)
            if prev and not _needs_new_token(prev.token):
                token = prev.token
            else:
                token = generate_agent_token()
        prepared.append(agent.model_copy(update={"token": token}))

    return AgentsFile(agents=prepared)

