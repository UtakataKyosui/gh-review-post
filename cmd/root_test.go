package cmd_test

import (
	"bytes"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
)

func TestRootCommand_NoArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	root := cmd.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{})

	// Running with no args prints help and exits with code 0.
	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRootCommand_UnknownSubcommand(t *testing.T) {
	buf := &bytes.Buffer{}
	root := cmd.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"unknown"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}

func TestSubcommands_Registered(t *testing.T) {
	root := cmd.NewRootCmd()
	names := make(map[string]bool)
	for _, sub := range root.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{"submit", "reply", "threads"} {
		if !names[want] {
			t.Errorf("subcommand %q not registered", want)
		}
	}
}
