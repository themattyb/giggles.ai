// Pure, DOM-free logic for the meme browser. Kept separate from app.js so it
// can be unit-tested under Node without a browser/jsdom. Imported by app.js as
// an ES module.

// Escape a string for safe interpolation into HTML. Pure (no DOM) so it is
// usable and testable outside the browser.
export function escapeHtml(text) {
    const map = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };
    return String(text ?? '').replace(/[&<>"']/g, (ch) => map[ch]);
}

// Normalize a raw manifest entry into the shape the app expects. uploadedAt is
// coerced to a Date; missing fields get safe defaults.
export function normalizeMeme(raw, index = 0) {
    return {
        id: raw.id ?? index,
        url: raw.url ?? '',
        title: raw.title ?? 'Untitled',
        source: raw.source ?? 'Unknown',
        uploadedAt: raw.uploadedAt ? new Date(raw.uploadedAt) : new Date(0),
    };
}

// Parse a fetched manifest. Accepts either a bare array or an object with a
// `memes` array. Returns an array of normalized memes.
export function parseManifest(data) {
    const list = Array.isArray(data) ? data : Array.isArray(data?.memes) ? data.memes : [];
    return list.map((raw, i) => normalizeMeme(raw, i));
}

// Filter memes by a search term matched against title and source.
// An empty/whitespace term returns all memes.
export function filterMemes(memes, term) {
    const needle = (term ?? '').trim().toLowerCase();
    if (!needle) return [...memes];
    return memes.filter((meme) => {
        const haystack = `${meme.title} ${meme.source}`.toLowerCase();
        return haystack.includes(needle);
    });
}

// Return a new array sorted by the given order: 'newest', 'oldest', or
// 'random'. `rng` is injectable for deterministic tests (defaults to
// Math.random). Never mutates the input.
export function sortMemes(memes, order, rng = Math.random) {
    const out = [...memes];
    switch (order) {
        case 'newest':
            return out.sort((a, b) => b.uploadedAt.getTime() - a.uploadedAt.getTime());
        case 'oldest':
            return out.sort((a, b) => a.uploadedAt.getTime() - b.uploadedAt.getTime());
        case 'random':
            // Fisher–Yates shuffle.
            for (let i = out.length - 1; i > 0; i--) {
                const j = Math.floor(rng() * (i + 1));
                [out[i], out[j]] = [out[j], out[i]];
            }
            return out;
        default:
            return out;
    }
}

// Return the slice of memes for a 1-based page.
export function paginate(memes, page, perPage) {
    const start = (page - 1) * perPage;
    return memes.slice(start, start + perPage);
}

// Compute pagination UI state for a result set.
export function getPaginationInfo(total, page, perPage) {
    const totalPages = Math.ceil(total / perPage);
    return {
        totalPages,
        hasPrev: page > 1,
        hasNext: page < totalPages,
        label: `Page ${page} of ${totalPages || 1}`,
    };
}
