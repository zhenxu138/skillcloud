package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type AppOptions struct {
	Target string
	Scope  string
}

type AppModel struct {
	Target string
	Scope  string
	view   int
}

var appViews = []string{"Browse", "Enabled", "Add", "Changes", "Sync", "Settings"}

func NewAppModel(opts AppOptions) AppModel {
	return AppModel{Target: opts.Target, Scope: opts.Scope}
}

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.view = (m.view + 1) % len(appViews)
		case "shift+tab":
			m.view = (m.view + len(appViews) - 1) % len(appViews)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AppModel) ViewName() string {
	return appViews[m.view]
}

func (m AppModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "skillcloud | Target: %s | Scope: %s\n\n", m.Target, m.Scope)
	for i, name := range appViews {
		marker := " "
		if i == m.view {
			marker = ">"
		}
		fmt.Fprintf(&b, "%s %s\n", marker, name)
	}
	b.WriteString("\nTab switch view | q quit\n")
	if m.ViewName() == "Add" {
		addModel := NewAddModel(AddOptions{})
		b.WriteString("\n")
		b.WriteString(addModel.View())
	}
	return b.String()
}

func RunApp(opts AppOptions) error {
	_, err := tea.NewProgram(NewAppModel(opts)).Run()
	return err
}
