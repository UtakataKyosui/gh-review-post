package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

func newThreadsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "threads",
		Short: "List open review threads on a PR",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cliexit.NewUsage(cliexit.ErrCodeUsageBadArgs,
				errors.New("threads: not yet implemented"))
		},
	}
}
