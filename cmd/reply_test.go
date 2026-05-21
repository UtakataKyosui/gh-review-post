package cmd_test

import (
	"bytes"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
)

func TestReplyCmd_Registered(t *testing.T) {
	root := cmd.NewRootCmd()
	var found bool
	for _, sub := range root.Commands() {
		if sub.Name() == "reply" {
			found = true
		}
	}
	if !found {
		t.Fatal("reply command not registered")
	}
}

func TestReplyCmd_StubRuns(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"reply", "--pr", "42", "-R", "owner/repo"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	_ = root.Execute()
}
