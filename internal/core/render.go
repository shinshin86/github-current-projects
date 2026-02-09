package core

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

// RenderMarkdown produces the Markdown section for the given repos.
func RenderMarkdown(repos []githubapi.Repository, marker string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<!-- BEGIN %s -->\n", marker))
	sb.WriteString("## Current Projects\n\n")

	if len(repos) == 0 {
		sb.WriteString("_No public projects matched._\n")
	} else {
		for _, r := range repos {
			sb.WriteString(formatRepoLine(r))
			sb.WriteByte('\n')
		}
	}

	sb.WriteString(fmt.Sprintf("<!-- END %s -->\n", marker))
	return sb.String()
}

func formatRepoLine(r githubapi.Repository) string {
	name := escapeMarkdownInline(strings.TrimSpace(r.Name))
	link := sanitizeMarkdownURL(r.HTMLURL)
	language := escapeMarkdownInline(strings.TrimSpace(r.Language))
	description := escapeMarkdownInline(normalizeInlineText(r.Description))

	var parts []string
	if link != "" {
		parts = append(parts, fmt.Sprintf("- [%s](%s)", name, link))
	} else {
		parts = append(parts, fmt.Sprintf("- %s", name))
	}

	if language != "" {
		parts = append(parts, fmt.Sprintf("(%s)", language))
	}

	if description != "" {
		parts = append(parts, fmt.Sprintf("- %s", description))
	}

	return strings.Join(parts, " ")
}

var markdownInlineEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\\", "\\\\",
	"[", "\\[",
	"]", "\\]",
	"(", "\\(",
	")", "\\)",
	"`", "\\`",
)

func escapeMarkdownInline(s string) string {
	if s == "" {
		return s
	}
	return markdownInlineEscaper.Replace(s)
}

func normalizeInlineText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.Join(strings.Fields(s), " ")
}

func sanitizeMarkdownURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ReplaceAll(raw, "\n", "")

	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return ""
	}

	// Keep Markdown link syntax stable even when URL contains parentheses.
	return strings.NewReplacer("(", "%28", ")", "%29", " ", "%20").Replace(u.String())
}

// JSONOutput is a single repo entry in JSON output mode.
type JSONOutput struct {
	Name            string `json:"name"`
	HTMLURL         string `json:"html_url"`
	Description     string `json:"description"`
	Language        string `json:"language"`
	PushedAt        string `json:"pushed_at"`
	StargazersCount int    `json:"stargazers_count"`
}

// RenderJSON produces a JSON array string for the given repos.
func RenderJSON(repos []githubapi.Repository) (string, error) {
	out := make([]JSONOutput, len(repos))
	for i, r := range repos {
		out[i] = JSONOutput{
			Name:            r.Name,
			HTMLURL:         r.HTMLURL,
			Description:     r.Description,
			Language:        r.Language,
			PushedAt:        r.PushedAt.Format("2006-01-02T15:04:05Z"),
			StargazersCount: r.StargazersCount,
		}
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	return string(data) + "\n", nil
}
