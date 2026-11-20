package tui

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testGroups() []GroupOption {
	return []GroupOption{
		{ID: "groups/matt", Name: "matt", Category: "SOP", Title: "Matt", Description: "单人小项目", SkillCount: 31, HasAgents: true},
		{ID: "groups/devops", Name: "devops", Category: "自定义组", Title: "DevOps", Description: "运维与网络诊断", SkillCount: 6, HasAgents: true},
		{ID: "groups/empty", Name: "empty", Category: "自定义组", Title: "Empty", SkillCount: 0},
	}
}

func TestGroupModelInitialSelectionAndDiff(t *testing.T) {
	m := NewGroupModel(GroupOptions{
		Groups:           testGroups(),
		Target:           "codex",
		Scope:            "project",
		Mode:             "link",
		InitiallyEnabled: []string{"groups/matt"},
	})
	if !m.Selected["groups/matt"] || m.Cursor != 0 {
		t.Fatalf("initial state = %#v cursor %d", m.Selected, m.Cursor)
	}
	m.Toggle("groups/matt")
	m.Toggle("groups/devops")
	enable, disable := m.PendingDiff()
	if !reflect.DeepEqual(enable, []string{"groups/devops"}) {
		t.Fatalf("enable = %#v", enable)
	}
	if !reflect.DeepEqual(disable, []string{"groups/matt"}) {
		t.Fatalf("disable = %#v", disable)
	}
}

func TestGroupModelCancelDoesNotApply(t *testing.T) {
	m := NewGroupModel(GroupOptions{Groups: testGroups(), Target: "codex", Scope: "project", Mode: "link"})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if updated.(GroupModel).Result().Apply {
		t.Fatal("expected q to cancel")
	}
}

func TestGroupModelStartsAtFirstEnabled(t *testing.T) {
	m := NewGroupModel(GroupOptions{
		Groups:           testGroups(),
		Target:           "codex",
		Scope:            "project",
		Mode:             "link",
		InitiallyEnabled: []string{"groups/devops"},
	})
	if m.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.Cursor)
	}
}

func TestGroupModelSearchMatchesDescription(t *testing.T) {
	m := NewGroupModel(GroupOptions{Groups: testGroups(), Target: "codex", Scope: "project", Mode: "link"})
	m.Query = "运维"
	visible := m.visibleGroups()
	if len(visible) != 1 || visible[0].Name != "devops" {
		t.Fatalf("visible = %#v", visible)
	}
}

func TestGroupModelViewShowsRowsAndRules(t *testing.T) {
	m := NewGroupModel(GroupOptions{Groups: testGroups(), Target: "codex", Scope: "project", Mode: "link"})
	view := m.View()
	for _, want := range []string{"skillcloud · groups · codex · link", "matt", "SOP", "31 skills", "rules", "Pending:"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in:\n%s", want, view)
		}
	}
}

func TestGroupModelConfirmAndApply(t *testing.T) {
	m := NewGroupModel(GroupOptions{Groups: testGroups(), Target: "codex", Scope: "project", Mode: "link"})
	m.Toggle("groups/devops")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(GroupModel)
	if m.Screen != groupScreenConfirm {
		t.Fatalf("screen = %v, want confirm", m.Screen)
	}
	if !strings.Contains(m.View(), "devops (6 skills)") {
		t.Fatalf("confirm view missing group:\n%s", m.View())
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(GroupModel)
	result := m.Result()
	if !result.Apply || !reflect.DeepEqual(result.SelectedIDs, []string{"groups/devops"}) {
		t.Fatalf("result = %#v", result)
	}
}
