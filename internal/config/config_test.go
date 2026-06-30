package config

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestExpandHomeUsesHomeEnvironment(t *testing.T) {
	t.Setenv("HOME", filepath.Join("tmp", "home"))
	t.Setenv("USERPROFILE", filepath.Join("tmp", "profile"))

	got, err := ExpandHome("~/.skillcloud/repo")
	if err != nil {
		t.Fatalf("ExpandHome() error = %v", err)
	}
	wantHome := filepath.Join("tmp", "home")
	if runtime.GOOS == "windows" {
		wantHome = filepath.Join("tmp", "profile")
	}
	want := filepath.Join(wantHome, ".skillcloud", "repo")
	if got != want {
		t.Fatalf("ExpandHome() = %q, want %q", got, want)
	}
}

func TestExpandHomeFallsBackToUserProfile(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", filepath.Join("tmp", "profile"))

	got, err := ExpandHome("~\\.skillcloud\\repo")
	if err != nil {
		t.Fatalf("ExpandHome() error = %v", err)
	}
	want := filepath.Join("tmp", "profile", ".skillcloud", "repo")
	if got != want {
		t.Fatalf("ExpandHome() = %q, want %q", got, want)
	}
}

func TestHomeDirReturnsErrorWhenUnset(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	if _, err := HomeDir(); err == nil {
		t.Fatal("HomeDir() error = nil, want error")
	}
}

func TestExpandHomeReturnsErrorWhenHomeUnset(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	if _, err := ExpandHome("~/.skillcloud/repo"); err == nil {
		t.Fatal("ExpandHome() error = nil, want error")
	}
}

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig("git@github.com:me/skills.git")
	want := Config{
		RepoURL:     "git@github.com:me/skills.git",
		RepoDir:     "~/.skillcloud/repo",
		DefaultMode: "link",
		Targets: map[string]TargetPath{
			"codex": {
				Global:  "~/.codex/skills",
				Project: ".agents/skills",
			},
			"claude": {
				Global:  "~/.claude/skills",
				Project: ".claude/skills",
			},
			"hermes": {
				Global:  "~/.hermes/skills",
				Project: "skills",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultConfig() = %#v, want %#v", got, want)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	want := DefaultConfig("git@github.com:me/skills.git")

	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}
