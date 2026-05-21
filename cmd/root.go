package cmd

import (
	"github.com/spf13/cobra"

	"github.com/UtakataKyosui/gh-review-post/internal/auth"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gh-review-post",
		Short: "Post and respond to GitHub PR reviews from the CLI",
		Long: `gh-review-post is a gh extension that lets reviewers post structured
review comments and PR authors reply to them using YAML/JSON input.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		// PersistentPreRunE runs before every subcommand (except help/completion),
		// after flag parsing — so --json is already resolved here.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.CheckGH(); err != nil {
				return err
			}
			return auth.CheckGHVersion()
		},
	}

	root.PersistentFlags().Int("pr", 0, "PR number")
	root.PersistentFlags().StringP("repo", "R", "", "Repository in owner/repo format (detected from cwd if omitted)")
	root.PersistentFlags().Bool("json", false, "Output errors and results as JSON to stdout")
	root.PersistentFlags().BoolP("verbose", "v", false, "Log HTTP requests and responses to stderr")
	root.PersistentFlags().Bool("dry-run", false, "Validate input without sending API requests")

	_ = root.MarkPersistentFlagRequired("pr")

	root.AddCommand(newSubmitCmd())
	root.AddCommand(newReplyCmd())
	root.AddCommand(newThreadsCmd())

	return root
}
