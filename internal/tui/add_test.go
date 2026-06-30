package tui

import (
	"strings"
	"testing"
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
