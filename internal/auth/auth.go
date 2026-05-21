package auth

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	ghauth "github.com/cli/go-gh/v2/pkg/auth"

	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

const MinGHVersion = "2.40.0"

var versionRe = regexp.MustCompile(`gh version (\d+)\.(\d+)\.(\d+)`)

// CheckGH verifies that the gh binary is in PATH.
func CheckGH() error {
	_, err := exec.LookPath("gh")
	if err != nil {
		return cliexit.NewAuth(cliexit.ErrCodeAuthNoBinary,
			fmt.Errorf("gh binary not found in PATH; install GitHub CLI from https://cli.github.com"))
	}
	return nil
}

// CheckGHVersion verifies that the installed gh meets the minimum version requirement.
func CheckGHVersion() error {
	out, err := exec.Command("gh", "--version").Output()
	if err != nil {
		return cliexit.NewAuth(cliexit.ErrCodeAuthNoBinary,
			fmt.Errorf("failed to run gh --version: %w", err))
	}
	return ParseGHVersion(string(out))
}

// ParseGHVersion parses the output of "gh --version" and checks the minimum version.
// Exported for testing.
func ParseGHVersion(output string) error {
	m := versionRe.FindStringSubmatch(output)
	if m == nil {
		return cliexit.NewAuth(cliexit.ErrCodeAuthOldGH,
			fmt.Errorf("could not parse gh version from output: %q", output))
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])

	// require >= 2.40.0
	if major < 2 || (major == 2 && minor < 40) || (major == 2 && minor == 40 && patch < 0) {
		current := fmt.Sprintf("%d.%d.%d", major, minor, patch)
		return cliexit.NewAuth(cliexit.ErrCodeAuthOldGH,
			fmt.Errorf("gh version %s is too old; %s or later is required", current, MinGHVersion))
	}
	return nil
}

// Token returns the authentication token for the given host.
// Falls back to github.com if host is empty.
// The second return value of ghauth.TokenForHost is the token source (not an error).
func Token(host string) (string, error) {
	if host == "" {
		host = "github.com"
	}
	token, _ := ghauth.TokenForHost(host)
	if token == "" {
		return "", cliexit.NewAuth(cliexit.ErrCodeAuthNoToken,
			fmt.Errorf("not authenticated for %s; run `gh auth login` first", host))
	}
	return token, nil
}
