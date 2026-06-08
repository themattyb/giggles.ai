package crawler

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/giggles-ai/crawler/internal/s3"
	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
)

// Response size caps to guard against memory exhaustion from oversized or
// malicious responses.
const (
	maxHTMLSize   = 10 << 20 // 10 MB per HTML page
	maxImageSize  = 50 << 20 // 50 MB per image
	maxRobotsSize = 1 << 20  // 1 MB per robots.txt
)

// Config holds crawler configuration
type Config struct {
	Workers            int
	Delay              time.Duration
	MaxPages           int
	StartURLs          []string // Multiple starting URLs
	UserAgent          string
	S3Client           *s3.Client
	LocalDir           string // Local directory to save images
	InsecureSkipVerify bool   // Skip TLS certificate verification (use only for testing)
}

// Stats holds crawler statistics
type Stats struct {
	PagesCrawled     int
	ImagesFound      int
	ImagesDownloaded int
	ImagesUploaded   int
	Errors           int
	Duration         time.Duration
}

// Crawler represents a web crawler
type Crawler struct {
	config            Config
	client            *http.Client
	visited           map[string]bool
	visitedMu         sync.RWMutex
	robotsCache       map[string]*robotstxt.RobotsData
	robotsMu          sync.RWMutex
	queue             chan string
	stats             Stats
	statsMu           sync.RWMutex
	wg                sync.WaitGroup
	startTime         time.Time
	queueClosed       bool
	queueMu           sync.Mutex
	discoveredDomains map[string]bool // Track discovered domains
	domainsMu         sync.RWMutex
}

// New creates a new crawler instance
func New(config Config) (*Crawler, error) {
	// Validate start URLs
	if len(config.StartURLs) == 0 {
		return nil, fmt.Errorf("at least one start URL is required")
	}
	for _, startURL := range config.StartURLs {
		_, err := url.Parse(startURL)
		if err != nil {
			return nil, fmt.Errorf("invalid start URL %s: %w", startURL, err)
		}
	}

	// Create local directory if specified
	if config.LocalDir != "" {
		err := os.MkdirAll(config.LocalDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create local directory %s: %w", config.LocalDir, err)
		}
		log.Printf("Local image directory created: %s", config.LocalDir)
	}

	transport := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}
	if config.InsecureSkipVerify {
		log.Println("WARNING: TLS certificate verification is disabled")
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	c := &Crawler{
		config: config,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		visited:           make(map[string]bool),
		robotsCache:       make(map[string]*robotstxt.RobotsData),
		queue:             make(chan string, config.MaxPages*2),
		startTime:         time.Now(),
		discoveredDomains: make(map[string]bool),
	}

	// Initialize discovered domains with start URLs
	for _, startURL := range config.StartURLs {
		if domain := c.extractDomain(startURL); domain != "" {
			c.domainsMu.Lock()
			c.discoveredDomains[domain] = true
			c.domainsMu.Unlock()
		}
	}

	return c, nil
}

// Run starts the crawler
func (c *Crawler) Run() (*Stats, error) {
	// Add all start URLs to queue
	for _, startURL := range c.config.StartURLs {
		select {
		case c.queue <- startURL:
			log.Printf("Added start URL to queue: %s", startURL)
		default:
			log.Printf("Queue full, skipping URL: %s", startURL)
		}
	}

	// Start worker goroutines
	for i := 0; i < c.config.Workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	// Start coordinator to close queue when done
	go c.coordinator()

	// Wait for all workers to finish
	c.wg.Wait()
	c.closeQueue()

	// Calculate duration
	c.statsMu.Lock()
	c.stats.Duration = time.Since(c.startTime)
	c.statsMu.Unlock()

	return &c.stats, nil
}

// coordinator monitors the crawler and closes the queue when appropriate
func (c *Crawler) coordinator() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.statsMu.RLock()
		pagesCrawled := c.stats.PagesCrawled
		c.statsMu.RUnlock()

		// Close queue if we've reached max pages
		if pagesCrawled >= c.config.MaxPages {
			c.closeQueue()
			return
		}

		// Check if queue is empty and we've processed at least one page
		// This handles the case where the first page fails
		if len(c.queue) == 0 && pagesCrawled > 0 {
			// Wait a bit more to see if any new URLs are added
			time.Sleep(2 * time.Second)
			if len(c.queue) == 0 {
				c.closeQueue()
				return
			}
		}
	}
}

// closeQueue safely closes the queue
func (c *Crawler) closeQueue() {
	c.queueMu.Lock()
	defer c.queueMu.Unlock()
	if !c.queueClosed {
		close(c.queue)
		c.queueClosed = true
	}
}

// worker processes URLs from the queue
func (c *Crawler) worker(id int) {
	defer c.wg.Done()

	for urlStr := range c.queue {
		// Check if we've reached max pages
		c.statsMu.RLock()
		if c.stats.PagesCrawled >= c.config.MaxPages {
			c.statsMu.RUnlock()
			return
		}
		c.statsMu.RUnlock()

		// Atomically check-and-mark visited in a single critical section so two
		// workers can't both pass the check and process the same URL.
		c.visitedMu.Lock()
		if c.visited[urlStr] {
			c.visitedMu.Unlock()
			continue
		}
		c.visited[urlStr] = true
		c.visitedMu.Unlock()

		// Check robots.txt
		if !c.canCrawl(urlStr) {
			log.Printf("[Worker %d] Blocked by robots.txt, skipping: %s", id, urlStr)
			// Mark as visited so we don't try again, but don't count as crawled
			// Continue to next URL in queue
			continue
		}

		// Respect delay
		time.Sleep(c.config.Delay)

		// Fetch and process page
		err := c.processPage(urlStr, id)
		if err != nil {
			log.Printf("[Worker %d] Error processing %s: %v", id, urlStr, err)
			c.statsMu.Lock()
			c.stats.Errors++
			c.statsMu.Unlock()
		}

		// Update pages crawled
		c.statsMu.Lock()
		c.stats.PagesCrawled++
		c.statsMu.Unlock()

		// Check if we should stop
		c.statsMu.RLock()
		if c.stats.PagesCrawled >= c.config.MaxPages {
			c.statsMu.RUnlock()
			return
		}
		c.statsMu.RUnlock()
	}
}

// canCrawl checks if a URL can be crawled according to robots.txt
func (c *Crawler) canCrawl(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	// Check cache first
	c.robotsMu.RLock()
	robotsData, exists := c.robotsCache[robotsURL]
	c.robotsMu.RUnlock()

	if !exists {
		// Fetch robots.txt
		robotsData = c.fetchRobotsTxt(robotsURL)
		if robotsData != nil {
			c.robotsMu.Lock()
			c.robotsCache[robotsURL] = robotsData
			c.robotsMu.Unlock()
		}
	}

	if robotsData == nil {
		// If we can't fetch robots.txt, allow crawling (fail open)
		return true
	}

	// Check if our user agent can access this path
	group := robotsData.FindGroup(c.config.UserAgent)
	if group == nil {
		// No specific rules for our user agent, check default
		group = robotsData.FindGroup("*")
	}

	if group != nil {
		// Test expects the request path (optionally with query), not the full
		// URL — passing the full URL means rules like "Disallow: /private"
		// never match and robots.txt is silently ignored.
		path := parsedURL.Path
		if path == "" {
			path = "/"
		}
		if parsedURL.RawQuery != "" {
			path += "?" + parsedURL.RawQuery
		}
		return group.Test(path)
	}

	// Default to allowing if no rules found
	return true
}

// fetchRobotsTxt fetches and parses robots.txt
func (c *Crawler) fetchRobotsTxt(robotsURL string) *robotstxt.RobotsData {
	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRobotsSize))
	if err != nil {
		return nil
	}

	robotsData, err := robotstxt.FromBytes(body)
	if err != nil {
		return nil
	}

	return robotsData
}

// processPage fetches a page and extracts images and links
func (c *Crawler) processPage(urlStr string, workerID int) error {
	log.Printf("[Worker %d] Processing: %s", workerID, urlStr)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status: %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return fmt.Errorf("not HTML content: %s", contentType)
	}

	// Parse HTML (capped to guard against oversized responses)
	doc, err := html.Parse(io.LimitReader(resp.Body, maxHTMLSize))
	if err != nil {
		return err
	}

	// Extract images and links
	baseURL, _ := url.Parse(urlStr)
	images := c.extractImages(doc, baseURL)
	links := c.extractLinks(doc, baseURL)

	// Process images
	for _, imgURL := range images {
		c.statsMu.Lock()
		c.stats.ImagesFound++
		c.statsMu.Unlock()

		if c.isMemeImage(imgURL) {
			c.downloadImage(imgURL, workerID)
		}
	}

	// Process links and discover new domains
	c.statsMu.RLock()
	if c.stats.PagesCrawled < c.config.MaxPages {
		for _, link := range links {
			// Check if this is a new domain
			domain := c.extractDomain(link)
			if domain != "" {
				c.domainsMu.RLock()
				isNewDomain := !c.discoveredDomains[domain]
				c.domainsMu.RUnlock()

				if isNewDomain {
					// New domain discovered!
					c.domainsMu.Lock()
					c.discoveredDomains[domain] = true
					c.domainsMu.Unlock()

					log.Printf("[Worker %d] Discovered new domain: %s (from: %s)", workerID, domain, link)

					// Add the root URL of the new domain to the queue
					parsedLink, err := url.Parse(link)
					if err == nil {
						rootURL := fmt.Sprintf("%s://%s/", parsedLink.Scheme, parsedLink.Host)
						select {
						case c.queue <- rootURL:
							log.Printf("[Worker %d] Added root URL of new domain to queue: %s", workerID, rootURL)
						default:
							// Queue is full, skip
						}
					}
				}
			}

			// Add the link itself to the queue
			select {
			case c.queue <- link:
			default:
				// Queue is full, skip
			}
		}
	}
	c.statsMu.RUnlock()

	return nil
}

// extractImages extracts image URLs from HTML
func (c *Crawler) extractImages(doc *html.Node, baseURL *url.URL) []string {
	var images []string
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" || attr.Key == "data-src" {
					imgURL := c.resolveURL(attr.Val, baseURL)
					if imgURL != "" {
						images = append(images, imgURL)
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}

	f(doc)
	return images
}

// extractLinks extracts link URLs from HTML
func (c *Crawler) extractLinks(doc *html.Node, baseURL *url.URL) []string {
	var links []string
	visited := make(map[string]bool)
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					linkURL := c.resolveURL(attr.Val, baseURL)
					if linkURL != "" && !visited[linkURL] {
						// Only add HTTP/HTTPS links
						if strings.HasPrefix(linkURL, "http://") || strings.HasPrefix(linkURL, "https://") {
							links = append(links, linkURL)
							visited[linkURL] = true
						}
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}

	f(doc)
	return links
}

// resolveURL resolves a relative URL against a base URL
func (c *Crawler) resolveURL(urlStr string, baseURL *url.URL) string {
	if urlStr == "" {
		return ""
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Resolve relative URLs
	resolved := baseURL.ResolveReference(parsed)

	// Only return HTTP/HTTPS URLs
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	return resolved.String()
}

// isMemeImage checks if an image URL looks like it might be a meme
// This is a simple heuristic - can be improved with ML or better filtering
func (c *Crawler) isMemeImage(imgURL string) bool {
	lower := strings.ToLower(imgURL)

	// Check file extension
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	hasImageExt := false
	for _, ext := range extensions {
		if strings.HasSuffix(lower, ext) {
			hasImageExt = true
			break
		}
	}

	if !hasImageExt {
		return false
	}

	// Accept images from known meme sites
	memeDomains := []string{
		"reddit.com", "imgur.com", "9gag.com", "knowyourmeme.com",
		"memegenerator.net", "memecenter.com",
	}
	for _, domain := range memeDomains {
		if strings.Contains(lower, domain) {
			return true
		}
	}

	// Keyword matching for AI memes. We match against whole tokens in the URL
	// path (split on non-alphanumeric boundaries) rather than raw substrings,
	// so short keywords like "ai" match "ai-meme.jpg" but not "contains.jpg".
	// The host is excluded to avoid matching e.g. the ".ai" TLD on every image.
	keywords := map[string]bool{
		"meme": true, "ai": true, "artificial": true, "intelligence": true,
		"chatgpt": true, "gpt": true, "dalle": true, "midjourney": true,
		"stable": true, "diffusion": true, "robot": true, "machine": true,
		"learning": true, "neural": true, "network": true, "deep": true,
		"llm": true, "generative": true,
	}

	pathToScan := lower
	if parsed, err := url.Parse(lower); err == nil {
		pathToScan = parsed.Path + " " + parsed.RawQuery
	}
	tokens := strings.FieldsFunc(pathToScan, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, token := range tokens {
		if keywords[token] {
			return true
		}
	}

	return false
}

// downloadImage downloads an image and optionally uploads to S3
func (c *Crawler) downloadImage(imgURL string, workerID int) {
	log.Printf("[Worker %d] Downloading image: %s", workerID, imgURL)

	req, err := http.NewRequest("GET", imgURL, nil)
	if err != nil {
		log.Printf("[Worker %d] Error creating request: %v", workerID, err)
		return
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[Worker %d] Error downloading image: %v", workerID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Worker %d] Non-200 status for image: %d", workerID, resp.StatusCode)
		return
	}

	// Read image data (capped to guard against oversized responses)
	imgData, err := io.ReadAll(io.LimitReader(resp.Body, maxImageSize))
	if err != nil {
		log.Printf("[Worker %d] Error reading image data: %v", workerID, err)
		return
	}

	// Check if it's actually an image
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		log.Printf("[Worker %d] Not an image: %s", workerID, contentType)
		return
	}

	c.statsMu.Lock()
	c.stats.ImagesDownloaded++
	c.statsMu.Unlock()

	// Generate a unique filename
	filename := c.generateFilename(imgURL, contentType)

	// Save to local directory if configured
	if c.config.LocalDir != "" {
		filePath := c.getUniqueFilePath(filename)
		err = os.WriteFile(filePath, imgData, 0644)
		if err != nil {
			log.Printf("[Worker %d] Error saving image locally: %v", workerID, err)
		} else {
			log.Printf("[Worker %d] Successfully saved image locally: %s", workerID, filePath)
		}
	}

	// Upload to S3 if configured
	if c.config.S3Client != nil {
		err = c.config.S3Client.UploadImage(filename, imgData, contentType)
		if err != nil {
			log.Printf("[Worker %d] Error uploading to S3: %v", workerID, err)
		} else {
			log.Printf("[Worker %d] Successfully uploaded to S3: %s", workerID, filename)
			c.statsMu.Lock()
			c.stats.ImagesUploaded++
			c.statsMu.Unlock()
		}
	}
}

// generateFilename generates a unique filename for an image
func (c *Crawler) generateFilename(imgURL, contentType string) string {
	parsedURL, err := url.Parse(imgURL)
	if err != nil {
		return fmt.Sprintf("meme_%d", time.Now().UnixNano())
	}

	// Extract filename from URL path
	path := parsedURL.Path
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]

	// If no filename in URL, generate one
	if filename == "" || !strings.Contains(filename, ".") {
		ext := ".jpg"
		if strings.Contains(contentType, "png") {
			ext = ".png"
		} else if strings.Contains(contentType, "gif") {
			ext = ".gif"
		} else if strings.Contains(contentType, "webp") {
			ext = ".webp"
		}
		filename = fmt.Sprintf("meme_%d%s", time.Now().UnixNano(), ext)
	}

	// Sanitize filename
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "%20", "_")

	return filename
}

// getUniqueFilePath returns a unique file path, appending a number if the file already exists
func (c *Crawler) getUniqueFilePath(filename string) string {
	basePath := filepath.Join(c.config.LocalDir, filename)

	// If file doesn't exist, return the original path
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	// File exists, try appending numbers
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	for i := 1; i < 10000; i++ {
		newFilename := fmt.Sprintf("%s_%d%s", name, i, ext)
		newPath := filepath.Join(c.config.LocalDir, newFilename)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback: use timestamp
	return filepath.Join(c.config.LocalDir, fmt.Sprintf("%s_%d%s", name, time.Now().UnixNano(), ext))
}

// extractDomain extracts the domain from a URL
func (c *Crawler) extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Return the host (domain) part
	host := parsed.Host

	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	return host
}
