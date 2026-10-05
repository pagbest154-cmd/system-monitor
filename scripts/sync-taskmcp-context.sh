#!/usr/bin/env bash
# Upload docs/taskmcp-project-context.md to TaskMCP project agent context.
# Env: TASKMCP_API_TOKEN (MCP key or session token from taskmcp.ru project settings),
#      TASKMCP_PROJECT_SLUG (default: sysmon),
#      TASKMCP_BASE_URL (default: https://taskmcp.ru)
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
context_file="${1:-$root/docs/taskmcp-project-context.md}"

token="${TASKMCP_API_TOKEN:-${TASKMCP_NOTIFY_TOKEN:-}}"
slug="${TASKMCP_PROJECT_SLUG:-sysmon}"
base="${TASKMCP_BASE_URL:-https://taskmcp.ru}"
base="${base%/}"

if [[ ! -f "$context_file" ]]; then
  echo "error: context file not found: $context_file" >&2
  exit 1
fi

if [[ -z "$token" ]]; then
  echo "error: set TASKMCP_API_TOKEN (or TASKMCP_NOTIFY_TOKEN)" >&2
  exit 1
fi

payload="$(python3 <<'PY'
import json
import pathlib
import os

path = pathlib.Path(os.environ["CONTEXT_FILE"])
text = path.read_text(encoding="utf-8")
print(json.dumps({"context": text, "agent_context": text}, ensure_ascii=False))
PY
)"
export CONTEXT_FILE="$context_file"

url="${base}/api/projects/${slug}"
echo "TaskMCP sync context: ${slug} (${#payload} bytes JSON) → PATCH ${url}"

curl -sS -f -X PATCH "$url" \
  -H "Authorization: Bearer ${token}" \
  -H "Content-Type: application/json; charset=utf-8" \
  -d "$payload"

echo
echo "OK — verify in Cursor: get_planning_context (excerpt should mention Go 1.0.N)."
