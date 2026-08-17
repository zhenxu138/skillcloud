package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"os"
	"path/filepath"
)

func TestAddModelShowsValidationError(t *testing.T) {
	m := NewAddModel(AddOptions{})
	m.SetValidationError("SKILL.md not found")
	view := m.View()
	if !strings.Contains(view, "SKILL.md not found") {
		t.Fatalf("View() = %q", view)
	}
}

func TestAddModelShowsConflict(t *testing.T) {
	m := NewAddModel(AddOptions{})
	m.SetConflict("skill coding/code-review already exists")
	view := m.View()
	if !strings.Contains(view, "already exists") {
		t.Fatalf("View() = %q", view)
	}
}

func TestAddModelShowsPreview(t *testing.T) {
	m := NewAddModel(AddOptions{})
	m.SetPreview(AddPreview{ID: "coding/code-review", Name: "code-review", Description: "Review code."})
	view := m.View()
	if !strings.Contains(view, "coding/code-review") || !strings.Contains(view, "Review code.") {
		t.Fatalf("View() = %q", view)
	}
}

func TestAddFormTypingFillsFieldsAndSubmits(t *testing.T) {
	m := NewAddFormModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/tmp/skill")})
	m = updated.(AddFormModel)
	if m.Source != "/tmp/skill" {
		t.Fatalf("Source = %q, want /tmp/skill", m.Source)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(AddFormModel)
	if m.field != addFieldID {
		t.Fatalf("field = %v, want addFieldID", m.field)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("coding/review")})
	m = updated.(AddFormModel)
	if m.ID != "coding/review" {
		t.Fatalf("ID = %q, want coding/review", m.ID)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AddFormModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AddFormModel)
	if m.field != addFieldSubmit {
		t.Fatalf("field = %v, want addFieldSubmit", m.field)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(AddFormModel)
	if m.Source != "/tmp/skill" || m.ID != "coding/review" {
		t.Fatalf("after submit Source=%q ID=%q", m.Source, m.ID)
	}
}

func TestAddFormEscCancels(t *testing.T) {
	m := NewAddFormModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = updated.(AddFormModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(AddFormModel)
	if !got.canceled {
		t.Fatal("expected esc to cancel")
	}
}

func TestAddFormSubmitRequiresBothFields(t *testing.T) {
	m := NewAddFormModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AddFormModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AddFormModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AddFormModel)
	if m.field != addFieldSubmit {
		t.Fatalf("field = %v, want addFieldSubmit", m.field)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(AddFormModel)
	if m.submitted {
		t.Fatal("expected enter on empty fields to not submit")
	}
}

func TestAddFormToggleReplace(t *testing.T) {
	m := NewAddFormModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AddFormModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AddFormModel)
	if m.field != addFieldReplace {
		t.Fatalf("field = %v, want addFieldReplace", m.field)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(AddFormModel)
	if !m.Replace {
		t.Fatal("expected space to toggle replace on")
	}
}

func TestSourceBrowserListsDirectoriesAndPicksCurrent(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	// SKILL.md present so the "use this directory" row appears in View.
	if err := os.WriteFile(filepath.Join(sub, "SKILL.md"), []byte("---\nname: x\ndescription: y\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewSourceBrowserModel(sub)
	if m.cwd != sub {
		t.Fatalf("cwd = %q, want %q", m.cwd, sub)
	}
	view := m.View()
	if !strings.Contains(view, "use this directory") {
		t.Fatalf("view missing use-this row: %s", view)
	}
	// Space picks the current directory.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	_ = updated.(SourceBrowserModel)
}

func TestSourceBrowserEntersSubdirectoryAndAscends(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "child")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	m := NewSourceBrowserModel(tmp)
	idx := -1
	for i, e := range m.entries {
		if e.name == "child" {
			idx = i
			break
		}
	}
	if idx == -1 {
		t.Fatalf("child not listed: %#v", m.entries)
	}
	m.cursor = idx
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(SourceBrowserModel)
	if m.cwd != sub {
		t.Fatalf("after enter cwd = %q, want %q", m.cwd, sub)
	}

	m.cursor = 0
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(SourceBrowserModel)
	if m.cwd != tmp {
		t.Fatalf("after .. cwd = %q, want %q", m.cwd, tmp)
	}
}

func TestAddFormBKeyOpensBrowser(t *testing.T) {
	m := NewAddFormModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	got := updated.(AddFormModel)
	if !got.wantBrowser {
		t.Fatal("expected b on source field to set wantBrowser")
	}
}
