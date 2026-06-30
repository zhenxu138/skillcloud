$ErrorActionPreference = "Stop"

$root = Resolve-Path "$PSScriptRoot\.."
$bin = Join-Path $root "skillcloud.exe"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("skillcloud-smoke-" + [System.Guid]::NewGuid().ToString())
$repo = Join-Path $tmp "repo"
$project = Join-Path $tmp "project"

New-Item -ItemType Directory -Force -Path $repo, $project | Out-Null
Copy-Item -Recurse -Path (Join-Path $root "testdata\sample-skills\skills") -Destination $repo

go build -o $bin ./cmd/skillcloud

$env:HOME = Join-Path $tmp "home"
$env:USERPROFILE = $env:HOME
New-Item -ItemType Directory -Force -Path $env:HOME | Out-Null

git -C $repo init
git -C $repo config user.email "skillcloud@example.invalid"
git -C $repo config user.name "Skillcloud Smoke"
git -C $repo add -A
git -C $repo commit -m "sample skills"

& $bin init $repo
& $bin pull
& $bin list

Push-Location $project
& $bin enable coding/code-review stock/risk-control --target codex --scope project --mode copy
if (!(Test-Path ".agents\skills\code-review\SKILL.md")) { throw "missing code-review" }
if (!(Test-Path ".agents\skills\risk-control\SKILL.md")) { throw "missing risk-control" }

$external = Join-Path $tmp "external-skill"
New-Item -ItemType Directory -Force -Path $external | Out-Null
$skillContent = @"
---
name: go-review
description: Review Go code changes.
---
Body
"@ -replace "`r`n", "`n"
[System.IO.File]::WriteAllText((Join-Path $external "SKILL.md"), $skillContent)

& $bin add $external --as coding/go-review
& $bin use coding/go-review --target codex --scope project --mode copy
if (!(Test-Path ".agents\skills\go-review\SKILL.md")) { throw "missing go-review" }
if (!(Test-Path ".agents\skills\go-review\.skillcloud-projection.json")) { throw "missing go-review projection manifest" }
& $bin unuse coding/go-review
if (Test-Path ".agents\skills\go-review") { throw "go-review should be removed" }
Pop-Location

Write-Host "smoke passed"

