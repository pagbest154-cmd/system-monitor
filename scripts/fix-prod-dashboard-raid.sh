#!/bin/sh
set -euo pipefail
p=/opt/system-monitor/config/dashboard.yaml
python3 <<'PY'
from pathlib import Path
p = Path("/opt/system-monitor/config/dashboard.yaml")
t = p.read_text()
# Remove wrongly indented trailing block if present
bad = "\n  - id: raid_status"
i = t.find(bad)
if i >= 0:
    t = t[:i].rstrip() + "\n"
block = """    - id: raid_status
      title: Состояние RAID
      type: raid
      sensors:
        - auto_mdadm
      row: 7
"""
if "raid_status" not in t:
    marker = "    - id: ram_gauge"
    j = t.rfind(marker)
    if j < 0:
        raise SystemExit("ram_gauge panel not found")
    k = t.find("\n    - id:", j + len(marker))
    if k < 0:
        k = len(t)
    t = t[:k] + "\n" + block + t[k:]
p.write_text(t)
print("dashboard.yaml updated")
PY
