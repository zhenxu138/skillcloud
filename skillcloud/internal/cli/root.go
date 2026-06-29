package cli

import (
	"github.com/skillcloud/skillcloud/internal/tui"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skillcloud",
		Short: "Sync and enable AI agent skills from a personal Git repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMainTUI(tui.AppOptions{Target: "codex", Scope: "project"})
		},
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newPullCommand())
	cmd.AddCommand(newConnectCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newPushCommand())
	cmd.AddCommand(newStatusCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newSearchCommand())
	cmd.AddCommand(newBrowseCommand())
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newEnableCommand())
	cmd.AddCommand(newUseCommand())
	cmd.AddCommand(newDisableCommand())
	cmd.AddCommand(newUnuseCommand())
	cmd.AddCommand(newApplyCommand())
	cmd.AddCommand(newDoctorCommand())
	cmd.AddCommand(newValidateCommand())

	return cmd
}
