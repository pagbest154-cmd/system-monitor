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

merge_remote_pool() {
  local base="${APT_REPO_MERGE_URL:-}"
  if [[ -z "$base" || "${APT_REPO_MERGE:-}" != "1" ]]; then
    return 0
  fi
  local tmp_pkg
  tmp_pkg="$(mktemp)"
  if ! curl -fsSL "$base/dists/$suite/$component/binary-$arch/Packages" -o "$tmp_pkg" 2>/dev/null; then
    if ! curl -fsSL "$base/dists/$suite/$component/binary-$arch/Packages.gz" | gunzip >"$tmp_pkg" 2>/dev/null; then
      rm -f "$tmp_pkg"
      echo "merge: no remote Packages at $base (first publish?)" >&2
      return 0
    fi
  fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line//$'\r'/}"
    case "$line" in
      Filename:*)
        local fn="${line#Filename: }"
        mkdir -p "$out_dir/$(dirname "$fn")"
        if [[ ! -f "$out_dir/$fn" ]]; then
          echo "merge: fetching $fn" >&2
          curl -fsSL "$base/$fn" -o "$out_dir/$fn" || echo "merge: failed $fn" >&2
        fi
        ;;
    esac
  done <"$tmp_pkg"
  rm -f "$tmp_pkg"
}

mkdir -p "$pool_dir" "$packages_dir"
merge_remote_pool
cp "$deb_file" "$pool_dir/"

# Run from repo root so Packages lists paths like pool/main/... (not dist/apt-repo/pool/...).
(
  cd "$out_dir"
  apt-ftparchive packages pool/main >"dists/$suite/$component/binary-$arch/Packages"
)
gzip -9c "$packages_file" > "$packages_file.gz"

cat >"$out_dir/apt-ftparchive.conf" <<EOF
APT::FTPArchive::Release::Origin "system-monitor-agent";
APT::FTPArchive::Release::Label "system-monitor-agent";
APT::FTPArchive::Release::Suite "$suite";
APT::FTPArchive::Release::Codename "$suite";
APT::FTPArchive::Release::Architectures "$arch";
EOF

(
  cd "$out_dir"
  apt-ftparchive -c=apt-ftparchive.conf release "dists/$suite" >"dists/$suite/Release"
)
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
