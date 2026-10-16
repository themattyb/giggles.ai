#!/usr/bin/env bash
#
# Publish giggles to staging: the GUI, a freshly generated manifest and the
# crawled images go to the cloudburstlabs portal's GIGGLES R2 bucket, which the
# portal serves at https://web.cloudburstlabs.workers.dev/giggles/ behind its
# own allowlist (see docs/WORKER-HOSTING.md §7). No portal deploy is needed.
#
# Usage:
#   scripts/publish-staging.sh            # upload to the real bucket
#   scripts/publish-staging.sh --local    # seed `wrangler dev`'s local R2 instead
#
# Wrangler runs from the cloudburstlabs checkout, so it uses that repo's pinned
# version, its login and (with --local) the same local state as `npm run dev`
# there. Override via env:
#
#   PORTAL_DIR       cloudburstlabs checkout   default: ../cloudburstlabs
#   GIGGLES_BUCKET   R2 bucket name            default: giggles-staging
#   IMAGES_DIR       images, absolute or relative to crawler/   default: found-images
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PORTAL_DIR="${PORTAL_DIR:-$REPO_ROOT/../cloudburstlabs}"
BUCKET="${GIGGLES_BUCKET:-giggles-staging}"
IMAGES_DIR="${IMAGES_DIR:-found-images}"
[[ "$IMAGES_DIR" = /* ]] || IMAGES_DIR="$REPO_ROOT/crawler/$IMAGES_DIR"

case "${1:-}" in
  "")       MODE=--remote ;;
  --local)  MODE=--local ;;
  *)        echo "usage: $0 [--local]" >&2; exit 2 ;;
esac

[[ -f "$PORTAL_DIR/wrangler.toml" ]] || { echo "No cloudburstlabs checkout at $PORTAL_DIR (set PORTAL_DIR)" >&2; exit 1; }
[[ -d "$IMAGES_DIR" ]] || { echo "No images at $IMAGES_DIR" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

put() { # key file content-type
  (cd "$PORTAL_DIR" && npx wrangler r2 object put "$BUCKET/$1" --file "$2" --content-type "$3" "$MODE") >/dev/null
  echo "  $1"
}

# The portal serves only these names and types; anything else would 404 there.
image_type() {
  case "$(printf '%s' "${1##*.}" | tr '[:upper:]' '[:lower:]')" in
    jpg|jpeg) echo image/jpeg ;;
    png)      echo image/png ;;
    gif)      echo image/gif ;;
    webp)     echo image/webp ;;
  esac
}

# Stage only the images the portal will serve (safe name, known type), then
# index exactly that set, so the manifest never lists an image that 404s.
mkdir "$TMP/img"
for f in "$IMAGES_DIR"/*; do
  name="$(basename "$f")"
  if [[ -f "$f" && -n "$(image_type "$name")" && "$name" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
    cp -p "$f" "$TMP/img/"   # -p keeps mtime, which the manifest uses as uploadedAt
  elif [[ -f "$f" && "$name" != .* ]]; then
    echo "  skipping $name (name or type not servable)" >&2
  fi
done

echo ">> Indexing images for /giggles/img ..."
(cd "$REPO_ROOT/crawler" && go run . -gen-manifest \
  -manifest-dir "$TMP/img" \
  -manifest-out "$TMP/memes.json" \
  -manifest-base-url /giggles/img)

# Images first and the manifest last, so a published manifest never lists an
# image that is not there yet.
echo ">> Uploading images ($MODE, bucket $BUCKET) ..."
for f in "$TMP/img"/*; do
  [[ -f "$f" ]] || continue
  name="$(basename "$f")"
  put "giggles/img/$name" "$f" "$(image_type "$name")"
done

echo ">> Uploading the GUI ..."
put giggles/site/index.html  "$REPO_ROOT/gui/index.html"  "text/html; charset=utf-8"
put giggles/site/app.js      "$REPO_ROOT/gui/app.js"      "text/javascript; charset=utf-8"
put giggles/site/memeLogic.js "$REPO_ROOT/gui/memeLogic.js" "text/javascript; charset=utf-8"
put giggles/site/styles.css  "$REPO_ROOT/gui/styles.css"  "text/css; charset=utf-8"
put giggles/site/memes.json  "$TMP/memes.json"           "application/json; charset=utf-8"

echo ">> Done."
