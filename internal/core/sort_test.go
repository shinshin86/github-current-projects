package core

import (
	"testing"
	"time"

	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

func TestSortByPushedAt(t *testing.T) {
	repos := []githubapi.Repository{
		{Name: "old", PushedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "new", PushedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "mid", PushedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
	}
	SortRepos(repos)

	expected := []string{"new", "mid", "old"}
	for i, e := range expected {
		if repos[i].Name != e {
			t.Errorf("position %d: got %s, want %s", i, repos[i].Name, e)
		}
	}
}

func TestSortByStarsWhenSamePushedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repos := []githubapi.Repository{
		{Name: "low-stars", PushedAt: ts, StargazersCount: 5},
		{Name: "high-stars", PushedAt: ts, StargazersCount: 100},
		{Name: "mid-stars", PushedAt: ts, StargazersCount: 50},
	}
	SortRepos(repos)

	expected := []string{"high-stars", "mid-stars", "low-stars"}
	for i, e := range expected {
		if repos[i].Name != e {
			t.Errorf("position %d: got %s, want %s", i, repos[i].Name, e)
		}
	}
}

func TestSortByNameWhenAllElseEqual(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repos := []githubapi.Repository{
		{Name: "charlie", PushedAt: ts, StargazersCount: 10},
		{Name: "alpha", PushedAt: ts, StargazersCount: 10},
		{Name: "bravo", PushedAt: ts, StargazersCount: 10},
	}
	SortRepos(repos)

	expected := []string{"alpha", "bravo", "charlie"}
	for i, e := range expected {
		if repos[i].Name != e {
			t.Errorf("position %d: got %s, want %s", i, repos[i].Name, e)
		}
	}
}

func TestSortByStarsPrimary(t *testing.T) {
	repos := []githubapi.Repository{
		{Name: "newer-low", PushedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), StargazersCount: 5},
		{Name: "older-high", PushedAt: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), StargazersCount: 50},
		{Name: "mid-mid", PushedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), StargazersCount: 20},
	}
	SortReposBy(repos, "stars")

	expected := []string{"older-high", "mid-mid", "newer-low"}
	for i, e := range expected {
		if repos[i].Name != e {
			t.Errorf("position %d: got %s, want %s", i, repos[i].Name, e)
		}
	}
}

func TestTopN(t *testing.T) {
	repos := []githubapi.Repository{
		{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}, {Name: "e"},
	}

	result := TopN(repos, 3)
	if len(result) != 3 {
		t.Errorf("TopN(5, 3) = %d items, want 3", len(result))
	}

	result = TopN(repos, 0)
	if len(result) != 5 {
		t.Errorf("TopN(5, 0) = %d items, want 5", len(result))
	}

	result = TopN(repos, 10)
	if len(result) != 5 {
		t.Errorf("TopN(5, 10) = %d items, want 5", len(result))
	}
}
