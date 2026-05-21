package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

func newReplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reply",
		Short: "Reply to review threads on a PR from a YAML/JSON file or stdin",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cliexit.NewUsage(cliexit.ErrCodeUsageBadArgs,
				errors.New("reply: not yet implemented"))
		},
	}
}
