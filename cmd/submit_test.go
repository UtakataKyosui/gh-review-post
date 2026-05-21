package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
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
	root.SetArgs([]string{"submit", "--pr", "42", "-R", "owner/repo", "--file", "x.yaml", "unexpected-arg"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when positional args given to submit")
	}
}

func TestSubmitCmd_MissingFile(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "42", "-R", "owner/repo"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
}

func TestSubmitCmd_DryRun(t *testing.T) {
	yaml := `event: COMMENT
body: test body
comments:
  - path: foo/bar.go
    line: 10
    body: a comment
`
	f, err := os.CreateTemp(t.TempDir(), "review-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(yaml); err != nil {
		t.Fatal(err)
	}
	f.Close()

	out := &bytes.Buffer{}
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "42", "-R", "owner/repo", "--file", f.Name(), "--dry-run"})
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"event": "COMMENT"`)) {
		t.Errorf("expected JSON payload in output, got: %s", out.String())
	}
}

func TestSubmitCmd_InvalidYAML(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "review-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("{{invalid yaml{{"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "42", "-R", "owner/repo", "--file", f.Name(), "--dry-run"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestSubmitCmd_JSONFile(t *testing.T) {
	content := `{"event":"COMMENT","body":"json body","comments":[{"path":"a.go","line":1,"body":"hi"}]}`
	f, err := os.CreateTemp(t.TempDir(), "review-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "42", "-R", "owner/repo", "--file", filepath.Clean(f.Name()), "--dry-run"})
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"event": "COMMENT"`)) {
		t.Errorf("expected JSON payload in output, got: %s", out.String())
	}
}
