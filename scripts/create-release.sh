#!/usr/bin/env bash
set -euo pipefail

tag="${1:?Usage: $0 <tag>   e.g. v0.0.3}"
force=false

if [[ "${2:-}" == "--force" ]]; then
  force=true
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "error: uncommitted changes — commit and push to main first" >&2
  exit 1
fi

branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$branch" != "main" ]]; then
  echo "warning: not on main (current: $branch)" >&2
fi

count="$(git rev-list --count HEAD)"
expected="v0.0.${count}"

echo "Release tag: $tag"
echo "Expected:    $expected (0.0.N = commit count)"
echo "Commit:      $(git rev-parse --short HEAD)"
echo "pyproject:   $(grep -m1 '^version' pyproject.toml)"
echo "debian:      $(head -1 debian/changelog)"
echo

if [[ "$tag" != "$expected" ]]; then
  echo "warning: tag $tag does not match commit count ($expected)" >&2
fi

git fetch origin main 2>/dev/null || true
local_main="$(git rev-parse main)"
remote_main="$(git rev-parse origin/main 2>/dev/null || echo "$local_main")"
if [[ "$local_main" != "$remote_main" ]]; then
  echo "warning: local main differs from origin/main — push main before tagging" >&2
fi

export GIT_AUTHOR_NAME='pagbest154-cmd'
export GIT_AUTHOR_EMAIL='pagbest154-cmd@users.noreply.github.com'
export GIT_COMMITTER_NAME='pagbest154-cmd'
export GIT_COMMITTER_EMAIL='pagbest154-cmd@users.noreply.github.com'

git tag -fa "$tag" -m "Release $tag"

if $force; then
  git push --force origin "$tag"
else
  git push origin "$tag"
fi

cat <<EOF

Tag $tag pushed. GitHub Actions will:
  - build system-monitor_*.deb
  - create GitHub Release
  - publish APT repo to GitHub Pages

Track: https://github.com/pagbest154-cmd/system-monitor/actions
EOF
