package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skillcloud/skillcloud/internal/agents"
	"github.com/skillcloud/skillcloud/internal/group"
	"github.com/skillcloud/skillcloud/internal/install"
	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
	"github.com/spf13/cobra"
)

func newGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage reusable groups of skills and their AGENTS.md rules",
	}
	cmd.AddCommand(newGroupListCommand())
	cmd.AddCommand(newGroupShowCommand())
	cmd.AddCommand(newGroupUseCommand())
	cmd.AddCommand(newGroupRemoveCommand())
	return cmd
}

func newGroupShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <group>",
		Short: "Show a group's skills, entry point, and AGENTS.md rules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, _, _, err := loadGroups()
			if err != nil {
				return err
			}
			g, err := group.Find(groups, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "id: %s\nname: %s\ntitle: %s\n", g.ID, g.Name, groupTitle(g))
			if g.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "description: %s\n", g.Description)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "entry: %s\nagents: %s\nskills: %d\n\n", g.Entry, g.AgentsPath, len(g.Members))
			for _, member := range g.Members {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", member.As, member.Skill.ID, member.Skill.Description)
			}
			return nil
		},
	}
}

func newGroupListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List skill groups in the repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, _, _, err := loadGroups()
			if err != nil {
				return err
			}
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			projectConfig, err := project.Load(root)
			if err != nil {
				return err
			}

			enabled := map[string][]string{}
			for targetName, targetConfig := range projectConfig.Targets {
				for _, groupRef := range targetConfig.Groups {
					enabled[groupRef.ID] = append(enabled[groupRef.ID], targetName)
				}
			}
			for id := range enabled {
				sort.Strings(enabled[id])
			}

			fmt.Fprintln(cmd.OutOrStdout(), "ID\tNAME\tSKILLS\tENTRY\tAGENTS\tENABLED")
			for _, g := range groups {
				agentsLabel := "-"
				if g.AgentsPath != "" {
					agentsLabel = "yes"
				}
				enabledLabel := "-"
				if targets := enabled[g.ID]; len(targets) > 0 {
					enabledLabel = strings.Join(targets, ",")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%d\t%s\t%s\t%s\n", g.ID, g.Name, len(g.Members), g.Entry, agentsLabel, enabledLabel)
			}
			printListFooter(cmd, fmt.Sprintf("(๑•̀ㅂ•́)و✧ 共 %d 个组。用 `skillcloud group show <组名>` 看成员，`skillcloud group use <组名>` 一次加载技能和规则。", len(groups)))
			return nil
		},
	}
}

func newGroupUseCommand() *cobra.Command {
	var targetName string
	var scope string
	var mode string
	var noAgents bool

	cmd := &cobra.Command{
		Use:   "use <group>",
		Short: "Enable every skill in a group and its AGENTS.md rules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if scope != "project" {
				return fmt.Errorf("group use currently supports project scope only")
			}
			if mode != "link" && mode != "copy" {
				return fmt.Errorf("unknown mode %q", mode)
			}

			groups, _, _, err := loadGroups()
			if err != nil {
				return err
			}
			g, err := group.Find(groups, args[0])
			if err != nil {
				return err
			}
			if len(g.Members) == 0 && (g.AgentsPath == "" || noAgents) {
				return fmt.Errorf("group %q has no skills and no AGENTS.md rules", g.Name)
			}

			root, err := os.Getwd()
			if err != nil {
				return err
			}
			projectConfig, err := project.Load(root)
			if err != nil {
				return err
			}
			globalConfig, _, err := loadGlobalConfigAndRepoDir()
			if err != nil {
				return err
			}
			destRoot, err := targetDestRoot(globalConfig, targetName, "project", root)
			if err != nil {
				return err
			}

			existing := map[string]bool{}
			existingAlias := map[string]string{}
			for _, ref := range projectConfig.Targets[targetName].Skills {
				existing[ref.ID] = true
				existingAlias[ref.ID] = ref.As
			}
			selected := make([]skill.Skill, 0, len(g.Members))
			selectedIDs := map[string]bool{}
			for _, member := range g.Members {
				s := member.Skill
				s.Name = member.As
				if alias := existingAlias[s.ID]; alias != "" {
					s.Name = alias
				}
				selected = append(selected, s)
				selectedIDs[s.ID] = true
			}

			plan, err := install.PlanEnable(selected, targetName, "project", mode, destRoot, projectConfig)
			if err != nil {
				return err
			}
			if err := install.CheckReplace(plan.Actions); err != nil {
				return err
			}
			if err := install.Execute(plan.Actions); err != nil {
				return err
			}

			updated := plan.ProjectConfig
			targetConfig := updated.Targets[targetName]
			for i := range targetConfig.Skills {
				ref := &targetConfig.Skills[i]
				if !selectedIDs[ref.ID] || existing[ref.ID] || ref.Group != "" {
					continue
				}
				ref.Group = g.ID
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
				return err
			}

			agentsUpdated := false
			if !noAgents && g.AgentsPath != "" {
				data, err := os.ReadFile(g.AgentsPath)
				if err != nil {
					return err
				}
				if err := agents.Upsert(filepath.Join(root, "AGENTS.md"), g.ID, groupTitle(g), string(data)); err != nil {
					return err
				}
				agentsUpdated = true
			}

			fmt.Fprintf(cmd.OutOrStdout(), "enabled %d skill(s) from group %s for %s\n", len(selected), g.Name, targetName)
			if agentsUpdated {
				fmt.Fprintln(cmd.OutOrStdout(), "updated AGENTS.md with group rules; start a new Codex session to load them")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&targetName, "target", "codex", "target agent: codex, claude, hermes")
	cmd.Flags().StringVar(&scope, "scope", "project", "scope: project")
	cmd.Flags().StringVar(&mode, "mode", "link", "mode: link or copy")
	cmd.Flags().BoolVar(&noAgents, "no-agents", false, "do not write the group's AGENTS.md rules")
	return cmd
}

func newGroupRemoveCommand() *cobra.Command {
	var targetName string
	var keepAgents bool

	cmd := &cobra.Command{
		Use:   "remove <group>",
		Short: "Disable a group's skills and remove its AGENTS.md rules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, _, _, err := loadGroups()
			if err != nil {
				return err
			}
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			projectConfig, err := project.Load(root)
			if err != nil {
				return err
			}

			g, err := group.Find(groups, args[0])
			if err != nil {
				g, err = findConfiguredGroup(projectConfig, args[0])
				if err != nil {
					return err
				}
			}
			globalConfig, _, err := loadGlobalConfigAndRepoDir()
			if err != nil {
				return err
			}
			if targetName != "" {
				if _, ok := projectConfig.Targets[targetName]; !ok {
					return fmt.Errorf("target %q is not configured in this project", targetName)
				}
			}

			removed := 0
			for name, targetConfig := range projectConfig.Targets {
				if targetName != "" && name != targetName {
					continue
				}
				destRoot, err := targetDestRoot(globalConfig, name, "project", root)
				if err != nil {
					return err
				}
				keptSkills := make([]project.SkillRef, 0, len(targetConfig.Skills))
				for _, ref := range targetConfig.Skills {
					if ref.Group != g.ID {
						keptSkills = append(keptSkills, ref)
						continue
					}
					if ref.As == "" {
						return fmt.Errorf("skill %q has no projection alias", ref.ID)
					}
					dest := filepath.Join(destRoot, ref.As)
					if _, err := os.Lstat(dest); err == nil {
						expected := install.ProjectionManifest{SourceID: ref.ID, Target: name, Scope: "project"}
						if err := install.CanRemoveProjection(dest, expected); err != nil {
							return fmt.Errorf("remove %s: %w", dest, err)
						}
						if err := install.RemoveProjection(dest); err != nil {
							return err
						}
						removed++
					} else if !os.IsNotExist(err) {
						return err
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
				return err
			}
			if !keepAgents {
				if err := agents.Remove(filepath.Join(root, "AGENTS.md"), g.ID); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %d skill(s) from group %s\n", removed, g.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&targetName, "target", "", "only remove from this target (default: all targets)")
	cmd.Flags().BoolVar(&keepAgents, "keep-agents", false, "keep the group's AGENTS.md block")
	return cmd
}

func loadGroups() ([]group.Group, string, []skill.Skill, error) {
	_, repoDir, err := loadGlobalConfigAndRepoDir()
	if err != nil {
		return nil, "", nil, err
	}
	index, err := skill.BuildIndex(repoDir)
	if err != nil {
		return nil, "", nil, err
	}
	groups, err := group.Discover(repoDir, index.Skills)
	if err != nil {
		return nil, "", nil, err
	}
	return groups, repoDir, index.Skills, nil
}

func findConfiguredGroup(cfg project.Config, query string) (group.Group, error) {
	for _, targetConfig := range cfg.Targets {
		for _, ref := range targetConfig.Groups {
			if ref.ID == query || ref.Name == query {
				name := ref.Name
				if name == "" {
					name = ref.ID
				}
				return group.Group{ID: ref.ID, Name: name}, nil
			}
		}
	}
	return group.Group{}, fmt.Errorf("group %q is not enabled and no GROUP.yaml was found", query)
}

func groupTitle(g group.Group) string {
	if g.Title != "" {
		return g.Title
	}
	return g.Name
}

func refreshGroupAgents(root string, cfg project.Config, groups []group.Group) error {
	path := filepath.Join(root, "AGENTS.md")
	configured := map[string]bool{}
	for _, targetConfig := range cfg.Targets {
		for _, groupRef := range targetConfig.Groups {
			g, err := group.Find(groups, groupRef.ID)
			if err != nil && groupRef.Name != "" {
				g, err = group.Find(groups, groupRef.Name)
			}
			if err != nil {
				return fmt.Errorf("configured group %q: %w", groupRef.ID, err)
			}
			configured[g.ID] = true
			if g.AgentsPath == "" || groupRef.NoAgents {
				if err := agents.Remove(path, g.ID); err != nil {
					return err
				}
				continue
			}
			data, err := os.ReadFile(g.AgentsPath)
			if err != nil {
				return err
			}
			if err := agents.Upsert(path, g.ID, groupTitle(g), string(data)); err != nil {
				return err
			}
		}
	}
	stale, err := agents.List(path)
	if err != nil {
		return err
	}
	for _, id := range stale {
		if configured[id] {
			continue
		}
		if err := agents.Remove(path, id); err != nil {
			return err
		}
	}
	return nil
}
