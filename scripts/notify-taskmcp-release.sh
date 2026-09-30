#!/usr/bin/env bash
# POST release version to TaskMCP → Telegram (incoming version webhook).
# Usage: notify-taskmcp-release.sh <version>
# Env: TASKMCP_NOTIFY_TOKEN (required), TASKMCP_PROJECT_SLUG (default: sysmon),
#      TASKMCP_NOTIFY_URL (default: https://taskmcp.ru),
#      GITHUB_RELEASE_URL (optional, default: github.com/.../releases/tag/v<version>)
set -euo pipefail

version="${1:?version required, e.g. 1.0.54}"

token="${TASKMCP_NOTIFY_TOKEN:-}"
slug="${TASKMCP_PROJECT_SLUG:-sysmon}"
base="${TASKMCP_NOTIFY_URL:-https://taskmcp.ru}"
base="${base%/}"
repo="${GITHUB_REPOSITORY:-pagbest154-cmd/system-monitor}"
release_url="${GITHUB_RELEASE_URL:-https://github.com/${repo}/releases/tag/v${version}}"

if [[ -z "$token" ]]; then
  echo "TASKMCP_NOTIFY_TOKEN not set — skip TaskMCP notify"
  exit 0
fi

export NOTIFY_VERSION="$version"
export NOTIFY_RELEASE_URL="$release_url"
payload="$(python3 <<'PY'
import json
import os

version = os.environ["NOTIFY_VERSION"]
release_url = os.environ["NOTIFY_RELEASE_URL"]
description = f"Вышла новая версия system-monitor {version}\n\n{release_url}"
print(json.dumps({"version": version, "description": description}, ensure_ascii=False))
PY
)"
url="${base}/api/projects/${slug}/notify"

echo "TaskMCP notify: ${slug} version=${version}"
curl -sS -f -X POST "$url" \
  -H "Authorization: Bearer ${token}" \
  -H "Content-Type: application/json; charset=utf-8" \
  -d "$payload"
