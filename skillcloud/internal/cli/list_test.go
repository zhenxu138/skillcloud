package cli

import (
	"bytes"
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

