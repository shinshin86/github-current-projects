package core

import (
	"testing"
	"time"

	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

func makeRepos() []githubapi.Repository {
	return []githubapi.Repository{
		{
			Name:            "public-repo",
			HTMLURL:         "https://github.com/u/public-repo",
			Fork:            false,
			Archived:        false,
			Private:         false,
			StargazersCount: 10,
			PushedAt:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:            "forked-repo",
			HTMLURL:         "https://github.com/u/forked-repo",
			Fork:            true,
			Archived:        false,
			Private:         false,
			StargazersCount: 5,
			PushedAt:        time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:            "archived-repo",
			HTMLURL:         "https://github.com/u/archived-repo",
			Fork:            false,
			Archived:        true,
			Private:         false,
			StargazersCount: 50,
			PushedAt:        time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:            "private-repo",
			HTMLURL:         "https://github.com/u/private-repo",
			Fork:            false,
			Archived:        false,
			Private:         true,
			StargazersCount: 0,
			PushedAt:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:            "low-star-repo",
			HTMLURL:         "https://github.com/u/low-star-repo",
			Fork:            false,
			Archived:        false,
			Private:         false,
			StargazersCount: 1,
			PushedAt:        time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		},
	}
}

func TestFilterDefaults(t *testing.T) {
	repos := makeRepos()
	filtered := FilterRepos(repos, FilterOptions{
		Now: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	})

	// Should exclude: forked, archived, private
	if len(filtered) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(filtered))
	}
	names := []string{filtered[0].Name, filtered[1].Name}
	if names[0] != "public-repo" || names[1] != "low-star-repo" {
		t.Errorf("unexpected repos: %v", names)
	}
}

func TestFilterIncludeForks(t *testing.T) {
	repos := makeRepos()
	filtered := FilterRepos(repos, FilterOptions{
		IncludeForks: true,
		Now:          time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	})

	if len(filtered) != 3 {
		t.Fatalf("expected 3 repos with forks, got %d", len(filtered))
	}
}

func TestFilterIncludeArchived(t *testing.T) {
	repos := makeRepos()
	filtered := FilterRepos(repos, FilterOptions{
		IncludeArchived: true,
		Now:             time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	})

	if len(filtered) != 3 {
		t.Fatalf("expected 3 repos with archived, got %d", len(filtered))
	}
}

func TestFilterMinStars(t *testing.T) {
	repos := makeRepos()
	filtered := FilterRepos(repos, FilterOptions{
		MinStars: 10,
		Now:      time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	})

	if len(filtered) != 1 {
		t.Fatalf("expected 1 repo with min-stars=10, got %d", len(filtered))
	}
	if filtered[0].Name != "public-repo" {
		t.Errorf("expected public-repo, got %s", filtered[0].Name)
	}
}

func TestFilterSinceDays(t *testing.T) {
	repos := makeRepos()
	filtered := FilterRepos(repos, FilterOptions{
		SinceDays: 30,
		Now:       time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	})

	// Only public-repo pushed on 2025-01-15 is within 30 days of 2025-01-20
	// low-star-repo pushed on 2024-12-01 is also within 30 days? 2025-01-20 - 30 = 2024-12-21
	// So low-star-repo (2024-12-01) is before 2024-12-21, excluded
	if len(filtered) != 1 {
		t.Fatalf("expected 1 repo within 30 days, got %d", len(filtered))
	}
	if filtered[0].Name != "public-repo" {
		t.Errorf("expected public-repo, got %s", filtered[0].Name)
	}
}

func TestFilterSinceDaysZeroDisabled(t *testing.T) {
	repos := makeRepos()
	filtered := FilterRepos(repos, FilterOptions{
		SinceDays: 0,
		Now:       time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	})

	// SinceDays=0 means no date filter, so public-repo + low-star-repo
	if len(filtered) != 2 {
		t.Fatalf("expected 2 repos with since-days=0, got %d", len(filtered))
	}
}

func TestFilterExcludesPrivate(t *testing.T) {
	repos := []githubapi.Repository{
		{Name: "private", Private: true},
		{Name: "public", Private: false},
	}
	filtered := FilterRepos(repos, FilterOptions{})
	if len(filtered) != 1 || filtered[0].Name != "public" {
		t.Errorf("expected only public repo, got %v", filtered)
	}
}

func TestFilterRequireDescription(t *testing.T) {
	repos := []githubapi.Repository{
		{Name: "no-desc", Description: ""},
		{Name: "blank-desc", Description: "   "},
		{Name: "has-desc", Description: "ok"},
	}
	filtered := FilterRepos(repos, FilterOptions{RequireDescription: true})
	if len(filtered) != 1 || filtered[0].Name != "has-desc" {
		t.Errorf("expected only has-desc, got %v", filtered)
	}
}

func TestFilterTagsAny(t *testing.T) {
	repos := []githubapi.Repository{
		{Name: "a", Topics: []string{"go", "cli"}},
		{Name: "b", Topics: []string{"web"}},
		{Name: "c", Topics: nil},
	}
	filtered := FilterRepos(repos, FilterOptions{
		Tags: []string{"cli", "tools"},
	})
	if len(filtered) != 1 || filtered[0].Name != "a" {
		t.Errorf("expected only a, got %v", filtered)
	}
}

func TestFilterTagsAll(t *testing.T) {
	repos := []githubapi.Repository{
		{Name: "a", Topics: []string{"go", "cli"}},
		{Name: "b", Topics: []string{"go"}},
	}
	filtered := FilterRepos(repos, FilterOptions{
		Tags:         []string{"go", "cli"},
		TagsMatchAll: true,
	})
	if len(filtered) != 1 || filtered[0].Name != "a" {
		t.Errorf("expected only a, got %v", filtered)
	}
}
