package core

import (
	"strings"
	"testing"
	"time"

	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

func TestRenderMarkdownFull(t *testing.T) {
	repos := []githubapi.Repository{
		{
			Name:        "awesome",
			HTMLURL:     "https://github.com/u/awesome",
			Description: "An awesome project",
			Language:    "Go",
		},
	}
	result := RenderMarkdown(repos, "CURRENT PROJECTS")

	if !strings.Contains(result, "<!-- BEGIN CURRENT PROJECTS -->") {
		t.Error("missing BEGIN marker")
	}
	if !strings.Contains(result, "<!-- END CURRENT PROJECTS -->") {
		t.Error("missing END marker")
	}
	if !strings.Contains(result, "## Current Projects") {
		t.Error("missing heading")
	}
	if !strings.Contains(result, "- [awesome](https://github.com/u/awesome) (Go) - An awesome project") {
		t.Error("missing repo line")
	}
}

func TestRenderMarkdownNoLanguage(t *testing.T) {
	repos := []githubapi.Repository{
		{
			Name:        "no-lang",
			HTMLURL:     "https://github.com/u/no-lang",
			Description: "Has no language",
			Language:    "",
		},
	}
	result := RenderMarkdown(repos, "CURRENT PROJECTS")

	if strings.Contains(result, "()") {
		t.Error("should not contain empty parentheses for language")
	}
	if !strings.Contains(result, "- [no-lang](https://github.com/u/no-lang) - Has no language") {
		t.Errorf("unexpected line format: %s", result)
	}
}

func TestRenderMarkdownNoDescription(t *testing.T) {
	repos := []githubapi.Repository{
		{
			Name:     "no-desc",
			HTMLURL:  "https://github.com/u/no-desc",
			Language: "Rust",
		},
	}
	result := RenderMarkdown(repos, "CURRENT PROJECTS")

	if !strings.Contains(result, "- [no-desc](https://github.com/u/no-desc) (Rust)") {
		t.Errorf("unexpected line format: %s", result)
	}
	// Should not have trailing " - "
	if strings.Contains(result, "- \n") {
		t.Error("should not have trailing dash")
	}
}

func TestRenderMarkdownNoLanguageNoDescription(t *testing.T) {
	repos := []githubapi.Repository{
		{
			Name:    "bare",
			HTMLURL: "https://github.com/u/bare",
		},
	}
	result := RenderMarkdown(repos, "CURRENT PROJECTS")

	if !strings.Contains(result, "- [bare](https://github.com/u/bare)\n") {
		t.Errorf("unexpected line format: %s", result)
	}
}

func TestRenderMarkdownEmpty(t *testing.T) {
	result := RenderMarkdown(nil, "CURRENT PROJECTS")

	if !strings.Contains(result, "_No public projects matched._") {
		t.Error("missing no-results message")
	}
}

func TestRenderJSON(t *testing.T) {
	repos := []githubapi.Repository{
		{
			Name:            "test",
			HTMLURL:         "https://github.com/u/test",
			Description:     "desc",
			Language:        "Go",
			PushedAt:        time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			StargazersCount: 42,
		},
	}
	result, err := RenderJSON(repos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `"name": "test"`) {
		t.Error("missing name field")
	}
	if !strings.Contains(result, `"stargazers_count": 42`) {
		t.Error("missing stargazers_count field")
	}
	if !strings.Contains(result, `"pushed_at": "2025-01-15T10:00:00Z"`) {
		t.Error("missing or incorrect pushed_at field")
	}
}

func TestRenderJSONEmpty(t *testing.T) {
	result, err := RenderJSON([]githubapi.Repository{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "[]") {
		t.Errorf("expected empty array, got: %s", result)
	}
}

func TestRenderMarkdownEscapesUserControlledFields(t *testing.T) {
	repos := []githubapi.Repository{
		{
			Name:        "x](javascript:alert(1))",
			HTMLURL:     "https://github.com/u/repo(with)paren",
			Description: "line1\n<script>alert(1)</script>",
			Language:    "Go",
		},
	}

	result := RenderMarkdown(repos, "CURRENT PROJECTS")

	if strings.Contains(result, "<script>") {
		t.Fatalf("script tag should be escaped: %s", result)
	}
	if !strings.Contains(result, "&lt;script&gt;alert\\(1\\)&lt;/script&gt;") {
		t.Fatalf("escaped description not found: %s", result)
	}
	if !strings.Contains(result, "repo%28with%29paren") {
		t.Fatalf("URL parentheses should be escaped: %s", result)
	}
}

func TestRenderMarkdownRejectsNonHTTPLinkScheme(t *testing.T) {
	repos := []githubapi.Repository{
		{
			Name:        "bad-link",
			HTMLURL:     "javascript:alert(1)",
			Description: "desc",
		},
	}

	result := RenderMarkdown(repos, "CURRENT PROJECTS")

	if strings.Contains(result, "javascript:alert") {
		t.Fatalf("unsafe scheme should not be rendered as link: %s", result)
	}
	if !strings.Contains(result, "- bad-link - desc") {
		t.Fatalf("expected non-link fallback line: %s", result)
	}
}
