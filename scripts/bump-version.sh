#!/usr/bin/env bash
# Версия 0.0.N, где N = число коммитов в текущей ветке.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

count="$(git rev-list --count HEAD)"
version="0.0.${count}"

echo "Version: $version (commits: $count)"

sed -i "s/^version = .*/version = \"${version}\"/" pyproject.toml
sed -i "s/^__version__ = .*/__version__ = \"${version}\"/" system_monitor/__init__.py

echo "Updated pyproject.toml and system_monitor/__init__.py"
