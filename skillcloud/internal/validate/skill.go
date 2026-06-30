package validate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const maxSkillNameLength = 64
const maxSkillDescriptionLength = 1024

var skillNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

type SkillMetadata struct {
	Name        string
	Description string
}

func Skill(skillDir string) (SkillMetadata, []error) {
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if os.IsNotExist(err) {
		return SkillMetadata{}, []error{fmt.Errorf("SKILL.md not found")}
	}
	if err != nil {
		return SkillMetadata{}, []error{err}
	}
	return SkillMarkdown(data)
}

func SkillMarkdown(data []byte) (SkillMetadata, []error) {
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	// Normalize CRLF to LF so frontmatter detection and YAML parsing are line-ending agnostic.
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return SkillMetadata{}, []error{fmt.Errorf("no YAML frontmatter found")}
	}

	parts := bytes.SplitN(data, []byte("\n---"), 2)
	if len(parts) != 2 {
		return SkillMetadata{}, []error{fmt.Errorf("invalid frontmatter format")}
	}
	frontmatterText := bytes.TrimPrefix(parts[0], []byte("---\n"))

	var raw map[string]any
	if err := yaml.Unmarshal(frontmatterText, &raw); err != nil {
		return SkillMetadata{}, []error{fmt.Errorf("invalid YAML in frontmatter: %w", err)}
	}
	if raw == nil {
		return SkillMetadata{}, []error{fmt.Errorf("frontmatter must be a YAML dictionary")}
	}

	allowed := map[string]bool{
		"name":                    true,
		"description":             true,
		"license":                 true,
		"allowed-tools":           true,
		"metadata":                true,
		"disable-model-invocation": true,
		"argument-hint":           true,
	}
	var unexpected []string
	for key := range raw {
		if !allowed[key] {
			unexpected = append(unexpected, key)
		}
	}
	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		return SkillMetadata{}, []error{fmt.Errorf("unexpected key(s) in SKILL.md frontmatter: %s", strings.Join(unexpected, ", "))}
	}

	name, ok := raw["name"].(string)
	if !ok {
		return SkillMetadata{}, []error{fmt.Errorf("missing or non-string 'name' in frontmatter")}
	}
	description, ok := raw["description"].(string)
	if !ok {
		return SkillMetadata{}, []error{fmt.Errorf("missing or non-string 'description' in frontmatter")}
	}

	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if err := validateSkillName(name); err != nil {
		return SkillMetadata{}, []error{err}
	}
	if err := validateSkillDescription(description); err != nil {
		return SkillMetadata{}, []error{err}
	}
	return SkillMetadata{Name: name, Description: description}, nil
}

func validateSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("missing 'name' in frontmatter")
	}
	if len(name) > maxSkillNameLength {
		return fmt.Errorf("name is too long (%d characters), maximum is %d", len(name), maxSkillNameLength)
	}
	if !skillNamePattern.MatchString(name) {
		return fmt.Errorf("name %q should be hyphen-case with lowercase letters, digits, and hyphens only", name)
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Contains(name, "--") {
		return fmt.Errorf("name %q cannot start/end with hyphen or contain consecutive hyphens", name)
	}
	return nil
}

func validateSkillDescription(description string) error {
	if description == "" {
		return fmt.Errorf("missing 'description' in frontmatter")
	}
	if strings.ContainsAny(description, "<>") {
		return fmt.Errorf("description cannot contain angle brackets (< or >)")
	}
	if len(description) > maxSkillDescriptionLength {
		return fmt.Errorf("description is too long (%d characters), maximum is %d", len(description), maxSkillDescriptionLength)
	}
	return nil
}
