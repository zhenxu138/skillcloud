package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skillcloud/skillcloud/internal/skill"
)

func TestInspectReportsMissingSkill(t *testing.T) {
	cfg := Config{Targets: map[string]TargetConfig{
		"codex": {Skills: []SkillRef{{ID: "writing/old-prd-review", As: "old-prd-review"}}},
	}}
	report := Inspect(cfg, []skill.Skill{}, t.TempDir(), "codex", func(dir string) error {
		return os.ErrNotExist
	})
	if len(report.Missing) != 1 || report.Missing[0].ID != "writing/old-prd-review" {
		t.Fatalf("Missing = %#v", report.Missing)
	}
}

func TestInspectReportsUnmanagedLocalSkill(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, ".agents", "skills", "manual")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("manual"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Targets: map[string]TargetConfig{"codex": {}}}
	report := Inspect(cfg, []skill.Skill{}, filepath.Join(root, ".agents", "skills"), "codex", func(dir string) error {
		return os.ErrNotExist
	})
	if len(report.Unmanaged) != 1 || report.Unmanaged[0] != "manual" {
		t.Fatalf("Unmanaged = %#v", report.Unmanaged)
	}
}

func TestInspectIgnoresManagedProjection(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, ".agents", "skills", "code-review")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Targets: map[string]TargetConfig{
		"codex": {Skills: []SkillRef{{ID: "coding/code-review", As: "code-review"}}},
	}}
	managedDir := dest
	report := Inspect(cfg, []skill.Skill{{ID: "coding/code-review"}}, filepath.Join(root, ".agents", "skills"), "codex", func(dir string) error {
		if dir == managedDir {
			return nil
		}
		return os.ErrNotExist
	})
	if len(report.Missing) != 0 || len(report.Unmanaged) != 0 {
		t.Fatalf("report = %#v", report)
	}
}
