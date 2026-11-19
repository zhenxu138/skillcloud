# skillcloud

Read [.codex/agent.md](.codex/agent.md) before changing this repository; it is the canonical instruction set for agents.

Critical rules:

- Run `gofmt`, `go test ./...`, `go vet ./...`, and `go build -o skillcloud ./cmd/skillcloud` before reporting completion.
- Never overwrite an unmanaged skill projection; use `install.CheckReplace`.
- Keep group `AGENTS.md` blocks reversible and preserve user content outside the markers.
- Keep `README.md` and `README.zh-CN.md` in sync.
- Do not push or publish without explicit instruction.
