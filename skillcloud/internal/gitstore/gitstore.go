package gitstore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Status struct {
	Dirty bool
	Lines []string
}

func ParseStatusOutput(out string) Status {
	out = strings.TrimRight(out, "\r\n")
	if out == "" {
		return Status{}
	}

	lines := strings.Split(out, "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	return Status{
		Dirty: len(lines) > 0,
		Lines: lines,
	}
}

type Store struct {
	RepoDir string
	RepoURL string
}

func (s Store) Clone(ctx context.Context) error {
	if _, err := os.Stat(s.RepoDir); err == nil {
		return s.verifyExistingRepo(ctx)
	} else if !os.IsNotExist(err) {
		return err
	}

	if parent := filepath.Dir(s.RepoDir); parent != "." {
		if err := os.MkdirAll(parent, 0755); err != nil {
			return err
		}
	}
	return runGit(ctx, "", "clone", s.RepoURL, s.RepoDir)
}

func (s Store) Pull(ctx context.Context) error {
	return runGit(ctx, s.RepoDir, "pull", "--ff-only")
}

func (s Store) Push(ctx context.Context, message string) error {
	status, err := s.Status(ctx)
	if err != nil {
		return err
	}
	if status.Dirty {
		if err := runGit(ctx, s.RepoDir, "add", "-A"); err != nil {
			return err
		}
		if err := runGit(ctx, s.RepoDir, "commit", "-m", message); err != nil {
			return err
		}
	}
	return runGit(ctx, s.RepoDir, "push")
}

func (s Store) Status(ctx context.Context) (Status, error) {
	out, err := runGitOutput(ctx, s.RepoDir, "status", "--short")
	if err != nil {
		return Status{}, err
	}
	return ParseStatusOutput(out), nil
}

func runGit(ctx context.Context, dir string, args ...string) error {
	_, err := runGitOutput(ctx, dir, args...)
	return err
}

func (s Store) verifyExistingRepo(ctx context.Context) error {
	if _, err := runGitOutput(ctx, s.RepoDir, "rev-parse", "--git-dir"); err != nil {
		return fmt.Errorf("%s is not a git repository: %w", s.RepoDir, err)
	}

	origin, err := runGitOutput(ctx, s.RepoDir, "remote", "get-url", "origin")
	if err != nil {
		return fmt.Errorf("%s has no origin remote: %w", s.RepoDir, err)
	}
	origin = strings.TrimSpace(origin)
	if origin != s.RepoURL {
		return fmt.Errorf("origin URL mismatch for %s: got %q, want %q", s.RepoDir, origin, s.RepoURL)
	}
	return nil
}

func runGitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
