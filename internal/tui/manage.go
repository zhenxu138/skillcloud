package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/skillcloud/skillcloud/internal/skill"
)

var (
	tuiTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	tuiCursor  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	tuiEnabled = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	tuiAdd     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	tuiRemove  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	tuiDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type ManageOptions struct {
	Skills           []skill.Skill
	Target           string
	Scope            string
	Mode             string
	InitiallyEnabled []string
	// HideEnabledCount is used by select-only flows that do not load the
	// project's existing enabled skills.
	HideEnabledCount bool
	// InitialFilterView preselects the all/enabled/disabled/changed filter.
	// 0=all, 1=enabled, 2=disabled, 3=changed. Defaults to all when zero.
	InitialFilterView int
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
	Skills           []skill.Skill
	Target           string
	Scope            string
	Mode             string
	Width            int
	Height           int
	HideEnabledCount bool
	Cursor           int
	Query            string
	Searching        bool
	FilterView       manageView
	Screen           manageScreen
	Selected         map[string]bool
	Initial          map[string]bool
	Applied          bool
}

func NewManageModel(opts ManageOptions) ManageModel {
	selected := map[string]bool{}
	initial := map[string]bool{}
	for _, id := range opts.InitiallyEnabled {
		selected[id] = true
		initial[id] = true
	}
	m := ManageModel{
		Skills:           opts.Skills,
		Target:           opts.Target,
		Scope:            opts.Scope,
		Mode:             opts.Mode,
		Width:            100,
		Height:           24,
		HideEnabledCount: opts.HideEnabledCount,
		Selected:         selected,
		Initial:          initial,
		FilterView:       manageView(opts.InitialFilterView),
	}
	if m.FilterView == viewAll {
		for i, s := range m.Skills {
			if m.Selected[s.ID] {
				m.Cursor = i
				break
			}
		}
	}
	return m
}

func (m ManageModel) Init() tea.Cmd { return nil }

func (m ManageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
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
	case "home", "g":
		m.Cursor = 0
	case "end", "G":
		if visible := m.visibleSkills(); len(visible) > 0 {
			m.Cursor = len(visible) - 1
		}
	case "pgup":
		m.Cursor -= m.pageSize()
		m.clampCursor()
	case "pgdown":
		m.Cursor += m.pageSize()
		m.clampCursor()
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
	visible := m.visibleSkills()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", tuiTitle.Render(fmt.Sprintf("skillcloud · %s · %s · %s", m.Target, m.Scope, m.Mode)))
	if m.Searching {
		fmt.Fprintf(&b, "Search: %s_\n", m.Query)
	} else if m.Query != "" {
		fmt.Fprintf(&b, "Filter: %s  (esc clears)\n", m.Query)
	}
	viewLine := fmt.Sprintf("View: %s · %d skill(s)", m.viewName(), len(visible))
	if !m.HideEnabledCount {
		viewLine += fmt.Sprintf(" · %d enabled", len(m.Initial))
	}
	fmt.Fprintf(&b, "%s\n", tuiDim.Render(viewLine))

	if len(visible) == 0 {
		b.WriteString("\nNo skills match this view or filter.\n\n")
	} else {
		start, end := m.window(len(visible))
		fmt.Fprintf(&b, "%s\n\n", tuiDim.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(visible))))
		for i := start; i < end; i++ {
			s := visible[i]
			cursor := " "
			if i == m.Cursor {
				cursor = tuiCursor.Render(">")
			}
			checked := tuiDim.Render("[ ]")
			if m.Selected[s.ID] {
				checked = tuiEnabled.Render("[x]")
			}
			marker := " "
			if m.Selected[s.ID] && !m.Initial[s.ID] {
				marker = tuiAdd.Render("+")
			} else if !m.Selected[s.ID] && m.Initial[s.ID] {
				marker = tuiRemove.Render("-")
			}
			group := padRight(groupLabel(s.ID), 12)
			name := padRight(skillLabel(s), 26)
			descWidth := m.Width - 47
			if descWidth < 12 {
				descWidth = 12
			}
			desc := truncateRunes(s.Description, descWidth)
			fmt.Fprintf(&b, "%s %s%s %s %s %s\n", cursor, marker, checked, group, name, tuiDim.Render(desc))
		}
		b.WriteString("\n")
	}

	enable, disable := m.PendingDiff()
	fmt.Fprintf(&b, "Pending: %s %d, %s %d\n", tuiAdd.Render("+"), len(enable), tuiRemove.Render("-"), len(disable))
	b.WriteString(tuiDim.Render("Space toggle · Enter apply · / search · Tab view") + "\n")
	b.WriteString(tuiDim.Render("PgUp/PgDn · Home/End · q cancel") + "\n")
	return b.String()
}

func (m ManageModel) confirmView() string {
	enable, disable := m.PendingDiff()
	var b strings.Builder
	b.WriteString(tuiTitle.Render("Apply skill changes?") + "\n\n")
	writeChangeSection(&b, "Enable", enable, "+", tuiAdd)
	writeChangeSection(&b, "Disable", disable, "-", tuiRemove)
	b.WriteString(tuiDim.Render("Enter confirm · Esc back · q cancel") + "\n")
	return b.String()
}

func writeChangeSection(b *strings.Builder, title string, ids []string, prefix string, style lipgloss.Style) {
	if len(ids) == 0 {
		return
	}
	fmt.Fprintf(b, "%s (%d)\n", title, len(ids))
	limit := len(ids)
	if limit > 8 {
		limit = 8
	}
	for _, id := range ids[:limit] {
		fmt.Fprintf(b, "  %s %s\n", style.Render(prefix), id)
	}
	if len(ids) > limit {
		fmt.Fprintf(b, "  %s\n", tuiDim.Render(fmt.Sprintf("… and %d more", len(ids)-limit)))
	}
	b.WriteString("\n")
}

func (m ManageModel) pageSize() int {
	if m.Height <= 0 {
		return 15
	}
	size := m.Height - 9
	if size < 5 {
		return 5
	}
	return size
}

func (m ManageModel) window(total int) (int, int) {
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

func groupLabel(id string) string {
	parts := strings.Split(strings.Trim(id, "/"), "/")
	if len(parts) >= 3 && parts[0] == "sop" {
		if parts[1] == "github-spec-kit" {
			return "spec-kit"
		}
		return parts[1]
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return "-"
}

func skillLabel(s skill.Skill) string {
	if strings.TrimSpace(s.Name) != "" {
		return s.Name
	}
	parts := strings.Split(strings.Trim(s.ID, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return s.ID
}

func padRight(value string, width int) string {
	runes := []rune(value)
	if width <= 0 {
		return ""
	}
	if len(runes) > width {
		return string(runes[:width-1]) + "…"
	}
	return value + strings.Repeat(" ", width-len(runes))
}

func truncateRunes(value string, width int) string {
	runes := []rune(strings.TrimSpace(value))
	if width <= 0 {
		return ""
	}
	if len(runes) <= width {
		return string(runes)
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "…"
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

// ViewEnabled returns the enabled-filter value for ManageOptions.InitialFilterView.
func ViewEnabled() int { return int(viewEnabled) }
