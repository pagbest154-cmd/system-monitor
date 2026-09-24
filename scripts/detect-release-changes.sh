#!/usr/bin/env bash
# Определяет, что пересобирать в release workflow:
#   hub     — Docker-образ hub
#   windows — Windows installer (.exe)
#   deb     — Linux agent (.deb) и APT repo
set -euo pipefail

CURRENT="${1:-${GITHUB_REF_NAME:-HEAD}}"
PREV="$(git tag --sort=-v:refname | awk -v c="$CURRENT" '$0==c {getline; print; exit}')"

hub=false
windows=false
deb=false

is_agent_core() {
  case "$1" in
    system_monitor/agent/*|system_monitor/collector/*|system_monitor/protocol/*|\
    system_monitor/fleet/*|system_monitor/config_loader.py|\
    system_monitor/disk_discovery.py|system_monitor/paths.py|\
    system_monitor/system_info.py|system_monitor/release_updates.py|\
    config/agent_sensors.yaml|requirements.txt)
      return 0
      ;;
  esac
  return 1
}

classify_file() {
  local file="$1"

  case "$file" in
    packaging/windows/*)
      windows=true
      return
      ;;
    debian/*|packaging/usrbin*)
      deb=true
      return
      ;;
  esac

  if is_agent_core "$file"; then
    windows=true
    deb=true
  fi

  case "$file" in
    system_monitor/agent/*)
      return
      ;;
  esac

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
  windows=true
  deb=true
  echo "No previous tag — full release build"
else
  echo "Comparing ${PREV}..HEAD"
  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    classify_file "$file"
    if $hub && $windows && $deb; then
      break
    fi
  done < <(git diff --name-only "$PREV" HEAD)

  $hub && echo "Hub (Docker) build required"
  $windows && echo "Windows agent build required"
  $deb && echo "Deb agent build required"
  if ! $hub && ! $windows && ! $deb; then
    echo "No component changes detected — release will contain notes only"
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
