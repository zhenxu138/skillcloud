package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/skillcloud/skillcloud/internal/project"
)

func TestCurrentCommandListsSkillsAndGroups(t *testing.T) {
	projectRoot := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	if err := project.Save(projectRoot, project.Config{Targets: map[string]project.TargetConfig{
		"codex": {
			Mode: "link",
			Skills: []project.SkillRef{
				{ID: "sop/demo/code-review", As: "code-review", Group: "skills/sop/demo"},
				{ID: "ops/collect", As: "collect"},
			},
			Groups: []project.GroupRef{{ID: "skills/sop/demo", Name: "demo"}},
		},
	}}); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"current"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	text := out.String()
	for _, want := range []string{
		"target: codex",
		"mode: link",
		"skills: 2",
		"code-review",
		"group=skills/sop/demo",
		"collect",
		"group=-",
		"groups: 1",
		"demo",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
}

func TestCurrentCommandOnBlankProject(t *testing.T) {
	projectRoot := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"current"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "no skills or groups configured") {
		t.Fatalf("output = %q", out.String())
	}
}
