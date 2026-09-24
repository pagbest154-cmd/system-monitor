#!/usr/bin/env bash
# Извлекает секцию версии из CHANGELOG.md для GitHub Release.
# Usage: extract-release-notes.sh <version> <output-file>
set -euo pipefail

version="${1:?version required, e.g. 0.0.1}"
out="${2:?output file required}"

root="$(cd "$(dirname "$0")/.." && pwd)"
changelog="$root/CHANGELOG.md"

if [[ ! -f "$changelog" ]]; then
  echo "error: CHANGELOG.md not found" >&2
  exit 1
fi

awk -v ver="$version" '
  BEGIN { found=0 }
  /^## \[/ {
    if (found) { exit }
    if ($0 ~ "\\[" ver "\\]") { found=1; print; next }
    next
  }
  found { print }
' "$changelog" > "$out"

if [[ ! -s "$out" ]]; then
  echo "warning: no CHANGELOG.md section for version $version, using fallback" >&2
  {
    echo "## [${version}]"
    echo
    echo "Release ${version}."
  } > "$out"
fi

echo "Release notes written to $out"
