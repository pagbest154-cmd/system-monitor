#!/usr/bin/env bash
# Версия 1.0.N после миграции на Go (N = число коммитов).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

count="$(git rev-list --count HEAD)"
version="1.0.${count}"

echo "Version: $version (commits: $count)"

sed -i "s/var Version = .*/var Version = \"${version}\"/" internal/version/version.go

if [[ -x "$root/scripts/sync-deb-version.sh" ]]; then
  "$root/scripts/sync-deb-version.sh" "$version"
else
  echo "warning: scripts/sync-deb-version.sh not found" >&2
fi

echo "Updated internal/version/version.go and debian/changelog"
