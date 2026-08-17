package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type AddOptions struct{}

type AddPreview struct {
	ID          string
	Name        string
	Description string
}

type AddModel struct {
	preview         *AddPreview
	validationError string
	conflict        string
}

func NewAddModel(opts AddOptions) AddModel {
	return AddModel{}
}

func (m *AddModel) SetValidationError(message string) {
	m.validationError = message
	m.conflict = ""
	m.preview = nil
}

func (m *AddModel) SetConflict(message string) {
	m.conflict = message
	m.validationError = ""
}

func (m *AddModel) SetPreview(preview AddPreview) {
	m.preview = &preview
	m.validationError = ""
	m.conflict = ""
}

func (m AddModel) View() string {
	var b strings.Builder
	b.WriteString("Add Skill\n\n")
	if m.validationError != "" {
		fmt.Fprintf(&b, "Validation error: %s\n", m.validationError)
		return b.String()
	}
	if m.conflict != "" {
		fmt.Fprintf(&b, "Conflict: %s\n", m.conflict)
		return b.String()
	}
	if m.preview != nil {
		fmt.Fprintf(&b, "ID: %s\nName: %s\nDescription: %s\n", m.preview.ID, m.preview.Name, m.preview.Description)
		return b.String()
	}
	b.WriteString("Enter a local skill directory to import.\n")
	return b.String()
}

// AddFormResult is what RunAddForm returns after the user fills or cancels the
// import form.
type AddFormResult struct {
	SourceDir string
	ID        string
	Replace   bool
	Canceled  bool
}

type addField int

const (
	addFieldSource addField = iota
	addFieldID
	addFieldReplace
	addFieldSubmit
)

// AddFormModel is an interactive form for importing a local skill directory
// into the library. It collects the source path and the library id, offers a
// replace toggle, and submits on Enter.
type AddFormModel struct {
	Source      string
	ID          string
	Replace     bool
	field       addField
	canceled    bool
	submitted   bool
	wantBrowser bool
}

func NewAddFormModel() AddFormModel {
	return AddFormModel{field: addFieldSource}
}

func (m AddFormModel) Init() tea.Cmd { return nil }

func (m AddFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	kmsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch kmsg.String() {
	case "ctrl+c", "esc":
		m.canceled = true
		return m, tea.Quit
	case "b":
		if m.field == addFieldSource {
			m.wantBrowser = true
			return m, tea.Quit
		}
	case "up", "k":
		if m.field > addFieldSource {
			m.field--
		}
	case "down", "tab", "j":
		if m.field < addFieldSubmit {
			m.field++
		}
	case "backspace":
		switch m.field {
		case addFieldSource:
			if len(m.Source) > 0 {
				m.Source = m.Source[:len(m.Source)-1]
			}
		case addFieldID:
			if len(m.ID) > 0 {
				m.ID = m.ID[:len(m.ID)-1]
			}
		}
	case "enter":
		switch m.field {
		case addFieldSource, addFieldID:
			if m.field < addFieldSubmit {
				m.field++
			}
		case addFieldReplace:
			m.Replace = !m.Replace
		case addFieldSubmit:
			if m.Source != "" && m.ID != "" {
				m.submitted = true
				return m, tea.Quit
			}
		}
	case " ":
		if m.field == addFieldReplace {
			m.Replace = !m.Replace
		} else if m.field == addFieldSource {
			m.Source += " "
		} else if m.field == addFieldID {
			m.ID += " "
		}
	default:
		if len(kmsg.Runes) > 0 && (m.field == addFieldSource || m.field == addFieldID) {
			for _, r := range kmsg.Runes {
				if m.field == addFieldSource {
					m.Source += string(r)
				} else {
					m.ID += string(r)
				}
			}
		}
	}
	return m, nil
}

func (m AddFormModel) View() string {
	var b strings.Builder
	b.WriteString("Add a skill to the library\n\n")
	b.WriteString(m.textField("Source dir", m.Source, addFieldSource))
	b.WriteString(m.textField("Skill id (--as)", m.ID, addFieldID))
	fmt.Fprintf(&b, "%s Replace existing   [%s]\n", m.rowMarker(addFieldReplace), m.replaceMark())
	fmt.Fprintf(&b, "%s Submit\n", m.rowMarker(addFieldSubmit))
	b.WriteString("\nUp/Down move | Enter next/toggle/submit | b browse | Esc cancel\n")
	if m.Source == "" || m.ID == "" {
		b.WriteString("Fill in source dir and skill id to submit.\n")
	}
	return b.String()
}

func (m AddFormModel) textField(label string, value string, field addField) string {
	cursor := "_"
	if value != "" {
		cursor = ""
	}
	return fmt.Sprintf("%s %-16s %s%s\n", m.rowMarker(field), label, value, cursor)
}

func (m AddFormModel) rowMarker(field addField) string {
	if m.field == field {
		return ">"
	}
	return " "
}

func (m AddFormModel) replaceMark() string {
	if m.Replace {
		return "x"
	}
	return " "
}

// RunAddForm launches the import form and returns the collected values. If the
// user canceled, Canceled is true.
func RunAddForm() (AddFormResult, error) {
	source := ""
	for {
		form := NewAddFormModel()
		form.Source = source
		result, err := tea.NewProgram(form).Run()
		if err != nil {
			return AddFormResult{}, err
		}
		model, ok := result.(AddFormModel)
		if !ok {
			return AddFormResult{}, fmt.Errorf("unexpected model type %T", result)
		}
		if model.canceled {
			return AddFormResult{Canceled: true}, nil
		}
		if model.wantBrowser {
			browser, berr := RunSourceBrowser(model.Source)
			if berr != nil {
				return AddFormResult{}, berr
			}
			if !browser.Canceled {
				source = browser.Path
			}
			continue
		}
		return AddFormResult{
			SourceDir: model.Source,
			ID:        model.ID,
			Replace:   model.Replace,
			Canceled:  model.canceled,
		}, nil
	}
}

// SourceBrowserResult is what RunSourceBrowser returns.
type SourceBrowserResult struct {
	Path     string
	Canceled bool
}

type browserEntry struct {
	name string
	dir  bool
}

// SourceBrowserModel lets the user navigate the filesystem and pick a skill
// source directory. Enter on a directory descends into it; Enter on ".."
// ascends; Enter on the explicit "use this directory" row picks the current
// directory as the source. It returns the absolute path.
type SourceBrowserModel struct {
	cwd      string
	entries  []browserEntry
	cursor   int
	picked   bool
	canceled bool
}

func NewSourceBrowserModel(startDir string) SourceBrowserModel {
	if startDir == "" {
		startDir, _ = os.Getwd()
	}
	m := SourceBrowserModel{cwd: startDir}
	m.reload()
	return m
}

func (m *SourceBrowserModel) reload() {
	entries, err := os.ReadDir(m.cwd)
	m.entries = m.entries[:0]
	if err == nil {
		var dirs, files []browserEntry
		for _, e := range entries {
			if e.IsDir() {
				if e.Name()[0] == '.' {
					continue
				}
				dirs = append(dirs, browserEntry{name: e.Name(), dir: true})
			} else {
				files = append(files, browserEntry{name: e.Name(), dir: false})
			}
		}
		sortBrowserEntries(dirs)
		sortBrowserEntries(files)
		m.entries = append(m.entries, browserEntry{name: "..", dir: true})
		m.entries = append(m.entries, dirs...)
		m.entries = append(m.entries, files...)
	} else {
		m.entries = append(m.entries, browserEntry{name: "..", dir: true})
	}
	if m.cursor >= len(m.entries) {
		m.cursor = len(m.entries) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func sortBrowserEntries(entries []browserEntry) {
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j-1].name > entries[j].name; j-- {
			entries[j-1], entries[j] = entries[j], entries[j-1]
		}
	}
}

func (m SourceBrowserModel) Init() tea.Cmd { return nil }

func (m SourceBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	kmsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch kmsg.String() {
	case "ctrl+c", "esc":
		m.canceled = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.cursor]
		if entry.name == ".." {
			parent := filepath.Dir(m.cwd)
			if parent != m.cwd {
				m.cwd = parent
			}
			m.cursor = 0
			m.reload()
		} else if entry.dir {
			m.cwd = filepath.Join(m.cwd, entry.name)
			m.cursor = 0
			m.reload()
		} else {
			// Selecting a file is a no-op; source must be a directory.
		}
	case " ", "s":
		// Space or s picks the current directory as the source.
		m.picked = true
		return m, tea.Quit
	case "q":
		m.canceled = true
		return m, tea.Quit
	}
	return m, nil
}

func (m SourceBrowserModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Browse for skill source\n%s\n\n", m.cwd)
	if _, err := os.Stat(filepath.Join(m.cwd, "SKILL.md")); err == nil {
		b.WriteString("> use this directory (Space)\n\n")
	}
	for i, e := range m.entries {
		marker := " "
		if i == m.cursor {
			marker = ">"
		}
		slash := ""
		if e.dir {
			slash = "/"
		}
		fmt.Fprintf(&b, "%s %s%s\n", marker, e.name, slash)
	}
	b.WriteString("\nUp/Down move | Enter open | Space pick | Esc back\n")
	return b.String()
}

// RunSourceBrowser launches the file browser and returns the selected path.
func RunSourceBrowser(startDir string) (SourceBrowserResult, error) {
	result, err := tea.NewProgram(NewSourceBrowserModel(startDir)).Run()
	if err != nil {
		return SourceBrowserResult{}, err
	}
	model, ok := result.(SourceBrowserModel)
	if !ok {
		return SourceBrowserResult{}, fmt.Errorf("unexpected model type %T", result)
	}
	return SourceBrowserResult{Path: model.cwd, Canceled: model.canceled}, nil
}
