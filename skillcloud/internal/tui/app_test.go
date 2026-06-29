package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppModelInitialViewShowsBrowse(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	view := m.View()
	if !strings.Contains(view, "Browse") || !strings.Contains(view, "codex") {
		t.Fatalf("View() = %q", view)
	}
}

func TestAppModelTabCyclesViews(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	got := updated.(AppModel)
	if got.ViewName() != "Enabled" {
		t.Fatalf("ViewName() = %q, want Enabled", got.ViewName())
	}
}
