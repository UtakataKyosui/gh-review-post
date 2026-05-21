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

	root.PersistentFlags().Int("pr", 0, "PR number (required)")
	root.PersistentFlags().StringP("repo", "R", "", "Repository in owner/repo format (detected from cwd if omitted)")
	root.PersistentFlags().Bool("json", false, "Output errors and results as JSON to stdout")
	root.PersistentFlags().BoolP("verbose", "v", false, "Log HTTP requests and responses to stderr")
	root.PersistentFlags().Bool("dry-run", false, "Validate input without sending API requests")

	root.AddCommand(newSubmitCmd())
	root.AddCommand(newReplyCmd())
	root.AddCommand(newThreadsCmd())

	return root
}
