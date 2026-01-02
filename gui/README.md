# Giggles.ai GUI

A modern, responsive web interface for searching and viewing AI memes collected by the crawler.

## Features

- 🔍 **Search Functionality**: Search memes by keywords
- 🎨 **Modern UI**: Clean, responsive design with blue and grey color scheme
- 📱 **Mobile Friendly**: Fully responsive layout
- 🖼️ **Image Modal**: Click any meme to view in full size
- 🔄 **Sorting**: Sort by newest, oldest, or random
- 📄 **Pagination**: Navigate through multiple pages of results

## Structure

```
gui/
├── index.html      # Main HTML structure
├── styles.css      # All styling and responsive design
├── app.js          # JavaScript application logic
└── README.md       # This file
```

## Usage

### Standalone

Simply open `index.html` in a web browser:

```bash
open gui/index.html  # macOS
xdg-open gui/index.html  # Linux
start gui/index.html  # Windows
```

### Integrated with Backend

The GUI is designed to work with a backend API. To connect it:

1. Update the `fetchMemesFromAPI()` method in `app.js` with your API endpoint
2. Ensure your API returns data in the expected format:

```json
{
  "memes": [
    {
      "id": 1,
      "url": "https://s3.amazonaws.com/bucket/memes/image.jpg",
      "title": "Meme Title",
      "source": "Reddit",
      "uploadedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 12
}
```

## Customization

### Colors

Edit CSS variables in `styles.css`:

```css
:root {
    --primary-blue: #4a90e2;
    --secondary-blue: #357abd;
    --light-grey: #f5f5f5;
    /* ... */
}
```

### Items Per Page

Change `itemsPerPage` in `app.js`:

```javascript
this.itemsPerPage = 12; // Change to desired number
```

### API Integration

Replace the mock data in `getMockMemes()` with actual API calls:

```javascript
async loadInitialMemes() {
    this.showLoading();
    try {
        const response = await fetch('/api/memes');
        const data = await response.json();
        this.currentMemes = data.memes;
        this.applySortAndFilter();
    } catch (error) {
        this.showError('Failed to load memes.');
    } finally {
        this.hideLoading();
    }
}
```

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Future Enhancements

- [ ] Infinite scroll instead of pagination
- [ ] Image lazy loading optimization
- [ ] Share functionality
- [ ] Download memes
- [ ] Filter by source
- [ ] Tag system
- [ ] Favorites/bookmarks

