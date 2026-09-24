#!/usr/bin/env bash
# Синхронизирует версию в debian/changelog (0.0.N-1).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

version="${1:-}"
if [[ -z "$version" ]]; then
  count="$(git rev-list --count HEAD)"
  version="0.0.${count}"
fi

deb_version="${version}-1"
today="$(date -R)"

if [[ ! -f debian/changelog ]]; then
  echo "error: debian/changelog not found" >&2
  exit 1
fi

current="$(head -1 debian/changelog)"
if [[ "$current" == "system-monitor-agent (${deb_version})"* ]]; then
  echo "debian/changelog already at ${deb_version}"
  exit 0
fi

tmp="$(mktemp)"
{
  echo "system-monitor-agent (${deb_version}) unstable; urgency=medium"
  echo
  echo "  * Release ${version}."
  echo
  echo " -- pagbest154-cmd <pagbest154-cmd@users.noreply.github.com>  ${today}"
  echo
  cat debian/changelog
} > "$tmp"
mv "$tmp" debian/changelog
echo "Updated debian/changelog to ${deb_version}"
