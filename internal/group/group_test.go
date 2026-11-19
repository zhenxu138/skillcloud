package group

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skillcloud/skillcloud/internal/skill"
)

func writeSkill(t *testing.T, repoDir string, id string, name string) {
	t.Helper()
	dir := filepath.Join(repoDir, "skills", filepath.FromSlash(id))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	data := "---\nname: " + name + "\ndescription: test\n---\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeManifest(t *testing.T, repoDir string, rel string, content string) {
	t.Helper()
	dir := filepath.Join(repoDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ManifestFile), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverCustomGroup(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, repo, "ops/foo", "foo")
	writeSkill(t, repo, "ops/bar", "bar")
	writeSkill(t, repo, "other/baz", "baz")
	writeManifest(t, repo, "groups/devops", `name: devops
title: DevOps
entry: ops/foo
agents: AGENTS.md
skills:
  - id: ops/foo
  - name: bar
    as: bar-tool
`)
	if err := os.WriteFile(filepath.Join(repo, "groups", "devops", "AGENTS.md"), []byte("Run checks.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := skill.Discover(repo)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := Discover(repo, skills)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("got %d groups", len(groups))
	}
	g := groups[0]
	if g.ID != "groups/devops" || g.Name != "devops" || g.Title != "DevOps" {
		t.Fatalf("group = %#v", g)
	}
	if len(g.Members) != 2 || g.Members[0].Skill.ID != "ops/foo" || g.Members[1].As != "bar-tool" {
		t.Fatalf("members = %#v", g.Members)
	}
	if g.AgentsPath == "" {
		t.Fatal("expected agents path")
	}
}

func TestDiscoverInlineGroupDefaultsToDirectory(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, repo, "sop/demo/a", "a")
	writeSkill(t, repo, "sop/demo/b", "b")
	writeSkill(t, repo, "sop/other/c", "c")
	writeManifest(t, repo, "skills/sop/demo", "name: demo\n")

	skills, err := skill.Discover(repo)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := Discover(repo, skills)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("got %d groups", len(groups))
	}
	if groups[0].ID != "skills/sop/demo" || groups[0].Name != "demo" {
		t.Fatalf("group = %#v", groups[0])
	}
	if len(groups[0].Members) != 2 {
		t.Fatalf("members = %#v", groups[0].Members)
	}
}

func TestDiscoverRejectsDuplicateAliases(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, repo, "a/review", "review")
	writeSkill(t, repo, "b/review", "review")
	writeManifest(t, repo, "groups/demo", `name: demo
skills:
  - id: a/review
  - id: b/review
`)

	skills, err := skill.Discover(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(repo, skills); err == nil {
		t.Fatal("expected duplicate alias error")
	}
}

func TestFindMatchesNameAndPath(t *testing.T) {
	groups := []Group{{ID: "skills/sop/demo", Name: "demo", Path: filepath.FromSlash("/repo/skills/sop/demo")}}
	for _, query := range []string{"demo", "sop/demo", "skills/sop/demo"} {
		g, err := Find(groups, query)
		if err != nil {
			t.Fatalf("Find(%q) error = %v", query, err)
		}
		if g.ID != "skills/sop/demo" {
			t.Fatalf("Find(%q) = %#v", query, g)
		}
	}
}
