package githubapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestFetchAllReposPaginated(t *testing.T) {
	page1, err := os.ReadFile("../../testdata/repos_page1.json")
	if err != nil {
		t.Fatalf("reading page1: %v", err)
	}
	page2, err := os.ReadFile("../../testdata/repos_page2.json")
	if err != nil {
		t.Fatalf("reading page2: %v", err)
	}

	mux := http.NewServeMux()
	var serverURL string

	mux.HandleFunc("/users/testuser/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Error("missing Accept header")
		}

		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "58")
		w.Header().Set("X-RateLimit-Limit", "60")

		switch page {
		case "", "1":
			w.Header().Set("Link", `<`+serverURL+`/users/testuser/repos?type=owner&per_page=100&page=2>; rel="next"`)
			if _, err := w.Write(page1); err != nil {
				t.Errorf("writing page1 response: %v", err)
			}
		case "2":
			// No Link header = last page
			if _, err := w.Write(page2); err != nil {
				t.Errorf("writing page2 response: %v", err)
			}
		default:
			t.Errorf("unexpected page: %s", page)
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	serverURL = server.URL

	client := NewClient(server.URL, "test-token", 0, nil)
	repos, err := client.FetchAllRepos("testuser")
	if err != nil {
		t.Fatalf("FetchAllRepos: %v", err)
	}

	// page1 has 3 repos, page2 has 2 repos = 5 total
	if len(repos) != 5 {
		t.Fatalf("expected 5 repos, got %d", len(repos))
	}

	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Name
	}
	expected := "awesome-project,forked-repo,archived-repo,small-tool,private-repo"
	got := strings.Join(names, ",")
	if got != expected {
		t.Errorf("repos = %s, want %s", got, expected)
	}
}

func TestFetchAllReposAuthHeader(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/testuser/repos", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-secret-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer my-secret-token")
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("[]")); err != nil {
			t.Errorf("writing empty response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "my-secret-token", 0, nil)
	repos, err := client.FetchAllRepos("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestFetchAllReposNoToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/testuser/repos", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("should not have Authorization header without token, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("[]")); err != nil {
			t.Errorf("writing empty response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "", 0, nil)
	_, err := client.FetchAllRepos("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchAllReposAPIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/testuser/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"message":"Not Found"}`)); err != nil {
			t.Errorf("writing error response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "", 0, nil)
	_, err := client.FetchAllRepos("testuser")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code: %v", err)
	}
}
