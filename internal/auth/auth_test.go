package auth_test

import (
	"testing"

	"github.com/UtakataKyosui/gh-review-post/internal/auth"
)

func TestCheckGH_Found(t *testing.T) {
	// gh binary is expected to be in PATH in CI and local dev environments.
	// If this fails, the gh CLI is not installed.
	err := auth.CheckGH()
	if err != nil {
		t.Skipf("gh not in PATH: %v", err)
	}
}

func TestCheckGH_NotFound(t *testing.T) {
	t.Setenv("PATH", "")
	err := auth.CheckGH()
	if err == nil {
		t.Fatal("expected error when gh not in PATH")
	}
	if err != auth.ErrGHNotFound {
		t.Fatalf("expected ErrGHNotFound, got: %v", err)
	}
}
