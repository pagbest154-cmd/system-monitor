#!/usr/bin/env bash
# Определяет, менялся ли код hub с предыдущего тега.
# Используется в release workflow: Docker пересобирается только при hub=true.
set -euo pipefail

CURRENT="${1:-${GITHUB_REF_NAME:-HEAD}}"
PREV="$(git tag --sort=-v:refname | awk -v c="$CURRENT" '$0==c {getline; print; exit}')"

hub=false

if [[ -z "$PREV" ]]; then
  hub=true
  echo "No previous tag — hub build required"
else
  echo "Comparing ${PREV}..HEAD for hub changes"
  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    [[ "$file" == system_monitor/agent/* ]] && continue

    if [[ "$file" =~ ^(Dockerfile|\.dockerignore|requirements\.txt|pyproject\.toml)$ ]] \
      || [[ "$file" =~ ^docker-compose ]] \
      || [[ "$file" =~ ^web/ ]] \
      || [[ "$file" =~ ^config/ ]] \
      || [[ "$file" =~ ^system_monitor/ ]]; then
      hub=true
      echo "Hub change detected: $file"
      break
    fi
  done < <(git diff --name-only "$PREV" HEAD)
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "hub=$hub" >> "$GITHUB_OUTPUT"
else
  echo "hub=$hub"
fi
