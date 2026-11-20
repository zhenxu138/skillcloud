<div align="center">

# skillcloud

自带你的 Agent Skills 仓库。同步一次，然后只给每个项目启用真正需要的 skill。

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square)](#安装)
[![Agents](https://img.shields.io/badge/agents-Codex%20%7C%20Claude%20Code%20%7C%20Hermes-blueviolet?style=flat-square)](#支持的-agent)

[English](README.md)

</div>

## 这是什么？

`skillcloud` 是一个开源 CLI，用来在多台设备、多个工作区、多个 AI 编程 agent 之间管理 Agent Skills。

这个仓库**只放 CLI 工具本身**。真正的 skills 应该放在你自己的独立 Git 仓库里：

```text
skillcloud CLI 仓库      -> 当前这个开源工具
你的 skills 仓库         -> 你的私有或公开 Agent Skills 库
你的项目                 -> 按需启用部分 skills
```

这样设计以后，工具就可以被任何人使用。公司可以把它指向内部 skills 仓库，个人可以指向私有 GitHub 仓库，社区也可以维护公开 skills catalog。

## 它解决什么问题？

当 skill 数量变多以后，管理成本会明显上升：

- 不同 agent 需要不同的 skill 目录；
- 很多 agent 只支持平铺目录，不适合整理大型 skill 库；
- 多设备手动复制 skill 很容易产生版本漂移；
- 单个项目不应该加载你写过的所有 skill。

`skillcloud` 的做法是：本地保留一份完整、分类清晰的 skills 仓库，然后只把项目启用的 skill 投影到目标 agent 需要的目录结构里。

## 核心模型

```text
Git 托管的 skills 仓库
        |
        | skillcloud update
        v
~/.skillcloud/repo
        |
        | skillcloud use coding/code-review --target codex --scope project
        v
my-project/.agents/skills/code-review
```

你的 skills 仓库可以保持多级分类：

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

而项目里只会出现已启用的 skill：

```text
my-project/
  .skillcloud.yaml
  .agents/skills/
    code-review -> ~/.skillcloud/repo/skills/coding/code-review
    risk-control -> ~/.skillcloud/repo/skills/stock/risk-control
```

## 功能特性

| 领域 | skillcloud 做什么 |
| --- | --- |
| 同步 | 拉取和推送 Git 托管的 skills 仓库。 |
| 导入 | 用 `skillcloud add` 将本地已有 skill 目录加入 library。 |
| 整理 | 用多级分类目录管理 skill，而不是全部放进一个大平铺目录。 |
| 选择 | 支持启用单个 skill、多个 skill，或 `category/*` 整个分类。 |
| 技能组 | 一次启用一组 skill，并管理分作用域的规则文件。 |
| TUI | 支持交互式浏览和选择 skill。 |
| 投影 | 将选中的 skill 链接或复制到 agent 指定目录。 |
| Agent | 内置 Codex、Claude Code、Hermes 的目标路径。 |
| 校验 | 检查 skills 仓库是否可发现、格式是否正确。 |
| 诊断 | 检查本机 Git 和配置状态，减少盲猜。 |

## 安装

从源码构建：

```bash
git clone git@github.com:zhenxu138/skillcloud.git
cd skillcloud
go build -o skillcloud ./cmd/skillcloud
```

然后把生成的二进制文件放到你的 `PATH` 中。

## 快速开始

先创建一个独立的 skills 仓库，可以是私有仓库，也可以是公开仓库：

```text
my-agent-skills/
  skills/
    coding/
      code-review/
        SKILL.md
```

用这个仓库连接 `skillcloud`：

```bash
skillcloud connect git@github.com:USER/my-agent-skills.git
```

拉取最新 skills：

```bash
skillcloud update
```

查看可用 skills：

```bash
skillcloud list
skillcloud search review
skillcloud
```

### 使用 TUI 管理项目 skills

```bash
skillcloud
```

默认管理 Codex 项目级 skills。

TUI 打开一个启动菜单。用上下方向键（或 j/k）移动，Enter 执行选中项，Tab 循环切换，q 或 Esc 退出。

| 菜单项 | 功能 |
| --- | --- |
| Browse | 交互式选择当前项目要启用的 skill |
| Groups | 加载或移除整个技能组及其作用域规则 |
| Enabled | 查看并取消当前项目已启用的 skill |
| Add | 将本地 skill 目录导入到 library |
| Changes | 显示 library 仓库改动和项目投影漂移 |
| Sync | 从远端拉取 skill 更新并重建索引 |
| Push | 提交并推送本地 skill 改动到远端 |
| Settings | 显示当前 skillcloud 配置 |

| 按键 | 行为 |
| --- | --- |
| `Space` | 切换当前 skill |
| `/` | 搜索 skills |
| `Tab` | 切换全部/已使用/未使用/已变更视图 |
| `Home` / `End` | 跳到第一个或最后一个 skill |
| `PgUp` / `PgDn` | 翻一页 |
| `Enter` | 查看摘要并应用变更 |
| `q` | 不保存退出 |

在 Browse 中，勾选的 skill 会在当前项目使用；列表会围绕光标滚动，显示 skill 所属的组/来源，并在存在已启用 skill 时从第一个已启用项开始。取消勾选已使用的 skill 会从项目移除它。在 Add 中，在 source 字段按 b 可浏览文件系统选择 skill 目录。

将下载的 skill 添加到 skills 仓库：

```bash
skillcloud add ./downloaded-skill --as coding/code-review
```

在当前项目使用 skills：

```bash
skillcloud use coding/code-review --target codex --scope project
```

停止在当前项目使用某个 skill：

```bash
skillcloud unuse coding/code-review
```

推送本地 skills 缓存里的修改：

```bash
skillcloud push -m "add code review skill"
```

接受 target 或 mode 的命令默认使用 `codex`、项目作用域和 `link` 安装模式。

普通用户不需要关心本地缓存；skillcloud 会在内部维护一份隐藏缓存，用来校验、索引和安全投影 skills。

## 技能组

技能组把任意一组 skill 和可选的 `AGENTS.md` 规则绑定在一起。skillcloud 会扫描 `groups/` 和 `skills/` 下的 `GROUP.yaml`，所以一个组可以混合使用不同分类里的 skill。

```text
groups/devops/
  GROUP.yaml
  AGENTS.md
```

```yaml
name: devops
category: 自定义组
title: 运维与网络诊断
description: 服务器巡检、网络排查与远程调试。
entry: ops/collect-env-info
agents: AGENTS.md
skills:
  - id: ops/clash-uas-mac-troubleshoot
  - id: ops/collect-env-info
  - name: remote-debug-helper
    as: remote-debug
```

- `skills` 可以写精确的 skill ID，也可以写唯一的 skill 名称；`as` 覆盖投影后的目录名。
- `category` 决定 `group list` 中的分类；未填写时按清单路径推断为 SOP、个人学习或自定义组。
- 如果省略 `skills` 且 `GROUP.yaml` 位于 `skills/` 下，组会自动包含该目录下的全部 skill。
- `title` 和 `description` 用来标识组；`description` 会作为简短的工作流/擅长说明显示在 `group list` 中。
- `agents` 指向一个 Markdown 文件。省略时，skillcloud 会依次检查组目录下的 `AGENTS.md` 和 `.codex/agent.md`。启用组时，内容写入 `.skillcloud/groups/<组>.md`，项目 `AGENTS.md` 只保留指向它的作用域索引。
- 技能组只支持项目作用域。`.skillcloud.yaml` 会记录组，并给组新增的 skill 打上 `group` 标记；移除组不会删掉之前手动启用的同名 skill。
- `--no-agents` 会记录组但不写入规则；后续 `apply` 也会保持抑制状态。
- Codex 在会话启动时读取 `AGENTS.md`，所以 `group use` 后需要新会话才能加载新规则。

```bash
skillcloud group list
skillcloud group show devops
skillcloud group use devops --target codex --mode link
skillcloud group remove devops
```

`skillcloud apply` 会重新应用已配置的 skill，并刷新每个组的规则文件以及项目 `AGENTS.md` 中的作用域索引。
用 `skillcloud current` 查看当前项目实际启用的 skill 和组。

两个组有同名技能时，给第二个组加前缀：

```bash
skillcloud group use work
skillcloud group use matt-pocock --prefix matt
```

第二个组会投影成 `matt-tdd`、`matt-code-review` 等。被两个组共用的 skill 会记录多个组归属，只有最后一个拥有它的组被移除时才会删除投影。

加载前可以用 `skillcloud group show <组>` 查看成员。要取消整个组，使用 `skillcloud group remove <组>`：它会删除该组的投影、配置记录、规则文件和索引项。只取消单个 skill 用 `skillcloud unuse <别名或ID>`；如果该 skill 属于某个组，优先用 `group remove`，这样组记录和规则也会一起清理。

组规则是分作用域的。项目根 `AGENTS.md` 只列出启用了哪些组、使用该组技能前该读哪个规则文件；不会把所有规则平铺到每个任务。工作 SOP 和个人学习组可以同时启用，任务跨组时再显式读取并协调各自规则。

## Skill 格式

每个 skill 都是一个包含 `SKILL.md` 的目录。

```markdown
---
name: code-review
description: Review code changes for bugs, regressions, and missing tests.
---

Use this skill when reviewing code changes.
```

skill ID 就是它在 `skills/` 目录下的路径：

```text
skills/coding/code-review/SKILL.md -> coding/code-review
skills/stock/risk-control/SKILL.md -> stock/risk-control
```

## 支持的 Agent

| Target | 全局作用域 | 项目作用域 |
| --- | --- | --- |
| Codex | `~/.codex/skills` | `.agents/skills` |
| Claude Code | `~/.claude/skills` | `.claude/skills` |
| Hermes | `~/.hermes/skills` | `skills` |

## 安装模式

| 模式 | 行为 | 适合场景 |
| --- | --- | --- |
| `link` | 创建软链接；如果软链接失败，则回退到复制。 | 希望项目快速跟随本地 skills 缓存更新。 |
| `copy` | 将 skill 目录复制到目标路径。 | 希望项目更自包含，或环境不适合软链接。 |

## 项目配置

项目启用信息记录在 `.skillcloud.yaml`：

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

- `id` 是 skills 仓库里的分类路径。
- `as` 是投影到 agent 目录后的平铺目录名。
- skill 上的 `group` / `groups` 记录组归属；多个组共用一个 skill 时使用列表形式。
- target 上的 `groups` 记录当前启用的组，以及需要维护的规则文件和索引。

## 命令

```bash
skillcloud connect <repo-url>
skillcloud update
skillcloud push -m "message"
skillcloud status

skillcloud list
skillcloud search <query>
skillcloud current            # 当前项目启用的 skill 和组
skillcloud                    # 启动主 TUI

skillcloud add <path> --as <skill-id>
skillcloud use <skill-id...> --target codex --scope project
skillcloud use coding/* --target claude --scope project
skillcloud use --select --target hermes --scope global
skillcloud unuse <alias-or-skill-id...> --target codex
skillcloud group list
skillcloud group show <group>
skillcloud group use <group> --target codex --mode link
skillcloud group remove <group> --target codex
skillcloud apply

# 兼容别名
skillcloud init <repo-url>    # connect 的别名
skillcloud pull               # update 的别名
skillcloud enable <skill-id...> --target codex --scope project
skillcloud disable <alias-or-skill-id...>

skillcloud validate
skillcloud doctor
```

## 仓库边界

`skillcloud` 不会内置默认 skills catalog。

这样可以保持 CLI 干净：

- 不把个人 prompt 混进工具仓库；
- 不把组织内部 skill 放进公开仓库；
- 不假设所有用户都想要同一套 agent 行为；
- 不需要为了维护私有 skills 而 fork CLI。

建议用一个仓库存放工具，用另一个仓库存放你的 skills。

## 路线图

- 支持 `work`、`personal`、`frontend` 等 profile 批量启用。
- 用模板创建新 skill。
- 安全地重命名和删除 skill。
- 安全地重命名和删除 skill。
- 增加 lockfile，记录来源 commit 和 hash。
- 增强脚本和 secret 扫描。
- 支持 Git LFS 或对象存储，用于大型 skill 资产。
- 支持更多 agent adapter。

## 开发

运行测试：

```bash
go test ./...
```

运行 Windows smoke test：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1
```

## 贡献

欢迎提交 issue 和 pull request。请尽量保持改动聚焦；行为变更请附带测试。

## 许可证

MIT License。详见 [LICENSE](LICENSE)。
