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

func loadCurrentIndex() (skill.Index, error) {
	indexPath, err := config.DefaultIndexPath()
	if err != nil {
		return skill.Index{}, err
	}
	return skill.LoadIndex(indexPath)
}

