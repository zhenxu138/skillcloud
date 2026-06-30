package gitstore

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseStatusOutputEmptyStringIsClean(t *testing.T) {
	got := ParseStatusOutput("")

	if got.Dirty {
		t.Fatal("ParseStatusOutput(\"\").Dirty = true, want false")
	}
	if len(got.Lines) != 0 {
		t.Fatalf("ParseStatusOutput(\"\").Lines length = %d, want 0", len(got.Lines))
	}
}

func TestParseStatusOutputDirtyLines(t *testing.T) {
	out := " M skills/coding/code-review/SKILL.md\n?? skills/new/SKILL.md\n"

	got := ParseStatusOutput(out)

	if !got.Dirty {
		t.Fatal("ParseStatusOutput(dirty output).Dirty = false, want true")
	}
	if len(got.Lines) != 2 {
		t.Fatalf("ParseStatusOutput(dirty output).Lines length = %d, want 2", len(got.Lines))
	}
}

func TestPushCleanAheadRepoPushesLocalCommit(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	remote := filepath.Join(tempDir, "remote.git")
	local := filepath.Join(tempDir, "local")

	runGitTest(t, "", "init", "--bare", remote)
	runGitTest(t, "", "clone", remote, local)
	configureGitUser(t, local)
	writeFile(t, filepath.Join(local, "SKILL.md"), "name: test\n")
	runGitTest(t, local, "add", "SKILL.md")
	runGitTest(t, local, "commit", "-m", "add skill")

	store := Store{RepoDir: local, RepoURL: remote}
	if err := store.Push(ctx, "update skills"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	out := runGitTest(t, "", "--git-dir", remote, "log", "--oneline", "--all")
	if !strings.Contains(out, "add skill") {
		t.Fatalf("remote log = %q, want pushed commit", out)
	}
}

func TestCloneExistingNonGitDirectoryErrors(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	store := Store{RepoDir: repoDir, RepoURL: filepath.Join(tempDir, "remote.git")}
	err := store.Clone(context.Background())

	if err == nil {
		t.Fatal("Clone() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("Clone() error = %q, want not a git repository", err)
	}
}

func TestCloneExistingRepoOriginMismatchErrors(t *testing.T) {
	tempDir := t.TempDir()
	remoteA := filepath.Join(tempDir, "remote-a.git")
	remoteB := filepath.Join(tempDir, "remote-b.git")
	local := filepath.Join(tempDir, "local")

	runGitTest(t, "", "init", "--bare", remoteA)
	runGitTest(t, "", "init", "--bare", remoteB)
	runGitTest(t, "", "clone", remoteA, local)

	store := Store{RepoDir: local, RepoURL: remoteB}
	err := store.Clone(context.Background())

	if err == nil {
		t.Fatal("Clone() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "origin URL mismatch") {
		t.Fatalf("Clone() error = %q, want origin URL mismatch", err)
	}
}

func configureGitUser(t *testing.T, dir string) {
	t.Helper()
	runGitTest(t, dir, "config", "user.email", "test@example.com")
	runGitTest(t, dir, "config", "user.name", "Test User")
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func runGitTest(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
