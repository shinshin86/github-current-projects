package githubapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RepoFetcher is the interface for fetching repositories.
type RepoFetcher interface {
	FetchAllRepos(user string) ([]Repository, error)
}

// Client communicates with the GitHub REST API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// NewClient creates a new GitHub API client.
func NewClient(baseURL, token string, timeout time.Duration, logger *log.Logger) *Client {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Logger: logger,
	}
}

// FetchAllRepos fetches all public repositories for a user, handling pagination.
func (c *Client) FetchAllRepos(user string) ([]Repository, error) {
	var allRepos []Repository
	url := fmt.Sprintf("%s/users/%s/repos?type=owner&per_page=100&page=1", c.BaseURL, user)

	for url != "" {
		repos, nextURL, err := c.fetchPage(url)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		url = nextURL
	}

	return allRepos, nil
}

func (c *Client) fetchPage(url string) ([]Repository, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetching repos from %s: %w", url, err)
	}
	defer resp.Body.Close()

	c.logRateLimit(resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, "", fmt.Errorf("decoding response: %w", err)
	}

	nextURL := ParseNextLink(resp.Header.Get("Link"))
	return repos, nextURL, nil
}

func (c *Client) logRateLimit(resp *http.Response) {
	rl := ParseRateLimit(resp.Header)
	if rl.Limit > 0 {
		c.Logger.Printf("Rate limit: %d/%d remaining, resets at %s",
			rl.Remaining, rl.Limit, rl.Reset.Format(time.RFC3339))
	}
}

// linkNextRe matches the "next" relation in a Link header.
var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

// ParseNextLink extracts the "next" URL from a GitHub Link header.
func ParseNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	matches := linkNextRe.FindStringSubmatch(linkHeader)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// ParseRateLimit extracts rate-limit info from response headers.
func ParseRateLimit(h http.Header) RateLimit {
	rl := RateLimit{}
	if v := h.Get("X-RateLimit-Remaining"); v != "" {
		rl.Remaining, _ = strconv.Atoi(v)
	}
	if v := h.Get("X-RateLimit-Limit"); v != "" {
		rl.Limit, _ = strconv.Atoi(v)
	}
	if v := h.Get("X-RateLimit-Reset"); v != "" {
		ts, _ := strconv.ParseInt(v, 10, 64)
		rl.Reset = time.Unix(ts, 0)
	}
	return rl
}
