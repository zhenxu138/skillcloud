package tui

import (
	"testing"

	"github.com/skillcloud/skillcloud/internal/skill"
)

func TestToggleSelection(t *testing.T) {
	m := NewModel([]skill.Skill{
		{ID: "coding/code-review", Name: "code-review"},
		{ID: "stock/risk-control", Name: "risk-control"},
	})
	m.Toggle("coding/code-review")
	if !m.Selected["coding/code-review"] {
		t.Fatal("expected selected skill")
	}
	m.Toggle("coding/code-review")
	if m.Selected["coding/code-review"] {
		t.Fatal("expected deselected skill")
	}
}

