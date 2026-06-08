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
├── index.html         # Main HTML structure
├── styles.css         # All styling and responsive design
├── app.js             # App: DOM wiring + data fetching (ES module)
├── memeLogic.js       # Pure logic: filter/sort/paginate/escape (ES module)
├── memeLogic.test.js  # Unit tests (Node built-in test runner)
├── memes.json         # Sample data manifest the app loads
├── package.json       # `npm test` script (no runtime dependencies)
└── README.md          # This file
```

## Usage

### Running locally

The app fetches its data with `fetch()`, which browsers block on the `file://`
protocol, so serve the directory over HTTP rather than opening the file
directly:

```bash
python -m http.server 8000   # then open http://localhost:8000/gui/
```

It loads `gui/memes.json` (a sample manifest ships in the repo) on startup.

## Data source

The GUI is backend-free: it loads a JSON **manifest** describing the available
memes. By default it fetches `./memes.json` relative to the page. To point it at
another location (e.g. an S3/CloudFront URL), set a global before the script
loads:

```html
<script>window.GIGGLES_MANIFEST_URL = 'https://cdn.example.com/memes.json';</script>
```

The manifest is either a bare array or an object with a `memes` array; each
entry has `id`, `url`, `title`, `source`, and `uploadedAt`:

```json
{
  "generatedAt": "2024-01-15T00:00:00Z",
  "memes": [
    {
      "id": "image.jpg",
      "url": "https://bucket.s3.amazonaws.com/memes/image.jpg",
      "title": "Meme Title",
      "source": "crawler",
      "uploadedAt": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### Generating the manifest from crawled images

The crawler can produce a manifest from a directory of downloaded images:

```bash
# Index ./found-images and write found-images/memes.json
cd crawler
go run . -gen-manifest -manifest-dir found-images -manifest-out ../gui/memes.json

# With an S3/CDN prefix so the GUI loads images from S3
go run . -gen-manifest -manifest-dir found-images \
  -manifest-base-url "https://giggles-memes.s3.us-east-1.amazonaws.com/memes" \
  -manifest-out ../gui/memes.json
```

## Testing

Pure logic lives in `memeLogic.js` and is unit-tested with Node's built-in test
runner (no dependencies to install):

```bash
cd gui
npm test        # or: node --test
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

### Changing the data source

See [Data source](#data-source) above — set `window.GIGGLES_MANIFEST_URL` or
replace `gui/memes.json`. No code changes are needed to repoint the app.

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


