package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skillcloud/skillcloud/internal/agents"
	"github.com/skillcloud/skillcloud/internal/group"
	"github.com/skillcloud/skillcloud/internal/project"
)

type groupRuleEntry struct {
	Group    group.Group
	RulePath string
}

func groupRuleRelPath(groupID string) string {
	safe := strings.NewReplacer("/", "-", `\`, "-", ":", "-", " ", "-").Replace(groupID)
	return filepath.ToSlash(filepath.Join(".skillcloud", "groups", safe+".md"))
}

func refreshGroupRules(root string, cfg project.Config, groups []group.Group) error {
	agentsPath := filepath.Join(root, "AGENTS.md")
	if err := agents.RemoveLegacyGroupBlocks(agentsPath); err != nil {
		return err
	}
	if err := agents.RemoveGroupIndex(agentsPath); err != nil {
		return err
	}

	byID := make(map[string]group.Group, len(groups))
	for _, g := range groups {
		byID[g.ID] = g
	}
	var entries []groupRuleEntry
	usedPaths := map[string]bool{}
	wanted := map[string]bool{}
	for _, targetConfig := range cfg.Targets {
		for _, ref := range targetConfig.Groups {
			if _, ok := wanted[ref.ID]; !ok || !ref.NoAgents {
				wanted[ref.ID] = !ref.NoAgents
			}
		}
	}
	var ids []string
	for id, want := range wanted {
		if want {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		g, ok := byID[id]
		if !ok || g.AgentsPath == "" {
			_ = os.Remove(filepath.Join(root, filepath.FromSlash(groupRuleRelPath(id))))
			continue
		}
		data, err := os.ReadFile(g.AgentsPath)
		if err != nil {
			return err
		}
		rel := groupRuleRelPath(g.ID)
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(abs, []byte(renderGroupRuleFile(g, string(data))), 0o644); err != nil {
			return err
		}
		entries = append(entries, groupRuleEntry{Group: g, RulePath: rel})
		usedPaths[rel] = true
	}

	ruleDir := filepath.Join(root, ".skillcloud", "groups")
	if dirEntries, err := os.ReadDir(ruleDir); err == nil {
		for _, entry := range dirEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			rel := filepath.ToSlash(filepath.Join(".skillcloud", "groups", entry.Name()))
			if !usedPaths[rel] {
				_ = os.Remove(filepath.Join(ruleDir, entry.Name()))
			}
		}
	}

	if len(entries) == 0 {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Group.Category != entries[j].Group.Category {
			return entries[i].Group.Category < entries[j].Group.Category
		}
		return entries[i].Group.Name < entries[j].Group.Name
	})
	return agents.UpsertGroupIndex(agentsPath, renderGroupIndex(entries))
}

func renderGroupIndex(entries []groupRuleEntry) string {
	var b strings.Builder
	b.WriteString("## Skill groups\n\n")
	b.WriteString("Group rules are scoped. Apply a group's rules only to work that uses that group's skills. If a task uses several groups, read each group's rules and reconcile conflicts explicitly.\n\n")
	for _, entry := range entries {
		g := entry.Group
		fmt.Fprintf(&b, "- **%s** (%s, %d skills) — read `%s` before using skills from this group.", g.Name, g.Category, len(g.Members), entry.RulePath)
		if g.Entry != "" {
			fmt.Fprintf(&b, " Entry: `%s`.", g.Entry)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func renderGroupRuleFile(g group.Group, content string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", groupTitle(g))
	fmt.Fprintf(&b, "Category: %s\nGroup: %s\nSkills: %d", g.Category, g.ID, len(g.Members))
	if g.Entry != "" {
		fmt.Fprintf(&b, "\nEntry: %s", g.Entry)
	}
	b.WriteString("\n\n")
	b.WriteString(stripLeadingHeading(strings.TrimSpace(content)))
	b.WriteString("\n")
	return b.String()
}

func stripLeadingHeading(content string) string {
	lines := strings.Split(content, "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "# ") {
		i++
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
		return strings.Join(lines[i:], "\n")
	}
	return content
}
