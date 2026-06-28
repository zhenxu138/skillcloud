package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectionManifestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	manifest := ProjectionManifest{
		SourceID:     "coding/code-review",
		Target:       "codex",
		Scope:        "project",
		Mode:         "copy",
		SourceCommit: "abc123",
	}
	if err := WriteProjectionManifest(dir, manifest); err != nil {
		t.Fatalf("WriteProjectionManifest() error = %v", err)
	}
	got, err := ReadProjectionManifest(dir)
	if err != nil {
		t.Fatalf("ReadProjectionManifest() error = %v", err)
	}
	if got.SourceID != manifest.SourceID || got.Target != manifest.Target || got.Scope != manifest.Scope || got.Mode != manifest.Mode {
		t.Fatalf("manifest = %#v, want %#v", got, manifest)
	}
}

func TestCanRemoveProjectionRejectsUnmanagedDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("local"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CanRemoveProjection(dir, ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project"}); err == nil {
		t.Fatal("expected unmanaged projection removal to be rejected")
	}
}

func TestCanRemoveProjectionAcceptsMatchingManifest(t *testing.T) {
	dir := t.TempDir()
	expected := ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "copy"}
	if err := WriteProjectionManifest(dir, expected); err != nil {
		t.Fatal(err)
	}
	if err := CanRemoveProjection(dir, expected); err != nil {
		t.Fatalf("CanRemoveProjection() error = %v", err)
	}
}
