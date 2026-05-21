package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

func newSubmitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "submit",
		Short: "Submit a PR review from a YAML/JSON file or stdin",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cliexit.NewUsage(cliexit.ErrCodeUsageBadArgs,
				errors.New("submit: not yet implemented"))
		},
	}
}
