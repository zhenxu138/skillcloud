package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRepoMissingSkillsDirectory(t *testing.T) {
	errs := Repo(t.TempDir())
	if len(errs) == 0 {
		t.Fatal("expected validation error")
	}
}

func TestValidateRepoWithSkill(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, "skills", "coding", "code-review")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: code-review\ndescription: Review code\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	errs := Repo(repo)
	if len(errs) != 0 {
		t.Fatalf("unexpected validation errors %#v", errs)
	}
}

