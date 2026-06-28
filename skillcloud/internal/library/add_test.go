package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSourceSkill(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: code-review\ndescription: Review code changes.\n---\nBody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "check.txt"), []byte("script"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPlanAddRejectsExistingDestinationByDefault(t *testing.T) {
	repo := t.TempDir()
	src := t.TempDir()
	writeSourceSkill(t, src)
	dst := filepath.Join(repo, "skills", "coding", "code-review")
	writeSourceSkill(t, dst)

	_, err := PlanAdd(AddOptions{RepoDir: repo, SourceDir: src, ID: "coding/code-review"})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("PlanAdd() error = %v, want already exists", err)
	}
}

func TestPlanAddAllowsReplace(t *testing.T) {
	repo := t.TempDir()
	src := t.TempDir()
	writeSourceSkill(t, src)
	dst := filepath.Join(repo, "skills", "coding", "code-review")
	writeSourceSkill(t, dst)

	plan, err := PlanAdd(AddOptions{RepoDir: repo, SourceDir: src, ID: "coding/code-review", Replace: true})
	if err != nil {
		t.Fatalf("PlanAdd() error = %v", err)
	}
	if !plan.Replace {
		t.Fatal("expected replace plan")
	}
}

func TestExecuteAddCopiesSkillAndPreservesSource(t *testing.T) {
	repo := t.TempDir()
	src := t.TempDir()
	writeSourceSkill(t, src)

	plan, err := PlanAdd(AddOptions{RepoDir: repo, SourceDir: src, ID: "coding/code-review"})
	if err != nil {
		t.Fatalf("PlanAdd() error = %v", err)
	}
	if err := ExecuteAdd(plan); err != nil {
		t.Fatalf("ExecuteAdd() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "skills", "coding", "code-review", "SKILL.md")); err != nil {
		t.Fatalf("expected copied SKILL.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "skills", "coding", "code-review", "scripts", "check.txt")); err != nil {
		t.Fatalf("expected copied resource file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
		t.Fatalf("source should remain in place: %v", err)
	}
}
