"""PyInstaller entry point — absolute imports so the frozen binary runs as a package."""
from __future__ import annotations

import sys
import traceback

from system_monitor.agent.__main__ import main

if __name__ == "__main__":
    try:
        main()
    except Exception:
        traceback.print_exc()
        sys.exit(1)
