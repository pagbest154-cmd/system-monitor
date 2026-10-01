#!/usr/bin/env bash
# Integration test: .deb install seeds agent.yaml and agent.token from env (postinst).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

HUB_URL='https://ci-hub.example:9443/v1'
AGENT_ID='ci-agent-id_42'
AGENT_TOKEN='ci-token-with-"quote-and-\backslash'

AGENT_YAML=/etc/system-monitor/agent.yaml
TOKEN_FILE=/etc/system-monitor/agent.token

cleanup() {
  if dpkg -s system-monitor-agent &>/dev/null; then
    sudo DEBIAN_FRONTEND=noninteractive dpkg -r -y system-monitor-agent || true
  fi
}
trap cleanup EXIT

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

cleanup
sudo rm -rf /etc/system-monitor

echo "Installing ${deb[0]} with HUB_URL / AGENT_ID / AGENT_TOKEN..."
sudo DEBIAN_FRONTEND=noninteractive \
  HUB_URL="$HUB_URL" \
  AGENT_ID="$AGENT_ID" \
  AGENT_TOKEN="$AGENT_TOKEN" \
  dpkg -i "${deb[0]}"

if [ ! -f "$AGENT_YAML" ]; then
  echo "Missing $AGENT_YAML" >&2
  exit 1
fi
if [ ! -f "$TOKEN_FILE" ]; then
  echo "Missing $TOKEN_FILE" >&2
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
