package project

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadProjectConfig(t *testing.T) {
	root := t.TempDir()
	cfg := Config{Targets: map[string]TargetConfig{
		"codex": {
			Mode: "link",
			Skills: []SkillRef{
				{ID: "coding/code-review", As: "code-review"},
			},
		},
	}}
	if err := Save(root, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Targets["codex"].Skills[0].ID != "coding/code-review" {
		t.Fatalf("unexpected config %#v", got)
	}
}

func TestLoadMissingProjectConfig(t *testing.T) {
	got, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Targets == nil {
		t.Fatal("expected initialized Targets map")
	}
}

func TestConfigPath(t *testing.T) {
	root := filepath.Join("C:", "repo")
	if ConfigPath(root) != filepath.Join(root, ".skillcloud.yaml") {
		t.Fatalf("unexpected config path")
	}
}

func TestSkillRefGroupOwnershipMigration(t *testing.T) {
	legacy := SkillRef{ID: "sop/work/shared", As: "shared", Group: "skills/sop/work"}
	if !legacy.HasGroup("skills/sop/work") {
		t.Fatal("legacy group field should be readable")
	}
	legacy.AddGroup("skills/sop/matt-pocock")
	if len(legacy.GroupIDs()) != 2 || !legacy.HasGroup("skills/sop/matt-pocock") {
		t.Fatalf("groups = %#v", legacy.GroupIDs())
	}
	legacy.RemoveGroup("skills/sop/work")
	if len(legacy.GroupIDs()) != 1 || legacy.GroupIDs()[0] != "skills/sop/matt-pocock" {
		t.Fatalf("groups = %#v", legacy.GroupIDs())
	}
	if legacy.Group != "" {
		t.Fatalf("legacy group field should be cleared, got %q", legacy.Group)
	}
}
