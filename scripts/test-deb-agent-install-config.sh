#!/usr/bin/env bash
# Integration test: .deb install seeds agent.yaml and agent.token from env (postinst).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

HUB_URL='https://ci-hub.example:9443/v1'
AGENT_ID='ci-agent-id_42'
# debconf-set-selections line must not contain raw quotes; special chars covered in Go unit tests.
AGENT_TOKEN='ci-token-secret-42'

AGENT_YAML=/etc/system-monitor/agent.yaml
TOKEN_FILE=/etc/system-monitor/agent.token

purge_agent_pkg() {
  if dpkg -s system-monitor-agent &>/dev/null; then
    sudo DEBIAN_FRONTEND=noninteractive apt-get purge -y system-monitor-agent || true
  fi
  sudo rm -rf /etc/system-monitor /run/system-monitor-agent
}
trap purge_agent_pkg EXIT

echo "Installing build dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq debhelper-compat build-essential devscripts apt-utils gzip

find debian packaging scripts -type f -exec sed -i 's/\r$//' {} +

chmod +x debian/rules
chmod +x debian/system-monitor-agent.postinst debian/system-monitor-agent.prerm debian/system-monitor-agent.postrm
chmod +x debian/system-monitor-agent.config

echo "Building .deb..."
dpkg-buildpackage -us -uc -b -d

shopt -s nullglob
deb=(../system-monitor-agent_*_amd64.deb)
if [ ${#deb[@]} -eq 0 ]; then
  deb=(../system-monitor-agent_*.deb)
fi
if [ ${#deb[@]} -eq 0 ]; then
  echo "No .deb package found" >&2
  exit 1
fi

purge_agent_pkg

echo "Installing ${deb[0]} (debconf preseed; dpkg strips custom env from maintainer scripts)..."
printf '%s\n' \
  "system-monitor-agent system-monitor-agent/hub-url string ${HUB_URL}" \
  "system-monitor-agent system-monitor-agent/agent-id string ${AGENT_ID}" \
  "system-monitor-agent system-monitor-agent/token password ${AGENT_TOKEN}" \
  | sudo debconf-set-selections
sudo DEBIAN_FRONTEND=noninteractive dpkg -i "${deb[0]}"

dump_install_debug() {
  echo "--- debug: install state ---" >&2
  sudo ls -la /etc/system-monitor/ 2>&1 || true
  sudo cat /etc/system-monitor/agent.yaml 2>&1 || true
  sudo ls -la /run/system-monitor-agent/ 2>&1 || true
  sudo debconf-get-selections 2>/dev/null | grep system-monitor-agent || true
}

if ! sudo test -f "$AGENT_YAML"; then
  echo "Missing $AGENT_YAML" >&2
  dump_install_debug
  exit 1
fi
if ! sudo test -f "$TOKEN_FILE"; then
  echo "Missing $TOKEN_FILE" >&2
  dump_install_debug
  exit 1
fi

hub_in_yaml="$(sudo grep -E '^hub_url:' "$AGENT_YAML" || true)"
if ! echo "$hub_in_yaml" | grep -q "$HUB_URL"; then
  echo "hub_url mismatch in agent.yaml: $hub_in_yaml" >&2
  dump_install_debug
  exit 1
fi

token_on_disk="$(sudo cat "$TOKEN_FILE")"
if [ "$token_on_disk" != "$AGENT_TOKEN" ]; then
  echo "token mismatch: got $(printf '%q' "$token_on_disk")" >&2
  exit 1
fi

sudo env PATH="$PATH" \
  AGENT_YAML="$AGENT_YAML" \
  TOKEN_FILE="$TOKEN_FILE" \
  HUB_URL="$HUB_URL" \
  AGENT_ID="$AGENT_ID" \
  AGENT_TOKEN="$AGENT_TOKEN" \
  go test -tags=debinstall_integration ./internal/packaging/agentinstall/ -run TestDebInstalledConfigOnDisk -count=1 -v

echo "deb install config integration OK"
