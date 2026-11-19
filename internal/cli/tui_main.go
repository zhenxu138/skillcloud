package cli

import (
	"github.com/skillcloud/skillcloud/internal/tui"
	"github.com/spf13/cobra"
)

// runMainTUI launches the launcher menu, runs the chosen action, then loops
// back to the menu so the user can pick another action. The menu and each
// action run as separate tea programs so they never share the terminal.
var runMainTUI = func(cmd *cobra.Command, opts tui.AppOptions) error {
	for {
		result, err := tui.RunApp(opts)
		if err != nil {
			return err
		}
		var status string
		switch result.Action {
		case tui.ActionQuit:
			return nil
		case tui.ActionBrowse:
			status, err = manageSkillFlow(opts.Target, opts.Scope, opts.Mode, 0)
		case tui.ActionGroups:
			status, err = manageGroupFlow(opts.Target, opts.Scope, opts.Mode)
		case tui.ActionEnabled:
			status, err = manageFlow(opts.Target, opts.Scope, opts.Mode, tui.ViewEnabled())
		case tui.ActionAdd:
			status, err = addSkillFromForm()
		case tui.ActionChanges:
			status, err = showChangesSummary()
		case tui.ActionSync:
			status, err = syncRepo()
		case tui.ActionPush:
			status, err = pushRepo()
		case tui.ActionSettings:
			status, err = showSettingsSummary()
		}
		if err != nil {
			status = err.Error()
		}
		opts = tui.AppOptions{Target: opts.Target, Scope: opts.Scope, Mode: opts.Mode, Status: status}
	}
}
