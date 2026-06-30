package target

import "path/filepath"

type Target struct {
	Name    string
	Global  string
	Project string
}

func BuiltIn(name string) (Target, bool) {
	targets := map[string]Target{
		"codex":  {Name: "codex", Global: "~/.codex/skills", Project: ".agents/skills"},
		"claude": {Name: "claude", Global: "~/.claude/skills", Project: ".claude/skills"},
		"hermes": {Name: "hermes", Global: "~/.hermes/skills", Project: "skills"},
	}
	t, ok := targets[name]
	return t, ok
}

func InstallPath(t Target, scope string, projectRoot string) string {
	if scope == "global" {
		return filepath.Clean(t.Global)
	}
	return filepath.Join(projectRoot, filepath.FromSlash(t.Project))
}

