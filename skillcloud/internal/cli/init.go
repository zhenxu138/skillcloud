package cli

import (
	"context"
	"fmt"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/gitstore"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init <repo-url>",
		Short: "Initialize skillcloud",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.DefaultConfig(args[0])
			repoDir, err := config.ExpandHome(cfg.RepoDir)
			if err != nil {
				return err
			}
			store := gitstore.Store{
				RepoDir: repoDir,
				RepoURL: cfg.RepoURL,
			}
			if err := store.Clone(context.Background()); err != nil {
				return err
			}
			if err := rebuildIndex(repoDir); err != nil {
				return err
			}

			configPath, err := config.DefaultConfigPath()
			if err != nil {
				return err
			}
			if err := config.Save(configPath, cfg); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "initialized %s\n", repoDir)
			return nil
		},
	}
}
