package install

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestRemoveProjectionRemovesDirectoryAndManifest(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "proj")
	if err := os.Mkdir(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "copy"}
	if err := WriteProjectionManifest(projDir, manifest); err != nil {
		t.Fatal(err)
	}
	if err := RemoveProjection(projDir); err != nil {
		t.Fatalf("RemoveProjection() error = %v", err)
	}
	if _, err := os.Stat(projDir); !os.IsNotExist(err) {
		t.Fatalf("expected projection dir to be removed, got err = %v", err)
	}
	if _, err := os.Stat(projectionManifestPath(projDir)); !os.IsNotExist(err) {
		t.Fatalf("expected manifest to be removed, got err = %v", err)
	}
}

func TestRemoveProjectionRemovesSymlinkAndSidecar(t *testing.T) {
	dir := t.TempDir()
	targetDir := filepath.Join(dir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(targetDir, linkDir); err != nil {
		t.Skip("skipping symlink test: ", err)
	}
	manifest := ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "copy"}
	if err := WriteProjectionManifest(linkDir, manifest); err != nil {
		t.Fatal(err)
	}
	sidecarPath := projectionManifestPath(linkDir)
	if _, err := os.Stat(sidecarPath); err != nil {
		t.Fatalf("sidecar manifest not created: %v", err)
	}
	if err := RemoveProjection(linkDir); err != nil {
		t.Fatalf("RemoveProjection() error = %v", err)
	}
	if _, err := os.Lstat(linkDir); !os.IsNotExist(err) {
		t.Fatalf("expected symlink to be removed, got err = %v", err)
	}
	if _, err := os.Stat(sidecarPath); !os.IsNotExist(err) {
		t.Fatalf("expected sidecar manifest to be removed, got err = %v", err)
	}
	if _, err := os.Stat(targetDir); err != nil {
		t.Fatalf("expected symlink target to remain, got err = %v", err)
	}
}

func TestRemoveProjectionKeepsManifestIfRemoveAllFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based removal failure not reliable on Windows")
	}
	dir := t.TempDir()
	projDir := filepath.Join(dir, "proj")
	if err := os.Mkdir(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(projDir, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "file.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "copy"}
	if err := WriteProjectionManifest(projDir, manifest); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(projDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(projDir, 0o755) })

	if err := RemoveProjection(projDir); err == nil {
		t.Fatal("expected RemoveProjection to fail")
	}
	if err := CanRemoveProjection(projDir, manifest); err != nil {
		t.Fatalf("CanRemoveProjection() error = %v; manifest should still be readable", err)
	}
}
