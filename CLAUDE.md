# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

giggles.ai is a community-driven platform for collecting and browsing AI memes. It consists of two independent components: a Go web crawler that finds and downloads meme images, and a static web GUI for searching/viewing them.

## Build & Run Commands

### Crawler (Go)

```bash
# Build
cd crawler && go build -o crawler .

# Run with S3 upload
./crawler -start-urls "url1,url2" -s3-bucket "bucket-name" -workers 5 -delay 2s -max-pages 100

# Run with local-only storage (no S3)
./crawler -start-urls "url1,url2" -local-dir "found-images"

# Run deduplication on downloaded images (exits after dedupe)
./crawler -dedupe -dedupe-dir "found-images"

# Download dependencies
go mod download
```

TLS certificate verification is on by default; pass `-insecure` only for testing. `-start-urls` (comma-separated) supersedes the deprecated `-start-url`.

### GUI (Static Frontend)

No build process. Serve or open files directly:
```bash
open gui/index.html            # macOS
python -m http.server 8000     # local dev server from project root
```

### Testing

Run the Go test suite from the `crawler/` directory:

```bash
cd crawler
go test ./...              # all packages
go test -race ./...        # with the race detector (recommended — the worker pool is concurrent)
go test -run TestCanCrawl ./internal/crawler/   # a single test
```

Coverage so far: unit tests for the pure crawler helpers (`isMemeImage`, `resolveURL`, `generateFilename`, `getUniqueFilePath`, `extractDomain`) in `internal/crawler/crawler_test.go`; `httptest`-based tests for `canCrawl` and an end-to-end crawl in `internal/crawler/integration_test.go`; and deduplicator tests in `dedupe_test.go`. The GUI (`gui/`) has no JS test harness yet — see `TODO.md`.

> Note: `crawler/.gitignore` anchors the binary pattern as `/crawler`. An unanchored `crawler` pattern silently excludes the entire `internal/crawler/` source directory from git — do not reintroduce it.

## Architecture

### Two Independent Components

**Crawler** (`crawler/`) — Go application using a worker pool concurrency model:
- `main.go` — CLI entry point, flag parsing, orchestration
- `dedupe.go` — SHA256-based image deduplication with JSON hash database (`.hashdb.json`)
- `internal/crawler/` — Core crawling logic: HTML parsing, robots.txt compliance, meme image detection heuristics, link extraction
- `internal/s3/` — AWS S3 client wrapper for image uploads
- A `coordinator()` goroutine feeds a channel-based task queue consumed by N `worker()` goroutines; `sync.RWMutex` guards the visited-URL set and `sync.WaitGroup` tracks completion
- All queue sends go through `enqueue()`, which shares `queueMu` with `closeQueue()` so a close can never race an in-flight send (a closed channel panics rather than taking a `select` `default`). Do not send to `c.queue` directly
- Graceful shutdown: `Stop()` closes a `done` channel that workers and the coordinator select on; `main.go` wires it to `SIGINT`/`SIGTERM` so stats still print on interrupt
- Caches robots.txt per domain (`canCrawl` / `fetchRobotsTxt`)
- `isMemeImage` decides what to download via filename-keyword + known-meme-domain heuristics — the central tuning point for crawl precision (its keyword matching has known false positives; see `TODO.md`)

**GUI** (`gui/`) — Vanilla HTML/CSS/JS, no framework:
- `app.js` — `MemeSearchApp` class handling search, pagination, sorting, filtering (currently uses mock data; future: API backend)
- `index.html` / `styles.css` — Responsive search interface with CSS custom properties for theming

**Landing Page** (root) — `index.html` + `style.css`, separate from the GUI

### Key Dependencies (Crawler)
- `github.com/aws/aws-sdk-go` — S3 integration
- `github.com/temoto/robotstxt` — robots.txt parsing
- `golang.org/x/net` — HTML tokenizer/parser

### AWS Configuration
Crawler reads credentials from environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` (default: us-east-1), `AWS_SESSION_TOKEN` (optional). See `crawler/credentials.example` for template.

## Known Issues & Planned Work

`TODO.md` is the tracked, prioritized backlog (security, correctness, feature gaps, and the planned test suite). Consult it before starting work — several entries cite exact files and line numbers (e.g. the visited-URL TOCTOU race, unbounded response reads, unescaped `meme.url` in the GUI). When testing is added, the "Testing" section enumerates the specific cases expected per function.

## Documentation

Extensive docs live in `docs/` — setup, architecture, crawler usage, GUI guide, development, contributing guidelines, roadmap, and changelog.
