package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skillcloud/skillcloud/internal/config"
)

func TestAddCommandCopiesSkillIntoLibrary(t *testing.T) {
	home := t.TempDir()
	repoDir := t.TempDir()
	src := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	writeTestSkill(t, src, ".", "code-review", "Review code.")

	cfg := config.DefaultConfig("local")
	cfg.RepoDir = repoDir
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", src, "--as", "coding/code-review"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "skills", "coding", "code-review", "SKILL.md")); err != nil {
		t.Fatalf("expected imported skill: %v", err)
	}
	if !strings.Contains(out.String(), "added coding/code-review") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestAddCommandRejectsInvalidSkill(t *testing.T) {
	home := t.TempDir()
	repoDir := t.TempDir()
	src := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfg := config.DefaultConfig("local")
	cfg.RepoDir = repoDir
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", src, "--as", "coding/code-review"})
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "SKILL.md not found") {
		t.Fatalf("Execute() error = %v, want SKILL.md not found", err)
	}
	if _, statErr := os.Stat(filepath.Join(repoDir, "skills", "coding", "code-review")); !os.IsNotExist(statErr) {
		t.Fatalf("invalid skill should not be copied, stat err = %v", statErr)
	}
}
