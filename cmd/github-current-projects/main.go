package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shinshin86/github-current-projects/internal/cli"
	"github.com/shinshin86/github-current-projects/internal/core"
	"github.com/shinshin86/github-current-projects/internal/githubapi"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	logger := log.New(os.Stderr, "[github-current-projects] ", log.LstdFlags)

	opts, err := cli.ParseArgs(args, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if cli.IsUsageError(err) {
			return 2
		}
		return 1
	}

	// Resolve token: CLI flag takes priority, then GITHUB_TOKEN env var
	token := opts.Token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	if err := cli.ValidateOptions(opts, token); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if cli.IsUsageError(err) {
			return 2
		}
		return 1
	}

	client := githubapi.NewClient(opts.BaseURL, token, 30*time.Second, logger)

	repos, err := client.FetchAllRepos(opts.User)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching repositories: %v\n", err)
		return 1
	}

	// Filter
	filtered := core.FilterRepos(repos, core.FilterOptions{
		IncludeForks:       opts.IncludeForks,
		IncludeArchived:    opts.IncludeArchived,
		MinStars:           opts.MinStars,
		SinceDays:          opts.SinceDays,
		RequireDescription: opts.RequireDescription,
		Tags:               opts.Tags,
		TagsMatchAll:       opts.TagMatch == "all",
	})

	// Sort
	core.SortReposBy(filtered, opts.Sort)

	// Top N
	filtered = core.TopN(filtered, opts.Top)

	// Render
	var output string
	switch opts.Format {
	case "json":
		output, err = core.RenderJSON(filtered)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering JSON: %v\n", err)
			return 1
		}
	default:
		output = core.RenderMarkdown(filtered, opts.Marker)
	}

	// If --readme is specified, patch the existing file
	if opts.ReadmePath != "" {
		existing, err := os.ReadFile(opts.ReadmePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading README %q: %v\n", opts.ReadmePath, err)
			return 1
		}

		result, err := core.PatchREADME(string(existing), output, opts.Marker, opts.AppendIfMissing)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error patching README: %v\n", err)
			return 1
		}

		outPath := opts.ReadmePath
		if opts.OutPath != "" {
			outPath = opts.OutPath
		}

		if err := os.WriteFile(outPath, []byte(result.Content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %q: %v\n", outPath, err)
			return 1
		}
		logger.Printf("README updated: %s", outPath)
		return 0
	}

	// Write output
	if opts.OutPath != "" {
		if err := os.WriteFile(opts.OutPath, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %q: %v\n", opts.OutPath, err)
			return 1
		}
		logger.Printf("Output written to: %s", opts.OutPath)
	} else {
		fmt.Print(output)
	}

	return 0
}
