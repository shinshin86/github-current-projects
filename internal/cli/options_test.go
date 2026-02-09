package cli

import (
	"bytes"
	"testing"
)

func TestParseArgsValid(t *testing.T) {
	args := []string{"--user", "testuser", "--top", "5", "--min-stars", "10", "--format", "json"}
	opts, err := ParseArgs(args, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.User != "testuser" {
		t.Errorf("User = %q, want %q", opts.User, "testuser")
	}
	if opts.Top != 5 {
		t.Errorf("Top = %d, want 5", opts.Top)
	}
	if opts.MinStars != 10 {
		t.Errorf("MinStars = %d, want 10", opts.MinStars)
	}
	if opts.Format != "json" {
		t.Errorf("Format = %q, want %q", opts.Format, "json")
	}
}

func TestParseArgsMissingUser(t *testing.T) {
	args := []string{"--top", "5"}
	_, err := ParseArgs(args, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing --user")
	}
	if !IsUsageError(err) {
		t.Errorf("expected UsageError, got %T", err)
	}
}

func TestParseArgsInvalidFormat(t *testing.T) {
	args := []string{"--user", "u", "--format", "xml"}
	_, err := ParseArgs(args, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !IsUsageError(err) {
		t.Errorf("expected UsageError, got %T", err)
	}
}

func TestParseArgsDefaults(t *testing.T) {
	args := []string{"--user", "u"}
	opts, err := ParseArgs(args, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Top != 10 {
		t.Errorf("Top = %d, want default 10", opts.Top)
	}
	if opts.MinStars != 0 {
		t.Errorf("MinStars = %d, want default 0", opts.MinStars)
	}
	if opts.IncludeForks {
		t.Error("IncludeForks should default to false")
	}
	if opts.IncludeArchived {
		t.Error("IncludeArchived should default to false")
	}
	if opts.SinceDays != 0 {
		t.Errorf("SinceDays = %d, want default 0", opts.SinceDays)
	}
	if opts.RequireDescription {
		t.Error("RequireDescription should default to false")
	}
	if opts.TagMatch != "any" {
		t.Errorf("TagMatch = %q, want default any", opts.TagMatch)
	}
	if opts.Sort != "pushed" {
		t.Errorf("Sort = %q, want default pushed", opts.Sort)
	}
	if opts.Marker != "CURRENT PROJECTS" {
		t.Errorf("Marker = %q, want default", opts.Marker)
	}
	if opts.Format != "markdown" {
		t.Errorf("Format = %q, want default markdown", opts.Format)
	}
	if opts.BaseURL != "https://api.github.com" {
		t.Errorf("BaseURL = %q, want default", opts.BaseURL)
	}
}

func TestParseArgsNegativeTop(t *testing.T) {
	args := []string{"--user", "u", "--top", "-1"}
	_, err := ParseArgs(args, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for negative --top")
	}
	if !IsUsageError(err) {
		t.Errorf("expected UsageError, got %T", err)
	}
}

func TestParseArgsTagMatchInvalid(t *testing.T) {
	args := []string{"--user", "u", "--tag-match", "sometimes"}
	_, err := ParseArgs(args, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for invalid --tag-match")
	}
	if !IsUsageError(err) {
		t.Errorf("expected UsageError, got %T", err)
	}
}

func TestParseArgsSortInvalid(t *testing.T) {
	args := []string{"--user", "u", "--sort", "recent"}
	_, err := ParseArgs(args, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for invalid --sort")
	}
	if !IsUsageError(err) {
		t.Errorf("expected UsageError, got %T", err)
	}
}

func TestParseArgsSortStars(t *testing.T) {
	args := []string{"--user", "u", "--sort", "stars"}
	opts, err := ParseArgs(args, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Sort != "stars" {
		t.Errorf("Sort = %q, want stars", opts.Sort)
	}
}

func TestParseArgsTags(t *testing.T) {
	args := []string{"--user", "u", "--topics", "go", "--topics", "cli"}
	opts, err := ParseArgs(args, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(opts.Tags) != 2 || opts.Tags[0] != "go" || opts.Tags[1] != "cli" {
		t.Errorf("Tags = %v, want [go cli]", opts.Tags)
	}
}

func TestParseArgsReadmeWithJSON(t *testing.T) {
	args := []string{"--user", "u", "--readme", "README.md", "--format", "json"}
	_, err := ParseArgs(args, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for --readme with --format json")
	}
	if !IsUsageError(err) {
		t.Errorf("expected UsageError, got %T", err)
	}
}

func TestParseArgsInsecureBaseURLWithToken(t *testing.T) {
	args := []string{"--user", "u", "--token", "t", "--base-url", "http://example.com"}
	_, err := ParseArgs(args, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for http base-url with token")
	}
	if !IsUsageError(err) {
		t.Errorf("expected UsageError, got %T", err)
	}
}

func TestValidateOptionsAllowsLocalhostHTTP(t *testing.T) {
	opts := &Options{
		User:    "u",
		Format:  "markdown",
		BaseURL: "http://localhost:8080",
	}
	if err := ValidateOptions(opts, "token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
