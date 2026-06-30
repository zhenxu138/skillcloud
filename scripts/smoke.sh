#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
repo="$tmp/repo"
project="$tmp/project"
bin="$tmp/skillcloud"

mkdir -p "$repo" "$project"
cp -R "$root/testdata/sample-skills/skills" "$repo/skills"

go build -o "$bin" ./cmd/skillcloud

export HOME="$tmp/home"
mkdir -p "$HOME"

git -C "$repo" init
git -C "$repo" config user.email "skillcloud@example.invalid"
git -C "$repo" config user.name "Skillcloud Smoke"
git -C "$repo" add -A
git -C "$repo" commit -m "sample skills"

"$bin" init "$repo"
"$bin" pull
"$bin" list
cd "$project"
"$bin" enable coding/code-review stock/risk-control --target codex --scope project --mode copy
test -f ".agents/skills/code-review/SKILL.md"
test -f ".agents/skills/risk-control/SKILL.md"

echo "smoke passed"

