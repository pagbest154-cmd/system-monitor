#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

echo "Installing build dependencies..."
sudo apt-get update
sudo apt-get install -y \
  debhelper-compat \
  build-essential \
  devscripts \
  python3 \
  python3-dev \
  python3-pip \
  python3-venv \
  python3-build \
  python3-wheel

chmod +x debian/rules packaging/usrbin-system-monitor-agent
chmod +x debian/system-monitor-agent.postinst debian/system-monitor-agent.prerm debian/system-monitor-agent.postrm
chmod +x debian/system-monitor-agent.config
dpkg-buildpackage -us -uc -b

echo
echo "Built packages:"
ls -1 ../system-monitor-agent_*.deb
