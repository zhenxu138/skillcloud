package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/tui"
)

func TestDiffBrowseSelection(t *testing.T) {
	enable, disable := diffBrowseSelection(
		[]string{"coding/code-review", "stock/risk-control"},
		[]string{"coding/code-review", "coding/go-code-review"},
	)

	if want := []string{"coding/go-code-review"}; !reflect.DeepEqual(enable, want) {
		t.Fatalf("enable = %#v, want %#v", enable, want)
	}
	if want := []string{"stock/risk-control"}; !reflect.DeepEqual(disable, want) {
		t.Fatalf("disable = %#v, want %#v", disable, want)
	}
}

func setupBrowseCommandTest(t *testing.T) (string, string, string) {
	t.Helper()
	home := t.TempDir()
	repoDir := t.TempDir()
	projectRoot := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	writeTestSkill(t, repoDir, filepath.Join("skills", "coding", "code-review"), "code-review", "Review code.")
	writeTestSkill(t, repoDir, filepath.Join("skills", "coding", "go-code-review"), "go-code-review", "Review Go code.")

	cfg := config.DefaultConfig("local")
	cfg.RepoDir = repoDir
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	return home, repoDir, projectRoot
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

func TestBrowseCommandAppliesEnable(t *testing.T) {
	_, _, projectRoot := setupBrowseCommandTest(t)
	withWorkingDir(t, projectRoot)

	oldRunner := runBrowseTUI
	runBrowseTUI = func(opts tui.ManageOptions) (tui.ManageResult, error) {
		if opts.Target != "codex" || opts.Scope != "project" || opts.Mode != "link" {
			t.Fatalf("unexpected options: %#v", opts)
		}
		return tui.ManageResult{
			Apply:       true,
			SelectedIDs: []string{"coding/code-review"},
			InitialIDs:  nil,
		}, nil
	}
	t.Cleanup(func() { runBrowseTUI = oldRunner })

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"browse"})
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
	if len(got) != 1 || got[0].ID != "coding/code-review" {
		t.Fatalf("unexpected project config refs %#v", got)
	}
}

func TestBrowseCommandAppliesDisable(t *testing.T) {
	_, _, projectRoot := setupBrowseCommandTest(t)
	withWorkingDir(t, projectRoot)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"enable", "coding/code-review", "--target", "codex", "--scope", "project", "--mode", "copy"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed enable error = %v", err)
	}

	oldRunner := runBrowseTUI
	runBrowseTUI = func(opts tui.ManageOptions) (tui.ManageResult, error) {
		return tui.ManageResult{
			Apply:       true,
			SelectedIDs: nil,
			InitialIDs:  []string{"coding/code-review"},
		}, nil
	}
	t.Cleanup(func() { runBrowseTUI = oldRunner })

	cmd = NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"browse"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectRoot, ".agents", "skills", "code-review")); !os.IsNotExist(err) {
		t.Fatalf("expected projected skill removed, stat err = %v", err)
	}
	projectConfig, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if got := projectConfig.Targets["codex"].Skills; len(got) != 0 {
		t.Fatalf("expected no configured skills, got %#v", got)
	}
}

func TestBrowseCommandCancelLeavesProjectUnchanged(t *testing.T) {
	_, _, projectRoot := setupBrowseCommandTest(t)
	withWorkingDir(t, projectRoot)

	oldRunner := runBrowseTUI
	runBrowseTUI = func(opts tui.ManageOptions) (tui.ManageResult, error) {
		return tui.ManageResult{
			Apply:       false,
			SelectedIDs: []string{"coding/code-review"},
			InitialIDs:  nil,
		}, nil
	}
	t.Cleanup(func() { runBrowseTUI = oldRunner })

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"browse"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectRoot, ".skillcloud.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected no project config after cancel, stat err = %v", err)
	}
}

func TestBrowseCommandRejectsGlobalScope(t *testing.T) {
	_, _, projectRoot := setupBrowseCommandTest(t)
	withWorkingDir(t, projectRoot)

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"browse", "--scope", "global"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected global scope error")
	}
	if got := err.Error(); got != "browse --scope global is not implemented yet; use skillcloud enable ... --scope global or skillcloud disable ..." {
		t.Fatalf("error = %q", got)
	}
}

func TestBrowseCommandEmptyIndexDoesNotLaunchTUI(t *testing.T) {
	home := t.TempDir()
	repoDir := t.TempDir()
	projectRoot := t.TempDir()
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
	withWorkingDir(t, projectRoot)

	oldRunner := runBrowseTUI
	runBrowseTUI = func(opts tui.ManageOptions) (tui.ManageResult, error) {
		t.Fatal("TUI should not run for an empty index")
		return tui.ManageResult{}, nil
	}
	t.Cleanup(func() { runBrowseTUI = oldRunner })

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"browse"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != "no skills found\n" {
		t.Fatalf("output = %q", got)
	}
}
