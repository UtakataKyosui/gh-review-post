package auth

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/cli/go-gh/v2/pkg/auth"
)

var ErrGHNotFound = errors.New("gh binary not found in PATH; install GitHub CLI from https://cli.github.com")

// CheckGH verifies that the gh binary is in PATH.
func CheckGH() error {
	_, err := exec.LookPath("gh")
	if err != nil {
		return ErrGHNotFound
	}
	return nil
}

// Token returns the authentication token for the given host.
// Falls back to github.com if host is empty.
func Token(host string) (string, error) {
	if host == "" {
		host = "github.com"
	}
	token, _ := auth.TokenForHost(host)
	if token == "" {
		return "", fmt.Errorf("not authenticated for %s; run `gh auth login` first", host)
	}
	return token, nil
}
