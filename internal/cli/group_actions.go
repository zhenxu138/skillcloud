package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillcloud/skillcloud/internal/group"
	"github.com/skillcloud/skillcloud/internal/install"
	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
)

type groupUseResult struct {
	Skills        int
	AgentsUpdated bool
}

func applyGroupUse(root string, g group.Group, targetName string, mode string, noAgents bool, prefix string) (groupUseResult, error) {
	if mode != "link" && mode != "copy" {
		return groupUseResult{}, fmt.Errorf("unknown mode %q", mode)
	}
	prefix = normalizeAliasPrefix(prefix)
	if strings.ContainsAny(prefix, "/\\") || strings.ContainsAny(prefix, " \t\n") {
		return groupUseResult{}, fmt.Errorf("prefix %q cannot contain path separators or whitespace", prefix)
	}
	if len(g.Members) == 0 && (g.AgentsPath == "" || noAgents) {
		return groupUseResult{}, fmt.Errorf("group %q has no skills and no AGENTS.md rules", g.Name)
	}
	projectConfig, err := project.Load(root)
	if err != nil {
		return groupUseResult{}, err
	}
	globalConfig, _, err := loadGlobalConfigAndRepoDir()
	if err != nil {
		return groupUseResult{}, err
	}
	destRoot, err := targetDestRoot(globalConfig, targetName, "project", root)
	if err != nil {
		return groupUseResult{}, err
	}

	existing := map[string]bool{}
	existingAlias := map[string]string{}
	usedAlias := map[string]string{}
	for _, ref := range projectConfig.Targets[targetName].Skills {
		existing[ref.ID] = true
		existingAlias[ref.ID] = ref.As
		if ref.As != "" {
			usedAlias[ref.As] = ref.ID
		}
	}
	selected := make([]skill.Skill, 0, len(g.Members))
	selectedIDs := map[string]bool{}
	for _, member := range g.Members {
		s := member.Skill
		alias := member.As
		if prefix != "" {
			alias = prefix + member.As
		}
		if currentAlias := existingAlias[s.ID]; currentAlias != "" {
			s.Name = currentAlias
		} else {
			s.Name = alias
		}
		selected = append(selected, s)
		selectedIDs[s.ID] = true
	}
	var conflicts []string
	for _, s := range selected {
		if existingID, ok := usedAlias[s.Name]; ok && existingID != s.ID {
			conflicts = append(conflicts, fmt.Sprintf("%s is already used by %s", s.Name, existingID))
		}
	}
	if len(conflicts) > 0 {
		suggestion := ""
		if prefix == "" {
			suggestion = fmt.Sprintf("; rerun with --prefix %s to load both groups", g.Name)
		}
		return groupUseResult{}, fmt.Errorf("alias conflict(s): %s%s", strings.Join(conflicts, "; "), suggestion)
	}

	plan, err := install.PlanEnable(selected, targetName, "project", mode, destRoot, projectConfig)
	if err != nil {
		return groupUseResult{}, err
	}
	if err := install.CheckReplace(plan.Actions); err != nil {
		return groupUseResult{}, err
	}
	if err := install.Execute(plan.Actions); err != nil {
		return groupUseResult{}, err
	}

	updated := plan.ProjectConfig
	targetConfig := updated.Targets[targetName]
	for i := range targetConfig.Skills {
		ref := &targetConfig.Skills[i]
		if !selectedIDs[ref.ID] {
			continue
		}
		if existing[ref.ID] {
			if len(ref.GroupIDs()) > 0 && !ref.HasGroup(g.ID) {
				ref.AddGroup(g.ID)
			}
			continue
		}
		ref.AddGroup(g.ID)
	}
	foundGroup := false
	for i := range targetConfig.Groups {
		if targetConfig.Groups[i].ID == g.ID {
			targetConfig.Groups[i].Name = g.Name
			targetConfig.Groups[i].NoAgents = noAgents
			foundGroup = true
		}
	}
	if !foundGroup {
		targetConfig.Groups = append(targetConfig.Groups, project.GroupRef{ID: g.ID, Name: g.Name, NoAgents: noAgents})
	}
	updated.Targets[targetName] = targetConfig
	if err := project.Save(root, updated); err != nil {
		return groupUseResult{}, err
	}
	groups, _, _, err := loadGroups()
	if err != nil {
		return groupUseResult{}, err
	}
	if err := refreshGroupRules(root, updated, groups); err != nil {
		return groupUseResult{}, err
	}

	return groupUseResult{Skills: len(selected), AgentsUpdated: !noAgents && g.AgentsPath != ""}, nil
}

func normalizeAliasPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return ""
	}
	if !strings.HasSuffix(prefix, "-") && !strings.HasSuffix(prefix, "_") && !strings.HasSuffix(prefix, ".") {
		prefix += "-"
	}
	return prefix
}

func applyGroupRemove(root string, g group.Group, targetName string) (int, error) {
	projectConfig, err := project.Load(root)
	if err != nil {
		return 0, err
	}
	globalConfig, _, err := loadGlobalConfigAndRepoDir()
	if err != nil {
		return 0, err
	}
	if targetName != "" {
		if _, ok := projectConfig.Targets[targetName]; !ok {
			return 0, fmt.Errorf("target %q is not configured in this project", targetName)
		}
	}

	removed := 0
	for name, targetConfig := range projectConfig.Targets {
		if targetName != "" && name != targetName {
			continue
		}
		destRoot, err := targetDestRoot(globalConfig, name, "project", root)
		if err != nil {
			return removed, err
		}
		keptSkills := make([]project.SkillRef, 0, len(targetConfig.Skills))
		for _, ref := range targetConfig.Skills {
			if !ref.HasGroup(g.ID) {
				keptSkills = append(keptSkills, ref)
				continue
			}
			ref.RemoveGroup(g.ID)
			if len(ref.GroupIDs()) > 0 {
				keptSkills = append(keptSkills, ref)
				continue
			}
			if ref.As == "" {
				return removed, fmt.Errorf("skill %q has no projection alias", ref.ID)
			}
			dest := filepath.Join(destRoot, ref.As)
			if _, err := os.Lstat(dest); err == nil {
				expected := install.ProjectionManifest{SourceID: ref.ID, Target: name, Scope: "project"}
				if err := install.CanRemoveProjection(dest, expected); err != nil {
					return removed, fmt.Errorf("remove %s: %w", dest, err)
				}
				if err := install.RemoveProjection(dest); err != nil {
					return removed, err
				}
				removed++
			} else if !os.IsNotExist(err) {
				return removed, err
			}
		}
		targetConfig.Skills = keptSkills
		keptGroups := make([]project.GroupRef, 0, len(targetConfig.Groups))
		for _, groupRef := range targetConfig.Groups {
			if groupRef.ID != g.ID {
				keptGroups = append(keptGroups, groupRef)
			}
		}
		targetConfig.Groups = keptGroups
		projectConfig.Targets[name] = targetConfig
	}
	if err := project.Save(root, projectConfig); err != nil {
		return removed, err
	}
	groups, _, _, err := loadGroups()
	if err != nil {
		return removed, err
	}
	if err := refreshGroupRules(root, projectConfig, groups); err != nil {
		return removed, err
	}
	return removed, nil
}
