package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/gitstore"
	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
	"github.com/spf13/cobra"
)

func newPullCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull skill updates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadConfiguredStore()
			if err != nil {
				return err
			}
			if err := store.Pull(context.Background()); err != nil {
				return err
			}
			if err := rebuildIndex(store.RepoDir); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "pulled %s\n", store.RepoDir)
			return nil
		},
	}
}

func newConnectCommand() *cobra.Command {
	cmd := newInitCommand()
	cmd.Use = "connect <repo-url>"
	cmd.Short = "Connect skillcloud to a cloud skill library"
	return cmd
}

func newUpdateCommand() *cobra.Command {
	cmd := newPullCommand()
	cmd.Use = "update"
	cmd.Short = "Update the local skill library cache"
	return cmd
}

func newPushCommand() *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push skill updates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadConfiguredStore()
			if err != nil {
				return err
			}
			if err := store.Push(context.Background(), message); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "pushed %s\n", store.RepoDir)
			return nil
		},
	}
	cmd.Flags().StringVarP(&message, "message", "m", "update skills", "Commit message")
	return cmd
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show skill repository status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, repoDir, err := loadGlobalConfigAndRepoDir()
			if err != nil {
				return err
			}
			store := gitstore.Store{RepoDir: repoDir, RepoURL: cfg.RepoURL}
			status, err := store.Status(context.Background())
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if !status.Dirty {
				fmt.Fprintln(out, "clean")
			} else {
				for _, line := range status.Lines {
					fmt.Fprintln(out, line)
				}
			}

			root, _ := os.Getwd()
			projectConfig, loadErr := project.Load(root)
			if loadErr == nil {
				index, indexErr := skill.BuildIndex(store.RepoDir)
				if indexErr == nil {
					for targetName := range projectConfig.Targets {
						destRoot, destErr := targetDestRoot(cfg, targetName, "project", root)
						if destErr != nil {
							continue
						}
						report := project.Inspect(projectConfig, index.Skills, destRoot, targetName)
						for _, ref := range report.Missing {
							fmt.Fprintf(out, "missing\t%s\t%s\n", targetName, ref.ID)
						}
						for _, alias := range report.Unmanaged {
							fmt.Fprintf(out, "unmanaged\t%s\t%s\n", targetName, alias)
						}
					}
				}
			}
			return nil
		},
	}
}

func loadConfiguredStore() (gitstore.Store, error) {
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		return gitstore.Store{}, err
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return gitstore.Store{}, err
	}
	repoDir, err := config.ExpandHome(cfg.RepoDir)
	if err != nil {
		return gitstore.Store{}, err
	}
	return gitstore.Store{
		RepoDir: repoDir,
		RepoURL: cfg.RepoURL,
	}, nil
}

func rebuildIndex(repoDir string) error {
	index, err := skill.BuildIndex(repoDir)
	if err != nil {
		return err
	}
	indexPath, err := config.DefaultIndexPath()
	if err != nil {
		return err
	}
	return skill.SaveIndex(indexPath, index)
}
