from __future__ import annotations

import json
import platform
import re
import subprocess
from pathlib import Path

import psutil

from .base import Sensor, SensorReading

_CPU_NAME_RE = re.compile(r"cpu|core|package|tctl|processor|xeon|ryzen", re.I)
_GPU_NAME_RE = re.compile(r"gpu|graphics|nvidia|geforce|radeon|video", re.I)


def _run_powershell_json(script: str, timeout: float = 8.0) -> list[dict]:
    try:
        result = subprocess.run(
            ["powershell", "-NoProfile", "-Command", script],
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False,
        )
        if result.returncode != 0 or not result.stdout.strip():
            return []
        data = json.loads(result.stdout)
        if isinstance(data, dict):
            return [data]
        if isinstance(data, list):
            return [item for item in data if isinstance(item, dict)]
    except (OSError, subprocess.TimeoutExpired, json.JSONDecodeError, ValueError):
        return []
    return []


def _read_linux_thermal() -> float | None:
    thermal_dir = Path("/sys/class/thermal")
    temps: list[float] = []
    if thermal_dir.exists():
        for zone in sorted(thermal_dir.glob("thermal_zone*")):
            temp_file = zone / "temp"
            if not temp_file.exists():
                continue
            try:
                raw = int(temp_file.read_text(encoding="utf-8").strip())
                temps.append(raw / 1000.0)
            except (OSError, ValueError):
                continue

    if not temps:
        try:
            sensors = psutil.sensors_temperatures()
        except (AttributeError, OSError):
            sensors = {}
        for entries in sensors.values():
            for entry in entries:
                if entry.current is not None:
                    temps.append(float(entry.current))

    return max(temps) if temps else None


def _read_hw_monitor(namespace: str) -> list[tuple[str, float]]:
    script = (
        f"Get-CimInstance -Namespace {namespace} -ClassName Sensor "
        "-ErrorAction SilentlyContinue | "
        "Where-Object { $_.SensorType -eq 'Temperature' -and $_.Value -ne $null } | "
        "Select-Object Name, Value | ConvertTo-Json -Compress"
    )
    readings: list[tuple[str, float]] = []
    for item in _run_powershell_json(script):
        name = str(item.get("Name") or "")
        try:
            value = float(item.get("Value"))
        except (TypeError, ValueError):
            continue
        if 0 < value < 150:
            readings.append((name, value))
    return readings


def _pick_cpu_temp(readings: list[tuple[str, float]]) -> float | None:
    cpu_values = [value for name, value in readings if _CPU_NAME_RE.search(name)]
    if cpu_values:
        return max(cpu_values)
    if readings:
        return max(value for _, value in readings)
    return None


def _pick_gpu_temp(readings: list[tuple[str, float]]) -> float | None:
    gpu_values = [value for name, value in readings if _GPU_NAME_RE.search(name)]
    if gpu_values:
        return max(gpu_values)
    return None


def _read_windows_acpi() -> float | None:
    try:
        import wmi  # type: ignore[import-untyped]

        client = wmi.WMI(namespace="root\\wmi")
        temps: list[float] = []
        for item in client.MSAcpi_ThermalZoneTemperature():
            if item.CurrentTemperature is not None:
                celsius = (item.CurrentTemperature / 10.0) - 273.15
                if 0 < celsius < 150:
                    temps.append(float(celsius))
        return max(temps) if temps else None
    except Exception:
        script = (
            "Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature "
            "| Select-Object -ExpandProperty CurrentTemperature"
        )
        try:
            result = subprocess.run(
                ["powershell", "-NoProfile", "-Command", script],
                capture_output=True,
                text=True,
                timeout=5,
                check=False,
            )
            values = []
            for line in result.stdout.splitlines():
                line = line.strip()
                if not line:
                    continue
                try:
                    raw = float(line)
                    celsius = (raw / 10.0) - 273.15
                    if 0 < celsius < 150:
                        values.append(celsius)
                except ValueError:
                    continue
            return max(values) if values else None
        except (OSError, subprocess.TimeoutExpired):
            return None


def _read_windows_perf_thermal() -> float | None:
    script = (
        "Get-CimInstance Win32_PerfFormattedData_Counters_ThermalZoneInformation "
        "-ErrorAction SilentlyContinue | "
        "Where-Object { $_.Temperature -gt 0 } | "
        "Select-Object -ExpandProperty Temperature"
    )
    try:
        result = subprocess.run(
            ["powershell", "-NoProfile", "-Command", script],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        values = []
        for line in result.stdout.splitlines():
            line = line.strip()
            if not line:
                continue
            try:
                value = float(line)
                if 0 < value < 150:
                    values.append(value)
            except ValueError:
                continue
        return max(values) if values else None
    except (OSError, subprocess.TimeoutExpired):
        return None


def _read_windows_cpu_thermal() -> float | None:
    for reader in (
        _read_windows_acpi,
        lambda: _pick_cpu_temp(_read_hw_monitor("root/LibreHardwareMonitor")),
        lambda: _pick_cpu_temp(_read_hw_monitor("root/OpenHardwareMonitor")),
        _read_windows_perf_thermal,
    ):
        value = reader()
        if value is not None:
            return value
    return None


def _read_nvidia_gpu_temp() -> float | None:
    try:
        result = subprocess.run(
            [
                "nvidia-smi",
                "--query-gpu=temperature.gpu",
                "--format=csv,noheader,nounits",
            ],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        values = []
        for line in result.stdout.splitlines():
            line = line.strip()
            if not line or line in {"[N/A]", "N/A"}:
                continue
            try:
                value = float(line)
                if 0 < value < 150:
                    values.append(value)
            except ValueError:
                continue
        return max(values) if values else None
    except (OSError, subprocess.TimeoutExpired):
        return None


def _read_windows_gpu_thermal() -> float | None:
    value = _read_nvidia_gpu_temp()
    if value is not None:
        return value
    return _pick_gpu_temp(_read_hw_monitor("root/LibreHardwareMonitor")) or _pick_gpu_temp(
        _read_hw_monitor("root/OpenHardwareMonitor")
    )


class TemperatureSensor(Sensor):
    def read(self) -> SensorReading:
        system = platform.system()
        if system == "Linux":
            value = _read_linux_thermal()
        elif system == "Windows":
            value = _read_windows_cpu_thermal()
        else:
            try:
                sensors = psutil.sensors_temperatures()
            except (AttributeError, OSError):
                sensors = {}
            temps: list[float] = []
            for entries in sensors.values():
                for entry in entries:
                    if entry.current is not None:
                        temps.append(float(entry.current))
            value = max(temps) if temps else None

        if value is None:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="unknown",
                error="Температура CPU недоступна (нужен LibreHardwareMonitor или датчики ACPI)",
            )

        value = round(value, 1)
        return SensorReading(
            sensor_id=self.id,
            value=value,
            status=self.evaluate_status(value),
        )


class GpuTemperatureSensor(Sensor):
    def read(self) -> SensorReading:
        system = platform.system()
        if system == "Windows":
            value = _read_windows_gpu_thermal()
        elif system == "Linux":
            value = _read_nvidia_gpu_temp() or _pick_gpu_temp(
                _read_hw_monitor("root/LibreHardwareMonitor")
            )
        else:
            value = _read_nvidia_gpu_temp()

        if value is None:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="unknown",
                error="Температура GPU недоступна",
            )

        value = round(value, 1)
        return SensorReading(
            sensor_id=self.id,
            value=value,
            status=self.evaluate_status(value),
        )
