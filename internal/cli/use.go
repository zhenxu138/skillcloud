package cli

import "github.com/spf13/cobra"

func newUseCommand() *cobra.Command {
	cmd := newEnableCommand()
	cmd.Use = "use <skill-id...>"
	cmd.Short = "Use one or more skills"
	return cmd
}

func newUnuseCommand() *cobra.Command {
	cmd := newDisableCommand()
	cmd.Use = "unuse <alias-or-skill-id...>"
	cmd.Short = "Stop using one or more projected skills from the current project"
	return cmd
}
