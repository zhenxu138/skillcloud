package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/skillcloud/skillcloud/internal/group"
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
			fmt.Fprintf(cmd.OutOrStdout(), "id: %s\nname: %s\ncategory: %s\ntitle: %s\n", g.ID, g.Name, g.Category, groupTitle(g))
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

			byCategory := map[string][]group.Group{}
			for _, g := range groups {
				byCategory[g.Category] = append(byCategory[g.Category], g)
			}
			categoryOrder := []string{"SOP", "个人学习", "自定义组", "其他"}
			var categories []string
			seenCategory := map[string]bool{}
			for _, category := range categoryOrder {
				if len(byCategory[category]) > 0 {
					categories = append(categories, category)
					seenCategory[category] = true
				}
			}
			var extra []string
			for category := range byCategory {
				if !seenCategory[category] {
					extra = append(extra, category)
				}
			}
			sort.Strings(extra)
			categories = append(categories, extra...)

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for i, category := range categories {
				if i > 0 {
					fmt.Fprintln(w)
				}
				fmt.Fprintf(w, "%s\n", category)
				fmt.Fprintln(w, "GROUP\tSKILLS\tENABLED\t工作流 / 擅长")
				for _, g := range byCategory[category] {
					description := g.Description
					if description == "" {
						description = "-"
					}
					enabledLabel := "-"
					if targets := enabled[g.ID]; len(targets) > 0 {
						enabledLabel = strings.Join(targets, ",")
					}
					fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", g.Name, len(g.Members), enabledLabel, description)
				}
			}
			w.Flush()
			return nil
		},
	}
}

func newGroupUseCommand() *cobra.Command {
	var targetName string
	var scope string
	var mode string
	var noAgents bool
	var aliasPrefix string

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
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			projectConfig, err := project.Load(root)
			if err != nil {
				return err
			}
			prefixNorm := normalizeAliasPrefix(aliasPrefix)
			if prefixNorm != "" {
				for _, ref := range projectConfig.Targets[targetName].Skills {
					if ref.HasGroup(g.ID) && !strings.HasPrefix(ref.As, prefixNorm) {
						return fmt.Errorf("group %s is already enabled without prefix %q; run `skillcloud group remove %s` first", g.Name, strings.TrimSuffix(prefixNorm, "-"), g.Name)
					}
				}
			}
			result, err := applyGroupUse(root, g, targetName, mode, noAgents, aliasPrefix)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "enabled %d skill(s) from group %s for %s\n", result.Skills, g.Name, targetName)
			if result.AgentsUpdated {
				fmt.Fprintln(cmd.OutOrStdout(), "updated group rule file and AGENTS.md index; start a new Codex session to load them")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&targetName, "target", "codex", "target agent: codex, claude, hermes")
	cmd.Flags().StringVar(&scope, "scope", "project", "scope: project")
	cmd.Flags().StringVar(&mode, "mode", "link", "mode: link or copy")
	cmd.Flags().BoolVar(&noAgents, "no-agents", false, "do not write the group's AGENTS.md rules")
	cmd.Flags().StringVar(&aliasPrefix, "prefix", "", "prefix projected skill aliases when groups overlap")
	return cmd
}

func newGroupRemoveCommand() *cobra.Command {
	var targetName string

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
			removed, err := applyGroupRemove(root, g, targetName)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %d skill(s) from group %s\n", removed, g.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&targetName, "target", "codex", "target agent: codex, claude, hermes")
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
