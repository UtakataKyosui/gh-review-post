package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSubmitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "submit <owner/repo> <pr-number>",
		Short: "Submit a PR review from a YAML/JSON file or stdin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "submit: not yet implemented")
			return nil
		},
	}
}
