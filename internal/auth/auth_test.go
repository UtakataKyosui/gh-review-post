package auth_test

import (
	"errors"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/internal/auth"
	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

func TestCheckGH_Found(t *testing.T) {
	err := auth.CheckGH()
	if err != nil {
		t.Skipf("gh not in PATH: %v", err)
	}
}

func TestCheckGH_NotFound(t *testing.T) {
	// Use a guaranteed-nonexistent directory so no stray gh binary is found.
	t.Setenv("PATH", t.TempDir())
	err := auth.CheckGH()
	if err == nil {
		t.Fatal("expected error when gh not in PATH")
	}
	var e *cliexit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *cliexit.Error, got: %T", err)
	}
	if e.Code != cliexit.ErrCodeAuthNoBinary {
		t.Errorf("Code = %q, want %q", e.Code, cliexit.ErrCodeAuthNoBinary)
	}
	if e.ExitCode != cliexit.CodeAuth {
		t.Errorf("ExitCode = %d, want %d", e.ExitCode, cliexit.CodeAuth)
	}
}

func TestCheckGHVersion_Parse(t *testing.T) {
	cases := []struct {
		output  string
		wantErr bool
	}{
		{"gh version 2.40.0 (2024-01-01)\nhttps://github.com/cli/cli/releases/tag/v2.40.0\n", false},
		{"gh version 2.50.0 (2025-01-01)\n", false},
		{"gh version 2.39.5 (2023-12-01)\n", true},
		{"gh version 2.0.0 (2022-01-01)\n", true},
	}
	for _, tc := range cases {
		t.Run(tc.output[:30], func(t *testing.T) {
			err := auth.ParseGHVersion(tc.output)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseGHVersion(%q) error = %v, wantErr %v", tc.output[:30], err, tc.wantErr)
			}
			if err != nil {
				var e *cliexit.Error
				if !errors.As(err, &e) {
					t.Errorf("expected *cliexit.Error, got %T", err)
				}
				if e.Code != cliexit.ErrCodeAuthOldGH {
					t.Errorf("Code = %q, want %q", e.Code, cliexit.ErrCodeAuthOldGH)
				}
			}
		})
	}
}

func TestCheckGHVersion_InvalidFormat(t *testing.T) {
	err := auth.ParseGHVersion("not a valid version string")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestToken_EmptyHostFallback(t *testing.T) {
	// Token("") and Token("github.com") must behave identically.
	// We can't control the auth state in CI, so we only verify the fallback doesn't panic.
	_, err1 := auth.Token("")
	_, err2 := auth.Token("github.com")

	// Both should either succeed or fail with the same error code.
	code := func(err error) cliexit.ErrCode {
		var e *cliexit.Error
		if errors.As(err, &e) {
			return e.Code
		}
		return ""
	}
	if code(err1) != code(err2) {
		t.Errorf("Token(\"\") code=%q, Token(\"github.com\") code=%q — should match", code(err1), code(err2))
	}
}

func TestToken_Unauthenticated(t *testing.T) {
	// Force unauthenticated by clearing all token env vars.
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	_, err := auth.Token("github.com")
	if err == nil {
		// Authenticated via keychain/config — skip, can't test unauthenticated path.
		t.Skip("authenticated via keychain; cannot test unauthenticated path")
	}
	var e *cliexit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *cliexit.Error, got %T", err)
	}
	if e.Code != cliexit.ErrCodeAuthNoToken {
		t.Errorf("Code = %q, want %q", e.Code, cliexit.ErrCodeAuthNoToken)
	}
}
