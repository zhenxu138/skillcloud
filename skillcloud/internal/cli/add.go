package cli

import (
	"fmt"

	"github.com/skillcloud/skillcloud/internal/library"
	"github.com/skillcloud/skillcloud/internal/skill"
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	var id string
	var replace bool
	var enable bool
	var targetName string
	var scope string
	var mode string

	cmd := &cobra.Command{
		Use:   "add <skill-dir>",
		Short: "Add an existing skill directory to the skill library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			globalConfig, repoDir, err := loadGlobalConfigAndRepoDir()
			if err != nil {
				return err
			}
			plan, err := library.PlanAdd(library.AddOptions{
				RepoDir:   repoDir,
				SourceDir: args[0],
				ID:        id,
				Replace:   replace,
			})
			if err != nil {
				return err
			}
			if err := library.ExecuteAdd(plan); err != nil {
				return err
			}
			if err := rebuildIndex(repoDir); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", plan.ID)

			if enable {
				selected := []skill.Skill{{
					ID:          plan.ID,
					Name:        plan.Metadata.Name,
					Description: plan.Metadata.Description,
					Path:        plan.DestDir,
				}}
				return runEnableWithConfig(cmd, selected, globalConfig, targetName, scope, mode)
			}
			_ = globalConfig
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "as", "", "library id, for example coding/code-review")
	cmd.Flags().BoolVar(&replace, "replace", false, "replace an existing library skill")
	cmd.Flags().BoolVar(&enable, "enable", false, "enable the imported skill after adding it")
	cmd.Flags().StringVar(&targetName, "target", "codex", "target agent: codex, claude, hermes")
	cmd.Flags().StringVar(&scope, "scope", "project", "scope: project or global")
	cmd.Flags().StringVar(&mode, "mode", "link", "mode: link or copy")
	_ = cmd.MarkFlagRequired("as")
	return cmd
}
