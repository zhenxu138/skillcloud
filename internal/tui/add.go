package tui

import (
	"fmt"
	"strings"
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
