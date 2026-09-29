#!/usr/bin/env bash
# Upload android/release.keystore to GitHub Actions secrets (run once: gh auth login).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

keystore="${1:-android/sysmon-release.jks}"
store_password="${ANDROID_KEYSTORE_PASSWORD:-sysmon-release-store}"
key_alias="${ANDROID_KEY_ALIAS:-sysmon}"
key_password="${ANDROID_KEY_PASSWORD:-sysmon-release-key}"

if [[ ! -f "$keystore" ]]; then
  echo "error: keystore not found: $keystore" >&2
  exit 1
fi

base64 -w0 "$keystore" | gh secret set ANDROID_KEYSTORE_BASE64
printf '%s' "$store_password" | gh secret set ANDROID_KEYSTORE_PASSWORD
printf '%s' "$key_alias" | gh secret set ANDROID_KEY_ALIAS
printf '%s' "$key_password" | gh secret set ANDROID_KEY_PASSWORD

cp "$keystore" android/release.keystore
echo "GitHub secrets updated. Copied keystore to android/release.keystore"
