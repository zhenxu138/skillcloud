package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skillcloud/skillcloud/internal/skill"
)

type Model struct {
	Skills   []skill.Skill
	Cursor   int
	Selected map[string]bool
	Done     bool
}

func NewModel(skills []skill.Skill) Model {
	return Model{Skills: skills, Selected: map[string]bool{}}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Skills)-1 {
				m.Cursor++
			}
		case " ":
			if len(m.Skills) > 0 {
				m.Toggle(m.Skills[m.Cursor].ID)
			}
		case "enter":
			m.Done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("Select skills. Space toggles, Enter applies, q quits.\n\n")
	for i, s := range m.Skills {
		cursor := " "
		if i == m.Cursor {
			cursor = ">"
		}
		checked := "[ ]"
		if m.Selected[s.ID] {
			checked = "[x]"
		}
		fmt.Fprintf(&b, "%s %s %s - %s\n", cursor, checked, s.ID, s.Description)
	}
	return b.String()
}

func (m *Model) Toggle(id string) {
	m.Selected[id] = !m.Selected[id]
}

func (m Model) SelectedIDs() []string {
	var ids []string
	for _, s := range m.Skills {
		if m.Selected[s.ID] {
			ids = append(ids, s.ID)
		}
	}
	return ids
}

func Select(skills []skill.Skill) ([]string, error) {
	program := tea.NewProgram(NewModel(skills))
	result, err := program.Run()
	if err != nil {
		return nil, err
	}
	model := result.(Model)
	return model.SelectedIDs(), nil
}

