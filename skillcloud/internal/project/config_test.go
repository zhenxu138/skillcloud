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

