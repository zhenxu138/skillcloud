package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skillcloud/skillcloud/internal/tui"
)

func TestRootCommandHasExpectedSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}

	for _, want := range []string{"init", "pull", "push", "status", "list", "search", "browse", "enable", "disable", "apply", "doctor", "validate"} {
		if !names[want] {
			t.Fatalf("missing subcommand %q", want)
		}
	}
}

func TestRootCommandRegistersFriendlyAliases(t *testing.T) {
	cmd := NewRootCommand()
	names := map[string]bool{}
	for _, child := range cmd.Commands() {
		names[child.Name()] = true
	}
	for _, name := range []string{"connect", "update", "use", "unuse"} {
		if !names[name] {
			t.Fatalf("expected root command to register %q", name)
		}
	}
}

func TestRootCommandHelp(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected help output")
	}
}

func TestSyncCommandsRejectExtraArgs(t *testing.T) {
	for _, args := range [][]string{
		{"pull", "extra"},
		{"push", "extra"},
		{"status", "extra"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := NewRootCommand()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() error = nil, want error")
			}
			if !strings.Contains(err.Error(), "extra") {
				t.Fatalf("Execute() error = %q, want extra args error", err)
			}
			if strings.Contains(err.Error(), ".skillcloud") {
				t.Fatalf("Execute() error = %q, want cobra arg validation before config load", err)
			}
		})
	}
}

func TestInitDoesNotSaveConfigWhenCloneFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cmd := NewRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"init", filepath.Join(home, "missing.git")})

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want clone error")
	}

	configPath := filepath.Join(home, ".skillcloud", "config.yaml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("config file exists after failed clone, Stat error = %v", err)
	}
}

func TestRootCommandNoArgsRunsMainTUI(t *testing.T) {
	oldRunner := runMainTUI
	called := false
	runMainTUI = func(opts tui.AppOptions) error {
		called = true
		if opts.Target != "codex" || opts.Scope != "project" {
			t.Fatalf("opts = %#v", opts)
		}
		return nil
	}
	t.Cleanup(func() { runMainTUI = oldRunner })

	cmd := NewRootCommand()
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("expected main TUI runner to be called")
	}
}
