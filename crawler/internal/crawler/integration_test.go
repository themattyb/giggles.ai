package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/temoto/robotstxt"
)

// newTestCrawler builds a Crawler suitable for unit tests that need an HTTP
// client and robots cache without going through New() (which requires start
// URLs and creates directories).
func newTestCrawler(userAgent string) *Crawler {
	return &Crawler{
		config:      Config{UserAgent: userAgent},
		client:      &http.Client{Timeout: 5 * time.Second},
		robotsCache: make(map[string]*robotstxt.RobotsData),
	}
}

func TestCanCrawl(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-agent: *\nDisallow: /private\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestCrawler("test-agent")

	if !c.canCrawl(srv.URL + "/public/page.html") {
		t.Error("expected /public to be crawlable")
	}
	if c.canCrawl(srv.URL + "/private/secret.html") {
		t.Error("expected /private to be blocked by robots.txt")
	}
}

func TestCanCrawlAllowsWhenRobotsMissing(t *testing.T) {
	// Server returns 404 for everything, including /robots.txt -> fail open.
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	c := newTestCrawler("test-agent")
	if !c.canCrawl(srv.URL + "/anything.html") {
		t.Error("expected crawling to be allowed when robots.txt returns 404")
	}
}

func TestCrawlerIntegration(t *testing.T) {
	const imageBody = "\xff\xd8\xff\xe0fake-jpeg-bytes"

	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r) // no robots restrictions
	})
	mux.HandleFunc("/ai-meme.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte(imageBody))
	})
	mux.HandleFunc("/page2.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>second page</body></html>`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// Includes both an image and an <a> link so the enqueue path is exercised.
		w.Write([]byte(`<html><body><img src="/ai-meme.jpg"><a href="/page2.html">next</a></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	c, err := New(Config{
		Workers:   1,
		Delay:     0,
		MaxPages:  5,
		StartURLs: []string{srv.URL + "/"},
		UserAgent: "test-agent",
		LocalDir:  dir,
		// InsecureSkipVerify intentionally left false; httptest.NewServer is plain HTTP.
	})
	if err != nil {
		t.Fatal(err)
	}

	stats, err := c.Run()
	if err != nil {
		t.Fatal(err)
	}

	// The start page plus the linked page should both be crawled.
	if stats.PagesCrawled < 2 {
		t.Errorf("expected at least 2 pages crawled, got %d", stats.PagesCrawled)
	}
	if stats.ImagesFound < 1 {
		t.Errorf("expected at least 1 image found, got %d", stats.ImagesFound)
	}
	if stats.ImagesDownloaded < 1 {
		t.Errorf("expected at least 1 image downloaded, got %d", stats.ImagesDownloaded)
	}

	// The downloaded image should be written to the local directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one file written to the local directory")
	}
}

func TestStopIsGraceful(t *testing.T) {
	// Each page links to a fresh deeper path, so the queue keeps growing and the
	// crawler would run indefinitely without intervention.
	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", http.NotFound)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><a href="%sx.html">next</a></body></html>`, r.URL.Path)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := New(Config{
		Workers:   2,
		Delay:     5 * time.Millisecond,
		MaxPages:  1_000_000,
		StartURLs: []string{srv.URL + "/"},
		UserAgent: "test-agent",
		LocalDir:  t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}

	finished := make(chan struct{})
	go func() {
		c.Run()
		close(finished)
	}()

	// Let it crawl briefly, then request a graceful stop.
	time.Sleep(150 * time.Millisecond)
	c.Stop()

	select {
	case <-finished:
		// Run returned -> graceful stop worked without panicking.
	case <-time.After(5 * time.Second):
		t.Fatal("crawler did not stop within 5s of Stop()")
	}
}
