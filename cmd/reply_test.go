package cmd_test

import (
	"bytes"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
)

func TestReplyCmd_RequiresTwoArgs(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"reply"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when args are missing")
	}
}

func TestReplyCmd_AcceptsTwoArgs(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"reply", "owner/repo", "42"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
