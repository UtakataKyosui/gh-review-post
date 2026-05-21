package cmd_test

import (
	"bytes"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
)

func TestThreadsCmd_Registered(t *testing.T) {
	root := cmd.NewRootCmd()
	var found bool
	for _, sub := range root.Commands() {
		if sub.Name() == "threads" {
			found = true
		}
	}
	if !found {
		t.Fatal("threads command not registered")
	}
}

func TestThreadsCmd_StubRuns(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"threads", "--pr", "42", "-R", "owner/repo"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	_ = root.Execute()
}
