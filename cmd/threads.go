package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newThreadsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "threads <owner/repo> <pr-number>",
		Short: "List open review threads on a PR",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "threads: not yet implemented")
			return nil
		},
	}
}
