package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppModelInitialViewShowsBrowse(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	view := m.View()
	if !strings.Contains(view, "Browse") || !strings.Contains(view, "codex") || !strings.Contains(view, "Choose skills for this project") {
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

func TestAppModelDownAdvancesCursorAndClamps(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	if got := m.ViewName(); got != "Browse" {
		t.Fatalf("initial view = %q, want Browse", got)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AppModel)
	if got := m.ViewName(); got != "Enabled" {
		t.Fatalf("after one down view = %q, want Enabled", got)
	}
	for i := 0; i < 10; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(AppModel)
	}
	if got := m.ViewName(); got != "Settings" {
		t.Fatalf("after many downs view = %q, want Settings (clamped)", got)
	}
}

func TestAppModelUpMovesCursorUpAndClamps(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	for i := 0; i < 3; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(AppModel)
	}
	if got := m.ViewName(); got != "Changes" {
		t.Fatalf("view = %q, want Changes", got)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(AppModel)
	if got := m.ViewName(); got != "Add" {
		t.Fatalf("after up view = %q, want Add", got)
	}
	for i := 0; i < 10; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = updated.(AppModel)
	}
	if got := m.ViewName(); got != "Browse" {
		t.Fatalf("after many ups view = %q, want Browse (clamped)", got)
	}
}

func TestAppModelEnterRecordsChosenAction(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AppModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(AppModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(AppModel)
	if m.chosen != ActionAdd {
		t.Fatalf("chosen = %v, want ActionAdd", m.chosen)
	}
}

func TestAppModelQuitSetsActionQuit(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(AppModel)
	if m.chosen != ActionQuit {
		t.Fatalf("chosen = %v, want ActionQuit", m.chosen)
	}
}

func TestAppModelEscQuits(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(AppModel)
	if got.chosen != ActionQuit {
		t.Fatalf("chosen = %v, want ActionQuit", got.chosen)
	}
}

func TestAppModelIncludesPushView(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project"})
	// Navigate to Push: Browse(0) -> Enabled(1) -> Add(2) -> Changes(3) -> Sync(4) -> Push(5)
	for i := 0; i < 5; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(AppModel)
	}
	if got := m.ViewName(); got != "Push" {
		t.Fatalf("view = %q, want Push", got)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(AppModel)
	if m.chosen != ActionPush {
		t.Fatalf("chosen = %v, want ActionPush", m.chosen)
	}
}

func TestAppModelShowsStatusFromPreviousAction(t *testing.T) {
	m := NewAppModel(AppOptions{Target: "codex", Scope: "project", Status: "applied skill changes: enabled 1, disabled 0"})
	if view := m.View(); !strings.Contains(view, "applied skill changes") {
		t.Fatalf("View() = %q, want status shown", view)
	}
}
