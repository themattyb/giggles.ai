# TODO

## Security

- [x] Enable TLS verification by default in crawler HTTP client (`crawler/internal/crawler/crawler.go:91-92`). Add an opt-in `--insecure` CLI flag for testing instead of hardcoding `InsecureSkipVerify: true`.
- [ ] Cap response body reads with `io.LimitReader` to prevent memory exhaustion from oversized responses (`crawler/internal/crawler/crawler.go`). Suggested limits: ~50MB for images in `downloadImage`, ~1MB for `fetchRobotsTxt`.
- [ ] Escape `meme.url` in `createMemeCard` (`gui/app.js:183`). Currently interpolated directly into `innerHTML` without escaping. Build the img element via `document.createElement` instead.
- [ ] Restrict `LoadCredentialsFromFile` (`crawler/internal/s3/client.go:92-117`) to only set AWS-specific environment variables, or remove the function entirely in favor of the AWS SDK's built-in credential file support.

## Bugs / Correctness

- [ ] Fix TOCTOU race condition on visited-URL check (`crawler/internal/crawler/crawler.go:200-211`). The RLock check and Lock set are separate critical sections, allowing two workers to process the same URL. Use a single Lock section that checks and sets atomically.
- [ ] Fix misleading S3 credential chain (`crawler/internal/s3/client.go:27-44`). `NewEnvCredentials()` doesn't fail at session creation, so the fallback session is dead code. Use the default credential chain instead.

## Feature Gaps

- [ ] Add domain scoping to the crawler. Currently follows links to any domain (`crawler/internal/crawler/crawler.go:380-403`). Add an `--allowed-domains` or `--same-domain` flag to constrain crawling.
- [ ] Handle OS signals (`SIGINT`/`SIGTERM`) for graceful shutdown so workers can drain and stats are printed on interrupt.
- [ ] Connect the GUI to real data. The frontend uses 3 hardcoded placeholder memes (`gui/app.js:91-117`) with no backend API. Build an API layer or have the GUI read from S3 directly.

## Testing

### Crawler unit tests (`crawler/internal/crawler/`)

- [ ] `isMemeImage`: returns true for keyword matches (`https://example.com/ai-meme.jpg`), meme domain matches (`https://imgur.com/photo.png`), and false for non-image extensions (`logo.svg`), URLs with no keywords (`https://example.com/photo.jpg`)
- [ ] `isMemeImage` false positive: the keyword `ai` matches any URL containing those two letters (e.g. `https://example.com/contains.jpg` matches on "ai" in "cont**ai**ns") — test documents this behavior and a fix should tighten matching
- [ ] `resolveURL`: resolves relative paths against a base URL, returns empty string for non-HTTP schemes (`mailto:`, `javascript:`), handles empty input
- [ ] `generateFilename`: extracts filename from URL path, generates a timestamped fallback when URL has no filename, maps content-type to correct extension (png, gif, webp), sanitizes spaces and `%20`
- [ ] `getUniqueFilePath`: returns original path when file doesn't exist, appends `_1`, `_2` etc. when file already exists
- [ ] `extractDomain`: strips port numbers, returns empty string for unparseable URLs
- [ ] `canCrawl`: respects a robots.txt that disallows the path, allows crawling when robots.txt returns 404, allows crawling when robots.txt has no matching rules

### Deduplication unit tests (`crawler/`)

- [ ] `CalculateHash`: two identical files produce the same SHA256 hash, two different files produce different hashes
- [ ] `ProcessFiles`: removes the newer duplicate when two files have the same content, keeps all files when no duplicates exist, writes a valid `.hashdb.json` after processing
- [ ] `LoadDatabase` / `SaveDatabase`: round-trips records through JSON correctly, `LoadDatabase` returns no error when the file doesn't exist yet

### Crawler integration test

- [ ] Stand up an `httptest.Server` serving a small HTML page with `<img>` and `<a>` tags, run the crawler against it with `InsecureSkipVerify: false`, and verify `Stats` fields (pages crawled, images found/downloaded) are correct

### GUI tests (`gui/`)

- [ ] `escapeHtml`: escapes `<`, `>`, `&`, and `"` characters
- [ ] `applySortAndFilter`: "newest" sort returns most recent first, search term filters by title and source, empty search returns all memes
- [ ] `updatePagination`: disables prev button on page 1, disables next button on last page, shows correct "Page X of Y" text
