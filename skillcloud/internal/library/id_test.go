package library

import (
	"path/filepath"
	"testing"
)

func TestValidateIDAcceptsCategoryPaths(t *testing.T) {
	tests := []string{
		"coding/code-review",
		"stock/risk-control",
		"writing/product/prd-review",
		"code-review",
	}
	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			if err := ValidateID(id); err != nil {
				t.Fatalf("ValidateID(%q) error = %v", id, err)
			}
		})
	}
}

func TestValidateIDRejectsUnsafePaths(t *testing.T) {
	tests := []string{
		"",
		"/coding/code-review",
		`C:\skills\code-review`,
		"coding//code-review",
		"coding/../code-review",
		"coding/.hidden",
		"coding/code_review",
		"coding/code review",
		"coding/code-review/",
		"coding\\code-review",
	}
	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			if err := ValidateID(id); err == nil {
				t.Fatalf("ValidateID(%q) expected error", id)
			}
		})
	}
}

func TestSkillDirForID(t *testing.T) {
	got, err := SkillDirForID("C:/repo", "coding/code-review")
	if err != nil {
		t.Fatalf("SkillDirForID() error = %v", err)
	}
	want := filepath.Join("C:/repo", "skills", "coding", "code-review")
	if got != want {
		t.Fatalf("SkillDirForID() = %q, want %q", got, want)
	}
}
