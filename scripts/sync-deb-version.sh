#!/usr/bin/env bash
# Синхронизирует версию в debian/changelog (<tag>-1).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

version="${1:-}"
if [[ -z "$version" ]]; then
  ref="${GITHUB_REF_NAME:-}"
  if [[ "$ref" =~ ^v[0-9] ]]; then
    version="${ref#v}"
  elif [[ -f internal/version/version.go ]]; then
    version="$(grep -oP 'Version = "\K[^"]+' internal/version/version.go || true)"
  fi
fi
if [[ -z "$version" ]]; then
  echo "error: version required (arg, GITHUB_REF_NAME, or internal/version/version.go)" >&2
  exit 1
fi
if [[ "$version" == v* ]]; then
  version="${version#v}"
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
