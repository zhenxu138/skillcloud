package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/spf13/cobra"
)

func newCurrentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the skills and groups enabled in the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			cfg, err := project.Load(root)
			if err != nil {
				return err
			}
			totalSkills := 0
			totalGroups := 0
			for _, target := range cfg.Targets {
				totalSkills += len(target.Skills)
				totalGroups += len(target.Groups)
			}
			if totalSkills == 0 && totalGroups == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "no skills or groups configured in %s\n", root)
				return nil
			}

			targets := make([]string, 0, len(cfg.Targets))
			for targetName := range cfg.Targets {
				targets = append(targets, targetName)
			}
			sort.Strings(targets)
			for i, targetName := range targets {
				if i > 0 {
					fmt.Fprintln(cmd.OutOrStdout())
				}
				target := cfg.Targets[targetName]
				fmt.Fprintf(cmd.OutOrStdout(), "target: %s\nmode: %s\nskills: %d\n", targetName, target.Mode, len(target.Skills))
				for _, ref := range target.Skills {
					groupLabel := ref.Group
					if groupLabel == "" {
						groupLabel = "-"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  %-28s %-48s group=%s\n", ref.As, ref.ID, groupLabel)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "groups: %d\n", len(target.Groups))
				for _, groupRef := range target.Groups {
					suffix := ""
					if groupRef.NoAgents {
						suffix = " (no agents)"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  %-28s %s%s\n", groupRef.Name, groupRef.ID, suffix)
				}
			}
			return nil
		},
	}
}
