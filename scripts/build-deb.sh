#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

echo "Installing build dependencies (Go agent .deb)..."
sudo apt-get update
sudo apt-get install -y \
  debhelper-compat \
  build-essential \
  devscripts \
  apt-utils \
  gzip

find debian packaging scripts -type f -exec sed -i 's/\r$//' {} +

chmod +x debian/rules packaging/usrbin-system-monitor-agent
chmod +x debian/system-monitor-agent.postinst debian/system-monitor-agent.prerm debian/system-monitor-agent.postrm
chmod +x debian/system-monitor-agent.config

if ! command -v go >/dev/null; then
  echo "error: Go 1.23+ required on PATH (same as CI)" >&2
  exit 1
fi

dpkg-buildpackage -us -uc -b -d

echo
echo "Built packages:"
ls -1 ../system-monitor-agent_*.deb
