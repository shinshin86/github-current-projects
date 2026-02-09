package core

import (
	"strings"
	"time"

	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

// FilterOptions controls which repositories pass filtering.
type FilterOptions struct {
	IncludeForks    bool
	IncludeArchived bool
	MinStars        int
	SinceDays       int
	RequireDescription bool
	Tags               []string
	TagsMatchAll       bool
	Now             time.Time // for testability; zero means use time.Now()
}

// FilterRepos returns only the repositories that match the filter criteria.
func FilterRepos(repos []githubapi.Repository, opts FilterOptions) []githubapi.Repository {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	var result []githubapi.Repository
	for _, r := range repos {
		if r.Private {
			continue
		}
		if r.Fork && !opts.IncludeForks {
			continue
		}
		if r.Archived && !opts.IncludeArchived {
			continue
		}
		if r.StargazersCount < opts.MinStars {
			continue
		}
		if opts.RequireDescription && strings.TrimSpace(r.Description) == "" {
			continue
		}
		if len(opts.Tags) > 0 && !matchTopics(r.Topics, opts.Tags, opts.TagsMatchAll) {
			continue
		}
		if opts.SinceDays > 0 {
			cutoff := now.AddDate(0, 0, -opts.SinceDays)
			if r.PushedAt.Before(cutoff) {
				continue
			}
		}
		result = append(result, r)
	}
	return result
}

func matchTopics(repoTopics, filterTags []string, matchAll bool) bool {
	if len(repoTopics) == 0 {
		return false
	}
	topicSet := make(map[string]struct{}, len(repoTopics))
	for _, t := range repoTopics {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		topicSet[t] = struct{}{}
	}
	if len(topicSet) == 0 {
		return false
	}

	matches := 0
	for _, tag := range filterTags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := topicSet[tag]; ok {
			matches++
			if !matchAll {
				return true
			}
		} else if matchAll {
			return false
		}
	}

	if matchAll {
		return matches > 0
	}
	return false
}
