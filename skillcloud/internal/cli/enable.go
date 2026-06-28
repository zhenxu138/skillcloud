package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/install"
	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
	"github.com/skillcloud/skillcloud/internal/tui"
	"github.com/spf13/cobra"
)

func defaultAlias(id string) string {
	parts := strings.Split(strings.Trim(id, "/"), "/")
	return parts[len(parts)-1]
}

func newEnableCommand() *cobra.Command {
	var targetName string
	var scope string
	var mode string
	var selectMode bool

	cmd := &cobra.Command{
		Use:   "enable <skill-id...>",
		Short: "Enable one or more skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			if selectMode {
				return runSelectEnable(cmd, targetName, scope, mode)
			}
			if len(args) == 0 {
				return fmt.Errorf("provide at least one skill id or use --select")
			}
			return runEnable(cmd, args, targetName, scope, mode)
		},
	}
	cmd.Flags().StringVar(&targetName, "target", "codex", "target agent: codex, claude, hermes")
	cmd.Flags().StringVar(&scope, "scope", "project", "scope: project or global")
	cmd.Flags().StringVar(&mode, "mode", "link", "mode: link or copy")
	cmd.Flags().BoolVar(&selectMode, "select", false, "select skills in a TUI")
	return cmd
}

func runEnable(cmd *cobra.Command, patterns []string, targetName string, scope string, mode string) error {
	if scope != "project" && scope != "global" {
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

	selected := resolvePatterns(index.Skills, patterns)
	if len(selected) == 0 {
		return fmt.Errorf("no skills matched %s", strings.Join(patterns, ", "))
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}
	projectConfig, err := project.Load(projectRoot)
	if err != nil {
		return err
	}
	destRoot, err := targetDestRoot(globalConfig, targetName, scope, projectRoot)
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
	if scope == "project" {
		if err := project.Save(projectRoot, plan.ProjectConfig); err != nil {
			return err
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "enabled %d skill(s) for %s\n", len(selected), targetName)
	return nil
}

func resolvePatterns(skills []skill.Skill, patterns []string) []skill.Skill {
	seen := map[string]bool{}
	var selected []skill.Skill
	for _, pattern := range patterns {
		for _, match := range skill.Resolve(skills, pattern) {
			if seen[match.ID] {
				continue
			}
			seen[match.ID] = true
			selected = append(selected, match)
		}
	}
	return selected
}

func targetDestRoot(cfg config.Config, targetName string, scope string, projectRoot string) (string, error) {
	targetConfig, ok := cfg.Targets[targetName]
	if !ok {
		return "", fmt.Errorf("unknown target %q", targetName)
	}
	if scope == "global" {
		return config.ExpandHome(targetConfig.Global)
	}
	return filepath.Join(projectRoot, filepath.FromSlash(targetConfig.Project)), nil
}

func loadGlobalConfigAndRepoDir() (config.Config, string, error) {
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		return config.Config{}, "", err
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return config.Config{}, "", err
	}
	repoDir, err := config.ExpandHome(cfg.RepoDir)
	if err != nil {
		return config.Config{}, "", err
	}
	return cfg, repoDir, nil
}

func runSelectEnable(cmd *cobra.Command, targetName string, scope string, mode string) error {
	index, err := loadCurrentIndex()
	if err != nil {
		return err
	}
	ids, err := tui.Select(index.Skills)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("no skills selected")
	}
	return runEnable(cmd, ids, targetName, scope, mode)
}

func newDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <alias-or-skill-id...>",
		Short: "Disable one or more projected skills from the current project",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			globalConfig, _, err := loadGlobalConfigAndRepoDir()
			if err != nil {
				return err
			}
			projectConfig, err := project.Load(root)
			if err != nil {
				return err
			}

			projectConfig, removed, err := disableProjectSkills(root, globalConfig, projectConfig, args)
			if err != nil {
				return err
			}
			if err := project.Save(root, projectConfig); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "disabled %d skill(s)\n", removed)
			return nil
		},
	}
}

func disableProjectSkills(root string, globalConfig config.Config, projectConfig project.Config, idsOrAliases []string) (project.Config, int, error) {
	remove := map[string]bool{}
	for _, arg := range idsOrAliases {
		remove[arg] = true
	}

	removed := 0
	for targetName, targetConfig := range projectConfig.Targets {
		destRoot, err := targetDestRoot(globalConfig, targetName, "project", root)
		if err != nil {
			return project.Config{}, removed, err
		}
		kept := targetConfig.Skills[:0]
		for _, ref := range targetConfig.Skills {
			if remove[ref.As] || remove[ref.ID] {
				removed++
				if err := os.RemoveAll(filepath.Join(destRoot, ref.As)); err != nil {
					return project.Config{}, removed, err
				}
				continue
			}
			kept = append(kept, ref)
		}
		targetConfig.Skills = kept
		projectConfig.Targets[targetName] = targetConfig
	}
	return projectConfig, removed, nil
}

func newApplyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Re-apply current project skill configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			globalConfig, repoDir, err := loadGlobalConfigAndRepoDir()
			if err != nil {
				return err
			}
			projectConfig, err := project.Load(root)
			if err != nil {
				return err
			}
			index, err := skill.BuildIndex(repoDir)
			if err != nil {
				return err
			}
			byID := map[string]skill.Skill{}
			for _, s := range index.Skills {
				byID[s.ID] = s
			}

			for targetName, targetConfig := range projectConfig.Targets {
				destRoot, err := targetDestRoot(globalConfig, targetName, "project", root)
				if err != nil {
					return err
				}
				var selected []skill.Skill
				for _, ref := range targetConfig.Skills {
					s, ok := byID[ref.ID]
					if !ok {
						return fmt.Errorf("configured skill %q not found", ref.ID)
					}
					s.Name = ref.As
					selected = append(selected, s)
				}
				plan, err := install.PlanEnable(selected, targetName, "project", targetConfig.Mode, destRoot, projectConfig)
				if err != nil {
					return err
				}
				if err := install.Execute(plan.Actions); err != nil {
					return err
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "applied project skills")
			return nil
		},
	}
}
