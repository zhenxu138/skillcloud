package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/project"
)

func setupGroupProject(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	repo := t.TempDir()
	projectRoot := t.TempDir()
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

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	return repo, projectRoot
}

func writeGroupFile(t *testing.T, repo string, rel string, content string) {
	t.Helper()
	path := filepath.Join(repo, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestGroupUseAndRemove(t *testing.T) {
	repo, projectRoot := setupGroupProject(t)
	writeTestSkill(t, repo, filepath.Join("skills", "ops", "foo"), "foo", "Foo skill.")
	writeTestSkill(t, repo, filepath.Join("skills", "ops", "bar"), "bar", "Bar skill.")
	writeGroupFile(t, repo, "groups/devops/GROUP.yaml", `name: devops
title: DevOps
entry: ops/foo
agents: AGENTS.md
skills:
  - id: ops/foo
  - id: ops/bar
    as: bar-tool
`)
	writeGroupFile(t, repo, "groups/devops/AGENTS.md", "# DevOps rules\n\nRun checks before reporting.\n")
	if err := os.WriteFile(filepath.Join(projectRoot, "AGENTS.md"), []byte("Manual note.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "group", "use", "devops", "--mode", "copy")
	if err != nil {
		t.Fatalf("group use error = %v\n%s", err, out)
	}
	if !strings.Contains(out, "enabled 2 skill(s) from group devops") {
		t.Fatalf("output = %q", out)
	}
	for _, name := range []string{"foo", "bar-tool"} {
		if _, err := os.Stat(filepath.Join(projectRoot, ".agents", "skills", name, "SKILL.md")); err != nil {
			t.Fatalf("expected projected skill %s: %v", name, err)
		}
	}

	projectConfig, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	target := projectConfig.Targets["codex"]
	if len(target.Skills) != 2 {
		t.Fatalf("skills = %#v", target.Skills)
	}
	for _, ref := range target.Skills {
		if ref.Group != "groups/devops" {
			t.Fatalf("skill %q group = %q", ref.ID, ref.Group)
		}
	}
	if len(target.Groups) != 1 || target.Groups[0].ID != "groups/devops" || target.Groups[0].Name != "devops" {
		t.Fatalf("groups = %#v", target.Groups)
	}

	agentsData, err := os.ReadFile(filepath.Join(projectRoot, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	agents := string(agentsData)
	for _, want := range []string{"Manual note.", "## Skill group: DevOps", "Run checks before reporting.", "skillcloud:group:groups/devops:begin"} {
		if !strings.Contains(agents, want) {
			t.Fatalf("missing %q in AGENTS.md:\n%s", want, agents)
		}
	}

	listOut, err := runCLI(t, "group", "list")
	if err != nil {
		t.Fatalf("group list error = %v", err)
	}
	if !strings.Contains(listOut, "groups/devops") || !strings.Contains(listOut, "codex") {
		t.Fatalf("list output = %q", listOut)
	}
	showOut, err := runCLI(t, "group", "show", "devops")
	if err != nil {
		t.Fatalf("group show error = %v", err)
	}
	for _, want := range []string{"title: DevOps", "ops/foo", "bar-tool"} {
		if !strings.Contains(showOut, want) {
			t.Fatalf("show output missing %q: %q", want, showOut)
		}
	}
	if _, err := runCLI(t, "group", "use", "devops"); err != nil {
		t.Fatalf("second group use should be idempotent: %v", err)
	}
	projectConfig, _ = project.Load(projectRoot)
	if len(projectConfig.Targets["codex"].Skills) != 2 || len(projectConfig.Targets["codex"].Groups) != 1 {
		t.Fatalf("group use duplicated config: %#v", projectConfig.Targets["codex"])
	}

	out, err = runCLI(t, "group", "remove", "devops")
	if err != nil {
		t.Fatalf("group remove error = %v\n%s", err, out)
	}
	if !strings.Contains(out, "removed 2 skill(s)") {
		t.Fatalf("output = %q", out)
	}
	for _, name := range []string{"foo", "bar-tool"} {
		if _, err := os.Lstat(filepath.Join(projectRoot, ".agents", "skills", name)); !os.IsNotExist(err) {
			t.Fatalf("projection %s should be removed, err = %v", name, err)
		}
	}
	projectConfig, err = project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(projectConfig.Targets["codex"].Skills) != 0 || len(projectConfig.Targets["codex"].Groups) != 0 {
		t.Fatalf("config not cleared: %#v", projectConfig.Targets["codex"])
	}
	agentsData, _ = os.ReadFile(filepath.Join(projectRoot, "AGENTS.md"))
	agents = string(agentsData)
	if strings.Contains(agents, "groups/devops") || strings.Contains(agents, "Run checks before reporting.") {
		t.Fatalf("group block remained:\n%s", agents)
	}
	if !strings.Contains(agents, "Manual note.") {
		t.Fatalf("manual content lost:\n%s", agents)
	}
}

func TestGroupUsePreservesManuallyEnabledSkill(t *testing.T) {
	repo, projectRoot := setupGroupProject(t)
	writeTestSkill(t, repo, filepath.Join("skills", "ops", "foo"), "foo", "Foo skill.")
	writeTestSkill(t, repo, filepath.Join("skills", "ops", "manual"), "manual", "Manual skill.")
	writeGroupFile(t, repo, "groups/demo/GROUP.yaml", `name: demo
skills:
  - id: ops/foo
`)

	if err := project.Save(projectRoot, project.Config{Targets: map[string]project.TargetConfig{
		"codex": {Mode: "copy", Skills: []project.SkillRef{{ID: "ops/manual", As: "manual"}}},
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "group", "use", "demo"); err != nil {
		t.Fatalf("group use error = %v", err)
	}
	if _, err := runCLI(t, "group", "remove", "demo"); err != nil {
		t.Fatalf("group remove error = %v", err)
	}
	projectConfig, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	skills := projectConfig.Targets["codex"].Skills
	if len(skills) != 1 || skills[0].ID != "ops/manual" || skills[0].Group != "" {
		t.Fatalf("manual skill not preserved: %#v", skills)
	}
}

func TestGroupUseRulesOnly(t *testing.T) {
	repo, projectRoot := setupGroupProject(t)
	writeGroupFile(t, repo, "groups/rules/GROUP.yaml", "name: rules\ntitle: Rules\n")
	writeGroupFile(t, repo, "groups/rules/.codex/agent.md", "Always run tests.\n")

	out, err := runCLI(t, "group", "use", "rules")
	if err != nil {
		t.Fatalf("group use error = %v\n%s", err, out)
	}
	if !strings.Contains(out, "enabled 0 skill(s)") {
		t.Fatalf("output = %q", out)
	}
	projectConfig, _ := project.Load(projectRoot)
	if len(projectConfig.Targets["codex"].Groups) != 1 {
		t.Fatalf("groups = %#v", projectConfig.Targets["codex"].Groups)
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Always run tests.") {
		t.Fatalf("rules missing:\n%s", string(data))
	}
}

func TestGroupUseRefusesUnmanagedProjection(t *testing.T) {
	repo, projectRoot := setupGroupProject(t)
	writeTestSkill(t, repo, filepath.Join("skills", "ops", "foo"), "foo", "Foo skill.")
	writeGroupFile(t, repo, "groups/demo/GROUP.yaml", `name: demo
skills:
  - id: ops/foo
`)
	unmanaged := filepath.Join(projectRoot, ".agents", "skills", "foo")
	if err := os.MkdirAll(unmanaged, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unmanaged, "SKILL.md"), []byte("manual"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := runCLI(t, "group", "use", "demo")
	if err == nil || !strings.Contains(err.Error(), "not managed by skillcloud") {
		t.Fatalf("expected unmanaged projection error, got %v", err)
	}
	data, err := os.ReadFile(filepath.Join(unmanaged, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "manual" {
		t.Fatalf("unmanaged skill changed: %q", string(data))
	}
}

func TestGroupUseNoAgentsPersistsThroughApply(t *testing.T) {
	repo, projectRoot := setupGroupProject(t)
	writeTestSkill(t, repo, filepath.Join("skills", "ops", "foo"), "foo", "Foo skill.")
	writeGroupFile(t, repo, "groups/demo/GROUP.yaml", `name: demo
agents: AGENTS.md
skills:
  - id: ops/foo
`)
	writeGroupFile(t, repo, "groups/demo/AGENTS.md", "Rule.\n")
	if _, err := runCLI(t, "group", "use", "demo", "--no-agents"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md should not be created, err = %v", err)
	}
	if _, err := runCLI(t, "apply"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("apply should keep rules suppressed, err = %v", err)
	}
}

func TestApplyRefreshesGroupAgents(t *testing.T) {
	repo, projectRoot := setupGroupProject(t)
	writeTestSkill(t, repo, filepath.Join("skills", "ops", "foo"), "foo", "Foo skill.")
	writeGroupFile(t, repo, "groups/demo/GROUP.yaml", `name: demo
agents: AGENTS.md
skills:
  - id: ops/foo
`)
	writeGroupFile(t, repo, "groups/demo/AGENTS.md", "Old rule.\n")
	if _, err := runCLI(t, "group", "use", "demo", "--mode", "copy"); err != nil {
		t.Fatal(err)
	}

	writeGroupFile(t, repo, "groups/demo/AGENTS.md", "New rule.\n")
	if _, err := runCLI(t, "apply"); err != nil {
		t.Fatalf("apply error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "New rule.") || strings.Contains(string(data), "Old rule.") {
		t.Fatalf("AGENTS.md not refreshed:\n%s", string(data))
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".agents", "skills", "foo", "SKILL.md")); err != nil {
		t.Fatalf("skill not re-applied: %v", err)
	}
}
