package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkillMarkdown(t *testing.T, dir string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSkillAcceptsSkillCreatorCompatibleFrontmatter(t *testing.T) {
	dir := t.TempDir()
	writeSkillMarkdown(t, dir, "---\nname: code-review\ndescription: Review code changes.\nlicense: MIT\nallowed-tools: Read, Grep\nmetadata:\n  short-description: Review code\n---\nBody\n")

	meta, errs := Skill(dir)
	if len(errs) != 0 {
		t.Fatalf("Skill() errors = %#v", errs)
	}
	if meta.Name != "code-review" {
		t.Fatalf("Name = %q, want code-review", meta.Name)
	}
	if meta.Description != "Review code changes." {
		t.Fatalf("Description = %q", meta.Description)
	}
}

func TestSkillRejectsMissingSkillMarkdown(t *testing.T) {
	_, errs := Skill(t.TempDir())
	if len(errs) == 0 || !strings.Contains(errs[0].Error(), "SKILL.md not found") {
		t.Fatalf("errors = %#v, want missing SKILL.md", errs)
	}
}

func TestSkillRejectsUnknownFrontmatterKey(t *testing.T) {
	dir := t.TempDir()
	writeSkillMarkdown(t, dir, "---\nname: code-review\ndescription: Review code.\nversion: 1.0.0\n---\nBody\n")

	_, errs := Skill(dir)
	if len(errs) == 0 || !strings.Contains(errs[0].Error(), "unexpected key") {
		t.Fatalf("errors = %#v, want unexpected key", errs)
	}
}

func TestSkillRejectsInvalidName(t *testing.T) {
	tests := []string{"CodeReview", "-code-review", "code--review", "code-review-", "code_review"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeSkillMarkdown(t, dir, "---\nname: "+name+"\ndescription: Review code.\n---\nBody\n")

			_, errs := Skill(dir)
			if len(errs) == 0 {
				t.Fatalf("expected invalid name error for %q", name)
			}
		})
	}
}

func TestSkillRejectsInvalidDescription(t *testing.T) {
	dir := t.TempDir()
	writeSkillMarkdown(t, dir, "---\nname: code-review\ndescription: \"Review <code>.\"\n---\nBody\n")

	_, errs := Skill(dir)
	if len(errs) == 0 || !strings.Contains(errs[0].Error(), "angle brackets") {
		t.Fatalf("errors = %#v, want angle bracket rejection", errs)
	}
}

func TestSkillAcceptsCommonSkillHubKeys(t *testing.T) {
	dir := t.TempDir()
	writeSkillMarkdown(t, dir, "---\nname: ask-matt\ndescription: Ask which skill fits.\ndisable-model-invocation: true\nargument-hint: \"What do you want to ask?\"\n---\nBody\n")

	meta, errs := Skill(dir)
	if len(errs) != 0 {
		t.Fatalf("Skill() errors = %#v", errs)
	}
	if meta.Name != "ask-matt" {
		t.Fatalf("Name = %q, want ask-matt", meta.Name)
	}
}

func TestSkillAcceptsCRLFLineEndings(t *testing.T) {
	dir := t.TempDir()
	writeSkillMarkdown(t, dir, "---\r\nname: code-review\r\ndescription: Review code.\r\n---\r\nBody\r\n")

	meta, errs := Skill(dir)
	if len(errs) != 0 {
		t.Fatalf("Skill() errors = %#v", errs)
	}
	if meta.Name != "code-review" {
		t.Fatalf("Name = %q, want code-review", meta.Name)
	}
}

