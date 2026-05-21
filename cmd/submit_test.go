package cmd_test

import (
	"bytes"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
)

func TestSubmitCmd_RequiresTwoArgs(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when args are missing")
	}
}

func TestSubmitCmd_AcceptsTwoArgs(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "owner/repo", "42"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	// Currently a stub — should not error out.
	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
