package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertCreatesAndReplacesBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(path, []byte("# Project\n\nKeep this.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Upsert(path, "groups/demo", "Demo", "# Demo rules\n\nUse checks.\n"); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"# Project", "Keep this.", BeginMarker("groups/demo"), "## Skill group: Demo", "Use checks.", EndMarker("groups/demo")} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "# Demo rules") {
		t.Fatalf("leading H1 should be stripped:\n%s", text)
	}

	if err := Upsert(path, "groups/demo", "Demo", "Updated rule.\n"); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	data, _ = os.ReadFile(path)
	text = string(data)
	if strings.Contains(text, "Use checks.") {
		t.Fatalf("old block content remained:\n%s", text)
	}
	if strings.Count(text, BeginMarker("groups/demo")) != 1 {
		t.Fatalf("expected exactly one block:\n%s", text)
	}
	if !strings.Contains(text, "Updated rule.") {
		t.Fatalf("updated rule missing:\n%s", text)
	}
}

func TestRemovePreservesOtherBlocksAndUserContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(path, []byte("User note.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(path, "groups/a", "A", "Rule A.\n"); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(path, "groups/b", "B", "Rule B.\n"); err != nil {
		t.Fatal(err)
	}
	if err := Remove(path, "groups/a"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	data, _ := os.ReadFile(path)
	text := string(data)
	if strings.Contains(text, "groups/a") || strings.Contains(text, "Rule A.") {
		t.Fatalf("group a block remained:\n%s", text)
	}
	for _, want := range []string{"User note.", BeginMarker("groups/b"), "Rule B."} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	ids, err := List(path)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != "groups/b" {
		t.Fatalf("List() = %#v", ids)
	}
}

func TestRemoveMissingBlockIsNoop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(path, []byte("User note.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Remove(path, "groups/missing"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "User note.\n" {
		t.Fatalf("file changed: %q", string(data))
	}
}

func TestUpsertRejectsUnterminatedBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	content := BeginMarker("groups/demo") + "\nrule\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(path, "groups/demo", "Demo", "new rule"); err == nil {
		t.Fatal("expected unterminated block error")
	}
	data, _ := os.ReadFile(path)
	if string(data) != content {
		t.Fatalf("file should remain unchanged, got %q", string(data))
	}
}

func TestLineEndingsArePreserved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(path, []byte("User note.\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(path, "groups/demo", "Demo", "Rule."); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "\r\n") || strings.Contains(string(data), "\n\n") && !strings.Contains(string(data), "\r\n\r\n") {
		t.Fatalf("expected CRLF output, got %q", string(data))
	}
}

func TestGroupIndexUpsertAndRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(path, []byte("User note.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := UpsertGroupIndex(path, "## Skill groups\n\n- work"); err != nil {
		t.Fatalf("UpsertGroupIndex() error = %v", err)
	}
	data, _ := os.ReadFile(path)
	text := string(data)
	if !strings.Contains(text, GroupIndexBegin) || !strings.Contains(text, GroupIndexEnd) || !strings.Contains(text, "User note.") {
		t.Fatalf("index not written:\n%s", text)
	}
	if err := UpsertGroupIndex(path, "## Skill groups\n\n- myself"); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	text = string(data)
	if strings.Contains(text, "- work") || !strings.Contains(text, "- myself") || strings.Count(text, GroupIndexBegin) != 1 {
		t.Fatalf("index not replaced:\n%s", text)
	}
	if err := RemoveGroupIndex(path); err != nil {
		t.Fatalf("RemoveGroupIndex() error = %v", err)
	}
	data, _ = os.ReadFile(path)
	text = string(data)
	if strings.Contains(text, GroupIndexBegin) || !strings.Contains(text, "User note.") {
		t.Fatalf("index not removed:\n%s", text)
	}
}

func TestRemoveLegacyGroupBlocksKeepsIndex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(path, []byte("User note.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(path, "skills/sop/work", "Work", "Old work rule.\n"); err != nil {
		t.Fatal(err)
	}
	if err := UpsertGroupIndex(path, "## Skill groups\n\n- work"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveLegacyGroupBlocks(path); err != nil {
		t.Fatalf("RemoveLegacyGroupBlocks() error = %v", err)
	}
	data, _ := os.ReadFile(path)
	text := string(data)
	if strings.Contains(text, "Old work rule.") || strings.Contains(text, BeginMarker("skills/sop/work")) {
		t.Fatalf("legacy block remained:\n%s", text)
	}
	if !strings.Contains(text, GroupIndexBegin) || !strings.Contains(text, "User note.") {
		t.Fatalf("index or user content removed:\n%s", text)
	}
}

func TestGroupIndexPreservesSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "AGENTS.real.md")
	link := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(target, []byte("User note.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
	if err := UpsertGroupIndex(link, "## Skill groups\n\n- work"); err != nil {
		t.Fatalf("UpsertGroupIndex() error = %v", err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("AGENTS.md symlink was replaced")
	}
	data, _ := os.ReadFile(target)
	if !strings.Contains(string(data), GroupIndexBegin) || !strings.Contains(string(data), "User note.") {
		t.Fatalf("target not updated:\n%s", string(data))
	}
}
