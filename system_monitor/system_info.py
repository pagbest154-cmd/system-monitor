from __future__ import annotations

import json
import platform
import re
import socket
import subprocess
import sys
import time
from pathlib import Path
from typing import Any

import psutil

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


def _bytes_to_gb(value: int | float | None) -> float | None:
    if value is None:
        return None
    return round(float(value) / (1024 ** 3), 1)


def _run_command(args: list[str], timeout: float = 5.0) -> str | None:
    try:
        output = subprocess.check_output(
            args,
            text=True,
            timeout=timeout,
            stderr=subprocess.DEVNULL,
        )
        return output.strip()
    except (OSError, subprocess.SubprocessError):
        return None


def _run_powershell(script: str, timeout: float = 8.0) -> Any | None:
    output = _run_command(
        ["powershell", "-NoProfile", "-Command", script],
        timeout=timeout,
    )
    if not output:
        return None
    try:
        return json.loads(output)
    except json.JSONDecodeError:
        return None


def _normalize_media_type(raw: str | None) -> str:
    if not raw:
        return "unknown"
    value = raw.strip().upper()
    if value in {"SSD", "HDD", "NVME", "UNKNOWN", "UNSPECIFIED"}:
        return "NVMe" if value == "NVME" else value.title() if value != "UNKNOWN" else "unknown"
    if "NVME" in value or "NVMe" in raw:
        return "NVMe"
    if "SSD" in value or "SOLID" in value:
        return "SSD"
    if "HDD" in value or "HARD" in value or value == "UNSPECIFIED":
        return "HDD" if value != "UNSPECIFIED" else "unknown"
    return raw.strip()


def _get_cpu_name() -> str:
    system = platform.system()
    if system == "Windows":
        try:
            import winreg

            key = winreg.OpenKey(
                winreg.HKEY_LOCAL_MACHINE,
                r"HARDWARE\DESCRIPTION\System\CentralProcessor\0",
            )
            name, _ = winreg.QueryValueEx(key, "ProcessorNameString")
            return str(name).strip()
        except OSError:
            pass
    elif system == "Linux":
        try:
            with open("/proc/cpuinfo", encoding="utf-8") as fh:
                for line in fh:
                    if line.lower().startswith("model name"):
                        return line.split(":", 1)[1].strip()
        except OSError:
            pass
    elif system == "Darwin":
        output = _run_command(["sysctl", "-n", "machdep.cpu.brand_string"], timeout=2)
        if output:
            return output

    return platform.processor() or "Неизвестный процессор"


def _get_os_label() -> str:
    system = platform.system()
    release = platform.release()
    if system == "Windows":
        return f"Windows {release}"
    if system == "Linux":
        return f"Linux {release}"
    if system == "Darwin":
        return f"macOS {platform.mac_ver()[0] or release}"
    return f"{system} {release}"


def _linux_rotational_map() -> dict[str, bool]:
    mapping: dict[str, bool] = {}
    block_dir = Path("/sys/block")
    if not block_dir.is_dir():
        return mapping
    for entry in block_dir.iterdir():
        rotational_path = entry / "queue" / "rotational"
        try:
            rotational = rotational_path.read_text(encoding="utf-8").strip()
            mapping[entry.name] = rotational == "1"
        except OSError:
            continue
    return mapping


def _partition_media_type(device: str, rotational: dict[str, bool]) -> str | None:
    match = re.search(r"(?:/dev/)?(nvme\d+n\d+|sd[a-z]|vd[a-z]|hd[a-z]|mmcblk\d+)", device)
    if not match:
        return None
    block = match.group(1)
    if block.startswith("nvme"):
        return "NVMe"
    if block.startswith("mmcblk"):
        return "SSD"
    is_hdd = rotational.get(block)
    if is_hdd is True:
        return "HDD"
    if is_hdd is False:
        return "SSD"
    return None


def _get_physical_drives() -> list[dict[str, Any]]:
    system = platform.system()
    drives: list[dict[str, Any]] = []

    if system == "Windows":
        data = _run_powershell(
            "Get-PhysicalDisk | Select-Object DeviceId, FriendlyName, MediaType, Size, "
            "HealthStatus, BusType, SerialNumber | ConvertTo-Json -Compress"
        )
        items = data if isinstance(data, list) else ([data] if data else [])
        for item in items:
            if not isinstance(item, dict):
                continue
            drives.append(
                {
                    "id": item.get("DeviceId"),
                    "name": item.get("FriendlyName") or f"Диск {item.get('DeviceId', '?')}",
                    "model": item.get("FriendlyName"),
                    "size_gb": _bytes_to_gb(item.get("Size")),
                    "media_type": _normalize_media_type(str(item.get("MediaType") or "")),
                    "interface": str(item.get("BusType") or ""),
                    "health": str(item.get("HealthStatus") or ""),
                    "serial": str(item.get("SerialNumber") or ""),
                }
            )
        if drives:
            return drives

        fallback = _run_powershell(
            "Get-CimInstance Win32_DiskDrive | Select-Object Index, Model, Size, "
            "InterfaceType, MediaType, SerialNumber | ConvertTo-Json -Compress"
        )
        items = fallback if isinstance(fallback, list) else ([fallback] if fallback else [])
        for item in items:
            if not isinstance(item, dict):
                continue
            drives.append(
                {
                    "id": item.get("Index"),
                    "name": item.get("Model") or f"Диск {item.get('Index', '?')}",
                    "model": item.get("Model"),
                    "size_gb": _bytes_to_gb(item.get("Size")),
                    "media_type": _normalize_media_type(str(item.get("MediaType") or "")),
                    "interface": str(item.get("InterfaceType") or ""),
                    "health": "",
                    "serial": str(item.get("SerialNumber") or "").strip(),
                }
            )
        return drives

    if system == "Linux":
        rotational = _linux_rotational_map()
        lsblk_json = _run_command(["lsblk", "-d", "-b", "-o", "NAME,MODEL,SIZE,ROTA,TYPE,TRAN", "-J"])
        if lsblk_json:
            try:
                payload = json.loads(lsblk_json)
                for block in payload.get("blockdevices", []):
                    if block.get("type") != "disk":
                        continue
                    name = block.get("name", "")
                    rota = block.get("rota")
                    if rota is True:
                        media = "HDD"
                    elif rota is False:
                        media = "NVMe" if str(block.get("tran") or "").lower() == "nvme" else "SSD"
                    else:
                        media = _partition_media_type(name, rotational) or "unknown"
                    drives.append(
                        {
                            "id": name,
                            "name": block.get("model") or name,
                            "model": block.get("model") or name,
                            "size_gb": _bytes_to_gb(block.get("size")),
                            "media_type": media,
                            "interface": str(block.get("tran") or ""),
                            "health": "",
                            "serial": "",
                        }
                    )
                if drives:
                    return drives
            except json.JSONDecodeError:
                pass

        for name, is_hdd in rotational.items():
            size_path = Path(f"/sys/block/{name}/size")
            try:
                sectors = int(size_path.read_text(encoding="utf-8").strip())
                size_bytes = sectors * 512
            except (OSError, ValueError):
                size_bytes = None
            model_path = Path(f"/sys/block/{name}/device/model")
            model = ""
            try:
                model = model_path.read_text(encoding="utf-8").strip()
            except OSError:
                model = name
            media = "HDD" if is_hdd else ("NVMe" if name.startswith("nvme") else "SSD")
            drives.append(
                {
                    "id": name,
                    "name": model or name,
                    "model": model or name,
                    "size_gb": _bytes_to_gb(size_bytes),
                    "media_type": media,
                    "interface": "",
                    "health": "",
                    "serial": "",
                }
            )
        return drives

    if system == "Darwin":
        output = _run_command(["diskutil", "list"], timeout=6)
        if not output:
            return drives
        current: dict[str, Any] | None = None
        for line in output.splitlines():
            if line.startswith("/dev/"):
                parts = line.split()
                if len(parts) >= 3:
                    current = {
                        "id": parts[0].replace("/dev/", ""),
                        "name": " ".join(parts[2:]),
                        "model": " ".join(parts[2:]),
                        "size_gb": None,
                        "media_type": "SSD" if "solid state" in line.lower() else "unknown",
                        "interface": "",
                        "health": "",
                        "serial": "",
                    }
                    drives.append(current)
            elif current and "*:" in line:
                size_match = re.search(r"\(([^)]+)\)", line)
                if size_match:
                    current["size_gb"] = _parse_size_label(size_match.group(1))
        return drives

    return drives


def _parse_size_label(label: str) -> float | None:
    match = re.match(r"([\d.]+)\s*([KMGT]?i?B?)", label.strip(), re.IGNORECASE)
    if not match:
        return None
    value = float(match.group(1))
    unit = match.group(2).upper()
    multipliers = {
        "B": 1,
        "KB": 1024,
        "KIB": 1024,
        "MB": 1024 ** 2,
        "MIB": 1024 ** 2,
        "GB": 1024 ** 3,
        "GIB": 1024 ** 3,
        "TB": 1024 ** 4,
        "TIB": 1024 ** 4,
    }
    for key, mult in multipliers.items():
        if unit.startswith(key.rstrip("B")):
            return round(value * mult / (1024 ** 3), 1)
    return None


def _get_partitions() -> list[dict[str, Any]]:
    system = platform.system()
    rotational = _linux_rotational_map() if system == "Linux" else {}
    partitions: list[dict[str, Any]] = []
    seen_mounts: set[str] = set()

    for part in psutil.disk_partitions(all=(system != "Windows")):
        mount = part.mountpoint
        if mount in seen_mounts:
            continue
        if part.fstype in _IGNORE_FSTYPES:
            continue
        try:
            usage = psutil.disk_usage(mount)
        except (OSError, PermissionError):
            continue
        seen_mounts.add(mount)
        media_type = _partition_media_type(part.device, rotational)
        partitions.append(
            {
                "device": part.device,
                "mountpoint": mount,
                "fstype": part.fstype,
                "opts": part.opts,
                "media_type": media_type,
                "total_gb": _bytes_to_gb(usage.total),
                "used_gb": _bytes_to_gb(usage.used),
                "free_gb": _bytes_to_gb(usage.free),
                "percent": round(usage.percent, 1),
            }
        )

    if system != "Windows":
        return partitions

    win_parts = _run_powershell(
        "Get-Partition | Where-Object { $_.DriveLetter } | "
        "Select-Object DiskNumber, DriveLetter, Size, Type, Guid | ConvertTo-Json -Compress"
    )
    if not win_parts:
        return partitions

    items = win_parts if isinstance(win_parts, list) else [win_parts]
    letter_map = {str(item.get("DriveLetter")).upper(): item for item in items if item.get("DriveLetter")}
    for entry in partitions:
        letter = entry["mountpoint"].rstrip("\\").replace(":", "").upper()
        if letter in letter_map:
            meta = letter_map[letter]
            entry["partition_type"] = str(meta.get("Type") or "")
            entry["disk_number"] = meta.get("DiskNumber")
            entry["size_gb"] = _bytes_to_gb(meta.get("Size")) or entry["total_gb"]

    return partitions


def _get_cuda_versions() -> dict[str, str | None]:
    cuda_version: str | None = None
    cuda_toolkit_version: str | None = None

    smi_output = _run_command(["nvidia-smi"], timeout=5)
    if smi_output:
        match = re.search(r"CUDA Version:\s*([\d.]+)", smi_output)
        if match:
            cuda_version = match.group(1)

    nvcc_output = _run_command(["nvcc", "--version"], timeout=5)
    if nvcc_output:
        match = re.search(r"release\s+([\d.]+)", nvcc_output, re.IGNORECASE)
        if match:
            cuda_toolkit_version = match.group(1)

    return {
        "cuda_version": cuda_version,
        "cuda_toolkit_version": cuda_toolkit_version,
    }


def _nvidia_gpu_stats() -> list[dict[str, Any]]:
    output = _run_command(
        [
            "nvidia-smi",
            "--query-gpu=name,memory.total,memory.used,memory.free,utilization.gpu,temperature.gpu,driver_version",
            "--format=csv,noheader,nounits",
        ],
        timeout=5,
    )
    if not output:
        return []

    cuda = _get_cuda_versions()
    gpus: list[dict[str, Any]] = []
    for line in output.splitlines():
        parts = [part.strip() for part in line.split(",")]
        if len(parts) < 7:
            continue
        total_mb = float(parts[1])
        used_mb = float(parts[2])
        gpus.append(
            {
                "name": parts[0],
                "vendor": "NVIDIA",
                "driver_version": parts[6],
                "cuda_version": cuda["cuda_version"],
                "cuda_toolkit_version": cuda["cuda_toolkit_version"],
                "memory_total_gb": round(total_mb / 1024, 1),
                "memory_used_gb": round(used_mb / 1024, 1),
                "memory_free_gb": round(float(parts[3]) / 1024, 1),
                "memory_percent": round(used_mb / total_mb * 100, 1) if total_mb else None,
                "utilization_percent": float(parts[4]),
                "temperature_c": float(parts[5]) if parts[5] not in {"[N/A]", "N/A"} else None,
                "resolution": None,
            }
        )
    return gpus


def _get_gpus() -> list[dict[str, Any]]:
    nvidia = _nvidia_gpu_stats()
    if nvidia:
        return nvidia

    system = platform.system()
    gpus: list[dict[str, Any]] = []

    if system == "Windows":
        data = _run_powershell(
            "Get-CimInstance Win32_VideoController | "
            "Select-Object Name, AdapterRAM, DriverVersion, VideoProcessor, "
            "CurrentHorizontalResolution, CurrentVerticalResolution, VideoModeDescription | "
            "ConvertTo-Json -Compress"
        )
        items = data if isinstance(data, list) else ([data] if data else [])
        for item in items:
            if not isinstance(item, dict):
                continue
            name = str(item.get("Name") or "GPU")
            if "microsoft" in name.lower() and "basic" in name.lower():
                continue
            width = item.get("CurrentHorizontalResolution")
            height = item.get("CurrentVerticalResolution")
            resolution = f"{width}x{height}" if width and height else None
            gpus.append(
                {
                    "name": name,
                    "vendor": name.split()[0] if name else "",
                    "driver_version": str(item.get("DriverVersion") or ""),
                    "memory_total_gb": _bytes_to_gb(item.get("AdapterRAM")),
                    "memory_used_gb": None,
                    "memory_free_gb": None,
                    "memory_percent": None,
                    "utilization_percent": None,
                    "temperature_c": None,
                    "resolution": resolution,
                    "video_processor": str(item.get("VideoProcessor") or ""),
                }
            )
        return gpus

    if system == "Linux":
        drm_cards = sorted(Path("/sys/class/drm").glob("card*/device/vendor"))
        for card_path in drm_cards:
            device_dir = card_path.parent
            vendor_id = _read_text(card_path)
            device_id = _read_text(device_dir / "device")
            name = _pci_device_name(vendor_id, device_id) or device_dir.name
            gpus.append(
                {
                    "name": name,
                    "vendor": _pci_vendor_name(vendor_id),
                    "driver_version": "",
                    "memory_total_gb": None,
                    "memory_used_gb": None,
                    "memory_free_gb": None,
                    "memory_percent": None,
                    "utilization_percent": None,
                    "temperature_c": None,
                    "resolution": None,
                }
            )
        if gpus:
            return gpus

    if system == "Darwin":
        output = _run_command(["system_profiler", "SPDisplaysDataType", "-json"], timeout=8)
        if output:
            try:
                payload = json.loads(output)
                displays = payload.get("SPDisplaysDataType", [])
                for item in displays:
                    vram = item.get("sppci_vram") or item.get("vram")
                    gpus.append(
                        {
                            "name": item.get("_name") or item.get("sppci_model") or "GPU",
                            "vendor": "",
                            "driver_version": "",
                            "memory_total_gb": _parse_size_label(str(vram)) if vram else None,
                            "memory_used_gb": None,
                            "memory_free_gb": None,
                            "memory_percent": None,
                            "utilization_percent": None,
                            "temperature_c": None,
                            "resolution": item.get("spdisplays_resolution"),
                        }
                    )
                return gpus
            except json.JSONDecodeError:
                pass

    return gpus


def _read_text(path: Path) -> str:
    try:
        return path.read_text(encoding="utf-8").strip()
    except OSError:
        return ""


def _pci_vendor_name(vendor_id: str) -> str:
    vendors = {
        "0x10de": "NVIDIA",
        "0x1002": "AMD",
        "0x8086": "Intel",
        "0x106b": "Apple",
    }
    return vendors.get(vendor_id.lower(), vendor_id)


def _pci_device_name(vendor_id: str, device_id: str) -> str | None:
    output = _run_command(["lspci", "-nn"], timeout=4)
    if not output:
        return None
    pattern = re.compile(
        rf".*\[({re.escape(vendor_id)}):({re.escape(device_id)})\].*",
        re.IGNORECASE,
    )
    for line in output.splitlines():
        if pattern.search(line):
            return line.split(":", 1)[-1].strip()
    return None


def _is_mac_address(value: str) -> bool:
    return bool(re.fullmatch(r"([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}", value))


def _get_network_interfaces() -> list[dict[str, Any]]:
    addrs = psutil.net_if_addrs()
    stats = psutil.net_if_stats()
    io = psutil.net_io_counters(pernic=True)
    interfaces: list[dict[str, Any]] = []

    for name in sorted(addrs):
        if name.lower() in {"lo", "loopback"}:
            continue
        nic_addrs = addrs.get(name, [])
        ipv4: list[str] = []
        ipv6: list[str] = []
        mac = ""
        for addr in nic_addrs:
            if addr.family == socket.AF_INET:
                ipv4.append(addr.address)
            elif addr.family == socket.AF_INET6:
                ipv6.append(addr.address)
            elif _is_mac_address(addr.address):
                mac = addr.address

        stat = stats.get(name)
        counters = io.get(name)
        interfaces.append(
            {
                "name": name,
                "ipv4": ipv4,
                "ipv6": ipv6,
                "mac": mac,
                "is_up": bool(stat.isup) if stat else False,
                "speed_mbps": stat.speed if stat and stat.speed > 0 else None,
                "bytes_sent_gb": _bytes_to_gb(counters.bytes_sent) if counters else None,
                "bytes_recv_gb": _bytes_to_gb(counters.bytes_recv) if counters else None,
            }
        )
    return interfaces


def _get_battery() -> dict[str, Any] | None:
    try:
        battery = psutil.sensors_battery()
    except (AttributeError, OSError):
        return None
    if battery is None:
        return None
    return {
        "percent": round(battery.percent, 1),
        "plugged": battery.power_plugged,
        "secsleft": battery.secsleft if battery.secsleft >= 0 else None,
    }


def _get_cpu_section() -> dict[str, Any]:
    cpu_freq = psutil.cpu_freq()
    per_cpu_freq = psutil.cpu_freq(percpu=True) or []
    per_core_percent = [round(v, 1) for v in psutil.cpu_percent(interval=None, percpu=True)]
    percent = round(sum(per_core_percent) / len(per_core_percent), 1) if per_core_percent else 0.0

    return {
        "name": _get_cpu_name(),
        "cores_physical": psutil.cpu_count(logical=False) or 0,
        "cores_logical": psutil.cpu_count(logical=True) or 0,
        "freq_mhz": round(cpu_freq.current) if cpu_freq and cpu_freq.current else None,
        "freq_min_mhz": round(cpu_freq.min) if cpu_freq and cpu_freq.min else None,
        "freq_max_mhz": round(cpu_freq.max) if cpu_freq and cpu_freq.max else None,
        "percent": percent,
        "per_core_percent": per_core_percent,
        "per_core_freq_mhz": [
            round(item.current) if item and item.current else None for item in per_cpu_freq
        ],
    }


def get_system_info() -> dict[str, Any]:
    memory = psutil.virtual_memory()
    swap = psutil.swap_memory()
    partitions = _get_partitions()

    return {
        "hostname": platform.node(),
        "os": _get_os_label(),
        "os_version": platform.version(),
        "platform": sys.platform,
        "architecture": platform.machine(),
        "python_version": platform.python_version(),
        "cpu": _get_cpu_section(),
        "memory": {
            "total_gb": _bytes_to_gb(memory.total),
            "used_gb": _bytes_to_gb(memory.used),
            "available_gb": _bytes_to_gb(memory.available),
            "free_gb": _bytes_to_gb(memory.free),
            "percent": round(memory.percent, 1),
        },
        "swap": {
            "total_gb": _bytes_to_gb(swap.total),
            "used_gb": _bytes_to_gb(swap.used),
            "free_gb": _bytes_to_gb(swap.free),
            "percent": round(swap.percent, 1),
        },
        "physical_drives": _get_physical_drives(),
        "partitions": partitions,
        "disks": partitions,
        "gpus": _get_gpus(),
        "network": _get_network_interfaces(),
        "battery": _get_battery(),
        "boot_time": psutil.boot_time(),
        "uptime_sec": int(time.time() - psutil.boot_time()),
    }
