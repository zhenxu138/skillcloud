package target

import (
	"path/filepath"
	"testing"
)

func TestBuiltIns(t *testing.T) {
	for _, name := range []string{"codex", "claude", "hermes"} {
		got, ok := BuiltIn(name)
		if !ok {
			t.Fatalf("missing target %s", name)
		}
		if got.Global == "" || got.Project == "" {
			t.Fatalf("target %s has empty paths", name)
		}
	}
}

func TestInstallPath(t *testing.T) {
	got := InstallPath(Target{Name: "codex", Project: ".agents/skills"}, "project", "C:/repo")
	want := filepath.Join("C:/repo", ".agents", "skills")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
