import { test } from 'node:test';
import assert from 'node:assert/strict';

import {
    escapeHtml,
    normalizeMeme,
    parseManifest,
    filterMemes,
    sortMemes,
    paginate,
    getPaginationInfo,
} from './memeLogic.js';

function meme(id, title, source, dateStr) {
    return normalizeMeme({ id, url: `http://x/${id}.jpg`, title, source, uploadedAt: dateStr });
}

test('escapeHtml escapes the dangerous characters', () => {
    assert.equal(escapeHtml(`<img src="x" onerror='alert(1)'>&`),
        '&lt;img src=&quot;x&quot; onerror=&#39;alert(1)&#39;&gt;&amp;');
    assert.equal(escapeHtml(''), '');
    assert.equal(escapeHtml(null), '');
});

test('parseManifest accepts both array and {memes:[]} shapes', () => {
    const arr = parseManifest([{ id: 1, title: 'a' }]);
    assert.equal(arr.length, 1);
    assert.equal(arr[0].title, 'a');

    const obj = parseManifest({ memes: [{ id: 2, title: 'b' }] });
    assert.equal(obj.length, 1);
    assert.equal(obj[0].title, 'b');

    assert.deepEqual(parseManifest(null), []);
    assert.deepEqual(parseManifest({}), []);
});

test('normalizeMeme fills defaults and coerces the date', () => {
    const m = normalizeMeme({ url: 'u' }, 7);
    assert.equal(m.id, 7);
    assert.equal(m.title, 'Untitled');
    assert.equal(m.source, 'Unknown');
    assert.ok(m.uploadedAt instanceof Date);
});

test('filterMemes matches title and source, empty term returns all', () => {
    const memes = [
        meme(1, 'ChatGPT meme', 'Reddit', '2024-01-03'),
        meme(2, 'Robot art', 'Imgur', '2024-01-02'),
        meme(3, 'Cat photo', 'Reddit', '2024-01-01'),
    ];
    assert.deepEqual(filterMemes(memes, 'chatgpt').map((m) => m.id), [1]);
    assert.deepEqual(filterMemes(memes, 'reddit').map((m) => m.id), [1, 3]); // matches source
    assert.deepEqual(filterMemes(memes, '').map((m) => m.id), [1, 2, 3]);
    assert.deepEqual(filterMemes(memes, '   ').map((m) => m.id), [1, 2, 3]);
});

test('sortMemes orders by date and does not mutate input', () => {
    const memes = [
        meme(1, 'a', 's', '2024-01-01'),
        meme(2, 'b', 's', '2024-01-03'),
        meme(3, 'c', 's', '2024-01-02'),
    ];
    const original = memes.map((m) => m.id);

    assert.deepEqual(sortMemes(memes, 'newest').map((m) => m.id), [2, 3, 1]);
    assert.deepEqual(sortMemes(memes, 'oldest').map((m) => m.id), [1, 3, 2]);
    assert.deepEqual(memes.map((m) => m.id), original, 'input not mutated');
});

test('sortMemes random returns a permutation (seeded rng)', () => {
    const memes = [meme(1, 'a', 's', '2024-01-01'), meme(2, 'b', 's', '2024-01-02'), meme(3, 'c', 's', '2024-01-03')];
    // Deterministic rng so the test is stable.
    let i = 0;
    const seq = [0.1, 0.9, 0.5, 0.3];
    const rng = () => seq[i++ % seq.length];
    const out = sortMemes(memes, 'random', rng);
    assert.equal(out.length, 3);
    assert.deepEqual(out.map((m) => m.id).sort(), [1, 2, 3]); // same set, possibly reordered
});

test('paginate returns the correct slice', () => {
    const items = Array.from({ length: 25 }, (_, i) => i);
    assert.deepEqual(paginate(items, 1, 12), items.slice(0, 12));
    assert.deepEqual(paginate(items, 2, 12), items.slice(12, 24));
    assert.deepEqual(paginate(items, 3, 12), [24]);
    assert.deepEqual(paginate(items, 4, 12), []);
});

test('getPaginationInfo computes pages and button state', () => {
    const first = getPaginationInfo(25, 1, 12);
    assert.equal(first.totalPages, 3);
    assert.equal(first.hasPrev, false);
    assert.equal(first.hasNext, true);
    assert.equal(first.label, 'Page 1 of 3');

    const last = getPaginationInfo(25, 3, 12);
    assert.equal(last.hasPrev, true);
    assert.equal(last.hasNext, false);

    const empty = getPaginationInfo(0, 1, 12);
    assert.equal(empty.totalPages, 0);
    assert.equal(empty.hasNext, false);
    assert.equal(empty.label, 'Page 1 of 1');
});
