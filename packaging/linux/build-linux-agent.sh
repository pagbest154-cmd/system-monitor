#!/bin/bash
set -euo pipefail

PACKAGING_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$PACKAGING_DIR/../.." && pwd)"
BUILD_DIR="$PACKAGING_DIR/build"
DIST_DIR="$PACKAGING_DIR/dist"
SPEC_FILE="$PACKAGING_DIR/system-monitor-agent.spec"
AGENT_BIN="$DIST_DIR/system-monitor-agent/system-monitor-agent"

echo "Building system-monitor-agent for Linux (PyInstaller)..."

rm -rf "$BUILD_DIR" "$DIST_DIR"
mkdir -p "$DIST_DIR"

cd "$ROOT"
python3 -m PyInstaller \
    --noconfirm \
    --clean \
    --distpath "$DIST_DIR" \
    --workpath "$BUILD_DIR" \
    "$SPEC_FILE"

if [ ! -x "$AGENT_BIN" ]; then
    echo "PyInstaller output not found: $AGENT_BIN" >&2
    exit 1
fi

echo "Smoke test: system-monitor-agent --version"
"$AGENT_BIN" --version

echo "Done: $AGENT_BIN"
