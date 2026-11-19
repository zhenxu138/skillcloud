package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skillcloud/skillcloud/internal/project"
)

func TestDisableDefaultsToCodex(t *testing.T) {
	repo, projectRoot := setupGroupProject(t)
	writeTestSkill(t, repo, filepath.Join("skills", "sop", "demo"), "demo", "Demo skill.")

	if _, err := runCLI(t, "use", "sop/demo", "--target", "codex", "--mode", "copy"); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "use", "sop/demo", "--target", "claude", "--mode", "copy"); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "disable", "demo"); err != nil {
		t.Fatal(err)
	}

	cfg, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Targets["codex"].Skills) != 0 {
		t.Fatalf("codex skills = %#v", cfg.Targets["codex"].Skills)
	}
	if len(cfg.Targets["claude"].Skills) != 1 {
		t.Fatalf("claude skills = %#v", cfg.Targets["claude"].Skills)
	}
	if _, err := os.Lstat(filepath.Join(projectRoot, ".agents", "skills", "demo")); !os.IsNotExist(err) {
		t.Fatalf("codex projection should be removed, err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".claude", "skills", "demo", "SKILL.md")); err != nil {
		t.Fatalf("claude projection should remain: %v", err)
	}
}
