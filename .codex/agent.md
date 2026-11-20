# skillcloud agent instructions

## Scope

This repository contains the `skillcloud` CLI. The skills themselves live in a separate repository; do not add personal skills or catalogs here.

## Architecture

- `cmd/skillcloud/` — binary entry point.
- `internal/cli/` — Cobra commands and command-level orchestration.
- `internal/config/` — global config, paths, home expansion.
- `internal/skill/` — skill discovery, index, exact/name/glob resolution.
- `internal/group/` — `GROUP.yaml` discovery and member resolution.
- `internal/agents/` — managed scoped group index in project `AGENTS.md` and legacy block cleanup.
- `internal/install/` — projection plans, replacement checks, link/copy execution.
- `internal/project/` — `.skillcloud.yaml` schema and status inspection.
- `internal/tui/` — Bubble Tea browser and management UI.
- `internal/validate/` — repository and skill validation.

## Required checks

Run these before reporting a code change as complete:

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
go build -o skillcloud ./cmd/skillcloud
```

For changes to skill discovery, groups, projections, or config, add or update a focused test instead of relying on a manual run alone.

## Invariants

- Never overwrite or delete an unmanaged skill destination. Replacement must go through `install.CheckReplace`; identical unmanaged content may be adopted, different content must be refused.
- Projections are recorded in sidecar manifests when linked and inside the directory when copied. Keep removal tied to a matching manifest.
- Group rules live in `.skillcloud/groups/<safe-group-id>.md`. Project `AGENTS.md` contains only a scoped index block delimited by `<!-- skillcloud:groups:begin -->` and `...:end -->`; do not flatten rule files back into root instructions. Remove legacy `<!-- skillcloud:group:<id>:begin -->` blocks when migrating.
- `group use` is project scope only. Groups may overlap; `--prefix` resolves alias collisions. A shared skill keeps every owner in `SkillRef.Groups` and must not be removed until the last owning group is removed. Keep the legacy single `group` field readable.
- Keep `.skillcloud.yaml` backward compatible. New optional fields use `omitempty`; existing configs must still load and apply.
- `apply` must refresh projections, group rule files, the scoped index, and remove orphaned rule files.
- Keep `README.md` and `README.zh-CN.md` in sync when commands, config, or behavior change.
- Do not push, tag, publish, or change external state without explicit user instruction.

## Style

Prefer the smallest complete change. Reuse the existing command, project, skill, and install packages rather than adding parallel abstractions. Keep CLI output plain and script-friendly unless the TUI is the intended interface.
