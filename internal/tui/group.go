package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GroupOption is one group row in the group manager.
type GroupOption struct {
	ID          string
	Name        string
	Category    string
	Title       string
	Description string
	SkillCount  int
	HasAgents   bool
}

// GroupOptions configures the group manager.
type GroupOptions struct {
	Groups           []GroupOption
	Target           string
	Scope            string
	Mode             string
	InitiallyEnabled []string
}

// GroupResult reports the groups the user wants enabled.
type GroupResult struct {
	Apply       bool
	SelectedIDs []string
	InitialIDs  []string
}

type groupScreen int

const (
	groupScreenList groupScreen = iota
	groupScreenConfirm
)

// GroupModel is the group manager TUI.
type GroupModel struct {
	Groups    []GroupOption
	Target    string
	Scope     string
	Mode      string
	Width     int
	Height    int
	Cursor    int
	Query     string
	Searching bool
	Screen    groupScreen
	Selected  map[string]bool
	Initial   map[string]bool
	Applied   bool
}

func NewGroupModel(opts GroupOptions) GroupModel {
	selected := map[string]bool{}
	initial := map[string]bool{}
	for _, id := range opts.InitiallyEnabled {
		selected[id] = true
		initial[id] = true
	}
	m := GroupModel{
		Groups:   opts.Groups,
		Target:   opts.Target,
		Scope:    opts.Scope,
		Mode:     opts.Mode,
		Width:    100,
		Height:   24,
		Selected: selected,
		Initial:  initial,
	}
	for i, g := range m.Groups {
		if m.Selected[g.ID] {
			m.Cursor = i
			break
		}
	}
	return m
}

func (m GroupModel) Init() tea.Cmd { return nil }

func (m GroupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.Screen == groupScreenConfirm {
			return m.updateConfirm(msg)
		}
		if m.Searching {
			return m.updateSearch(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m GroupModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.Applied = false
		return m, tea.Quit
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.visibleGroups())-1 {
			m.Cursor++
		}
	case "home", "g":
		m.Cursor = 0
	case "end", "G":
		if visible := m.visibleGroups(); len(visible) > 0 {
			m.Cursor = len(visible) - 1
		}
	case "pgup":
		m.Cursor -= m.pageSize()
		m.clampCursor()
	case "pgdown":
		m.Cursor += m.pageSize()
		m.clampCursor()
	case " ":
		visible := m.visibleGroups()
		if len(visible) > 0 {
			m.Toggle(visible[m.Cursor].ID)
		}
	case "/":
		m.Searching = true
	case "esc":
		m.Query = ""
		m.Cursor = 0
	case "enter":
		enable, disable := m.PendingDiff()
		if len(enable) == 0 && len(disable) == 0 {
			m.Applied = true
			return m, tea.Quit
		}
		m.Screen = groupScreenConfirm
	}
	return m, nil
}

func (m GroupModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m GroupModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.Applied = true
		return m, tea.Quit
	case "esc":
		m.Screen = groupScreenList
	case "q", "ctrl+c":
		m.Applied = false
		return m, tea.Quit
	}
	return m, nil
}

func (m GroupModel) View() string {
	if m.Screen == groupScreenConfirm {
		return m.confirmView()
	}
	return m.listView()
}

func (m GroupModel) listView() string {
	visible := m.visibleGroups()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", tuiTitle.Render(fmt.Sprintf("skillcloud · groups · %s · %s", m.Target, m.Mode)))
	if m.Searching {
		fmt.Fprintf(&b, "Search: %s_\n", m.Query)
	} else if m.Query != "" {
		fmt.Fprintf(&b, "Filter: %s  (esc clears)\n", m.Query)
	}
	fmt.Fprintf(&b, "%s\n", tuiDim.Render(fmt.Sprintf("View: All · %d group(s) · %d enabled", len(visible), len(m.Initial))))

	if len(visible) == 0 {
		b.WriteString("\nNo groups match this filter.\n\n")
	} else {
		start, end := m.window(len(visible))
		fmt.Fprintf(&b, "%s\n\n", tuiDim.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(visible))))
		for i := start; i < end; i++ {
			g := visible[i]
			cursor := " "
			if i == m.Cursor {
				cursor = tuiCursor.Render(">")
			}
			checked := tuiDim.Render("[ ]")
			if m.Selected[g.ID] {
				checked = tuiEnabled.Render("[x]")
			}
			name := padDisplay(g.Name, 16)
			category := padDisplay(g.Category, 10)
			skillCount := padDisplay(fmt.Sprintf("%d skills", g.SkillCount), 10)
			rules := "rules"
			if !g.HasAgents {
				rules = "no rules"
			}
			rules = padDisplay(rules, 8)
			descWidth := m.Width - 54
			if descWidth < 12 {
				descWidth = 12
			}
			description := g.Description
			if description == "" {
				description = g.Title
			}
			description = lipgloss.NewStyle().MaxWidth(descWidth).Render(description)
			fmt.Fprintf(&b, "%s %s %s %s %s %s %s\n", cursor, checked, name, category, skillCount, rules, tuiDim.Render(description))
		}
		b.WriteString("\n")
	}

	enable, disable := m.PendingDiff()
	fmt.Fprintf(&b, "Pending: %s %d, %s %d\n", tuiAdd.Render("+"), len(enable), tuiRemove.Render("-"), len(disable))
	b.WriteString(tuiDim.Render("Space toggle · Enter apply · / search · q cancel") + "\n")
	b.WriteString(tuiDim.Render("PgUp/PgDn · Home/End") + "\n")
	return b.String()
}

func padDisplay(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = lipgloss.NewStyle().MaxWidth(width).Render(value)
	if pad := width - lipgloss.Width(value); pad > 0 {
		value += strings.Repeat(" ", pad)
	}
	return value
}

func (m GroupModel) confirmView() string {
	enable, disable := m.PendingDiff()
	var b strings.Builder
	b.WriteString(tuiTitle.Render("Apply group changes?") + "\n\n")
	m.writeGroupSection(&b, "Enable", enable, "+", tuiAdd)
	m.writeGroupSection(&b, "Disable", disable, "-", tuiRemove)
	b.WriteString(tuiDim.Render("Enter confirm · Esc back · q cancel") + "\n")
	return b.String()
}

func (m GroupModel) writeGroupSection(b *strings.Builder, title string, ids []string, prefix string, style lipgloss.Style) {
	if len(ids) == 0 {
		return
	}
	fmt.Fprintf(b, "%s (%d)\n", title, len(ids))
	for _, id := range ids {
		label := id
		for _, g := range m.Groups {
			if g.ID == id {
				label = fmt.Sprintf("%s (%d skills)", g.Name, g.SkillCount)
				break
			}
		}
		fmt.Fprintf(b, "  %s %s\n", style.Render(prefix), label)
	}
	b.WriteString("\n")
}

func (m GroupModel) pageSize() int {
	if m.Height <= 0 {
		return 12
	}
	size := m.Height - 9
	if size < 5 {
		return 5
	}
	return size
}

func (m GroupModel) window(total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	size := m.pageSize()
	if total <= size {
		return 0, total
	}
	start := m.Cursor - size/2
	if start < 0 {
		start = 0
	}
	if start > total-size {
		start = total - size
	}
	return start, start + size
}

func (m GroupModel) visibleGroups() []GroupOption {
	query := strings.ToLower(strings.TrimSpace(m.Query))
	if query == "" {
		return m.Groups
	}
	var visible []GroupOption
	for _, g := range m.Groups {
		haystack := strings.ToLower(strings.Join([]string{g.ID, g.Name, g.Category, g.Title, g.Description}, " "))
		if strings.Contains(haystack, query) {
			visible = append(visible, g)
		}
	}
	return visible
}

func (m *GroupModel) clampCursor() {
	visible := m.visibleGroups()
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

func (m GroupModel) Toggle(id string) {
	m.Selected[id] = !m.Selected[id]
}

func (m GroupModel) PendingDiff() ([]string, []string) {
	var enable []string
	var disable []string
	for _, g := range m.Groups {
		if m.Selected[g.ID] && !m.Initial[g.ID] {
			enable = append(enable, g.ID)
		}
		if !m.Selected[g.ID] && m.Initial[g.ID] {
			disable = append(disable, g.ID)
		}
	}
	return enable, disable
}

func (m GroupModel) selectedIDs() []string {
	var ids []string
	for _, g := range m.Groups {
		if m.Selected[g.ID] {
			ids = append(ids, g.ID)
		}
	}
	return ids
}

func (m GroupModel) initialIDs() []string {
	var ids []string
	for _, g := range m.Groups {
		if m.Initial[g.ID] {
			ids = append(ids, g.ID)
		}
	}
	return ids
}

func (m GroupModel) Result() GroupResult {
	return GroupResult{
		Apply:       m.Applied,
		SelectedIDs: m.selectedIDs(),
		InitialIDs:  m.initialIDs(),
	}
}

func ManageGroups(opts GroupOptions) (GroupResult, error) {
	program := tea.NewProgram(NewGroupModel(opts))
	result, err := program.Run()
	if err != nil {
		return GroupResult{}, err
	}
	model := result.(GroupModel)
	return model.Result(), nil
}
