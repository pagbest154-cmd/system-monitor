#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <deb-file> <output-dir>" >&2
  exit 1
fi

deb_file="$1"
out_dir="$2"
suite="stable"
component="main"
arch="$(dpkg-deb -f "$deb_file" Architecture)"
version="$(dpkg-deb -f "$deb_file" Version)"

pool_dir="$out_dir/pool/main/s/system-monitor-agent"
packages_dir="$out_dir/dists/$suite/$component/binary-$arch"
packages_file="$packages_dir/Packages"

mkdir -p "$pool_dir" "$packages_dir"
cp "$deb_file" "$pool_dir/"

apt-ftparchive packages "$out_dir/pool/main" > "$packages_file"
gzip -9c "$packages_file" > "$packages_file.gz"

cat >"$out_dir/apt-ftparchive.conf" <<EOF
APT::FTPArchive::Release::Origin "system-monitor-agent";
APT::FTPArchive::Release::Label "system-monitor-agent";
APT::FTPArchive::Release::Suite "$suite";
APT::FTPArchive::Release::Codename "$suite";
APT::FTPArchive::Release::Architectures "$arch";
EOF

apt-ftparchive -c="$out_dir/apt-ftparchive.conf" release "$out_dir/dists/$suite" \
  > "$out_dir/dists/$suite/Release"
rm -f "$out_dir/apt-ftparchive.conf"

cat >"$out_dir/README.txt" <<EOF
Add this APT source:

  deb [trusted=yes] https://pagbest154-cmd.github.io/system-monitor/apt $suite $component

Then run:

  sudo apt update
  sudo apt install system-monitor-agent

Package version: $version
EOF

echo "APT repository created at: $out_dir"
