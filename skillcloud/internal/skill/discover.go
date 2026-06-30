package skill

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseSkillMarkdown extracts required skill metadata from YAML frontmatter.
func ParseSkillMarkdown(data []byte) (string, string, error) {
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	lines := bytes.SplitAfter(data, []byte("\n"))
	if len(lines) == 0 || trimLineEnding(lines[0]) != "---" {
		return "", "", fmt.Errorf("skill markdown missing YAML frontmatter")
	}

	var frontmatter []byte
	closed := false
	for _, line := range lines[1:] {
		if trimLineEnding(line) == "---" {
			closed = true
			break
		}
		frontmatter = append(frontmatter, line...)
	}
	if !closed {
		return "", "", fmt.Errorf("skill markdown frontmatter is not closed")
	}

	var meta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
		return "", "", err
	}
	if meta.Name == "" {
		return "", "", fmt.Errorf("skill markdown missing name")
	}
	if meta.Description == "" {
		return "", "", fmt.Errorf("skill markdown missing description")
	}
	return meta.Name, meta.Description, nil
}

func trimLineEnding(line []byte) string {
	return string(bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r")))
}

// Discover scans repoDir/skills for SKILL.md files and returns skills sorted by ID.
func Discover(repoDir string) ([]Skill, error) {
	skillsDir := filepath.Join(repoDir, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil, nil
	}
	var skills []Skill

	err := filepath.WalkDir(skillsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != skillsDir && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != "SKILL.md" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name, description, err := ParseSkillMarkdown(data)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		dir := filepath.Dir(path)
		rel, err := filepath.Rel(skillsDir, dir)
		if err != nil {
			return err
		}
		id := filepath.ToSlash(rel)
		skills = append(skills, Skill{
			ID:          id,
			Name:        name,
			Description: description,
			Path:        dir,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].ID < skills[j].ID
	})
	return skills, nil
}

// Resolve finds skills by exact ID, exact name, or prefix when pattern ends with /*.
func Resolve(skills []Skill, pattern string) []Skill {
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		var matched []Skill
		for _, skill := range skills {
			if strings.HasPrefix(skill.ID, prefix+"/") {
				matched = append(matched, skill)
			}
		}
		return matched
	}

	var matched []Skill
	for _, skill := range skills {
		if skill.ID == pattern || skill.Name == pattern {
			matched = append(matched, skill)
		}
	}
	return matched
}
