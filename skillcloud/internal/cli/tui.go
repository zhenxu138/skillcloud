package cli

import (
	"fmt"
	"os"

	"github.com/skillcloud/skillcloud/internal/install"
	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
	"github.com/skillcloud/skillcloud/internal/tui"
	"github.com/spf13/cobra"
)

var runBrowseTUI = tui.Manage

func newBrowseCommand() *cobra.Command {
	var targetName string
	var scope string
	var mode string

	cmd := &cobra.Command{
		Use:   "browse",
		Short: "Browse and manage project skills in a TUI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBrowse(cmd, targetName, scope, mode)
		},
	}
	cmd.Flags().StringVar(&targetName, "target", "codex", "target agent: codex, claude, hermes")
	cmd.Flags().StringVar(&scope, "scope", "project", "scope: project")
	cmd.Flags().StringVar(&mode, "mode", "link", "mode: link or copy")
	return cmd
}

func diffBrowseSelection(initialIDs []string, selectedIDs []string) ([]string, []string) {
	initial := map[string]bool{}
	selected := map[string]bool{}
	for _, id := range initialIDs {
		initial[id] = true
	}
	for _, id := range selectedIDs {
		selected[id] = true
	}

	var enable []string
	for _, id := range selectedIDs {
		if !initial[id] {
			enable = append(enable, id)
		}
	}

	var disable []string
	for _, id := range initialIDs {
		if !selected[id] {
			disable = append(disable, id)
		}
	}
	return enable, disable
}

func enabledSkillIDs(cfg project.Config, targetName string) []string {
	targetConfig := cfg.Targets[targetName]
	ids := make([]string, 0, len(targetConfig.Skills))
	for _, ref := range targetConfig.Skills {
		ids = append(ids, ref.ID)
	}
	return ids
}

func runBrowse(cmd *cobra.Command, targetName string, scope string, mode string) error {
	if scope != "project" {
		if scope == "global" {
			return fmt.Errorf("browse --scope global is not implemented yet; use skillcloud enable ... --scope global or skillcloud disable ...")
		}
		return fmt.Errorf("unknown scope %q", scope)
	}
	if mode != "link" && mode != "copy" {
		return fmt.Errorf("unknown mode %q", mode)
	}

	globalConfig, repoDir, err := loadGlobalConfigAndRepoDir()
	if err != nil {
		return err
	}
	index, err := skill.BuildIndex(repoDir)
	if err != nil {
		return err
	}
	if len(index.Skills) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no skills found")
		return nil
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}
	projectConfig, err := project.Load(root)
	if err != nil {
		return err
	}
	initialIDs := enabledSkillIDs(projectConfig, targetName)

	result, err := runBrowseTUI(tui.ManageOptions{
		Skills:           index.Skills,
		Target:           targetName,
		Scope:            scope,
		Mode:             mode,
		InitiallyEnabled: initialIDs,
	})
	if err != nil {
		return err
	}
	if !result.Apply {
		fmt.Fprintln(cmd.OutOrStdout(), "browse canceled")
		return nil
	}

	toEnable, toDisable := diffBrowseSelection(result.InitialIDs, result.SelectedIDs)
	if len(toEnable) == 0 && len(toDisable) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no skill changes")
		return nil
	}

	if len(toDisable) > 0 {
		var removed int
		projectConfig, removed, err = disableProjectSkills(root, globalConfig, projectConfig, toDisable)
		if err != nil {
			return err
		}
		if removed != len(toDisable) {
			return fmt.Errorf("disabled %d skill(s), expected %d", removed, len(toDisable))
		}
	}

	if len(toEnable) > 0 {
		selected := resolvePatterns(index.Skills, toEnable)
		if len(selected) != len(toEnable) {
			return fmt.Errorf("selected skill disappeared from repository")
		}
		destRoot, err := targetDestRoot(globalConfig, targetName, "project", root)
		if err != nil {
			return err
		}
		plan, err := install.PlanEnable(selected, targetName, scope, mode, destRoot, projectConfig)
		if err != nil {
			return err
		}
		if err := install.Execute(plan.Actions); err != nil {
			return err
		}
		projectConfig = plan.ProjectConfig
	}

	if err := project.Save(root, projectConfig); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "applied skill changes: enabled %d, disabled %d\n", len(toEnable), len(toDisable))
	return nil
}
