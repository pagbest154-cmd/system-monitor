from __future__ import annotations

import re

import psutil

from .config_loader import SensorConfig

_IGNORE_FSTYPES = frozenset(
    {
        "tmpfs",
        "devtmpfs",
        "squashfs",
        "overlay",
        "proc",
        "sysfs",
        "devfs",
        "autofs",
        "cgroup",
        "cgroup2",
        "pstore",
        "bpf",
        "tracefs",
        "debugfs",
        "securityfs",
        "configfs",
        "fusectl",
        "mqueue",
        "hugetlbfs",
    }
)


def _mount_to_sensor_id(mountpoint: str) -> str:
    cleaned = mountpoint.strip().rstrip("\\/").lower()
    if not cleaned:
        return "disk_auto_root"
    cleaned = cleaned.replace(":", "").replace("\\", "_").replace("/", "_")
    cleaned = re.sub(r"[^a-z0-9_]+", "_", cleaned)
    cleaned = re.sub(r"_+", "_", cleaned).strip("_")
    return f"disk_auto_{cleaned or 'root'}"


def discover_disk_sensors() -> list[SensorConfig]:
    import platform

    sensors: list[SensorConfig] = []
    seen_ids: set[str] = set()
    all_partitions = platform.system() != "Windows"

    for part in psutil.disk_partitions(all=all_partitions):
        if part.fstype in _IGNORE_FSTYPES:
            continue
        try:
            psutil.disk_usage(part.mountpoint)
        except (OSError, PermissionError):
            continue

        sensor_id = _mount_to_sensor_id(part.mountpoint)
        if sensor_id in seen_ids:
            suffix = 2
            while f"{sensor_id}_{suffix}" in seen_ids:
                suffix += 1
            sensor_id = f"{sensor_id}_{suffix}"
        seen_ids.add(sensor_id)

        label = part.mountpoint
        if part.device:
            label = f"{part.mountpoint} ({part.device})"

        sensors.append(
            SensorConfig(
                id=sensor_id,
                name=f"Диск {label}",
                type="system.disk_usage",
                enabled=True,
                interval_sec=30,
                unit="%",
                params={"path": part.mountpoint},
                warn_above=85,
                critical_above=95,
            )
        )

    return sensors
