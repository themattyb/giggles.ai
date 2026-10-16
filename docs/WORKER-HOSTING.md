# Hosting giggles.ai's GUI as a gated Worker (staging) - design notes

Read-only research, 2026-10-16. Sources: giggles.ai repo (README, CLAUDE.md, docs/GUI.md, SETUP.md, gui/*, crawler/internal/s3), cloudburstlabs `src/sso.ts`, REAtrends `site/src/admin/ssoGate.tsx`, Cloudflare docs.
Items marked (verify) were not confirmed in the docs fetched.

## 1. Workers Static Assets

**Serve a folder.** `wrangler.toml` gets
```toml
name = "giggles"
account_id = "45d0740fd203828ed967d25195341b87"   # same pinned account as reanews/web
main = "src/index.ts"
compatibility_date = "2026-10-01"
[assets]
directory = "./public"
binding = "ASSETS"
run_worker_first = true
```
Default (`run_worker_first = false`) serves a matching asset without invoking the Worker; the Worker only runs when nothing matches. That default would leave the files ungated.

**Run the Worker first.**
- `run_worker_first = true` invokes the Worker unconditionally for every request. This is the right form for "gate everything".
- Array form runs the Worker first only on matching paths; everything else stays asset-first. `*` is a deep match, `!` negates (negatives win, order irrelevant), max 100 entries. Example: `["/*", "!/health"]` is possible, but `/health` can simply be handled in the Worker, so `true` is simpler and fails closed.
- Array form is the foot-gun for gating: any path not listed is public.

**Serve the asset from the Worker.** `return env.ASSETS.fetch(request)`. It accepts a Request, URL or string, applies `html_handling` and `not_found_handling`, and uses only the pathname. `ASSETS` needs `binding = "ASSETS"` (type `Fetcher` in `Env`).

**Limits.** 20,000 files per Worker version (Free) / 100,000 (Paid; needs wrangler >= 4.34.0 for the higher count); 25 MiB per file. The GUI is about 8 small files, so the limits are irrelevant unless images are ever bundled.

**Wrangler.** No minimum is stated for static assets. The other repos use 4.x (REAtrends: `^4.151.0`), which is fine. `_headers` rules: max 100, 2,000 chars per line.

**Practical gotcha: the assets directory.** `gui/` holds files that must not be published: `memeLogic.test.js`, `package.json`, and the gitignored `memes.local.json` (which has local-only URLs, see section 2). Options: a build/copy step into `worker/public/`, or point `directory` at `../gui` and add an `.assetsignore` (verify the exact file name and semantics in docs). The copy-step route also lets the build swap in a production manifest.

**Docs:**
- https://developers.cloudflare.com/workers/static-assets/binding/
- https://developers.cloudflare.com/workers/static-assets/routing/worker-script/
- https://developers.cloudflare.com/workers/static-assets/headers/
- https://developers.cloudflare.com/workers/platform/limits/

## 2. What needs gating

**How the GUI loads data** (gui/app.js, docs/GUI.md): a static ES-module app. Manifest order: `window.GIGGLES_MANIFEST_URL` (no fallback) -> `./memes.local.json` -> `./memes.json`, all relative to the page and fetched with `cache: 'no-cache'`. Each entry is `{id, url, title, source, uploadedAt}`.

**Data today.**
- `memes.json` is committed sample data (3 `placehold.co` images). It would render on staging as-is.
- `memes.local.json` is generated and gitignored. Its URLs are `/crawler/found-images/<file>`, i.e. local-disk paths relative to serving the repo root. Those images are not in any Worker asset directory, so on the Worker they 404.
- The crawler can emit real URLs: `go run . -gen-manifest ... -manifest-base-url <prefix>` (manifest.go `joinURL`), so a manifest with S3 URLs is one flag away.

**S3 images** (crawler/internal/s3): key `memes/<filename>`, uploaded with ACL `public-read`, URL `https://<bucket>.s3.<region>.amazonaws.com/memes/<filename>`. They are public to anyone who has the URL. (Bucket-level Block Public Access may reject `public-read` ACLs on newer buckets; the owner should confirm the live state. ACL enforcement is off by default for new buckets, so uploads may fail or the objects may not actually be public. Check before relying on it.)

**What gating does and doesn't protect.**
- Protects: the HTML/JS/CSS, and `memes.json`, which is the index (titles, sources, the list of URLs). A visitor without the cookie can't discover what is in the bucket or enumerate it.
- Does not protect: the image bytes. Anyone with an S3 URL can fetch it. Keys are filenames derived from source names (guessable, e.g. `Dubai_chocolate_bar_meme.jpg`), and a public-read bucket may allow listing if the bucket policy permits `s3:ListBucket`. Obscurity only.
- The images are scraped third-party memes, so "protection" is more about not publicly republishing the collection staging-side than about secrecy. Worth stating that plainly in the spec.

**Options for images:**
1. Accept public S3 images (simplest; GUI/manifest gated, bytes not). Fine for staging if the owner agrees.
2. Worker proxy with presigned S3 URLs: bucket private, Worker signs short-lived GET URLs (SigV4 via `aws4fetch`, needs AWS keys as Worker secrets) or streams objects through the Worker. Gate holds for images; adds secrets and latency; images become a Worker route like `/img/<key>` that checks SSO first.
3. Move images to R2 (private bucket, binding `IMAGES`, Worker serves `/img/*` after the SSO check). Cleanest with Cloudflare, no AWS credentials in the Worker, no egress fees. Cost: the crawler must upload to R2 (S3-compatible API, so the Go AWS SDK works with a custom endpoint) or a sync step copies S3 -> R2.
4. Cookie caveat: `<img>` requests to S3 carry no `cbl_sso`, so only options 2/3 (same-origin or Worker-signed) can gate them.

## 3. Minimal Worker shape (sketch, not committed code)

Copy `sso.ts` byte-identical into `src/sso.ts` (the repos' tests check a shared fixture; also copy `sso-fixture.ts`/the test pattern). The REAtrends gate is the template, minus Hono.

```ts
import { parseSsoPublicKeys, verifySso, SSO_COOKIE, SSO_ISSUER } from './sso'

interface Env { ASSETS: Fetcher; SITE_ENV: string; SSO_PUBLIC_KEYS: string; ALLOWED_EMAILS: string }

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)
    if (url.pathname === '/health') return Response.json({ ok: true }, { headers: { 'Cache-Control': 'no-store' } })

    const keys = env.SITE_ENV === 'production' ? [] : parseSsoPublicKeys(env.SSO_PUBLIC_KEYS)
    const token = readCookie(request.headers.get('Cookie'), SSO_COOKIE)        // small parser, or hono/cookie
    const email = await verifySso(token, keys, Math.floor(Date.now() / 1000))

    if (!email) {                                                              // no/invalid cookie
      url.searchParams.delete('sso')
      const login = `${SSO_ISSUER}/auth/login?return=${encodeURIComponent(url.toString())}`
      if (request.method !== 'GET' || new URL(request.url).searchParams.get('sso') === '1')
        return signInPage(login)                                               // 401, loop guard
      return Response.redirect(login, 302)
    }
    if (!allowed(env.ALLOWED_EMAILS, email)) return notAuthorized(email)       // 403

    const res = await env.ASSETS.fetch(request)
    return withCache(new URL(request.url).pathname, res)
  },
}
```
Notes:
- Keep the same semantics as `ssoGate.tsx`: 403 for a verified but non-allowlisted email, 401 sign-in page for non-GET or `sso=1`, redirect otherwise. Note that ssoGate deletes `sso` before building `return`; the portal appends `sso=1` on the way back (see the spec for exact behavior).
- Fail closed: if `SSO_PUBLIC_KEYS` parses to zero keys (misconfig), `verifySso` returns null for everyone, so nothing is served. Do not add an "if no keys, open" branch.
- Redirecting asset requests (js/css/json) to the portal is useless in the browser; for non-navigation requests (`Sec-Fetch-Dest != document`) prefer a plain 401. Since the user hits `/` first and gets the cookie, subresources carry it (same-site on `cloudburstlabs.workers.dev`).
- `ALLOWED_EMAILS` = `tusculum99@gmail.com`, exact-match, case-insensitive; do not add domain allowlisting.
- Responses from `env.ASSETS.fetch` have immutable headers in some cases; clone with `new Response(res.body, res)` before setting headers.

**Cache headers.** Asset default is `public, max-age=0, must-revalidate`, which is wrong behind auth: a shared cache could serve a gated page to a non-authenticated client (Worker runs first, so Cloudflare's cache won't skip the Worker, but still do not rely on defaults).
- HTML, `memes.json`, `memes.local.json`: `private, no-store` (matches the REAtrends admin convention).
- JS/CSS: `private, no-cache` (revalidate, ETag from assets) or `private, max-age=300`. Never `public`. Filenames aren't hashed, so avoid long `max-age`.
- `/health`: `no-store`, public, no auth.
- `_headers` is applied to static-asset responses and its docs say headers are not applied to Worker-generated responses; whether it applies to `ASSETS.fetch` output via the Worker was not stated (verify), so set headers in the Worker.
- Add `X-Robots-Tag: noindex` on staging (same as REAnews).

## 4. Repo layout options

**A. `worker/` subfolder in giggles.ai (recommended).**
```
giggles.ai/
  crawler/ (Go, untouched)   gui/ (source of truth)   scripts/crawl.sh
  worker/  package.json  wrangler.toml  tsconfig.json  src/index.ts  src/sso.ts  test/  public/ (build output, gitignored)
```
- `worker/package.json` is separate from `gui/package.json` (which has no deps, `node --test`). No npm workspaces needed; Go and Node don't interact.
- Build step `worker/scripts/prepare-assets.sh`: copy `gui/{index.html,app.js,memeLogic.js,styles.css}` + chosen manifest to `worker/public/`, excluding tests and `package.json`. Avoids the `.assetsignore` question and keeps `memes.local.json` out.
- Manifest handoff: `scripts/crawl.sh` already regenerates `gui/memes.local.json`; extend (or add a sibling) to pass `-manifest-base-url https://<bucket>.s3.<region>.amazonaws.com/memes` and write `worker/public/memes.json`, then `wrangler deploy`. Manifest updates then need a redeploy (fine at crawler cadence). Alternative: manifest in R2/KV served via a Worker route (no redeploy), later.
- Matches the REAtrends/cloudburstlabs pattern (wrangler 4.x, vitest + `@cloudflare/vitest-plugin`).
- Root `index.html`/`style.css` landing page: either a second asset set or leave off staging. The GUI expects to be served at its own root, but the committed README/crawler-manifest base `/crawler/found-images` assumes repo-root serving; the Worker copy should place the GUI at `/`.

**B. Separate repo.** Cleaner CI, but the GUI and manifest generator drift apart; copying `sso.ts` already duplicates enough. Not worth it for one small static app.

**C. Worker `directory = "../gui"`.** Fewest files but needs `.assetsignore`, ships `memes.local.json` if present, and ties Worker deploys to the dev folder. Not recommended.

## 5. Production

- Same pattern as REAnews: staging at the top level of `wrangler.toml`, `[env.production]` repeats every var (Wrangler doesn't inherit `vars`); a `production` test loads it.
- `SSO_PUBLIC_KEYS` is empty in production (the SSO helpers return `[]` when `SITE_ENV=production`), so the staging cookie never opens production.
- Root custom domain `giggles.ai` via `routes = [{ pattern = "giggles.ai", custom_domain = true }]` once the zone is Active (REAnews keeps its `routes` commented out until then). Needs DNS/zone moved to Cloudflare.
- Then decide: (a) public: drop the gate in production (`SITE_ENV=production` skips SSO and serves assets with `public, max-age`), index it, drop noindex; or (b) its own sign-in (Google OAuth + session like REAnews admin; significantly more code). The README says "community-driven"/open source, so public is likely the destiny; the gate is a staging-only concern.
- Public production + public S3 images is consistent; private production images would need R2/proxy (section 2).
- Scraped third-party memes on a public ad-free or ad-supported site raise a copyright/takedown question the owner should consider before going public (source attribution, takedown path).

## 6. Open questions for the owner

1. Images: accept public S3 for staging, or proxy via Worker/R2 so the gate covers bytes too? (Also confirm the bucket is actually public-read today and not listable.)
2. Which manifest does staging serve: the 3-item `memes.json` sample, or a real S3-based one generated by the crawler? How is it refreshed (redeploy vs. R2/KV-served)?
3. Is the root landing page (`index.html`) part of this Worker, or only the GUI at `/`?
4. Layout: `worker/` inside giggles.ai (recommended) vs. another repo.
5. Allowlist: only `tusculum99@gmail.com`?
6. Will `giggles.ai` go public in production, or stay behind its own sign-in? Is the `giggles.ai` zone on Cloudflare yet?
7. Is the 4-hour/other crawl cadence (and who runs it, the Mac) going to push deploys, or should the Worker read the manifest at runtime?
8. Copyright/takedown stance for scraped memes before any public launch.
9. Non-navigation requests when unauthenticated: 401 vs. redirect (sketch proposes 401 for subresources).

## 7. Decisions (2026-10-16)

Owner answers to section 6. Where a decision changes the notes above, it overrides them.

1. **Hosting:** a `/giggles/*` route inside the cloudburstlabs `web` Worker, not a separate Worker. Protected by the portal's existing Google session, so section 3's `sso.ts`/`cbl_sso` sketch does not apply. GUI source stays in `gui/` here and is copied into cloudburstlabs at build time, so it can be extracted into its own Worker when `giggles.ai` goes public.
2. **Allowlist:** giggles gets its own narrower check, not the portal's (the portal admits all of `cloudburstlabs.com`). Allowed: `tusculum99@gmail.com`, `matt@cloudburstlabs.com`. Exact match, case-insensitive.
3. **Images:** private R2 bucket, served at `/giggles/img/*` after the session + allowlist check. Replaces "public S3": no S3 bucket exists; crawls have only run with `-local-dir`. The crawl script syncs `crawler/found-images/` to R2; the Go crawler stays unchanged.
4. **Manifest:** a real crawler-generated manifest with the `/giggles/img` base URL, stored in R2 next to the images so crawls never redeploy the portal. This replaces the earlier "redeploy" answer. Note: the GUI probes `memes.local.json` first, so the route must answer it with a 404 rather than a 401 or a redirect. Better: have the build set `window.GIGGLES_MANIFEST_URL` so there is only one fetch.
5. **Landing page:** not on staging. Only the GUI, at `/giggles/`.
6. **Unauthenticated subresources:** plain 401. Only page navigations redirect to sign-in.
7. **Production:** public at `giggles.ai`, with source attribution on each meme and a published takedown process before launch. The domain is not live on Cloudflare yet.
8. **Subpath test (local, 2026-10-16):** served the GUI with a crawler-generated `/giggles/img` manifest and loaded it in Chromium.
   - `/giggles/`: works. All assets and all 5 test images loaded with 200.
   - `/giggles` (no slash, served without redirecting): broken. `styles.css` and `app.js` resolve against `/` and return 404, so the page renders blank.
   - **Requirement:** the route must 301 `/giggles` to `/giggles/`. All GUI references are relative (`styles.css`, `app.js`, `./memes*.json`, module imports), so nothing else needs to change. Image URLs in the manifest are absolute (`/giggles/img/...`), so they work either way.
