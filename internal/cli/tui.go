package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/group"
	"github.com/skillcloud/skillcloud/internal/install"
	"github.com/skillcloud/skillcloud/internal/library"
	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
	"github.com/skillcloud/skillcloud/internal/tui"
	"github.com/spf13/cobra"
)

var runBrowseTUI = tui.Manage
var runGroupTUI = tui.ManageGroups

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
	msg, err := manageSkillFlow(targetName, scope, mode, 0)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), msg)
	return nil
}

// manageFlow is the Enabled menu entry: it launches the same interactive skill
// manager as Browse but pre-filtered to only enabled skills, so the user can
// quickly review and disable what they no longer need.
func manageFlow(targetName, scope, mode string, initialView int) (string, error) {
	return manageSkillFlow(targetName, scope, mode, initialView)
}

func manageGroupFlow(targetName, scope, mode string) (string, error) {
	if scope != "project" {
		return "", fmt.Errorf("group management currently supports project scope only")
	}
	if mode != "link" && mode != "copy" {
		return "", fmt.Errorf("unknown mode %q", mode)
	}

	groups, _, _, err := loadGroups()
	if err != nil {
		return "", err
	}
	if len(groups) == 0 {
		return "no groups found", nil
	}
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	projectConfig, err := project.Load(root)
	if err != nil {
		return "", err
	}

	options := make([]tui.GroupOption, 0, len(groups))
	byID := make(map[string]group.Group, len(groups))
	for _, g := range groups {
		description := g.Description
		if description == "" {
			description = g.Title
		}
		options = append(options, tui.GroupOption{
			ID:          g.ID,
			Name:        g.Name,
			Title:       g.Title,
			Description: description,
			SkillCount:  len(g.Members),
			HasAgents:   g.AgentsPath != "",
		})
		byID[g.ID] = g
	}
	enabled := make([]string, 0, len(projectConfig.Targets[targetName].Groups))
	for _, ref := range projectConfig.Targets[targetName].Groups {
		enabled = append(enabled, ref.ID)
	}

	result, err := runGroupTUI(tui.GroupOptions{
		Groups:           options,
		Target:           targetName,
		Scope:            scope,
		Mode:             mode,
		InitiallyEnabled: enabled,
	})
	if err != nil {
		return "", err
	}
	if !result.Apply {
		return "groups canceled", nil
	}
	toEnable, toDisable := diffBrowseSelection(result.InitialIDs, result.SelectedIDs)
	if len(toEnable) == 0 && len(toDisable) == 0 {
		return "no group changes", nil
	}

	for _, id := range toDisable {
		g, ok := byID[id]
		if !ok {
			return "", fmt.Errorf("group %q disappeared from repository", id)
		}
		if _, err := applyGroupRemove(root, g, targetName, false); err != nil {
			return "", err
		}
	}
	for _, id := range toEnable {
		g, ok := byID[id]
		if !ok {
			return "", fmt.Errorf("group %q disappeared from repository", id)
		}
		if _, err := applyGroupUse(root, g, targetName, mode, false, ""); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("applied group changes: enabled %d, disabled %d", len(toEnable), len(toDisable)), nil
}

// manageSkillFlow runs the interactive skill manager and applies the resulting
// selection, returning a human-readable status string instead of printing so
// the launcher menu can show it as the previous-action result.
func manageSkillFlow(targetName, scope, mode string, initialView int) (string, error) {
	if scope != "project" {
		if scope == "global" {
			return "", fmt.Errorf("browse --scope global is not implemented yet; use skillcloud enable ... --scope global or skillcloud disable ...")
		}
		return "", fmt.Errorf("unknown scope %q", scope)
	}
	if mode != "link" && mode != "copy" {
		return "", fmt.Errorf("unknown mode %q", mode)
	}

	globalConfig, repoDir, err := loadGlobalConfigAndRepoDir()
	if err != nil {
		return "", err
	}
	index, err := skill.BuildIndex(repoDir)
	if err != nil {
		return "", err
	}
	if len(index.Skills) == 0 {
		return "no skills found", nil
	}

	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	projectConfig, err := project.Load(root)
	if err != nil {
		return "", err
	}
	initialIDs := enabledSkillIDs(projectConfig, targetName)

	result, err := runBrowseTUI(tui.ManageOptions{
		Skills:            index.Skills,
		Target:            targetName,
		Scope:             scope,
		Mode:              mode,
		InitiallyEnabled:  initialIDs,
		InitialFilterView: initialView,
	})
	if err != nil {
		return "", err
	}
	if !result.Apply {
		return "browse canceled", nil
	}

	toEnable, toDisable := diffBrowseSelection(result.InitialIDs, result.SelectedIDs)
	if len(toEnable) == 0 && len(toDisable) == 0 {
		return "no skill changes", nil
	}

	if len(toDisable) > 0 {
		var removed int
		projectConfig, removed, err = disableProjectSkills(root, globalConfig, projectConfig, toDisable, targetName)
		if err != nil {
			return "", err
		}
		if removed != len(toDisable) {
			return "", fmt.Errorf("disabled %d skill(s), expected %d", removed, len(toDisable))
		}
	}

	if len(toEnable) > 0 {
		selected := resolvePatterns(index.Skills, toEnable)
		if len(selected) != len(toEnable) {
			return "", fmt.Errorf("selected skill disappeared from repository")
		}
		destRoot, err := targetDestRoot(globalConfig, targetName, "project", root)
		if err != nil {
			return "", err
		}
		plan, err := install.PlanEnable(selected, targetName, scope, mode, destRoot, projectConfig)
		if err != nil {
			return "", err
		}
		if err := install.Execute(plan.Actions); err != nil {
			return "", err
		}
		projectConfig = plan.ProjectConfig
	}

	if err := project.Save(root, projectConfig); err != nil {
		return "", err
	}
	return fmt.Sprintf("applied skill changes: enabled %d, disabled %d", len(toEnable), len(toDisable)), nil
}

// syncRepo pulls skill updates and rebuilds the index.
func syncRepo() (string, error) {
	store, err := loadConfiguredStore()
	if err != nil {
		return "", err
	}
	if err := store.Pull(context.Background()); err != nil {
		return "", err
	}
	if err := rebuildIndex(store.RepoDir); err != nil {
		return "", err
	}
	return fmt.Sprintf("pulled %s", store.RepoDir), nil
}

// pushRepo commits any uncommitted skill edits and pushes them to the remote.
func pushRepo() (string, error) {
	store, err := loadConfiguredStore()
	if err != nil {
		return "", err
	}
	if err := store.Push(context.Background(), "update skills from TUI"); err != nil {
		return "", err
	}
	return fmt.Sprintf("pushed %s", store.RepoDir), nil
}

// showChangesSummary reports uncommitted edits in the skill library plus any
// project projection drift (missing or unmanaged skills) for the current
// project, mirroring skillcloud status.
func showChangesSummary() (string, error) {
	store, err := loadConfiguredStore()
	if err != nil {
		return "", err
	}
	status, err := store.Status(context.Background())
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if !status.Dirty {
		b.WriteString("repo is clean")
	} else {
		b.WriteString("changes:\n  ")
		b.WriteString(strings.Join(status.Lines, "\n  "))
	}

	// Project projection drift, mirroring skillcloud status.
	root, _ := os.Getwd()
	projectConfig, loadErr := project.Load(root)
	if loadErr == nil {
		index, indexErr := skill.BuildIndex(store.RepoDir)
		if indexErr == nil {
			cfg, _, cfgErr := loadGlobalConfigAndRepoDir()
			if cfgErr != nil {
				cfg = config.Config{}
			}
			for targetName := range projectConfig.Targets {
				if cfg.Targets == nil {
					continue
				}
				destRoot, destErr := targetDestRoot(cfg, targetName, "project", root)
				if destErr != nil {
					continue
				}
				report := project.Inspect(projectConfig, index.Skills, destRoot, targetName, readProjectionManifest)
				for _, ref := range report.Missing {
					fmt.Fprintf(&b, "\n  missing\t%s\t%s", targetName, ref.ID)
				}
				for _, alias := range report.Unmanaged {
					fmt.Fprintf(&b, "\n  unmanaged\t%s\t%s", targetName, alias)
				}
			}
		}
	}
	return b.String(), nil
}

// showSettingsSummary reports the current skillcloud configuration.
func showSettingsSummary() (string, error) {
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		return "", err
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return "", err
	}
	repoDir, err := config.ExpandHome(cfg.RepoDir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "config: %s", configPath)
	fmt.Fprintf(&b, "\nrepo_url: %s", cfg.RepoURL)
	fmt.Fprintf(&b, "\nrepo_dir: %s", repoDir)
	fmt.Fprintf(&b, "\ndefault_mode: %s", cfg.DefaultMode)
	b.WriteString("\ntargets:")
	if len(cfg.Targets) == 0 {
		b.WriteString(" (none)")
	}
	for name, tp := range cfg.Targets {
		fmt.Fprintf(&b, "\n  %s: project=%s global=%s", name, tp.Project, tp.Global)
	}
	return b.String(), nil
}

// addSkillFromForm runs the interactive import form, then adds the skill to the
// library and rebuilds the index.
func addSkillFromForm() (string, error) {
	formResult, err := tui.RunAddForm()
	if err != nil {
		return "", err
	}
	if formResult.Canceled {
		return "add canceled", nil
	}
	if formResult.SourceDir == "" || formResult.ID == "" {
		return "add canceled: source dir and skill id are required", nil
	}
	_, repoDir, err := loadGlobalConfigAndRepoDir()
	if err != nil {
		return "", err
	}
	plan, err := library.PlanAdd(library.AddOptions{
		RepoDir:   repoDir,
		SourceDir: formResult.SourceDir,
		ID:        formResult.ID,
		Replace:   formResult.Replace,
	})
	if err != nil {
		return fmt.Sprintf("add failed: %s", err), nil
	}
	if err := library.ExecuteAdd(plan); err != nil {
		return "", err
	}
	if err := rebuildIndex(repoDir); err != nil {
		return "", err
	}
	return fmt.Sprintf("added %s", plan.ID), nil
}
