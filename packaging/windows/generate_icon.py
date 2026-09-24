#!/usr/bin/env python3
"""Generate packaging/windows/app-icon.ico from shared branding."""

from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from system_monitor.agent.branding import save_app_icon  # noqa: E402

OUTPUT = Path(__file__).resolve().parent / "app-icon.ico"


def main() -> None:
    save_app_icon(OUTPUT)
    print(f"Wrote {OUTPUT}")


if __name__ == "__main__":
    main()
