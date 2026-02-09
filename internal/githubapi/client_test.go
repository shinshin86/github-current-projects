package githubapi

import (
	"net/http"
	"testing"
	"time"
)

func TestParseNextLink(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "standard next link",
			header:   `<https://api.github.com/user/repos?page=2>; rel="next", <https://api.github.com/user/repos?page=5>; rel="last"`,
			expected: "https://api.github.com/user/repos?page=2",
		},
		{
			name:     "no next link",
			header:   `<https://api.github.com/user/repos?page=5>; rel="last"`,
			expected: "",
		},
		{
			name:     "empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "next only",
			header:   `<https://api.github.com/user/repos?page=3>; rel="next"`,
			expected: "https://api.github.com/user/repos?page=3",
		},
		{
			name:     "mixed relations with extra spaces",
			header:   `<https://api.github.com/user/repos?page=1>; rel="prev", <https://api.github.com/user/repos?page=3>;  rel="next"`,
			expected: "https://api.github.com/user/repos?page=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseNextLink(tt.header)
			if got != tt.expected {
				t.Errorf("ParseNextLink(%q) = %q, want %q", tt.header, got, tt.expected)
			}
		})
	}
}

func TestParseRateLimit(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Remaining", "42")
	h.Set("X-RateLimit-Limit", "60")
	h.Set("X-RateLimit-Reset", "1700000000")

	rl := ParseRateLimit(h)

	if rl.Remaining != 42 {
		t.Errorf("Remaining = %d, want 42", rl.Remaining)
	}
	if rl.Limit != 60 {
		t.Errorf("Limit = %d, want 60", rl.Limit)
	}
	expected := time.Unix(1700000000, 0)
	if !rl.Reset.Equal(expected) {
		t.Errorf("Reset = %v, want %v", rl.Reset, expected)
	}
}

func TestParseRateLimitEmpty(t *testing.T) {
	h := http.Header{}
	rl := ParseRateLimit(h)

	if rl.Remaining != 0 || rl.Limit != 0 {
		t.Errorf("Expected zero values for empty headers, got %+v", rl)
	}
}
