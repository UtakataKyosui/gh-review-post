package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/UtakataKyosui/gh-review-post/cmd"
	"github.com/UtakataKyosui/gh-review-post/internal/auth"
)

func main() {
	if err := auth.CheckGH(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		var exitErr *exitError
		if errors.As(err, &exitErr) {
			fmt.Fprintf(os.Stderr, "error: %v\n", exitErr.err)
			os.Exit(exitErr.code)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// exitError carries an exit code alongside the underlying error.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string             { return e.err.Error() }
func (e *exitError) Unwrap() error             { return e.err }
func newExitError(code int, err error) *exitError { return &exitError{code: code, err: err} }
