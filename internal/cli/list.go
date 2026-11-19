package cli

import (
	"fmt"

	"github.com/skillcloud/skillcloud/internal/config"
	"github.com/skillcloud/skillcloud/internal/skill"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List indexed skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := loadCurrentIndex()
			if err != nil {
				return err
			}
			for _, s := range index.Skills {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", s.ID, s.Name, s.Description)
			}
			printListFooter(cmd, fmt.Sprintf("(｡･ω･｡)ﾉ 共 %d 个 skill。用 `skillcloud group list` 查看技能组，`skillcloud current` 查看当前项目。", len(index.Skills)))
			return nil
		},
	}
}

func newSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search indexed skills",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := loadCurrentIndex()
			if err != nil {
				return err
			}
			for _, s := range index.Search(args[0]) {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", s.ID, s.Name, s.Description)
			}
			return nil
		},
	}
}

func printListFooter(cmd *cobra.Command, message string) {
	fmt.Fprintln(cmd.ErrOrStderr())
	fmt.Fprintln(cmd.ErrOrStderr(), message)
}

func loadCurrentIndex() (skill.Index, error) {
	if _, repoDir, err := loadGlobalConfigAndRepoDir(); err == nil {
		if index, buildErr := skill.BuildIndex(repoDir); buildErr == nil {
			return index, nil
		}
	}
	indexPath, err := config.DefaultIndexPath()
	if err != nil {
		return skill.Index{}, err
	}
	return skill.LoadIndex(indexPath)
}
