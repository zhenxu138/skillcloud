package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// AppOptions configures the top-level skillcloud menu.
type AppOptions struct {
	Target string
	Scope  string
	Mode   string
	// Status is a message from the previously completed action, shown in
	// the menu so the user can see what just happened before picking again.
	Status string
}

// AppAction identifies which menu entry the user chose to run.
type AppAction int

const (
	ActionQuit AppAction = iota
	ActionBrowse
	ActionEnabled
	ActionAdd
	ActionChanges
	ActionSync
	ActionPush
	ActionSettings
)

var appViews = []string{"Browse", "Enabled", "Add", "Changes", "Sync", "Push", "Settings"}

var appDescriptions = []string{
	"Choose skills for this project",
	"Review and remove enabled skills",
	"Import a local skill directory",
	"Show repo changes and projection drift",
	"Pull updates from the remote",
	"Push local skill edits",
	"Show target, scope and mode",
}

var appActions = []AppAction{
	ActionBrowse,
	ActionEnabled,
	ActionAdd,
	ActionChanges,
	ActionSync,
	ActionPush,
	ActionSettings,
}

// AppModel is the main skillcloud launcher menu. Up/Down moves the selection,
// Enter runs the highlighted entry, and the program exits so the chosen action
// can take over the terminal.
type AppModel struct {
	Target string
	Scope  string
	Mode   string
	Status string
	view   int
	chosen AppAction
}

func NewAppModel(opts AppOptions) AppModel {
	return AppModel{Target: opts.Target, Scope: opts.Scope, Mode: opts.Mode, Status: opts.Status}
}

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.view > 0 {
				m.view--
			}
		case "down", "j":
			if m.view < len(appViews)-1 {
				m.view++
			}
		case "tab":
			m.view = (m.view + 1) % len(appViews)
		case "shift+tab":
			m.view = (m.view + len(appViews) - 1) % len(appViews)
		case "enter":
			if m.view >= 0 && m.view < len(appActions) {
				m.chosen = appActions[m.view]
			}
			return m, tea.Quit
		case "esc":
			m.chosen = ActionQuit
			return m, tea.Quit
		case "q", "ctrl+c":
			m.chosen = ActionQuit
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
	fmt.Fprintf(&b, "%s\n", tuiTitle.Render("skillcloud"))
	fmt.Fprintf(&b, "%s\n\n", tuiDim.Render(fmt.Sprintf("Target: %s · Scope: %s · Mode: %s", m.Target, m.Scope, m.Mode)))
	for i, name := range appViews {
		marker := " "
		if i == m.view {
			marker = tuiCursor.Render(">")
		}
		description := ""
		if i < len(appDescriptions) {
			description = appDescriptions[i]
		}
		fmt.Fprintf(&b, "%s %-10s %s\n", marker, name, tuiDim.Render(description))
	}
	if strings.TrimSpace(m.Status) != "" {
		fmt.Fprintf(&b, "\n%s\n", tuiEnabled.Render(m.Status))
	}
	b.WriteString("\n" + tuiDim.Render("Up/Down select · Enter run · Tab cycle · q quit") + "\n")
	return b.String()
}

// AppResult reports which menu entry the user selected.
type AppResult struct {
	Action AppAction
}

// RunApp launches the launcher menu and returns the chosen action. It does not
// run the action itself; the caller dispatches on AppResult.Action so that each
// action can take over the terminal in its own tea program.
func RunApp(opts AppOptions) (AppResult, error) {
	result, err := tea.NewProgram(NewAppModel(opts)).Run()
	if err != nil {
		return AppResult{}, err
	}
	model, ok := result.(AppModel)
	if !ok {
		return AppResult{}, fmt.Errorf("unexpected model type %T", result)
	}
	return AppResult{Action: model.chosen}, nil
}
