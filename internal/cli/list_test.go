package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/skill"
)

func TestListCommandPrintsIndexedSkills(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	saveTestIndex(t)

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "coding/code-review") {
		t.Fatalf("list output = %q, want skill id", out.String())
	}
}

func TestSearchCommandPrintsMatches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	saveTestIndex(t)

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"search", "risk"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.Contains(out.String(), "coding/code-review") {
		t.Fatalf("search output = %q, did not expect code-review", out.String())
	}
	if !strings.Contains(out.String(), "stock/risk-control") {
		t.Fatalf("search output = %q, want risk-control", out.String())
	}
}

func TestListCommandBuildsFreshIndexAndKeepsFooterOnStderr(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfg := config.DefaultConfig("local")
	cfg.RepoDir = repo
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(repo, "skills", "sop", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: demo\ndescription: Demo skill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	indexPath, err := config.DefaultIndexPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := skill.SaveIndex(indexPath, skill.Index{}); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "sop/demo") {
		t.Fatalf("stdout = %q, want fresh repo skill", stdout.String())
	}
	if strings.Contains(stdout.String(), "共 ") {
		t.Fatalf("footer leaked into stdout: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "共 1 个 skill") {
		t.Fatalf("stderr = %q, want footer", stderr.String())
	}
}

func saveTestIndex(t *testing.T) {
	t.Helper()
	indexPath, err := config.DefaultIndexPath()
	if err != nil {
		t.Fatal(err)
	}
	index := skill.Index{Skills: []skill.Skill{
		{ID: "coding/code-review", Name: "code-review", Description: "Review code changes"},
		{ID: "stock/risk-control", Name: "risk-control", Description: "Check trading risk"},
	}}
	if err := skill.SaveIndex(indexPath, index); err != nil {
		t.Fatal(err)
	}
}
