package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newReplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reply <owner/repo> <pr-number>",
		Short: "Reply to review threads on a PR from a YAML/JSON file or stdin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "reply: not yet implemented")
			return nil
		},
	}
}
