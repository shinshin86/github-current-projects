package core

import (
	"sort"
	"strings"

	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

// SortRepos sorts repositories by pushed_at desc, then stars desc, then name asc.
func SortRepos(repos []githubapi.Repository) {
	SortReposBy(repos, "pushed")
}

// SortReposBy sorts repositories by the requested order.
// Supported modes: "pushed" (default), "stars".
func SortReposBy(repos []githubapi.Repository, mode string) {
	sort.SliceStable(repos, func(i, j int) bool {
		a, b := repos[i], repos[j]
		switch mode {
		case "stars":
			// Primary: stars descending
			if a.StargazersCount != b.StargazersCount {
				return a.StargazersCount > b.StargazersCount
			}
			// Secondary: pushed_at descending
			if !a.PushedAt.Equal(b.PushedAt) {
				return a.PushedAt.After(b.PushedAt)
			}
		default:
			// Primary: pushed_at descending
			if !a.PushedAt.Equal(b.PushedAt) {
				return a.PushedAt.After(b.PushedAt)
			}
			// Secondary: stars descending
			if a.StargazersCount != b.StargazersCount {
				return a.StargazersCount > b.StargazersCount
			}
		}
		// Tertiary: name ascending (case-insensitive)
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
}

// TopN returns the first n items, or all if n <= 0 or n > len(repos).
func TopN(repos []githubapi.Repository, n int) []githubapi.Repository {
	if n <= 0 || n > len(repos) {
		return repos
	}
	return repos[:n]
}
