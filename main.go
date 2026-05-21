package main

import (
	"errors"
	"os"
	"strings"

	"github.com/UtakataKyosui/gh-review-post/cmd"
	"github.com/UtakataKyosui/gh-review-post/internal/auth"
	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

func main() {
	if err := auth.CheckGH(); err != nil {
		cliexit.Render(err, false, os.Stdout, os.Stderr)
		os.Exit(cliexit.ExitCodeOf(err))
	}
	if err := auth.CheckGHVersion(); err != nil {
		cliexit.Render(err, false, os.Stdout, os.Stderr)
		os.Exit(cliexit.ExitCodeOf(err))
	}

	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		wrapped := cobraUsageErr(err)
		asJSON, _ := root.PersistentFlags().GetBool("json")
		cliexit.Render(wrapped, asJSON, os.Stdout, os.Stderr)
		os.Exit(cliexit.ExitCodeOf(wrapped))
	}
}

// cobraUsageErr wraps cobra's own usage/unknown-command errors into *cliexit.Error
// so they get exit code 2 (CodeUsage) instead of the generic 1 (CodeGeneral).
func cobraUsageErr(err error) error {
	if err == nil {
		return nil
	}
	var ce *cliexit.Error
	if errors.As(err, &ce) {
		return err
	}
	msg := err.Error()
	if strings.HasPrefix(msg, "unknown command") ||
		strings.HasPrefix(msg, "unknown flag") ||
		strings.HasPrefix(msg, "unknown shorthand flag") ||
		strings.HasPrefix(msg, "accepts") ||
		strings.HasPrefix(msg, "required flag") {
		return cliexit.NewUsage(cliexit.ErrCodeUsageBadArgs, err)
	}
	return cliexit.NewGeneral(err)
}
