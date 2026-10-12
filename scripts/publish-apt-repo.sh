#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 <deb-file> [<deb-file> ...] <output-dir>" >&2
  echo "   or: $0 <deb-file> <output-dir>" >&2
  exit 1
}

if [ "$#" -lt 2 ]; then
  usage
fi

out_dir="${!#}"
if [ ! -d "$out_dir" ] && [ "$#" -eq 2 ]; then
  # legacy: <deb> <out-dir> (out-dir may not exist yet)
  deb_files=("$1")
  out_dir="$2"
elif [ -d "$out_dir" ] || [ "$#" -ge 3 ]; then
  deb_files=("${@:1:$#-1}")
else
  deb_files=("$1")
  out_dir="$2"
fi

suite="stable"
component="main"
pool_dir="$out_dir/pool/main/s/system-monitor-agent"

merge_remote_pool() {
  local arch="$1"
  local base="${APT_REPO_MERGE_URL:-}"
  if [[ -z "$base" || "${APT_REPO_MERGE:-}" != "1" ]]; then
    return 0
  fi
  local tmp_pkg
  tmp_pkg="$(mktemp)"
  if ! curl -fsSL "$base/dists/$suite/$component/binary-$arch/Packages" -o "$tmp_pkg" 2>/dev/null; then
    if ! curl -fsSL "$base/dists/$suite/$component/binary-$arch/Packages.gz" | gunzip >"$tmp_pkg" 2>/dev/null; then
      rm -f "$tmp_pkg"
      echo "merge: no remote Packages at $base for $arch (first publish?)" >&2
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

mkdir -p "$pool_dir"

declare -A seen_arch=()
version=""
for deb_file in "${deb_files[@]}"; do
  if [ ! -f "$deb_file" ]; then
    echo "Missing deb: $deb_file" >&2
    exit 1
  fi
  arch="$(dpkg-deb -f "$deb_file" Architecture)"
  version="$(dpkg-deb -f "$deb_file" Version)"
  merge_remote_pool "$arch"
  cp "$deb_file" "$pool_dir/"
  seen_arch["$arch"]=1
done

for arch in "${!seen_arch[@]}"; do
  packages_dir="$out_dir/dists/$suite/$component/binary-$arch"
  packages_file="$packages_dir/Packages"
  mkdir -p "$packages_dir"
  (
    cd "$out_dir"
    apt-ftparchive packages pool/main >"dists/$suite/$component/binary-$arch/Packages"
  )
  gzip -9c "$packages_file" > "$packages_file.gz"
done

arch_list="$(printf '%s\n' "${!seen_arch[@]}" | sort | tr '\n' ' ' | sed 's/ $//')"

cat >"$out_dir/apt-ftparchive.conf" <<EOF
APT::FTPArchive::Release::Origin "system-monitor-agent";
APT::FTPArchive::Release::Label "system-monitor-agent";
APT::FTPArchive::Release::Suite "$suite";
APT::FTPArchive::Release::Codename "$suite";
APT::FTPArchive::Release::Architectures "$arch_list";
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
Architectures: $arch_list
EOF

echo "APT repository created at: $out_dir (arch: $arch_list)"
