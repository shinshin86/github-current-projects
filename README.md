# github-current-projects

English | [日本語](README_ja.md)

A CLI tool that fetches a GitHub user's public repositories, filters and sorts them, and generates a "Current Projects" section in Markdown. It can also automatically embed the section into an existing README.

## Features

- Fetches repository list from GitHub REST API with full pagination support
- Filters by star count, push date, fork/archived status
- Outputs in Markdown or JSON
- Safely replaces marker sections in an existing README
- Preserves line endings (LF/CRLF)
- Zero external dependencies (standard library only)
- Test-friendly design (`--base-url` allows swapping the API endpoint)

## Setup

### Install

```bash
go install github.com/shinshin86/github-current-projects/cmd/github-current-projects@latest
```

### Build from Source

```bash
git clone https://github.com/shinshin86/github-current-projects.git
cd github-current-projects
make build
```

## Usage

> **Note**  
> By default, only the top 10 repositories are shown (`--top 10`). Use `--top 0` to show all.  
> This default keeps output from getting too long, helps avoid unnecessary rate-limit hits, and is safer for users.

### Basic (Markdown output to stdout)

The tool works without a token for fetching public repositories. However, the unauthenticated rate limit (60 req/h) applies.

```bash
github-current-projects --user YOUR_USERNAME
```

### With Token for Higher Rate Limits

Providing a token enables the authenticated rate limit (5,000 req/h). There are two ways to specify a token; the `--token` flag takes priority over the environment variable.
For security, when a token is set, `--base-url` must use `https` (plain `http` is allowed only for `localhost`).

**1. Environment variable (recommended):**

```bash
export GITHUB_TOKEN=ghp_xxxx
github-current-projects --user YOUR_USERNAME
```

**2. `--token` flag:**

```bash
github-current-projects --user YOUR_USERNAME --token ghp_xxxx
```

### With Filter Options

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --top 5 \
  --min-stars 10 \
  --since-days 90 \
  --include-forks
```

### Sort by Stars

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --sort stars
```

### Show All Repos With Descriptions, Sorted by Stars

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --require-description \
  --top 0 \
  --sort stars
```

### With Topic Filters

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --require-description \
  --topics go \
  --topics cli \
  --tag-match all
```

### JSON Output

```bash
github-current-projects --user YOUR_USERNAME --format json
```

### Output to File

```bash
github-current-projects --user YOUR_USERNAME --out projects.md
```

### Update Marker Section in an Existing README

Place the following markers in your README.md beforehand:

```markdown
<!-- BEGIN CURRENT PROJECTS -->
<!-- END CURRENT PROJECTS -->
```

Then run:

```bash
github-current-projects --user YOUR_USERNAME --readme README.md
```

Note: `--readme` supports Markdown output only (it cannot be combined with `--format json`).

To append the section if markers are not found:

```bash
github-current-projects --user YOUR_USERNAME --readme README.md --append-if-missing
```

## CLI Options

| Option | Description | Default |
|---|---|---|
| `--user` | GitHub username (required) | - |
| `--token` | GitHub personal access token | env `GITHUB_TOKEN` |
| `--top` | Number of repos to display | 10 |
| `--min-stars` | Minimum star count | 0 |
| `--include-forks` | Include forked repositories | false |
| `--include-archived` | Include archived repositories | false |
| `--since-days` | Only repos pushed within N days (0 = no limit) | 0 |
| `--require-description` | Only include repos with a description | false |
| `--topics` | Filter by GitHub topic (repeatable) | - |
| `--tag-match` | Topic match mode (`any` / `all`) | `any` |
| `--sort` | Sort order (`pushed` / `stars`) | `pushed` |
| `--readme` | Path to an existing README.md for patching | - |
| `--out` | Output file path (default: stdout) | - |
| `--marker` | Marker name for the README section | `CURRENT PROJECTS` |
| `--format` | Output format (`markdown` / `json`) | `markdown` |
| `--base-url` | GitHub API base URL | `https://api.github.com` |
| `--append-if-missing` | Append section if markers are not found | false |

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | API error or file I/O failure |
| 2 | Invalid arguments |

## Scheduled Updates with GitHub Actions

Example workflow to periodically update a profile README (`username/username` repository):

```yaml
name: Update Current Projects

on:
  schedule:
    - cron: '0 0 * * *'  # Daily at UTC 00:00
  workflow_dispatch:       # Allow manual trigger

permissions:
  contents: write

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install github-current-projects
        run: go install github.com/shinshin86/github-current-projects/cmd/github-current-projects@latest

      - name: Update README
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          github-current-projects \
            --user ${{ github.repository_owner }} \
            --readme README.md \
            --top 10 \
            --min-stars 1

      - name: Commit and push
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git diff --quiet README.md || (git add README.md && git commit -m "Update current projects" && git push)
```

### About Tokens

- The `GITHUB_TOKEN` provided automatically by Actions is sufficient for reading public repositories
- If you need access to private repositories, register a Personal Access Token with the `repo` scope as a secret
- **Tokens are never logged.** The CLI does not write the token to stdout or stderr

## Common Errors

### `GitHub API returned status 403`

You have hit the rate limit. Specify `--token` to use the authenticated rate (5,000 req/h).

### `marker "CURRENT PROJECTS" not found in README`

The following markers are missing from your README.md:

```markdown
<!-- BEGIN CURRENT PROJECTS -->
<!-- END CURRENT PROJECTS -->
```

Add the markers manually, or use the `--append-if-missing` flag.

### `GitHub API returned status 404`

The specified username does not exist or may be misspelled.

## Output Examples

### Markdown

```markdown
<!-- BEGIN CURRENT PROJECTS -->
## Current Projects

- [awesome-project](https://github.com/user/awesome-project) (Go) - An awesome project
- [web-app](https://github.com/user/web-app) (TypeScript) - A modern web application
- [dotfiles](https://github.com/user/dotfiles) - My dotfiles
<!-- END CURRENT PROJECTS -->
```

### JSON

```json
[
  {
    "name": "awesome-project",
    "html_url": "https://github.com/user/awesome-project",
    "description": "An awesome project",
    "language": "Go",
    "pushed_at": "2025-01-15T10:00:00Z",
    "stargazers_count": 100
  }
]
```

## For Developers

### Run Tests

```bash
make test
```

### Run Tests with Race Detector

```bash
make test-race
```

### Coverage Report

```bash
make cover
# Generates coverage.html
```

### Format and Static Analysis

```bash
make lint
```

### Project Structure

```
cmd/github-current-projects/main.go  # Entry point
internal/
  cli/         # CLI argument parsing
  core/        # Filter, sort, rendering, patching (business logic)
  githubapi/   # GitHub API client (pagination, type definitions)
testdata/      # Test fixtures
.github/workflows/ci.yml  # CI configuration
```

## License

MIT
