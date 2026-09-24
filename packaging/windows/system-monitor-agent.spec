# -*- mode: python ; coding: utf-8 -*-
from pathlib import Path

SPEC_DIR = Path(SPECPATH)
ROOT = SPEC_DIR.parent.parent

block_cipher = None

a = Analysis(
    [str(SPEC_DIR / "agent_entry.py")],
    pathex=[str(ROOT)],
    binaries=[],
    datas=[
        (str(ROOT / "config" / "agent_sensors.yaml"), "config"),
    ],
    hiddenimports=[
        "system_monitor",
        "system_monitor.agent",
        "system_monitor.agent.__main__",
        "system_monitor.agent.runner",
        "system_monitor.agent.tray",
        "system_monitor.agent.settings_ui",
        "system_monitor.agent.status",
        "system_monitor.agent.service_control",
        "system_monitor.agent.updates",
        "system_monitor.agent.transport",
        "system_monitor.agent.transport.base",
        "system_monitor.agent.transport.http",
        "system_monitor.agent.transport.grpc",
        "system_monitor.collector",
        "system_monitor.collector.runner",
        "system_monitor.collector.temperature",
        "system_monitor.config_loader",
        "system_monitor.disk_discovery",
        "system_monitor.fleet",
        "system_monitor.fleet.service",
        "system_monitor.paths",
        "system_monitor.protocol",
        "system_monitor.protocol.models",
        "system_monitor.system_info",
        "wmi",
        "win32com",
        "win32com.client",
        "pythoncom",
        "pywintypes",
        "win32timezone",
        "httpx",
        "httpcore",
        "anyio",
        "sniffio",
        "certifi",
        "yaml",
        "pydantic",
        "psutil",
        "pystray",
        "PIL",
        "PIL.Image",
        "PIL.ImageDraw",
        "tkinter",
        "tkinter.ttk",
        "tkinter.messagebox",
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=["fastapi", "uvicorn", "grpc"],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name="system-monitor-agent",
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=False,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)

coll = COLLECT(
    exe,
    a.binaries,
    a.zipfiles,
    a.datas,
    strip=False,
    upx=False,
    upx_exclude=[],
    name="system-monitor-agent",
)
