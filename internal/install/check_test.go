package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckReplaceAllowsMissingDestination(t *testing.T) {
	root := t.TempDir()
	action := Action{
		Dest:       filepath.Join(root, "code-review"),
		Projection: &ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project"},
	}
	if err := CheckReplace([]Action{action}); err != nil {
		t.Fatalf("CheckReplace() error = %v", err)
	}
}

func TestCheckReplaceAcceptsManagedProjection(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "code-review")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteProjectionManifest(dest, ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "copy"}); err != nil {
		t.Fatal(err)
	}
	action := Action{
		Dest:       dest,
		Projection: &ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project"},
	}
	if err := CheckReplace([]Action{action}); err != nil {
		t.Fatalf("CheckReplace() error = %v", err)
	}
}

func TestCheckReplaceRejectsUnmanagedDirectory(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "code-review")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("manual"), 0o644); err != nil {
		t.Fatal(err)
	}
	action := Action{
		Dest:       dest,
		Projection: &ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project"},
	}
	if err := CheckReplace([]Action{action}); err == nil {
		t.Fatal("expected unmanaged destination error")
	}
	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); err != nil {
		t.Fatalf("manual skill should remain: %v", err)
	}
}

func TestCheckReplaceRejectsDifferentProjection(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "code-review")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteProjectionManifest(dest, ProjectionManifest{SourceID: "other/review", Target: "codex", Scope: "project", Mode: "copy"}); err != nil {
		t.Fatal(err)
	}
	action := Action{
		Dest:       dest,
		Projection: &ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project"},
	}
	if err := CheckReplace([]Action{action}); err == nil {
		t.Fatal("expected projection mismatch error")
	}
}

func TestCheckReplaceAdoptsIdenticalCopy(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dest := filepath.Join(root, "dest")
	for _, dir := range []string{src, dest} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: review\ndescription: test\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	action := Action{
		Source:     src,
		Dest:       dest,
		Projection: &ProjectionManifest{SourceID: "coding/review", Target: "codex", Scope: "project"},
	}
	if err := CheckReplace([]Action{action}); err != nil {
		t.Fatalf("identical copy should be adoptable: %v", err)
	}
}

func TestCheckReplaceRejectsDifferentCopy(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dest := filepath.Join(root, "dest")
	for _, dir := range []string{src, dest} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("source"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("different"), 0o644); err != nil {
		t.Fatal(err)
	}
	action := Action{
		Source:     src,
		Dest:       dest,
		Projection: &ProjectionManifest{SourceID: "coding/review", Target: "codex", Scope: "project"},
	}
	if err := CheckReplace([]Action{action}); err == nil {
		t.Fatal("expected different copy rejection")
	}
}
