# TODO

## Security

- [x] Enable TLS verification by default in crawler HTTP client (`crawler/internal/crawler/crawler.go:91-92`). Add an opt-in `--insecure` CLI flag for testing instead of hardcoding `InsecureSkipVerify: true`.
- [x] Cap response body reads with `io.LimitReader` to prevent memory exhaustion from oversized responses (`crawler/internal/crawler/crawler.go`). Caps added: 50MB images (`downloadImage`), 10MB HTML (`processPage`), 1MB `fetchRobotsTxt`.
- [x] Escape `meme.url` in `createMemeCard` (`gui/app.js`). Card is now built imperatively via `document.createElement` with `.src`/`.textContent`, so no field can break out of the attribute context.
- [x] Remove `LoadCredentialsFromFile` (`crawler/internal/s3/client.go`). It was dead code (never called) that called `os.Setenv` for arbitrary keys; the SDK default credential chain already reads `~/.aws/credentials`.

## Bugs / Correctness

- [x] Fix TOCTOU race condition on visited-URL check (`crawler/internal/crawler/crawler.go`). Check and set now happen in a single `Lock` critical section.
- [x] Fix misleading S3 credential chain (`crawler/internal/s3/client.go`). Now uses `session.NewSessionWithOptions` with `SharedConfigEnable` and the SDK default credential chain.
- [x] Fix `isMemeImage` over-broad matching (`crawler/internal/crawler/crawler.go`). Keyword matching is now token-based against the URL path (not raw substrings), so `ai` no longer matches `contains.jpg`; duplicate `deep`/`learning` keywords removed; host excluded so the `.ai` TLD doesn't match every image.
- [x] Fix sort dropping the active search filter (`gui/app.js`). The search term is now stored in `this.searchTerm`; changing the sort order re-applies it instead of resetting to all memes.
- [x] Fix robots.txt rules being silently ignored (`crawler/internal/crawler/crawler.go`). `canCrawl` passed the full URL to `group.Test()`, which expects a request path, so `Disallow:` prefixes never matched. Now passes `parsedURL.Path` (+ query). Caught by `TestCanCrawl`.
- [x] Fix `.gitignore` excluding source (`crawler/.gitignore`). The unanchored `crawler` pattern matched the `internal/crawler/` directory, so `crawler.go` had never been committed. Anchored to `/crawler` (binary only).
- [x] Fix send-on-closed-channel panic (`crawler/internal/crawler/crawler.go`). The coordinator could `close(c.queue)` while a worker was executing `select { case c.queue <- link: default: }` in `processPage` — a closed channel panics instead of taking the `default`. Sends now go through `enqueue`, which shares `queueMu` with `closeQueue` so a close can never race an in-flight send. Covered by `TestStopIsGraceful` under `-race`.

## Feature Gaps

- [x] Add domain scoping to the crawler. Added `-same-domain` (restrict to the start URLs' domains) and `-allowed-domains` (explicit comma-separated allow-list); discovered links are filtered via `isAllowedDomain` before enqueue. Empty list preserves the old "follow anywhere" behavior. Covered by `TestIsAllowedDomain` / `TestNewSameDomainBuildsAllowList`.
- [x] Handle OS signals (`SIGINT`/`SIGTERM`) for graceful shutdown so workers can drain and stats are printed on interrupt. `main.go` installs a handler that calls `Crawler.Stop()`, which closes a `done` channel; workers and the coordinator select on it and exit cleanly, then `Run` returns and stats print as normal.
- [x] Connect the GUI to real data. The GUI now fetches a `memes.json` manifest (default `./memes.json`, override via `window.GIGGLES_MANIFEST_URL`) instead of hardcoded mock data, with loading/error states. The crawler generates the manifest via `-gen-manifest` (`crawler/manifest.go`), optionally with an S3/CDN base URL. Pure data logic extracted to `gui/memeLogic.js`.

## Testing

### Crawler unit tests (`crawler/internal/crawler/`)

- [x] `isMemeImage`: returns true for keyword matches (`https://example.com/ai-meme.jpg`), meme domain matches (`https://imgur.com/photo.png`), and false for non-image extensions (`logo.svg`), URLs with no keywords (`https://example.com/photo.jpg`)
- [x] `isMemeImage` (regression test for the tightened matching): `https://example.com/contains.jpg` returns **false** (no bare-substring `ai` match), `https://example.com/ai-meme.jpg` returns true (token match), and an image on a `.ai` host with no keyword in the path returns false
- [x] `resolveURL`: resolves relative paths against a base URL, returns empty string for non-HTTP schemes (`mailto:`, `javascript:`), handles empty input
- [x] `generateFilename`: extracts filename from URL path, generates a timestamped fallback when URL has no filename, maps content-type to correct extension (png, gif, webp), sanitizes spaces and `%20`
- [x] `getUniqueFilePath`: returns original path when file doesn't exist, appends `_1`, `_2` etc. when file already exists
- [x] `extractDomain`: strips port numbers, returns empty string for unparseable URLs
- [x] `canCrawl`: respects a robots.txt that disallows the path, allows crawling when robots.txt returns 404, allows crawling when robots.txt has no matching rules (`integration_test.go`)

### Deduplication unit tests (`crawler/`)

- [x] `CalculateHash`: two identical files produce the same SHA256 hash, two different files produce different hashes
- [x] `ProcessFiles`: removes the newer duplicate when two files have the same content, keeps all files when no duplicates exist, writes a valid `.hashdb.json` after processing
- [x] `LoadDatabase` / `SaveDatabase`: round-trips records through JSON correctly, `LoadDatabase` returns no error when the file doesn't exist yet

### Crawler integration test

- [x] Stand up an `httptest.Server` serving a small HTML page with `<img>` and `<a>` tags, run the crawler against it with `InsecureSkipVerify: false`, and verify `Stats` fields (pages crawled, images found/downloaded) are correct (`integration_test.go`)

### GUI tests (`gui/`)

- [x] `escapeHtml`: escapes `<`, `>`, `&`, `"`, and `'` characters (`gui/memeLogic.test.js`)
- [x] `filterMemes` / `sortMemes`: newest/oldest ordering, search filters by title and source, empty search returns all memes, sort doesn't mutate input (`gui/memeLogic.test.js`)
- [x] `getPaginationInfo` / `paginate`: prev disabled on page 1, next disabled on last page, correct "Page X of Y" text, correct page slices (`gui/memeLogic.test.js`)
- Note: tests run via Node's built-in runner (`cd gui && npm test`); logic was extracted from `app.js` into `gui/memeLogic.js` to be DOM-free and testable.
