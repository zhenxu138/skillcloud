<div align="center">

# skillcloud

Bring your own Agent Skills repository. Sync it once, enable only what each project needs.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square)](#installation)
[![Agents](https://img.shields.io/badge/agents-Codex%20%7C%20Claude%20Code%20%7C%20Hermes-blueviolet?style=flat-square)](#supported-agents)

[中文文档](README.zh-CN.md)

</div>

## What Is It?

`skillcloud` is an open-source CLI for managing Agent Skills across machines, workspaces, and AI coding agents.

This repository is **only the CLI**. Your actual skills live in a separate Git repository that you control:

```text
skillcloud CLI repo        -> this open-source tool
your skills repo           -> your private or public Agent Skills library
your projects              -> opt in to selected skills
```

That separation keeps the tool reusable for everyone. A company can point `skillcloud` at an internal skills repository. An individual can point it at a private GitHub repo. An open-source community can publish a shared skills catalog.

## The Problem

Agent Skills become hard to manage once you have more than a handful:

- different agents expect different directories;
- most agent skill folders are flat, which is bad for organizing a large library;
- copying skills across devices creates version drift;
- projects should not load every skill you have ever written.

`skillcloud` keeps a full categorized skills repository locally, then projects selected skills into the directory shape each agent expects.

## Core Idea

```text
Git-hosted skills repository
        |
        | skillcloud update
        v
~/.skillcloud/repo
        |
        | skillcloud use coding/code-review --target codex --scope project
        v
my-project/.agents/skills/code-review
```

Your skills repo can stay organized:

```text
skills/
  coding/
    code-review/
      SKILL.md
    tdd/
      SKILL.md
  stock/
    risk-control/
      SKILL.md
  writing/
    prd-review/
      SKILL.md
```

Your project receives only the skills it enables:

```text
my-project/
  .skillcloud.yaml
  .agents/skills/
    code-review -> ~/.skillcloud/repo/skills/coding/code-review
    risk-control -> ~/.skillcloud/repo/skills/stock/risk-control
```

## Features

| Area | What skillcloud does |
| --- | --- |
| Sync | Pull and push a Git-hosted skills repository. |
| Import | Add existing local skill directories to your library with `skillcloud add`. |
| Organization | Keep skills in nested categories instead of one flat folder. |
| Selection | Enable explicit skills, many skills, or `category/*`. |
| Groups | Enable a named set of skills and its managed `AGENTS.md` rules. |
| TUI | Browse and select skills interactively. |
| Projection | Link or copy selected skills into agent-specific directories. |
| Agents | Built-in target paths for Codex, Claude Code, and Hermes. |
| Validation | Check whether a skills repository is discoverable and well-formed. |
| Doctor | Check local Git/config setup before debugging by guesswork. |

## Installation

Build from source:

```bash
git clone git@github.com:zhenxu138/skillcloud.git
cd skillcloud
go build -o skillcloud ./cmd/skillcloud
```

Move the resulting binary to a directory on your `PATH`.

## Quick Start

Create a separate skills repository first. It can be private or public:

```text
my-agent-skills/
  skills/
    coding/
      code-review/
        SKILL.md
```

Connect `skillcloud` to that repository:

```bash
skillcloud connect git@github.com:USER/my-agent-skills.git
```

Pull the latest skills:

```bash
skillcloud update
```

Inspect available skills:

```bash
skillcloud list
skillcloud search review
skillcloud
```

### Manage project skills with the TUI

```bash
skillcloud
```

By default this manages Codex project skills.

The TUI opens a launcher menu. Move with Up/Down (or j/k), run an entry with Enter, cycle with Tab, and quit with q or Esc.

| Entry | What it does |
| --- | --- |
| Browse | Select skills for the current project interactively |
| Enabled | Review and disable enabled skills for the current project |
| Add | Import a local skill directory into the library |
| Changes | Show library repo changes and project projection drift |
| Sync | Pull skill updates from the remote and rebuild the index |
| Push | Commit and push local skill edits to the remote |
| Settings | Show the current skillcloud configuration |

| Key | Action |
| --- | --- |
| `Space` | Toggle the current skill |
| `/` | Search skills |
| `Tab` | Cycle all/enabled/disabled/changed views |
| `Enter` | Review and apply changes |
| `q` | Quit without saving |

In Browse, checked skills are used for the current project. Unchecking an already used skill removes it from the project. In Add, press b on the source field to browse the filesystem for a skill directory.

Add a downloaded skill to your skills repository:

```bash
skillcloud add ./downloaded-skill --as coding/code-review
```

Use skills for the current project:

```bash
skillcloud use coding/code-review --target codex --scope project
```

Stop using a skill for the current project:

```bash
skillcloud unuse coding/code-review
```

Push changes made inside the local skills cache:

```bash
skillcloud push -m "add code review skill"
```

## Skill Groups

A group bundles any set of skills with optional `AGENTS.md` rules. Groups are discovered from `GROUP.yaml` files under `groups/` or `skills/`, so a group can reuse skills from different categories.

```text
groups/devops/
  GROUP.yaml
  AGENTS.md
```

```yaml
name: devops
title: DevOps
description: Server inspection, network diagnosis, and remote debugging.
entry: ops/collect-env-info
agents: AGENTS.md
skills:
  - id: ops/clash-uas-mac-troubleshoot
  - id: ops/collect-env-info
  - name: remote-debug-helper
    as: remote-debug
```

- `skills` accepts exact skill IDs or unique skill names. `as` overrides the projected directory name.
- If `skills` is omitted and `GROUP.yaml` is inside `skills/`, every skill under that directory is included.
- `title` and `description` label the group; `description` is the short workflow/strength summary shown by `group list`.
- `agents` points to a Markdown file. If omitted, skillcloud checks `AGENTS.md` and then `.codex/agent.md` in the group directory. The content is written into the project `AGENTS.md` inside a managed block, and removed when the group is removed.
- Group enablement is project scope. `.skillcloud.yaml` records the group and tags the skills it added, so removing the group does not remove skills that were already enabled manually.
- `--no-agents` records the group while suppressing its rules; later `apply` runs keep them suppressed.
- Codex loads `AGENTS.md` at session start, so start a new session after `group use` to pick up new rules.

```bash
skillcloud group list
skillcloud group show devops
skillcloud group use devops --target codex --mode link
skillcloud group remove devops
```

`skillcloud apply` re-applies the configured skills and refreshes every group's managed `AGENTS.md` block.
Use `skillcloud current` to see the exact skills and groups enabled in the current project.

To inspect a group before loading it, use `skillcloud group show <group>`. To unload a whole group, use `skillcloud group remove <group>`; this removes its projections, its config record, and its managed `AGENTS.md` block. To remove one skill, use `skillcloud unuse <alias-or-id>`; for a group-owned skill, prefer `group remove` so the group record and rules are cleaned up too.

## Skill Format

Each skill is a folder containing `SKILL.md`.

```markdown
---
name: code-review
description: Review code changes for bugs, regressions, and missing tests.
---

Use this skill when reviewing code changes.
```

The skill ID is its path under `skills/`:

```text
skills/coding/code-review/SKILL.md -> coding/code-review
skills/stock/risk-control/SKILL.md -> stock/risk-control
```

## Supported Agents

| Target | Global scope | Project scope |
| --- | --- | --- |
| Codex | `~/.codex/skills` | `.agents/skills` |
| Claude Code | `~/.claude/skills` | `.claude/skills` |
| Hermes | `~/.hermes/skills` | `skills` |

## Install Modes

| Mode | Behavior | Use when |
| --- | --- | --- |
| `link` | Create a symlink. If linking fails, fall back to copy. | You want updates from the local skills cache to appear quickly. |
| `copy` | Copy the skill directory into the target path. | You want a more self-contained project setup. |

## Project Config

Project enablement is recorded in `.skillcloud.yaml`:

```yaml
targets:
  codex:
    mode: "link"
    skills:
      - id: "coding/code-review"
        as: "code-review"
      - id: "stock/risk-control"
        as: "risk-control"
        group: "groups/devops"
    groups:
      - id: "groups/devops"
        name: "devops"
```

- `id` is the categorized path in your skills repository.
- `as` is the flat directory name created for the agent.
- `group` marks a skill added by `skillcloud group use`.
- `groups` records the enabled groups whose `AGENTS.md` rules are managed.

## Commands

```bash
skillcloud connect <repo-url>
skillcloud update
skillcloud push -m "message"
skillcloud status

skillcloud list
skillcloud search <query>
skillcloud current            # skills and groups enabled here
skillcloud                    # launch main TUI

skillcloud add <path> --as <skill-id>
skillcloud use <skill-id...> --target codex --scope project
skillcloud use coding/* --target claude --scope project
skillcloud use --select --target hermes --scope global
skillcloud unuse <alias-or-skill-id...>
skillcloud group list
skillcloud group show <group>
skillcloud group use <group> --target codex --mode link
skillcloud group remove <group>
skillcloud apply

# compatibility aliases
skillcloud init <repo-url>    # alias for connect
skillcloud pull               # alias for update
skillcloud enable <skill-id...> --target codex --scope project
skillcloud disable <alias-or-skill-id...>

skillcloud validate
skillcloud doctor
```

## Repository Boundaries

`skillcloud` intentionally does not bundle a default skills catalog.

This keeps the CLI clean:

- no personal prompts mixed into the tool;
- no organization-specific skills in the public repo;
- no assumption that every user wants the same agent behavior;
- no need to fork the CLI just to maintain a private skills library.

Use one repository for the tool and another repository for your skills.

## Roadmap

- Profile-based bulk enablement such as `work`, `personal`, or `frontend`.
- Create new skills from templates.
- Rename and remove skills safely.
- Lockfile support for source commits and hashes.
- Stronger script and secret scanning.
- Git LFS or object storage for large skill assets.
- More agent adapters.

## Development

Run tests:

```bash
go test ./...
```

Run the Windows smoke test:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1
```

## Contributing

Issues and pull requests are welcome. Please keep changes focused and include tests for behavior changes.

## License

MIT License. See [LICENSE](LICENSE).
