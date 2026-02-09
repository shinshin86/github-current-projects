package githubapi

import "time"

// Repository represents the minimal GitHub repository fields we need.
type Repository struct {
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	HTMLURL         string    `json:"html_url"`
	Description     string    `json:"description"`
	Topics          []string  `json:"topics"`
	Fork            bool      `json:"fork"`
	Archived        bool      `json:"archived"`
	Private         bool      `json:"private"`
	Language        string    `json:"language"`
	StargazersCount int       `json:"stargazers_count"`
	PushedAt        time.Time `json:"pushed_at"`
}

// RateLimit holds rate-limit information from response headers.
type RateLimit struct {
	Remaining int
	Limit     int
	Reset     time.Time
}
