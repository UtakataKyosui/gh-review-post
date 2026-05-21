package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
)

func TestRootCommand_NoArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	root := cmd.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{})

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

func TestGlobalFlags_Present(t *testing.T) {
	root := cmd.NewRootCmd()
	flags := []string{"pr", "repo", "json", "verbose", "dry-run"}
	for _, name := range flags {
		if root.PersistentFlags().Lookup(name) == nil {
			t.Errorf("global flag --%s not registered", name)
		}
	}
}

func TestGlobalFlags_Help(t *testing.T) {
	var buf bytes.Buffer
	root := cmd.NewRootCmd()
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})
	_ = root.Execute()
	help := buf.String()
	for _, flag := range []string{"--pr", "-R", "--json", "--verbose", "--dry-run"} {
		if !strings.Contains(help, flag) {
			t.Errorf("help text missing flag %q", flag)
		}
	}
}
