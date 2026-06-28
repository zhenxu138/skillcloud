package install

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteCopy(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("---\nname: x\ndescription: y\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Execute([]Action{{Mode: "copy", Source: src, Dest: dst}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "SKILL.md")); err != nil {
		t.Fatalf("expected copied SKILL.md: %v", err)
	}
}

func TestExecuteReplacesExistingDestination(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Execute([]Action{{Mode: "copy", Source: src, Dest: dst}}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected old destination content removed, got err %v", err)
	}
}

func TestExecuteWritesProjectionManifest(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "copy"}
	err := Execute([]Action{{Mode: "copy", Source: src, Dest: dst, Projection: &manifest}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, ProjectionManifestFile)); err != nil {
		t.Fatalf("expected projection manifest: %v", err)
	}
}

func TestExecuteLinkFallbackCopiesAndRecordsCopyMode(t *testing.T) {
	origSymlink := osSymlink
	osSymlink = func(src, dst string) error { return errors.New("forced symlink failure") }
	defer func() { osSymlink = origSymlink }()

	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("---\nname: x\ndescription: y\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest := ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "link"}
	err := Execute([]Action{{Mode: "link", Source: src, Dest: dst, Projection: &manifest}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "SKILL.md")); err != nil {
		t.Fatalf("expected copied SKILL.md: %v", err)
	}
	got, err := ReadProjectionManifest(dst)
	if err != nil {
		t.Fatalf("ReadProjectionManifest() error = %v", err)
	}
	if got.Mode != "copy" {
		t.Fatalf("manifest mode = %q, want %q", got.Mode, "copy")
	}
}

func TestExecuteLinkWritesProjectionManifestOutsideSource(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst", "code-review")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("link me"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Verify symlinks are supported in this test environment.
	probeSrc := filepath.Join(root, "probe-src")
	probeDst := filepath.Join(root, "probe-dst")
	if err := os.MkdirAll(probeSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(probeSrc, probeDst); err != nil {
		t.Skipf("symlinks not supported in test environment: %v", err)
	}
	_ = os.Remove(probeDst)

	manifest := ProjectionManifest{SourceID: "coding/code-review", Target: "codex", Scope: "project", Mode: "link"}
	if err := Execute([]Action{{Mode: "link", Source: src, Dest: dst, Projection: &manifest}}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	info, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("Lstat(%q) error = %v", dst, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %q to be a symlink", dst)
	}

	sidecar := filepath.Join(filepath.Dir(dst), ".code-review"+ProjectionManifestFile)
	if _, err := os.Stat(sidecar); err != nil {
		t.Fatalf("expected sidecar manifest in projection directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(src, ProjectionManifestFile)); !os.IsNotExist(err) {
		t.Fatalf("manifest should not be written into library source tree")
	}

	got, err := ReadProjectionManifest(dst)
	if err != nil {
		t.Fatalf("ReadProjectionManifest() error = %v", err)
	}
	if got.Mode != "link" {
		t.Fatalf("manifest mode = %q, want %q", got.Mode, "link")
	}
}

