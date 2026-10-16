#!/usr/bin/env bash
#
# Crawl, then regenerate the GUI manifest in one step.
#
# The crawler downloads images and BUILDS THE GUI INDEX as two separate
# operations. Run only the crawl and the GUI keeps showing stale data; this
# wrapper does both so they never drift.
#
# Usage:
#   scripts/crawl.sh -start-urls "https://example.com" -max-pages 50
#
# Any flags you pass are forwarded to the crawler verbatim, EXCEPT local image
# storage and manifest output, which this script owns. Override those via env:
#
#   IMAGES_DIR         where images are downloaded (relative to crawler/)
#                      default: found-images
#   MANIFEST_OUT       manifest path written for the GUI
#                      default: ../gui/memes.local.json   (gitignored local artifact)
#   MANIFEST_BASE_URL  URL prefix prepended to each image filename
#                      default: /crawler/found-images     (correct when serving
#                      the repo root, e.g. `python3 -m http.server` from the root)
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT/crawler"

IMAGES_DIR="${IMAGES_DIR:-found-images}"
MANIFEST_OUT="${MANIFEST_OUT:-../gui/memes.local.json}"
MANIFEST_BASE_URL="${MANIFEST_BASE_URL:-/crawler/found-images}"

echo ">> Crawling (images -> crawler/${IMAGES_DIR}) ..."
go run . -local-dir "$IMAGES_DIR" "$@"

echo ">> Indexing for the GUI (-> ${MANIFEST_OUT}) ..."
go run . -gen-manifest \
  -manifest-dir "$IMAGES_DIR" \
  -manifest-out "$MANIFEST_OUT" \
  -manifest-base-url "$MANIFEST_BASE_URL"

echo ">> Done. Serve the repo root and open http://localhost:8000/gui/index.html"
echo "   e.g.  python3 -m http.server 8000"
