package repo

import (
	"fmt"

	ghrepo "github.com/cli/go-gh/v2/pkg/repository"

	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
)

// Resolve returns the repository from the -R flag or cwd git remote.
func Resolve(flagRepo string) (ghrepo.Repository, error) {
	if flagRepo != "" {
		r, err := ghrepo.Parse(flagRepo)
		if err != nil {
			return ghrepo.Repository{}, cliexit.NewUsage(cliexit.ErrCodeUsageBadArgs,
				fmt.Errorf("invalid repository %q: %w", flagRepo, err))
		}
		return r, nil
	}
	r, err := ghrepo.Current()
	if err != nil {
		return ghrepo.Repository{}, cliexit.NewUsage(cliexit.ErrCodeUsageNoRepo,
			fmt.Errorf("could not detect repository from current directory; use -R owner/repo: %w", err))
	}
	return r, nil
}
