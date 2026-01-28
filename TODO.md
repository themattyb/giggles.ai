# TODO

## Security

- [ ] Enable TLS verification by default in crawler HTTP client (`crawler/internal/crawler/crawler.go:91-92`). Add an opt-in `--insecure` CLI flag for testing instead of hardcoding `InsecureSkipVerify: true`.
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

- [ ] Add Go unit tests for core crawler logic: `isMemeImage`, `resolveURL`, `generateFilename`, `CalculateHash`, and deduplication.
- [ ] Add JavaScript tests for `MemeSearchApp` search, sort, pagination, and `escapeHtml`.
