package group

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skillcloud/skillcloud/internal/skill"
	"gopkg.in/yaml.v3"
)

// ManifestFile marks a directory as a reusable skill group.
const ManifestFile = "GROUP.yaml"

// Member selects one skill for a group. A member can use an exact ID, a unique
// skill name, or both. As overrides the projected directory name.
type Member struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	As   string `yaml:"as"`
}

func (m *Member) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		m.ID = strings.TrimSpace(value.Value)
		if m.ID == "" {
			return fmt.Errorf("group member cannot be empty")
		}
		return nil
	}
	var raw struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
		As   string `yaml:"as"`
	}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	m.ID = strings.TrimSpace(raw.ID)
	m.Name = strings.TrimSpace(raw.Name)
	m.As = strings.TrimSpace(raw.As)
	if m.ID == "" && m.Name == "" {
		return fmt.Errorf("group member requires id or name")
	}
	return nil
}

// Manifest is the GROUP.yaml schema.
type Manifest struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Entry       string   `yaml:"entry"`
	Agents      string   `yaml:"agents"`
	Skills      []Member `yaml:"skills"`
}

// MemberRef is a resolved group member.
type MemberRef struct {
	Skill skill.Skill
	As    string
}

// Group is a reusable set of skills with optional AGENTS.md constraints.
type Group struct {
	ID          string
	Name        string
	Title       string
	Description string
	Entry       string
	Path        string
	AgentsPath  string
	Members     []MemberRef
}

// Discover finds GROUP.yaml files under groups/ and skills/. Groups under
// groups/ must list members explicitly; groups under skills/ default to every
// skill inside their directory when skills is omitted.
func Discover(repoDir string, skills []skill.Skill) ([]Group, error) {
	if len(skills) == 0 {
		var err error
		skills, err = skill.Discover(repoDir)
		if err != nil {
			return nil, err
		}
	}

	roots := []struct {
		path   string
		inline bool
	}{
		{path: filepath.Join(repoDir, "groups"), inline: false},
		{path: filepath.Join(repoDir, "skills"), inline: true},
	}

	var groups []Group
	for _, root := range roots {
		if _, err := os.Stat(root.path); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(root.path, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if path != root.path && strings.HasPrefix(entry.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Name() != ManifestFile {
				return nil
			}
			g, err := loadGroup(repoDir, root.inline, path, skills)
			if err != nil {
				return err
			}
			groups = append(groups, g)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	seenNames := map[string]string{}
	for _, g := range groups {
		if prev, ok := seenNames[g.Name]; ok {
			return nil, fmt.Errorf("duplicate group name %q in %s and %s", g.Name, prev, g.Path)
		}
		seenNames[g.Name] = g.Path
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })
	return groups, nil
}

func loadGroup(repoDir string, inline bool, manifestPath string, skills []skill.Skill) (Group, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return Group{}, err
	}
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return Group{}, fmt.Errorf("%s: %w", manifestPath, err)
	}

	dir := filepath.Dir(manifestPath)
	relRepo, err := filepath.Rel(repoDir, dir)
	if err != nil {
		return Group{}, err
	}
	relRepo = filepath.ToSlash(relRepo)
	relSkills := strings.TrimPrefix(relRepo, "skills/")
	if relSkills == relRepo {
		relSkills = ""
	}

	name := strings.TrimSpace(manifest.Name)
	if name == "" {
		name = filepath.Base(dir)
	}
	members, err := resolveMembers(manifest, dir, inline, relSkills, skills)
	if err != nil {
		return Group{}, fmt.Errorf("%s: %w", manifestPath, err)
	}

	agentsPath := ""
	if strings.TrimSpace(manifest.Agents) != "" {
		agentsPath, err = resolveAgentsPath(dir, manifest.Agents)
		if err != nil {
			return Group{}, fmt.Errorf("%s: %w", manifestPath, err)
		}
	} else {
		candidate := filepath.Join(dir, "AGENTS.md")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			agentsPath = candidate
		}
	}

	return Group{
		ID:          relRepo,
		Name:        name,
		Title:       strings.TrimSpace(manifest.Title),
		Description: strings.TrimSpace(manifest.Description),
		Entry:       strings.TrimSpace(manifest.Entry),
		Path:        dir,
		AgentsPath:  agentsPath,
		Members:     members,
	}, nil
}

func resolveMembers(manifest Manifest, dir string, inline bool, relSkills string, skills []skill.Skill) ([]MemberRef, error) {
	var refs []MemberRef
	seenIDs := map[string]bool{}
	aliases := map[string]string{}
	add := func(s skill.Skill, alias string) error {
		if seenIDs[s.ID] {
			return fmt.Errorf("duplicate group member %q", s.ID)
		}
		seenIDs[s.ID] = true
		alias = strings.TrimSpace(alias)
		if alias == "" {
			alias = s.Name
		}
		if alias == "" {
			alias = filepath.Base(filepath.FromSlash(s.ID))
		}
		if prev, ok := aliases[alias]; ok {
			return fmt.Errorf("duplicate alias %q for %s and %s; set as to disambiguate", alias, prev, s.ID)
		}
		aliases[alias] = s.ID
		refs = append(refs, MemberRef{Skill: s, As: alias})
		return nil
	}

	if len(manifest.Skills) > 0 {
		for _, member := range manifest.Skills {
			s, err := resolveMember(skills, member)
			if err != nil {
				return nil, err
			}
			if err := add(s, member.As); err != nil {
				return nil, err
			}
		}
		return refs, nil
	}

	if !inline {
		return nil, nil
	}
	prefix := strings.TrimSuffix(relSkills, "/") + "/"
	for _, s := range skills {
		if s.ID == relSkills || strings.HasPrefix(s.ID, prefix) {
			if err := add(s, ""); err != nil {
				return nil, err
			}
		}
	}
	return refs, nil
}

func resolveMember(skills []skill.Skill, member Member) (skill.Skill, error) {
	if member.ID != "" {
		for _, s := range skills {
			if s.ID == member.ID {
				return s, nil
			}
		}
	}
	lookup := member.Name
	if lookup == "" {
		lookup = member.ID
	}
	if lookup != "" {
		var matches []skill.Skill
		for _, s := range skills {
			if s.Name == lookup {
				matches = append(matches, s)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			ids := make([]string, 0, len(matches))
			for _, s := range matches {
				ids = append(ids, s.ID)
			}
			return skill.Skill{}, fmt.Errorf("skill name %q is ambiguous: %s", lookup, strings.Join(ids, ", "))
		}
	}
	return skill.Skill{}, fmt.Errorf("skill %q not found", lookup)
}

func resolveAgentsPath(dir string, rel string) (string, error) {
	path := filepath.Join(dir, filepath.FromSlash(rel))
	relCheck, err := filepath.Rel(dir, path)
	if err != nil {
		return "", err
	}
	if relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("agents path escapes the group directory: %s", rel)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("agents path is a directory: %s", rel)
	}
	return path, nil
}

// Find resolves a group by unique name, relative path, or full manifest path.
func Find(groups []Group, query string) (Group, error) {
	q := strings.TrimSpace(strings.TrimSuffix(query, string(filepath.Separator)))
	q = filepath.ToSlash(q)
	var matches []Group
	for _, g := range groups {
		short := strings.TrimPrefix(g.ID, "skills/")
		if q == g.Name || q == g.ID || q == short || q == filepath.Base(g.Path) {
			matches = append(matches, g)
		}
	}
	if len(matches) == 0 {
		return Group{}, fmt.Errorf("group %q not found", query)
	}
	if len(matches) > 1 {
		ids := make([]string, 0, len(matches))
		for _, g := range matches {
			ids = append(ids, g.ID)
		}
		return Group{}, fmt.Errorf("group %q is ambiguous: %s", query, strings.Join(ids, ", "))
	}
	return matches[0], nil
}
