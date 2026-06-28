package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/project"
)

func TestUseCommandCopiesSkillAndWritesProjectConfig(t *testing.T) {
	home := t.TempDir()
	repoDir := t.TempDir()
	projectRoot := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	writeTestSkill(t, repoDir, filepath.Join("skills", "coding", "code-review"), "code-review", "Review code.")
	cfg := config.DefaultConfig("local")
	cfg.RepoDir = repoDir
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"use", "coding/code-review", "--target", "codex", "--scope", "project", "--mode", "copy"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".agents", "skills", "code-review", "SKILL.md")); err != nil {
		t.Fatalf("expected projected skill: %v", err)
	}
}

func TestDefaultAlias(t *testing.T) {
	tests := map[string]string{
		"coding/code-review": "code-review",
		"stock/risk-control": "risk-control",
	}
	for id, want := range tests {
		if got := defaultAlias(id); got != want {
			t.Fatalf("defaultAlias(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestEnableCommandCopiesSkillAndWritesProjectConfig(t *testing.T) {
	home := t.TempDir()
	repoDir := t.TempDir()
	projectRoot := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	writeTestSkill(t, repoDir, filepath.Join("skills", "coding", "code-review"), "code-review", "Review code.")
	cfg := config.DefaultConfig("local")
	cfg.RepoDir = repoDir
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"enable", "coding/code-review", "--target", "codex", "--scope", "project", "--mode", "copy"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectRoot, ".agents", "skills", "code-review", "SKILL.md")); err != nil {
		t.Fatalf("expected projected skill: %v", err)
	}
	projectConfig, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	got := projectConfig.Targets["codex"].Skills
	if len(got) != 1 || got[0].ID != "coding/code-review" || got[0].As != "code-review" {
		t.Fatalf("unexpected project config refs %#v", got)
	}
}

func writeTestSkill(t *testing.T, repoDir string, relDir string, name string, description string) {
	t.Helper()
	dir := filepath.Join(repoDir, relDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	data := "---\nname: " + name + "\ndescription: " + description + "\n---\nBody\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDisableRefusesUnmanagedProjection(t *testing.T) {
	home := t.TempDir()
	projectRoot := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfg := config.DefaultConfig("local")
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	projectConfig := project.Config{Targets: map[string]project.TargetConfig{
		"codex": {
			Mode:   "copy",
			Skills: []project.SkillRef{{ID: "coding/code-review", As: "code-review"}},
		},
	}}
	if err := project.Save(projectRoot, projectConfig); err != nil {
		t.Fatal(err)
	}
	unmanaged := filepath.Join(projectRoot, ".agents", "skills", "code-review")
	if err := os.MkdirAll(unmanaged, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unmanaged, "SKILL.md"), []byte("local"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, removed, err := disableProjectSkills(projectRoot, cfg, projectConfig, []string{"coding/code-review"})
	if err == nil {
		t.Fatal("expected unmanaged projection error")
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0", removed)
	}
	if _, err := os.Stat(filepath.Join(unmanaged, "SKILL.md")); err != nil {
		t.Fatalf("unmanaged skill should remain: %v", err)
	}
}

