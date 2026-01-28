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

# Run deduplication on downloaded images
./crawler -dedupe -dedupe-dir "found-images"

# Download dependencies
go mod download
```

### GUI (Static Frontend)

No build process. Serve or open files directly:
```bash
open gui/index.html            # macOS
python -m http.server 8000     # local dev server from project root
```

### Testing

No formal test suite exists yet. Go tests would use standard `go test ./...` from the `crawler/` directory.

## Architecture

### Two Independent Components

**Crawler** (`crawler/`) — Go application using a worker pool concurrency model:
- `main.go` — CLI entry point, flag parsing, orchestration
- `dedupe.go` — SHA256-based image deduplication with JSON hash database (`.hashdb.json`)
- `internal/crawler/` — Core crawling logic: HTML parsing, robots.txt compliance, meme image detection heuristics, link extraction
- `internal/s3/` — AWS S3 client wrapper for image uploads
- Uses goroutine worker pool with channel-based task queue, `sync.RWMutex`/`sync.WaitGroup` for synchronization
- Caches robots.txt per domain

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

## Documentation

Extensive docs live in `docs/` — setup, architecture, crawler usage, GUI guide, contributing guidelines, roadmap, and changelog.
