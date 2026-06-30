package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skillcloud/skillcloud/internal/skill"
)

type ManageOptions struct {
	Skills           []skill.Skill
	Target           string
	Scope            string
	Mode             string
	InitiallyEnabled []string
}

type ManageResult struct {
	Apply       bool
	SelectedIDs []string
	InitialIDs  []string
}

type manageView int

const (
	viewAll manageView = iota
	viewEnabled
	viewDisabled
	viewChanged
)

type manageScreen int

const (
	screenList manageScreen = iota
	screenConfirm
)

type ManageModel struct {
	Skills     []skill.Skill
	Target     string
	Scope      string
	Mode       string
	Cursor     int
	Query      string
	Searching  bool
	FilterView manageView
	Screen     manageScreen
	Selected   map[string]bool
	Initial    map[string]bool
	Applied    bool
}

func NewManageModel(opts ManageOptions) ManageModel {
	selected := map[string]bool{}
	initial := map[string]bool{}
	for _, id := range opts.InitiallyEnabled {
		selected[id] = true
		initial[id] = true
	}
	return ManageModel{
		Skills:   opts.Skills,
		Target:   opts.Target,
		Scope:    opts.Scope,
		Mode:     opts.Mode,
		Selected: selected,
		Initial:  initial,
	}
}

func (m ManageModel) Init() tea.Cmd { return nil }

func (m ManageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Screen == screenConfirm {
			return m.updateConfirm(msg)
		}
		if m.Searching {
			return m.updateSearch(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m ManageModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.Applied = false
		return m, tea.Quit
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.visibleSkills())-1 {
			m.Cursor++
		}
	case " ":
		visible := m.visibleSkills()
		if len(visible) > 0 {
			m.Toggle(visible[m.Cursor].ID)
		}
	case "/":
		m.Searching = true
	case "esc":
		m.Query = ""
		m.Cursor = 0
	case "tab":
		m.FilterView = (m.FilterView + 1) % 4
		m.clampCursor()
	case "enter":
		enable, disable := m.PendingDiff()
		if len(enable) == 0 && len(disable) == 0 {
			m.Applied = true
			return m, tea.Quit
		}
		m.Screen = screenConfirm
	}
	return m, nil
}

func (m ManageModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.Applied = false
		return m, tea.Quit
	case "esc":
		m.Searching = false
		m.Query = ""
		m.Cursor = 0
	case "enter":
		m.Searching = false
	case "backspace":
		if len(m.Query) > 0 {
			m.Query = m.Query[:len(m.Query)-1]
			m.Cursor = 0
		}
	default:
		if len(msg.Runes) > 0 {
			m.Query += string(msg.Runes)
			m.Cursor = 0
		}
	}
	m.clampCursor()
	return m, nil
}

func (m ManageModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.Applied = true
		return m, tea.Quit
	case "esc":
		m.Screen = screenList
	case "q", "ctrl+c":
		m.Applied = false
		return m, tea.Quit
	}
	return m, nil
}

func (m ManageModel) View() string {
	if m.Screen == screenConfirm {
		return m.confirmView()
	}
	return m.listView()
}

func (m ManageModel) listView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Target: %s | Scope: %s | Mode: %s\n", m.Target, m.Scope, m.Mode)
	if m.Searching || m.Query != "" {
		fmt.Fprintf(&b, "Filter: %s\n", m.Query)
	}
	fmt.Fprintf(&b, "View: %s\n\n", m.viewName())

	visible := m.visibleSkills()
	if len(visible) == 0 {
		b.WriteString("No matching skills\n\n")
	} else {
		for i, s := range visible {
			cursor := " "
			if i == m.Cursor {
				cursor = ">"
			}
			checked := "[ ]"
			if m.Selected[s.ID] {
				checked = "[x]"
			}
			marker := " "
			if m.Selected[s.ID] && !m.Initial[s.ID] {
				marker = "+"
			}
			if !m.Selected[s.ID] && m.Initial[s.ID] {
				marker = "-"
			}
			fmt.Fprintf(&b, "%s %s%s %-28s %s\n", cursor, marker, checked, s.ID, s.Description)
		}
		b.WriteString("\n")
	}

	enable, disable := m.PendingDiff()
	fmt.Fprintf(&b, "Pending: enable %d, disable %d\n", len(enable), len(disable))
	b.WriteString("Space toggle | / search | Tab view | Enter apply | q cancel\n")
	return b.String()
}

func (m ManageModel) confirmView() string {
	enable, disable := m.PendingDiff()
	var b strings.Builder
	b.WriteString("Apply skill changes?\n\n")
	if len(enable) > 0 {
		b.WriteString("Enable:\n")
		for _, id := range enable {
			fmt.Fprintf(&b, "  + %s\n", id)
		}
		b.WriteString("\n")
	}
	if len(disable) > 0 {
		b.WriteString("Disable:\n")
		for _, id := range disable {
			fmt.Fprintf(&b, "  - %s\n", id)
		}
		b.WriteString("\n")
	}
	b.WriteString("Enter confirm | Esc back | q cancel\n")
	return b.String()
}

func (m ManageModel) Toggle(id string) {
	m.Selected[id] = !m.Selected[id]
}

func (m ManageModel) PendingDiff() ([]string, []string) {
	var enable []string
	var disable []string
	for _, s := range m.Skills {
		if m.Selected[s.ID] && !m.Initial[s.ID] {
			enable = append(enable, s.ID)
		}
		if !m.Selected[s.ID] && m.Initial[s.ID] {
			disable = append(disable, s.ID)
		}
	}
	return enable, disable
}

func (m ManageModel) Result() ManageResult {
	return ManageResult{
		Apply:       m.Applied,
		SelectedIDs: m.selectedIDs(),
		InitialIDs:  m.initialIDs(),
	}
}

func (m ManageModel) selectedIDs() []string {
	var ids []string
	for _, s := range m.Skills {
		if m.Selected[s.ID] {
			ids = append(ids, s.ID)
		}
	}
	return ids
}

func (m ManageModel) initialIDs() []string {
	var ids []string
	for _, s := range m.Skills {
		if m.Initial[s.ID] {
			ids = append(ids, s.ID)
		}
	}
	return ids
}

func (m ManageModel) visibleSkills() []skill.Skill {
	query := strings.ToLower(strings.TrimSpace(m.Query))
	var visible []skill.Skill
	for _, s := range m.Skills {
		if !m.matchesView(s.ID) {
			continue
		}
		if query != "" && !skillMatchesQuery(s, query) {
			continue
		}
		visible = append(visible, s)
	}
	return visible
}

func (m ManageModel) matchesView(id string) bool {
	changed := m.Selected[id] != m.Initial[id]
	switch m.FilterView {
	case viewEnabled:
		return m.Selected[id]
	case viewDisabled:
		return !m.Selected[id]
	case viewChanged:
		return changed
	default:
		return true
	}
}

func skillMatchesQuery(s skill.Skill, query string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		s.ID,
		s.Name,
		strings.ReplaceAll(s.ID, "/", " "),
		s.Description,
	}, " "))
	return strings.Contains(haystack, query)
}

func (m *ManageModel) clampCursor() {
	visible := m.visibleSkills()
	if len(visible) == 0 {
		m.Cursor = 0
		return
	}
	if m.Cursor >= len(visible) {
		m.Cursor = len(visible) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
}

func (m ManageModel) viewName() string {
	switch m.FilterView {
	case viewEnabled:
		return "Enabled"
	case viewDisabled:
		return "Disabled"
	case viewChanged:
		return "Changed"
	default:
		return "All"
	}
}

func Manage(opts ManageOptions) (ManageResult, error) {
	program := tea.NewProgram(NewManageModel(opts))
	result, err := program.Run()
	if err != nil {
		return ManageResult{}, err
	}
	model := result.(ManageModel)
	return model.Result(), nil
}
