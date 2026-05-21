package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gh-review-post",
		Short: "Post and respond to GitHub PR reviews from the CLI",
		Long: `gh-review-post is a gh extension that lets reviewers post structured
review comments and PR authors reply to them using YAML/JSON input.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(newSubmitCmd())
	root.AddCommand(newReplyCmd())
	root.AddCommand(newThreadsCmd())

	return root
}
