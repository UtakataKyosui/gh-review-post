package cmd_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/cmd"
	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
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

// writeYAMLFile writes content to a temp YAML file and returns its path.
func writeYAMLFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "review-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

// runSubmitDryRun executes submit --dry-run with the given file and returns the error.
func runSubmitDryRun(t *testing.T, filePath string) error {
	t.Helper()
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "1", "-R", "owner/repo", "--file", filePath, "--dry-run"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	return root.Execute()
}

// assertValidationError fails if err is not a *cliexit.Error with CodeValidation.
func assertValidationError(t *testing.T, err error, wantCode cliexit.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var e *cliexit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *cliexit.Error, got %T: %v", err, err)
	}
	if e.ExitCode != cliexit.CodeValidation {
		t.Errorf("ExitCode = %d, want %d", e.ExitCode, cliexit.CodeValidation)
	}
	if e.Code != wantCode {
		t.Errorf("Code = %q, want %q", e.Code, wantCode)
	}
}

// --- Issue #15: body suggestion block validation ---

func TestSubmitCmd_Validate_BodyHasSuggestion(t *testing.T) {
	content := "event: COMMENT\nbody: |\n  ここを直してください。\n  ```suggestion\n  fixed line\n  ```\ncomments: []\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	assertValidationError(t, err, cliexit.ErrCodeBodyHasSuggestion)

	var e *cliexit.Error
	errors.As(err, &e)
	lines, ok := e.Details["lines"]
	if !ok {
		t.Fatal("details missing 'lines' key")
	}
	if lines == nil {
		t.Fatal("details.lines is nil")
	}
}

func TestSubmitCmd_Validate_BodyHasSuggestion_TaggedFence(t *testing.T) {
	// Language tag in body still violates (body bans all suggestion fences).
	content := "event: COMMENT\nbody: \"```suggestion:go\\nx := 1\\n```\"\ncomments: []\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	assertValidationError(t, err, cliexit.ErrCodeBodyHasSuggestion)
}

func TestSubmitCmd_Validate_BodyNoSuggestion_Passes(t *testing.T) {
	content := "event: COMMENT\nbody: 全体的に LGTM。\ncomments: []\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	if err != nil {
		t.Fatalf("expected no error for clean body, got: %v", err)
	}
}

// --- Issue #16: comment suggestion format validation ---

func TestSubmitCmd_Validate_SuggestionFormat_LanguageTag(t *testing.T) {
	content := "event: COMMENT\nbody: ok\ncomments:\n  - path: cmd/submit.go\n    line: 10\n    body: \"```suggestion:go\\nx := 1\\n```\"\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	assertValidationError(t, err, cliexit.ErrCodeSuggestionFormat)

	var e *cliexit.Error
	errors.As(err, &e)
	vs, ok := e.Details["violations"]
	if !ok {
		t.Fatal("details missing 'violations' key")
	}
	violations, ok := vs.([]map[string]any)
	if !ok || len(violations) == 0 {
		t.Fatalf("expected non-empty violations, got %v", vs)
	}
	if violations[0]["reason"] != "language_tag" {
		t.Errorf("reason = %v, want language_tag", violations[0]["reason"])
	}
}

func TestSubmitCmd_Validate_SuggestionFormat_MissingClose(t *testing.T) {
	content := "event: COMMENT\nbody: ok\ncomments:\n  - path: a.go\n    line: 1\n    body: \"```suggestion\\nmissing close\"\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	assertValidationError(t, err, cliexit.ErrCodeSuggestionFormat)

	var e *cliexit.Error
	errors.As(err, &e)
	vs := e.Details["violations"].([]map[string]any)
	if vs[0]["reason"] != "missing_close" {
		t.Errorf("reason = %v, want missing_close", vs[0]["reason"])
	}
}

func TestSubmitCmd_Validate_SuggestionFormat_ValidSuggestion_Passes(t *testing.T) {
	content := "event: COMMENT\nbody: ok\ncomments:\n  - path: cmd/submit.go\n    line: 100\n    body: \"```suggestion\\nreturn nil\\n```\"\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	if err != nil {
		t.Fatalf("expected no error for valid suggestion, got: %v", err)
	}
}

func TestSubmitCmd_Validate_SuggestionFormat_MultipleViolations(t *testing.T) {
	content := "event: COMMENT\nbody: ok\ncomments:\n  - path: a.go\n    line: 1\n    body: \"```suggestion:go\\nx\\n```\"\n  - path: b.go\n    line: 2\n    body: \"```suggestion\\nno close\"\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	assertValidationError(t, err, cliexit.ErrCodeSuggestionFormat)

	var e *cliexit.Error
	errors.As(err, &e)
	vs := e.Details["violations"].([]map[string]any)
	if len(vs) != 2 {
		t.Errorf("expected 2 violations, got %d", len(vs))
	}
}

// --- Issue #17: aggregate violations and no API call ---

func TestSubmitCmd_Validate_AggregateViolations(t *testing.T) {
	// Both body and comment have violations; must return VALIDATION_FAILED with 2 entries.
	content := "event: COMMENT\nbody: \"```suggestion\\nbad\\n```\"\ncomments:\n  - path: a.go\n    line: 1\n    body: \"```suggestion:go\\nx\\n```\"\n"
	path := writeYAMLFile(t, content)
	err := runSubmitDryRun(t, path)
	assertValidationError(t, err, cliexit.ErrCodeValidation)

	var e *cliexit.Error
	errors.As(err, &e)
	vs, ok := e.Details["violations"].([]any)
	if !ok {
		t.Fatalf("expected []any violations, got %T", e.Details["violations"])
	}
	if len(vs) != 2 {
		t.Errorf("expected 2 violations, got %d", len(vs))
	}
}

func TestSubmitCmd_Validate_AggregateViolations_JSONOutput(t *testing.T) {
	content := "event: COMMENT\nbody: \"```suggestion\\nbad\\n```\"\ncomments:\n  - path: a.go\n    line: 1\n    body: \"```suggestion:go\\nx\\n```\"\n"
	path := writeYAMLFile(t, content)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "1", "-R", "owner/repo", "--file", path, "--dry-run", "--json"})
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.Execute()

	if err == nil {
		t.Fatal("expected error")
	}

	// Render is called by main; here we call it manually to simulate.
	cliexit.Render(err, true, stdout, stderr)
	var out struct {
		Error struct {
			Code    string         `json:"code"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal(stdout.Bytes(), &out); jsonErr != nil {
		t.Fatalf("stdout not valid JSON: %v\nraw: %s", jsonErr, stdout.String())
	}
	if out.Error.Code != string(cliexit.ErrCodeValidation) {
		t.Errorf("code = %q, want %q", out.Error.Code, cliexit.ErrCodeValidation)
	}
}

// TestSubmitCmd_Validate_NoAPICall verifies that validation failure stops execution before
// reaching the API call. Since the function returns a validation error (exit 4) before
// building the API client, no HTTP request is made.
func TestSubmitCmd_Validate_NoAPICall(t *testing.T) {
	content := "event: COMMENT\nbody: \"```suggestion\\nbad\\n```\"\ncomments: []\n"
	path := writeYAMLFile(t, content)

	// Run without --dry-run; validation must fail before auth/API path.
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"submit", "--pr", "1", "-R", "owner/repo", "--file", path})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()

	var e *cliexit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *cliexit.Error, got %T: %v", err, err)
	}
	if e.ExitCode != cliexit.CodeValidation {
		t.Errorf("ExitCode = %d, want %d (CodeValidation)", e.ExitCode, cliexit.CodeValidation)
	}
	// If we reached here with CodeValidation, the function returned before auth/API calls.
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
