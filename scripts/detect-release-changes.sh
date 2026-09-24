#!/usr/bin/env bash
# Определяет, что пересобирать в release workflow:
#   hub     — Docker-образ hub (только при изменениях hub-кода)
#   windows — Windows installer (.exe)
#   deb     — Linux agent (.deb) и APT repo
#
# .deb и .exe всегда собираются вместе на каждом релизе.
set -euo pipefail

CURRENT="${1:-${GITHUB_REF_NAME:-HEAD}}"
PREV="$(git tag --sort=-v:refname | awk -v c="$CURRENT" '$0==c {getline; print; exit}')"

hub=false
windows=true
deb=true

classify_file() {
  local file="$1"

  if [[ "$file" =~ ^(Dockerfile|\.dockerignore|requirements\.txt|pyproject\.toml)$ ]] \
    || [[ "$file" =~ ^docker-compose ]] \
    || [[ "$file" =~ ^web/ ]] \
    || [[ "$file" =~ ^config/ ]] \
    || [[ "$file" =~ ^system_monitor/ ]]; then
    hub=true
  fi
}

if [[ -z "$PREV" ]]; then
  hub=true
  echo "No previous tag — full release build"
else
  echo "Comparing ${PREV}..HEAD"
  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    classify_file "$file"
    if $hub; then
      break
    fi
  done < <(git diff --name-only "$PREV" HEAD)

  $hub && echo "Hub (Docker) build required"
  echo "Agent (.deb + Windows .exe) build required (always on release)"
  if ! $hub; then
    echo "Hub unchanged — Docker image will not be rebuilt"
  fi
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "hub=$hub" >> "$GITHUB_OUTPUT"
  echo "windows=$windows" >> "$GITHUB_OUTPUT"
  echo "deb=$deb" >> "$GITHUB_OUTPUT"
else
  echo "hub=$hub"
  echo "windows=$windows"
  echo "deb=$deb"
fi
