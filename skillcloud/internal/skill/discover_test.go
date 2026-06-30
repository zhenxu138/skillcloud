package skill

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverReturnsSkillsSortedByID(t *testing.T) {
	repoDir := t.TempDir()
	writeSkillMarkdown(t, repoDir, "stock", "risk-control", "Risk Control", "Manage portfolio risk.")
	writeSkillMarkdown(t, repoDir, "coding", "code-review", "Code Review", "Review code changes.")

	got, err := Discover(repoDir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	want := []Skill{
		{
			ID:          "coding/code-review",
			Name:        "Code Review",
			Description: "Review code changes.",
			Path:        filepath.Join(repoDir, "skills", "coding", "code-review"),
		},
		{
			ID:          "stock/risk-control",
			Name:        "Risk Control",
			Description: "Manage portfolio risk.",
			Path:        filepath.Join(repoDir, "skills", "stock", "risk-control"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Discover() = %#v, want %#v", got, want)
	}
}

func TestDiscoverSkipsHiddenDirectorySkills(t *testing.T) {
	repoDir := t.TempDir()
	writeSkillMarkdown(t, repoDir, ".draft", "foo", "Draft Skill", "Not ready.")
	writeSkillMarkdown(t, repoDir, "stock", "risk-control", "Risk Control", "Manage portfolio risk.")

	got, err := Discover(repoDir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	want := []Skill{
		{
			ID:          "stock/risk-control",
			Name:        "Risk Control",
			Description: "Manage portfolio risk.",
			Path:        filepath.Join(repoDir, "skills", "stock", "risk-control"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Discover() = %#v, want %#v", got, want)
	}
}

func TestDiscoverSkipsMalformedHiddenDirectorySkills(t *testing.T) {
	repoDir := t.TempDir()
	writeRawSkillMarkdown(t, repoDir, []string{".draft", "foo"}, "not frontmatter")
	writeSkillMarkdown(t, repoDir, "stock", "risk-control", "Risk Control", "Manage portfolio risk.")

	got, err := Discover(repoDir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	want := []Skill{
		{
			ID:          "stock/risk-control",
			Name:        "Risk Control",
			Description: "Manage portfolio risk.",
			Path:        filepath.Join(repoDir, "skills", "stock", "risk-control"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Discover() = %#v, want %#v", got, want)
	}
}

func TestParseSkillMarkdownAcceptsCRLFFrontmatter(t *testing.T) {
	name, description, err := ParseSkillMarkdown([]byte("---\r\nname: Risk Control\r\ndescription: Manage portfolio risk.\r\n---\r\nBody\r\n"))
	if err != nil {
		t.Fatalf("ParseSkillMarkdown() error = %v", err)
	}
	if name != "Risk Control" {
		t.Fatalf("ParseSkillMarkdown() name = %q, want %q", name, "Risk Control")
	}
	if description != "Manage portfolio risk." {
		t.Fatalf("ParseSkillMarkdown() description = %q, want %q", description, "Manage portfolio risk.")
	}
}

func TestParseSkillMarkdownAcceptsUTF8BOM(t *testing.T) {
	name, description, err := ParseSkillMarkdown([]byte("\xef\xbb\xbf---\nname: Code Review\ndescription: Review code changes.\n---\nBody\n"))
	if err != nil {
		t.Fatalf("ParseSkillMarkdown() error = %v", err)
	}
	if name != "Code Review" {
		t.Fatalf("ParseSkillMarkdown() name = %q, want %q", name, "Code Review")
	}
	if description != "Review code changes." {
		t.Fatalf("ParseSkillMarkdown() description = %q, want %q", description, "Review code changes.")
	}
}

func TestParseSkillMarkdownRequiresClosingDelimiterLine(t *testing.T) {
	_, description, err := ParseSkillMarkdown([]byte("---\nname: Risk Control\ndescription: |\n  Manage portfolio risk.\n  --- not a delimiter\n---\nBody\n"))
	if err != nil {
		t.Fatalf("ParseSkillMarkdown() error = %v", err)
	}
	if !strings.Contains(description, "--- not a delimiter") {
		t.Fatalf("ParseSkillMarkdown() description = %q, want fake delimiter text", description)
	}
}

func TestParseSkillMarkdownRequiresNameAndDescription(t *testing.T) {
	_, _, err := ParseSkillMarkdown([]byte("---\nname: Missing Description\n---\nBody\n"))
	if err == nil {
		t.Fatal("ParseSkillMarkdown() error = nil, want error")
	}
}

func TestResolveMatchesExactIDNameAndPrefix(t *testing.T) {
	skills := []Skill{
		{ID: "coding", Name: "Coding"},
		{ID: "coding/code-review", Name: "Code Review"},
		{ID: "stock/risk-control", Name: "Risk Control"},
	}

	if got := Resolve(skills, "stock/risk-control"); !reflect.DeepEqual(got, []Skill{skills[2]}) {
		t.Fatalf("Resolve(ID) = %#v, want %#v", got, []Skill{skills[2]})
	}
	if got := Resolve(skills, "Code Review"); !reflect.DeepEqual(got, []Skill{skills[1]}) {
		t.Fatalf("Resolve(Name) = %#v, want %#v", got, []Skill{skills[1]})
	}
	if got := Resolve(skills, "coding/*"); !reflect.DeepEqual(got, []Skill{skills[1]}) {
		t.Fatalf("Resolve(prefix) = %#v, want %#v", got, []Skill{skills[1]})
	}
}

func writeSkillMarkdown(t *testing.T, repoDir string, parts ...string) {
	t.Helper()

	name := parts[len(parts)-2]
	description := parts[len(parts)-1]
	data := "---\nname: " + name + "\ndescription: " + description + "\n---\nBody\n"
	writeRawSkillMarkdown(t, repoDir, parts[:len(parts)-2], data)
}

func writeRawSkillMarkdown(t *testing.T, repoDir string, parts []string, data string) {
	t.Helper()

	dirParts := append([]string{repoDir, "skills"}, parts...)
	dir := filepath.Join(dirParts...)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
