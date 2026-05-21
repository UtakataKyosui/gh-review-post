package repo_test

import (
	"errors"
	"testing"

	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
	"github.com/UtakataKyosui/gh-review-post/internal/repo"
)

func TestResolve_FlagTakesPriority(t *testing.T) {
	r, err := repo.Resolve("owner/myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Owner != "owner" || r.Name != "myrepo" {
		t.Errorf("got %s/%s, want owner/myrepo", r.Owner, r.Name)
	}
}

func TestResolve_InvalidFlag(t *testing.T) {
	_, err := repo.Resolve("notavalidrepo")
	if err == nil {
		t.Fatal("expected error for invalid repo format")
	}
}

func TestResolve_EmptyFlag_NotInGitRepo(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := repo.Resolve("")
	if err == nil {
		t.Fatal("expected error outside git repo")
	}
	var e *cliexit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *cliexit.Error, got %T", err)
	}
	if e.Code != cliexit.ErrCodeUsageNoRepo {
		t.Errorf("Code = %q, want %q", e.Code, cliexit.ErrCodeUsageNoRepo)
	}
}
