package tui

import (
	"fmt"
	"reflect"
	"strings"
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

func manySkills(n int) []skill.Skill {
	skills := make([]skill.Skill, 0, n)
	for i := 0; i < n; i++ {
		skills = append(skills, skill.Skill{
			ID:          fmt.Sprintf("sop/matt-pocock/skill-%02d", i),
			Name:        fmt.Sprintf("skill-%02d", i),
			Description: "Description for skill",
		})
	}
	return skills
}

func TestManageModelWindowKeepsCursorVisible(t *testing.T) {
	m := NewManageModel(ManageOptions{Skills: manySkills(50), Target: "codex", Scope: "project", Mode: "link"})
	m.Height = 12
	m.Cursor = 40
	start, end := m.window(len(m.visibleSkills()))
	if start > 40 || 40 >= end {
		t.Fatalf("window = %d-%d, cursor 40 not visible", start, end)
	}
}

func TestManageModelHomeEndAndPageKeys(t *testing.T) {
	m := NewManageModel(ManageOptions{Skills: manySkills(50), Target: "codex", Scope: "project", Mode: "link"})
	m.Height = 12

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = updated.(ManageModel)
	if m.Cursor != 49 {
		t.Fatalf("end cursor = %d, want 49", m.Cursor)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = updated.(ManageModel)
	if m.Cursor != 0 {
		t.Fatalf("home cursor = %d, want 0", m.Cursor)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(ManageModel)
	if m.Cursor != m.pageSize() {
		t.Fatalf("pgdown cursor = %d, want %d", m.Cursor, m.pageSize())
	}
}

func TestManageModelViewShowsGroupAndWindow(t *testing.T) {
	m := NewManageModel(ManageOptions{Skills: manySkills(30), Target: "codex", Scope: "project", Mode: "link"})
	m.Height = 12
	m.Cursor = 20
	view := m.View()
	for _, want := range []string{"skillcloud · codex · project · link", "Showing", "matt-pocock", "Pending:"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in view:\n%s", want, view)
		}
	}
	if strings.Contains(view, "skill-00") {
		t.Fatalf("view should render a window, got:\n%s", view)
	}
}

func TestManageModelConfirmCapsLongLists(t *testing.T) {
	m := NewManageModel(ManageOptions{Skills: manySkills(20), Target: "codex", Scope: "project", Mode: "link"})
	for _, s := range m.Skills {
		m.Selected[s.ID] = true
	}
	view := m.confirmView()
	if !strings.Contains(view, "and 12 more") {
		t.Fatalf("expected capped list, got:\n%s", view)
	}
}

func TestGroupLabelAndSkillLabel(t *testing.T) {
	if got := groupLabel("sop/matt-pocock/code-review"); got != "matt-pocock" {
		t.Fatalf("groupLabel matts = %q", got)
	}
	if got := groupLabel("sop/github-spec-kit/speckit-plan"); got != "spec-kit" {
		t.Fatalf("groupLabel spec-kit = %q", got)
	}
	if got := groupLabel("myself/learn-lite"); got != "myself" {
		t.Fatalf("groupLabel myself = %q", got)
	}
	if got := skillLabel(skill.Skill{ID: "sop/matt-pocock/code-review", Name: "code-review"}); got != "code-review" {
		t.Fatalf("skillLabel = %q", got)
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

func TestManageModelStartsAtFirstEnabledSkill(t *testing.T) {
	m := NewManageModel(ManageOptions{
		Skills:           testSkills(),
		Target:           "codex",
		Scope:            "project",
		Mode:             "link",
		InitiallyEnabled: []string{"coding/go-code-review"},
	})
	if m.Cursor != 1 {
		t.Fatalf("cursor = %d, want first enabled skill index 1", m.Cursor)
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

func TestManageModelInitialFilterViewEnabled(t *testing.T) {
	m := NewManageModel(ManageOptions{
		Skills:            testSkills(),
		Target:            "codex",
		Scope:             "project",
		Mode:              "link",
		InitiallyEnabled:  []string{"coding/code-review", "stock/risk-control"},
		InitialFilterView: ViewEnabled(),
	})
	if m.viewName() != "Enabled" {
		t.Fatalf("initial view = %q, want Enabled", m.viewName())
	}
	if got := len(m.visibleSkills()); got != 2 {
		t.Fatalf("enabled visible count = %d, want 2", got)
	}
	for _, s := range m.visibleSkills() {
		if !m.Selected[s.ID] {
			t.Fatalf("visible skill %q is not enabled", s.ID)
		}
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
