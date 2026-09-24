#!/usr/bin/env bash
# Backward-compatible wrapper — use detect-release-changes.sh for full output.
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
"$script_dir/detect-release-changes.sh" "$@"
