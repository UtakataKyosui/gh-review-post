package cliexit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

func TestExitCodes(t *testing.T) {
	cases := []struct {
		name     string
		err      *cliexit.Error
		wantCode int
	}{
		{"NewGeneral", cliexit.NewGeneral(errors.New("oops")), cliexit.CodeGeneral},
		{"NewUsage", cliexit.NewUsage(cliexit.ErrCodeUsageBadArgs, errors.New("bad args")), cliexit.CodeUsage},
		{"NewAuth", cliexit.NewAuth(cliexit.ErrCodeAuthNoBinary, errors.New("no gh")), cliexit.CodeAuth},
		{"NewValidation", cliexit.NewValidation(cliexit.ErrCodeValidation, errors.New("invalid"), nil), cliexit.CodeValidation},
		{"NewAPI", cliexit.NewAPI(cliexit.ErrCodeAPI, errors.New("rate limit")), cliexit.CodeAPI},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.ExitCode != tc.wantCode {
				t.Errorf("ExitCode = %d, want %d", tc.err.ExitCode, tc.wantCode)
			}
		})
	}
}

func TestErrorInterface(t *testing.T) {
	inner := errors.New("inner error")
	e := cliexit.NewGeneral(inner)
	if !strings.Contains(e.Error(), "inner error") {
		t.Errorf("Error() = %q, want to contain inner error message", e.Error())
	}
	if !errors.Is(e, inner) {
		t.Error("errors.Is failed for wrapped error")
	}
}

func TestRenderPlain(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	e := cliexit.NewAuth(cliexit.ErrCodeAuthNoBinary, errors.New("gh not found"))
	cliexit.Render(e, false, &stdout, &stderr)
	if !strings.HasPrefix(stderr.String(), "error: ") {
		t.Errorf("stderr = %q, want prefix 'error: '", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty in plain mode, got %q", stdout.String())
	}
}

func TestRenderJSON(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	details := map[string]any{"path": "src/a.go", "line": 42}
	e := cliexit.NewValidation(cliexit.ErrCodeValidation, errors.New("line not in diff"), details)
	cliexit.Render(e, true, &stdout, &stderr)
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty in JSON mode, got %q", stderr.String())
	}
	var out struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nraw: %s", err, stdout.String())
	}
	if out.Error.Code != string(cliexit.ErrCodeValidation) {
		t.Errorf("code = %q, want %q", out.Error.Code, cliexit.ErrCodeValidation)
	}
}

func TestRenderNonExitError(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	plain := errors.New("unexpected panic")
	cliexit.Render(plain, false, &stdout, &stderr)
	if !strings.Contains(stderr.String(), "unexpected panic") {
		t.Errorf("stderr = %q, want to contain original message", stderr.String())
	}
}

func TestExitCodeOf(t *testing.T) {
	if cliexit.ExitCodeOf(cliexit.NewAuth(cliexit.ErrCodeAuthNoToken, errors.New("x"))) != cliexit.CodeAuth {
		t.Error("ExitCodeOf auth error should be CodeAuth")
	}
	if cliexit.ExitCodeOf(errors.New("generic")) != cliexit.CodeGeneral {
		t.Error("ExitCodeOf generic error should be CodeGeneral")
	}
	if cliexit.ExitCodeOf(nil) != cliexit.CodeSuccess {
		t.Error("ExitCodeOf nil should be CodeSuccess")
	}
}
