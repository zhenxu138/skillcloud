package tui

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skillcloud/skillcloud/internal/skill"
)

func testSkills() []skill.Skill {
	return []skill.Skill{
		{ID: "coding/code-review", Name: "code-review", Description: "Review code changes"},
		{ID: "coding/go-code-review", Name: "go-code-review", Description: "Review Go code"},
		{ID: "stock/risk-control", Name: "risk-control", Description: "Trading risk checklist"},
	}
}

func TestManageModelInitialSelectionAndDiff(t *testing.T) {
	m := NewManageModel(ManageOptions{
		Skills:           testSkills(),
		Target:           "codex",
		Scope:            "project",
		Mode:             "link",
		InitiallyEnabled: []string{"coding/code-review", "stock/risk-control"},
	})

	if !m.Selected["coding/code-review"] || !m.Selected["stock/risk-control"] {
		t.Fatalf("expected initial skills selected: %#v", m.Selected)
	}

	m.Toggle("coding/code-review")
	m.Toggle("coding/go-code-review")

	gotEnable, gotDisable := m.PendingDiff()
	wantEnable := []string{"coding/go-code-review"}
	wantDisable := []string{"coding/code-review"}
	if !reflect.DeepEqual(gotEnable, wantEnable) {
		t.Fatalf("enable diff = %#v, want %#v", gotEnable, wantEnable)
	}
	if !reflect.DeepEqual(gotDisable, wantDisable) {
		t.Fatalf("disable diff = %#v, want %#v", gotDisable, wantDisable)
	}
}

func TestManageModelCancelDoesNotApply(t *testing.T) {
	m := NewManageModel(ManageOptions{
		Skills:           testSkills(),
		Target:           "codex",
		Scope:            "project",
		Mode:             "link",
		InitiallyEnabled: []string{"coding/code-review"},
	})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	got := updated.(ManageModel).Result()
	if got.Apply {
		t.Fatal("expected q to cancel without apply")
	}
}

func TestManageModelSearchMatchesIDNamePathAndDescription(t *testing.T) {
	tests := map[string]string{
		"coding":  "coding/code-review",
		"go-code": "coding/go-code-review",
		"trading": "stock/risk-control",
	}

	for query, wantID := range tests {
		m := NewManageModel(ManageOptions{Skills: testSkills(), Target: "codex", Scope: "project", Mode: "link"})
		m.Query = query
		visible := m.visibleSkills()
		if len(visible) == 0 {
			t.Fatalf("query %q returned no skills", query)
		}
		found := false
		for _, s := range visible {
			if s.ID == wantID {
				found = true
			}
		}
		if !found {
			t.Fatalf("query %q did not include %q: %#v", query, wantID, visible)
		}
	}
}

func TestManageModelTabCyclesViews(t *testing.T) {
	m := NewManageModel(ManageOptions{
		Skills:           testSkills(),
		Target:           "codex",
		Scope:            "project",
		Mode:             "link",
		InitiallyEnabled: []string{"coding/code-review"},
	})
	m.Toggle("coding/go-code-review")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(ManageModel)
	if m.viewName() != "Enabled" {
		t.Fatalf("first tab view = %q, want Enabled", m.viewName())
	}
	if got := len(m.visibleSkills()); got != 2 {
		t.Fatalf("enabled visible count = %d, want 2", got)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(ManageModel)
	if m.viewName() != "Disabled" {
		t.Fatalf("second tab view = %q, want Disabled", m.viewName())
	}
	if got := len(m.visibleSkills()); got != 1 {
		t.Fatalf("disabled visible count = %d, want 1", got)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(ManageModel)
	if m.viewName() != "Changed" {
		t.Fatalf("third tab view = %q, want Changed", m.viewName())
	}
	if got := len(m.visibleSkills()); got != 1 {
		t.Fatalf("changed visible count = %d, want 1", got)
	}
}

func TestManageModelEnterRequiresConfirmationWhenChanged(t *testing.T) {
	m := NewManageModel(ManageOptions{
		Skills:           testSkills(),
		Target:           "codex",
		Scope:            "project",
		Mode:             "link",
		InitiallyEnabled: []string{"coding/code-review"},
	})
	m.Toggle("coding/go-code-review")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(ManageModel)
	if m.Screen != screenConfirm {
		t.Fatalf("screen = %v, want confirm", m.Screen)
	}
	if m.Result().Apply {
		t.Fatal("expected first enter to wait for confirmation")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(ManageModel)
	if !m.Result().Apply {
		t.Fatal("expected second enter to apply")
	}
}
