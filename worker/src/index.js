// giggles.cloudburstlabs.workers.dev: a short hostname for giggles staging.
// It serves nothing itself. Every request is redirected to the cloudburstlabs
// portal's /giggles/ route, which holds the content behind Google sign-in and
// the giggles allowlist (see docs/WORKER-HOSTING.md §7). Keep it that way:
// adding content here would put it outside the sign-in wall.
const TARGET = 'https://web.cloudburstlabs.workers.dev/giggles'

export default {
  fetch(request) {
    const url = new URL(request.url)
    // 302, not 301: staging's home may move, and browsers cache a 301 forever.
    return new Response(null, {
      status: 302,
      headers: {
        Location: `${TARGET}${url.pathname}${url.search}`,
        'Cache-Control': 'no-store',
        'X-Robots-Tag': 'noindex',
      },
    })
  },
}
