package cli

import (
	"fmt"
	"os/exec"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/validate"
	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check local skillcloud setup",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := exec.LookPath("git"); err != nil {
				return fmt.Errorf("git not found in PATH")
			}
			configPath, err := config.DefaultConfigPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			repoDir, err := config.ExpandHome(cfg.RepoDir)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "config: %s\nrepo: %s\ngit: ok\n", configPath, repoDir)
			return nil
		},
	}
}

func newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the local skill repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.DefaultConfigPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			repoDir, err := config.ExpandHome(cfg.RepoDir)
			if err != nil {
				return err
			}
			errs := validate.Repo(repoDir)
			for _, e := range errs {
				fmt.Fprintln(cmd.OutOrStdout(), e)
			}
			if len(errs) > 0 {
				return fmt.Errorf("validation failed")
			}
			fmt.Fprintln(cmd.OutOrStdout(), "validation passed")
			return nil
		},
	}
}

