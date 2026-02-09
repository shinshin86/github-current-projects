package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// Options holds all CLI options.
type Options struct {
	User               string
	Token              string
	Top                int
	MinStars           int
	IncludeForks       bool
	IncludeArchived    bool
	SinceDays          int
	RequireDescription bool
	Tags               []string
	TagMatch           string
	Sort               string
	ReadmePath         string
	OutPath            string
	Marker             string
	Format             string
	BaseURL            string
	AppendIfMissing    bool
}

// ParseArgs parses command-line arguments.
// Returns options and nil on success.
// Returns nil and an error on failure, with exit code hint:
//   - ErrUsage for usage/argument errors (exit 2)
func ParseArgs(args []string, stderr io.Writer) (*Options, error) {
	fs := flag.NewFlagSet("github-current-projects", flag.ContinueOnError)
	fs.SetOutput(stderr)

	opts := &Options{}
	fs.StringVar(&opts.User, "user", "", "GitHub username (required)")
	fs.StringVar(&opts.Token, "token", "", "GitHub personal access token")
	fs.IntVar(&opts.Top, "top", 10, "Number of repos to show")
	fs.IntVar(&opts.MinStars, "min-stars", 0, "Minimum star count")
	fs.BoolVar(&opts.IncludeForks, "include-forks", false, "Include forked repositories")
	fs.BoolVar(&opts.IncludeArchived, "include-archived", false, "Include archived repositories")
	fs.IntVar(&opts.SinceDays, "since-days", 0, "Only repos pushed within N days (0 = no limit)")
	fs.BoolVar(&opts.RequireDescription, "require-description", false, "Only include repos with a description")
	fs.StringVar(&opts.TagMatch, "tag-match", "any", "Topic match mode: any or all")
	fs.StringVar(&opts.Sort, "sort", "pushed", "Sort order: pushed or stars")
	fs.Func("topics", "Filter by GitHub topic (repeatable)", func(v string) error {
		v = strings.TrimSpace(v)
		if v == "" {
			return errors.New("--topics must not be empty")
		}
		opts.Tags = append(opts.Tags, v)
		return nil
	})
	fs.StringVar(&opts.ReadmePath, "readme", "", "Path to existing README.md for patching")
	fs.StringVar(&opts.OutPath, "out", "", "Output file path (default: stdout)")
	fs.StringVar(&opts.Marker, "marker", "CURRENT PROJECTS", "Marker name for README section")
	fs.StringVar(&opts.Format, "format", "markdown", "Output format: markdown or json")
	fs.StringVar(&opts.BaseURL, "base-url", "https://api.github.com", "GitHub API base URL")
	fs.BoolVar(&opts.AppendIfMissing, "append-if-missing", false, "Append section if markers not found in README")

	if err := fs.Parse(args); err != nil {
		return nil, &UsageError{Err: err}
	}

	if opts.User == "" {
		return nil, &UsageError{Err: errors.New("--user is required")}
	}

	if opts.Format != "markdown" && opts.Format != "json" {
		return nil, &UsageError{Err: fmt.Errorf("--format must be 'markdown' or 'json', got %q", opts.Format)}
	}

	if opts.Sort != "pushed" && opts.Sort != "stars" {
		return nil, &UsageError{Err: fmt.Errorf("--sort must be 'pushed' or 'stars', got %q", opts.Sort)}
	}

	if opts.TagMatch != "any" && opts.TagMatch != "all" {
		return nil, &UsageError{Err: fmt.Errorf("--tag-match must be 'any' or 'all', got %q", opts.TagMatch)}
	}

	if opts.Top < 0 {
		return nil, &UsageError{Err: fmt.Errorf("--top must be non-negative, got %d", opts.Top)}
	}

	if err := ValidateOptions(opts, opts.Token); err != nil {
		return nil, err
	}

	return opts, nil
}

// UsageError indicates a usage/argument error (exit code 2).
type UsageError struct {
	Err error
}

func (e *UsageError) Error() string {
	return e.Err.Error()
}

func (e *UsageError) Unwrap() error {
	return e.Err
}

// IsUsageError checks if an error is a UsageError.
func IsUsageError(err error) bool {
	var ue *UsageError
	return errors.As(err, &ue)
}

// ValidateOptions validates option combinations that may depend on the token value.
func ValidateOptions(opts *Options, token string) error {
	if opts.ReadmePath != "" && opts.Format == "json" {
		return &UsageError{Err: errors.New("--readme cannot be used with --format json")}
	}
	if token != "" && !isSecureBaseURL(opts.BaseURL) {
		return &UsageError{Err: errors.New("--base-url must use https when a token is set (http is allowed only for localhost)")}
	}
	return nil
}

func isSecureBaseURL(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return true
	case "http":
		return isLocalhost(u.Hostname())
	default:
		return false
	}
}

func isLocalhost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
