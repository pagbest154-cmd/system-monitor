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

chmod +x debian/rules packaging/usrbin-system-monitor
chmod +x debian/postinst debian/prerm debian/postrm
dpkg-buildpackage -us -uc -b

echo
echo "Built packages:"
ls -1 ../system-monitor_*.deb
