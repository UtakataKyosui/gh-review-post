package cmd_test

import (
	"bytes"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
)

func TestSubmitCmd_Registered(t *testing.T) {
	root := cmd.NewRootCmd()
	var found bool
	for _, sub := range root.Commands() {
		if sub.Name() == "submit" {
			found = true
		}
	}
	if !found {
		t.Fatal("submit command not registered")
	}
}

func TestSubmitCmd_NoPositionalArgs(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "42", "-R", "owner/repo", "unexpected-arg"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when positional args given to submit")
	}
}

func TestSubmitCmd_StubRuns(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "42", "-R", "owner/repo"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	// Stub returns an error ("not yet implemented") — that is expected behaviour.
	_ = root.Execute()
}
