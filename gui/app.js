// Giggles.ai Meme Search Application
// Client-side app that loads a meme manifest (JSON) and lets the user search,
// sort, filter, and page through the results. The manifest is produced by the
// crawler (see `crawler -gen-manifest`) and served alongside these files, or
// from S3. Pure logic lives in memeLogic.js so it can be unit-tested.

import {
    parseManifest,
    filterMemes,
    sortMemes,
    paginate,
    getPaginationInfo,
} from './memeLogic.js';

// Where to load memes from.
//
// Precedence:
//   1. window.GIGGLES_MANIFEST_URL — explicit override (e.g. an S3 URL). When
//      set, ONLY this URL is used; no fallback.
//   2. ./memes.local.json — the crawler-generated manifest (gitignored). Present
//      when you've actually crawled images locally; preferred so you see them.
//   3. ./memes.json — the committed, portable sample manifest, so a fresh
//      checkout still renders something even before any crawl has run.
const MANIFEST_URLS = window.GIGGLES_MANIFEST_URL
    ? [window.GIGGLES_MANIFEST_URL]
    : ['./memes.local.json', './memes.json'];

class MemeSearchApp {
    constructor(manifestUrls = MANIFEST_URLS) {
        this.manifestUrls = Array.isArray(manifestUrls) ? manifestUrls : [manifestUrls];
        this.currentPage = 1;
        this.itemsPerPage = 12;
        this.currentMemes = [];
        this.filteredMemes = [];
        this.sortOrder = 'newest';
        this.searchTerm = '';

        this.initializeElements();
        this.attachEventListeners();
        this.loadInitialMemes();
    }

    initializeElements() {
        this.searchInput = document.getElementById('searchInput');
        this.searchButton = document.getElementById('searchButton');
        this.sortSelect = document.getElementById('sortSelect');
        this.loadingIndicator = document.getElementById('loadingIndicator');
        this.errorMessage = document.getElementById('errorMessage');
        this.memeGrid = document.getElementById('memeGrid');
        this.noResults = document.getElementById('noResults');
        this.pagination = document.getElementById('pagination');
        this.prevButton = document.getElementById('prevButton');
        this.nextButton = document.getElementById('nextButton');
        this.pageInfo = document.getElementById('pageInfo');
        this.imageModal = document.getElementById('imageModal');
        this.modalImage = document.getElementById('modalImage');
        this.modalSource = document.getElementById('modalSource');
        this.modalClose = document.querySelector('.modal-close');
    }

    attachEventListeners() {
        // Search functionality
        this.searchButton.addEventListener('click', () => this.performSearch());
        this.searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.performSearch();
            }
        });

        // Sort functionality
        this.sortSelect.addEventListener('change', (e) => {
            this.sortOrder = e.target.value;
            this.currentPage = 1;
            this.applySortAndFilter();
        });

        // Pagination
        this.prevButton.addEventListener('click', () => this.goToPreviousPage());
        this.nextButton.addEventListener('click', () => this.goToNextPage());

        // Modal
        this.modalClose.addEventListener('click', () => this.closeModal());
        this.imageModal.addEventListener('click', (e) => {
            if (e.target === this.imageModal) {
                this.closeModal();
            }
        });

        // Close modal on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && !this.imageModal.classList.contains('hidden')) {
                this.closeModal();
            }
        });
    }

    // Try each candidate manifest URL in order, returning the parsed JSON of the
    // first one that loads. A missing (e.g. 404) or invalid manifest falls
    // through to the next candidate; only when all fail do we throw.
    async fetchManifest() {
        const errors = [];
        for (const url of this.manifestUrls) {
            try {
                const response = await fetch(url, { cache: 'no-cache' });
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}`);
                }
                return await response.json();
            } catch (error) {
                errors.push(`${url} (${error.message})`);
            }
        }
        throw new Error(`No manifest could be loaded. Tried: ${errors.join(', ')}`);
    }

    async loadInitialMemes() {
        this.showLoading();

        try {
            const data = await this.fetchManifest();
            this.currentMemes = parseManifest(data);

            if (this.currentMemes.length === 0) {
                this.showError('No memes are available yet. Run the crawler to collect some!');
            }
            this.applySortAndFilter();
        } catch (error) {
            this.currentMemes = [];
            this.filteredMemes = [];
            this.renderMemes();
            this.showError(
                'Could not load memes. ' +
                'Make sure a memes.json manifest is served here (see docs/GUI.md).'
            );
            console.error('Error loading memes:', error);
        } finally {
            this.hideLoading();
        }
    }

    performSearch() {
        this.searchTerm = this.searchInput.value.trim().toLowerCase();
        this.currentPage = 1;
        this.applySortAndFilter();
    }

    applySortAndFilter() {
        // Filter then sort using the pure logic helpers. Reading the search term
        // from state keeps it applied when only the sort order changes.
        const filtered = filterMemes(this.currentMemes, this.searchTerm);
        this.filteredMemes = sortMemes(filtered, this.sortOrder);
        this.renderMemes();
    }

    renderMemes() {
        this.memeGrid.innerHTML = '';

        if (this.filteredMemes.length === 0) {
            this.noResults.classList.remove('hidden');
            this.pagination.classList.add('hidden');
            return;
        }

        this.noResults.classList.add('hidden');
        this.pagination.classList.remove('hidden');

        const pageMemes = paginate(this.filteredMemes, this.currentPage, this.itemsPerPage);
        pageMemes.forEach((meme) => {
            this.memeGrid.appendChild(this.createMemeCard(meme));
        });

        this.updatePagination();
    }

    createMemeCard(meme) {
        const card = document.createElement('div');
        card.className = 'meme-card';
        card.addEventListener('click', () => this.openModal(meme));

        // Build the DOM imperatively rather than via innerHTML so that meme.url
        // and other fields can't break out of the attribute context and inject
        // markup. Assigning to .src treats the value strictly as a URL.
        const img = document.createElement('img');
        img.className = 'meme-image';
        img.loading = 'lazy';
        img.alt = meme.title;
        img.src = meme.url;
        img.addEventListener('error', function () {
            this.src = 'https://via.placeholder.com/400/f5f5f5/999999?text=Image+Not+Available';
            this.alt = 'Image not available';
        });

        const info = document.createElement('div');
        info.className = 'meme-info';

        const title = document.createElement('div');
        title.className = 'meme-title';
        title.textContent = meme.title;

        const source = document.createElement('div');
        source.className = 'meme-source';
        source.textContent = `Source: ${meme.source}`;

        info.appendChild(title);
        info.appendChild(source);
        card.appendChild(img);
        card.appendChild(info);

        return card;
    }

    openModal(meme) {
        this.modalImage.src = meme.url;
        this.modalImage.alt = meme.title;
        this.modalSource.textContent = `Source: ${meme.source} | ${meme.uploadedAt.toLocaleDateString()}`;
        this.imageModal.classList.remove('hidden');
        document.body.style.overflow = 'hidden'; // Prevent background scrolling
    }

    closeModal() {
        this.imageModal.classList.add('hidden');
        document.body.style.overflow = ''; // Restore scrolling
    }

    updatePagination() {
        const info = getPaginationInfo(
            this.filteredMemes.length,
            this.currentPage,
            this.itemsPerPage
        );
        this.pageInfo.textContent = info.label;
        this.prevButton.disabled = !info.hasPrev;
        this.nextButton.disabled = !info.hasNext;
    }

    goToPreviousPage() {
        if (this.currentPage > 1) {
            this.currentPage--;
            this.renderMemes();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }
    }

    goToNextPage() {
        const { totalPages } = getPaginationInfo(
            this.filteredMemes.length,
            this.currentPage,
            this.itemsPerPage
        );
        if (this.currentPage < totalPages) {
            this.currentPage++;
            this.renderMemes();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }
    }

    showLoading() {
        this.loadingIndicator.classList.remove('hidden');
        this.errorMessage.classList.add('hidden');
    }

    hideLoading() {
        this.loadingIndicator.classList.add('hidden');
    }

    showError(message) {
        this.errorMessage.textContent = message;
        this.errorMessage.classList.remove('hidden');
    }
}

// Initialize the application when the DOM is ready.
document.addEventListener('DOMContentLoaded', () => {
    new MemeSearchApp();
});

export { MemeSearchApp };
