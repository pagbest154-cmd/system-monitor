#!/usr/bin/env bash
# Sync android/gradle.properties appVersion with hub release version (1.0.N).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
version="${1:-}"
props="$root/android/gradle.properties"

if [[ -z "$version" ]]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

if [[ ! -f "$props" ]]; then
  echo "error: $props not found" >&2
  exit 1
fi

version="${version#v}"
numeric_base="${version%%-*}"
patch="${numeric_base##*.}"
if [[ ! "$patch" =~ ^[0-9]+$ ]]; then
  echo "error: cannot derive appVersionCode from version '$version'" >&2
  exit 1
fi

sed -i "s/^appVersion=.*/appVersion=${version}/" "$props"
sed -i "s/^appVersionCode=.*/appVersionCode=${patch}/" "$props"
echo "Updated android/gradle.properties: appVersion=${version}, appVersionCode=${patch}"
