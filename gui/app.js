// Giggles.ai Meme Search Application
// This is a client-side application that searches and displays AI memes

class MemeSearchApp {
    constructor() {
        this.currentPage = 1;
        this.itemsPerPage = 12;
        this.currentMemes = [];
        this.filteredMemes = [];
        this.sortOrder = 'newest';
        
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

    async loadInitialMemes() {
        // In a real implementation, this would fetch from an API
        // For now, we'll use mock data or fetch from S3 if configured
        this.showLoading();
        
        try {
            // TODO: Replace with actual API endpoint
            // const response = await fetch('/api/memes');
            // const data = await response.json();
            
            // Mock data for demonstration
            this.currentMemes = this.getMockMemes();
            this.applySortAndFilter();
        } catch (error) {
            this.showError('Failed to load memes. Please try again later.');
            console.error('Error loading memes:', error);
        } finally {
            this.hideLoading();
        }
    }

    getMockMemes() {
        // Mock data - replace with actual API call
        // In production, this would fetch from your backend API that queries S3
        return [
            {
                id: 1,
                url: 'https://via.placeholder.com/400/4a90e2/ffffff?text=AI+Meme+1',
                title: 'ChatGPT Meme',
                source: 'Reddit',
                uploadedAt: new Date('2024-01-15')
            },
            {
                id: 2,
                url: 'https://via.placeholder.com/400/357abd/ffffff?text=AI+Meme+2',
                title: 'DALL-E Art',
                source: 'Imgur',
                uploadedAt: new Date('2024-01-14')
            },
            {
                id: 3,
                url: 'https://via.placeholder.com/400/6b6c76/ffffff?text=AI+Meme+3',
                title: 'Robot Humor',
                source: '9GAG',
                uploadedAt: new Date('2024-01-13')
            }
        ];
    }

    performSearch() {
        const searchTerm = this.searchInput.value.trim().toLowerCase();
        this.currentPage = 1;
        this.applySortAndFilter(searchTerm);
    }

    applySortAndFilter(searchTerm = '') {
        // Filter memes by search term
        this.filteredMemes = this.currentMemes.filter(meme => {
            if (!searchTerm) return true;
            const searchableText = `${meme.title} ${meme.source}`.toLowerCase();
            return searchableText.includes(searchTerm);
        });

        // Sort memes
        this.filteredMemes.sort((a, b) => {
            switch (this.sortOrder) {
                case 'newest':
                    return b.uploadedAt - a.uploadedAt;
                case 'oldest':
                    return a.uploadedAt - b.uploadedAt;
                case 'random':
                    return Math.random() - 0.5;
                default:
                    return 0;
            }
        });

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

        // Calculate pagination
        const startIndex = (this.currentPage - 1) * this.itemsPerPage;
        const endIndex = startIndex + this.itemsPerPage;
        const pageMemes = this.filteredMemes.slice(startIndex, endIndex);

        // Render meme cards
        pageMemes.forEach(meme => {
            const card = this.createMemeCard(meme);
            this.memeGrid.appendChild(card);
        });

        // Update pagination
        this.updatePagination();
    }

    createMemeCard(meme) {
        const card = document.createElement('div');
        card.className = 'meme-card';
        card.addEventListener('click', () => this.openModal(meme));

        card.innerHTML = `
            <img src="${meme.url}" alt="${meme.title}" class="meme-image" loading="lazy">
            <div class="meme-info">
                <div class="meme-title">${this.escapeHtml(meme.title)}</div>
                <div class="meme-source">Source: ${this.escapeHtml(meme.source)}</div>
            </div>
        `;

        // Handle image load errors
        const img = card.querySelector('.meme-image');
        img.addEventListener('error', function() {
            this.src = 'https://via.placeholder.com/400/f5f5f5/999999?text=Image+Not+Available';
            this.alt = 'Image not available';
        });

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
        const totalPages = Math.ceil(this.filteredMemes.length / this.itemsPerPage);
        
        this.pageInfo.textContent = `Page ${this.currentPage} of ${totalPages || 1}`;
        
        this.prevButton.disabled = this.currentPage === 1;
        this.nextButton.disabled = this.currentPage >= totalPages || totalPages === 0;
    }

    goToPreviousPage() {
        if (this.currentPage > 1) {
            this.currentPage--;
            this.renderMemes();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }
    }

    goToNextPage() {
        const totalPages = Math.ceil(this.filteredMemes.length / this.itemsPerPage);
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

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Method to fetch memes from API (to be implemented with backend)
    async fetchMemesFromAPI(searchTerm = '') {
        try {
            // TODO: Replace with actual API endpoint
            const apiUrl = '/api/memes';
            const params = new URLSearchParams();
            if (searchTerm) {
                params.append('search', searchTerm);
            }
            params.append('page', this.currentPage);
            params.append('limit', this.itemsPerPage);
            params.append('sort', this.sortOrder);

            const response = await fetch(`${apiUrl}?${params}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data.memes || [];
        } catch (error) {
            console.error('Error fetching memes:', error);
            throw error;
        }
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new MemeSearchApp();
});

