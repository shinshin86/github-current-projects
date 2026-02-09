package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shinshin86/github-current-projects/internal/cli"
	"github.com/shinshin86/github-current-projects/internal/core"
	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

// TestEndToEndMarkdown exercises the full pipeline: fetch -> filter -> sort -> render.
func TestEndToEndMarkdown(t *testing.T) {
	page1, err := os.ReadFile("testdata/repos_page1.json")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	page2, err := os.ReadFile("testdata/repos_page2.json")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	mux := http.NewServeMux()
	var serverURL string
	mux.HandleFunc("/users/testuser/repos", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "", "1":
			w.Header().Set("Link", `<`+serverURL+`/users/testuser/repos?type=owner&per_page=100&page=2>; rel="next"`)
			if _, err := w.Write(page1); err != nil {
				t.Errorf("writing page1 response: %v", err)
			}
		case "2":
			if _, err := w.Write(page2); err != nil {
				t.Errorf("writing page2 response: %v", err)
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	serverURL = server.URL

	client := githubapi.NewClient(server.URL, "", 5*time.Second, nil)
	repos, err := client.FetchAllRepos("testuser")
	if err != nil {
		t.Fatalf("FetchAllRepos: %v", err)
	}

	opts := &cli.Options{
		User:    "testuser",
		Top:     10,
		Format:  "markdown",
		Marker:  "CURRENT PROJECTS",
		BaseURL: server.URL,
	}

	filtered := core.FilterRepos(repos, core.FilterOptions{
		IncludeForks:    opts.IncludeForks,
		IncludeArchived: opts.IncludeArchived,
		MinStars:        opts.MinStars,
		SinceDays:       opts.SinceDays,
	})

	core.SortRepos(filtered)
	filtered = core.TopN(filtered, opts.Top)

	output := core.RenderMarkdown(filtered, opts.Marker)

	// Should contain awesome-project and small-tool (both public, non-fork, non-archived)
	// Should NOT contain forked-repo, archived-repo, private-repo
	if !strings.Contains(output, "[awesome-project]") {
		t.Error("missing awesome-project")
	}
	if !strings.Contains(output, "[small-tool]") {
		t.Error("missing small-tool")
	}
	if strings.Contains(output, "forked-repo") {
		t.Error("forked-repo should be excluded")
	}
	if strings.Contains(output, "archived-repo") {
		t.Error("archived-repo should be excluded")
	}
	if strings.Contains(output, "private-repo") {
		t.Error("private-repo should be excluded")
	}

	// awesome-project has higher stars (100) than small-tool (2), same pushed_at
	// so awesome-project should come first
	awIdx := strings.Index(output, "awesome-project")
	stIdx := strings.Index(output, "small-tool")
	if awIdx > stIdx {
		t.Error("awesome-project should appear before small-tool (higher stars)")
	}
}

// TestEndToEndJSON tests the full pipeline with JSON output.
func TestEndToEndJSON(t *testing.T) {
	page1, err := os.ReadFile("testdata/repos_page1.json")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/users/testuser/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(page1); err != nil {
			t.Errorf("writing page1 response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := githubapi.NewClient(server.URL, "", 5*time.Second, nil)
	repos, err := client.FetchAllRepos("testuser")
	if err != nil {
		t.Fatalf("FetchAllRepos: %v", err)
	}

	filtered := core.FilterRepos(repos, core.FilterOptions{
		MinStars: 50,
	})
	core.SortRepos(filtered)

	output, err := core.RenderJSON(filtered)
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	if !strings.Contains(output, `"name": "awesome-project"`) {
		t.Error("missing awesome-project in JSON output")
	}
	if !strings.Contains(output, `"stargazers_count": 100`) {
		t.Error("missing stargazers_count in JSON output")
	}
}

// TestEndToEndPatchREADME tests fetching repos and patching a README.
func TestEndToEndPatchREADME(t *testing.T) {
	page1, err := os.ReadFile("testdata/repos_page1.json")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/users/testuser/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(page1); err != nil {
			t.Errorf("writing page1 response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := githubapi.NewClient(server.URL, "", 5*time.Second, nil)
	repos, err := client.FetchAllRepos("testuser")
	if err != nil {
		t.Fatalf("FetchAllRepos: %v", err)
	}

	filtered := core.FilterRepos(repos, core.FilterOptions{})
	core.SortRepos(filtered)

	markdown := core.RenderMarkdown(filtered, "CURRENT PROJECTS")

	readme, err := os.ReadFile("testdata/readme_with_markers.md")
	if err != nil {
		t.Fatalf("reading readme testdata: %v", err)
	}

	result, err := core.PatchREADME(string(readme), markdown, "CURRENT PROJECTS", false)
	if err != nil {
		t.Fatalf("PatchREADME: %v", err)
	}

	if !strings.Contains(result.Content, "# My Profile") {
		t.Error("header should be preserved")
	}
	if !strings.Contains(result.Content, "## About Me") {
		t.Error("footer should be preserved")
	}
	if !strings.Contains(result.Content, "[awesome-project]") {
		t.Error("new content should be present")
	}
	if strings.Contains(result.Content, "[old-project]") {
		t.Error("old content should be replaced")
	}
}
