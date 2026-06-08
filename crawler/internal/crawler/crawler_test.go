package crawler

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsMemeImage(t *testing.T) {
	c := &Crawler{}

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"keyword in filename", "https://example.com/ai-meme.jpg", true},
		{"chatgpt keyword", "https://example.com/chatgpt-fail.png", true},
		{"meme domain", "https://imgur.com/photo.png", true},
		{"non-image extension", "https://example.com/logo.svg", false},
		{"image but no keyword", "https://example.com/photo.jpg", false},
		// Regression: the old substring match treated "ai" inside "contains" as
		// a hit. Token matching must not.
		{"substring false positive", "https://example.com/contains.jpg", false},
		// Regression: ".ai" TLD must not match every image on the host.
		{"ai tld in host only", "https://giggles.ai/photo.jpg", false},
		{"ai tld but keyword in path", "https://giggles.ai/robot-meme.jpg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.isMemeImage(tt.url); got != tt.want {
				t.Errorf("isMemeImage(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	c := &Crawler{}
	base, _ := url.Parse("https://example.com/dir/page.html")

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"relative path", "../img.jpg", "https://example.com/img.jpg"},
		{"absolute path", "/a/b.png", "https://example.com/a/b.png"},
		{"non-http scheme", "mailto:x@y.com", ""},
		{"javascript scheme", "javascript:alert(1)", ""},
		{"empty input", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.resolveURL(tt.in, base); got != tt.want {
				t.Errorf("resolveURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	c := &Crawler{}

	t.Run("extracts filename from path", func(t *testing.T) {
		got := c.generateFilename("https://example.com/path/funny.jpg", "image/jpeg")
		if got != "funny.jpg" {
			t.Errorf("got %q, want funny.jpg", got)
		}
	})

	t.Run("sanitizes encoded spaces", func(t *testing.T) {
		got := c.generateFilename("https://example.com/my%20meme.jpg", "image/jpeg")
		if got != "my_meme.jpg" {
			t.Errorf("got %q, want my_meme.jpg", got)
		}
	})

	t.Run("timestamped fallback maps content type to extension", func(t *testing.T) {
		cases := map[string]string{
			"image/png":  ".png",
			"image/gif":  ".gif",
			"image/webp": ".webp",
			"image/jpeg": ".jpg",
		}
		for ct, ext := range cases {
			got := c.generateFilename("https://example.com/", ct)
			if !strings.HasPrefix(got, "meme_") || !strings.HasSuffix(got, ext) {
				t.Errorf("contentType %q: got %q, want meme_*%s", ct, got, ext)
			}
		}
	})
}

func TestGetUniqueFilePath(t *testing.T) {
	dir := t.TempDir()
	c := &Crawler{config: Config{LocalDir: dir}}

	// File doesn't exist yet -> original path.
	want := filepath.Join(dir, "a.jpg")
	if got := c.getUniqueFilePath("a.jpg"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	// Create it, then expect _1, then _2.
	if err := os.WriteFile(want, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	want1 := filepath.Join(dir, "a_1.jpg")
	if got := c.getUniqueFilePath("a.jpg"); got != want1 {
		t.Fatalf("got %q, want %q", got, want1)
	}

	if err := os.WriteFile(want1, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	want2 := filepath.Join(dir, "a_2.jpg")
	if got := c.getUniqueFilePath("a.jpg"); got != want2 {
		t.Fatalf("got %q, want %q", got, want2)
	}
}

func TestIsAllowedDomain(t *testing.T) {
	// No allow-list -> everything is in scope.
	open := &Crawler{allowedDomains: map[string]bool{}}
	if !open.isAllowedDomain("https://anything.com/x.jpg") {
		t.Error("expected all domains allowed when allow-list is empty")
	}

	// With an allow-list, only listed hosts are in scope (port is ignored).
	scoped := &Crawler{allowedDomains: map[string]bool{"example.com": true}}
	if !scoped.isAllowedDomain("https://example.com/x.jpg") {
		t.Error("expected example.com to be allowed")
	}
	if !scoped.isAllowedDomain("https://example.com:8080/x.jpg") {
		t.Error("expected example.com with port to be allowed")
	}
	if scoped.isAllowedDomain("https://other.com/x.jpg") {
		t.Error("expected other.com to be rejected")
	}
}

func TestNewSameDomainBuildsAllowList(t *testing.T) {
	c, err := New(Config{
		StartURLs:  []string{"https://example.com/start"},
		SameDomain: true,
		MaxPages:   1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !c.isAllowedDomain("https://example.com/a.jpg") {
		t.Error("expected start-URL domain to be allowed under -same-domain")
	}
	if c.isAllowedDomain("https://evil.com/a.jpg") {
		t.Error("expected a different domain to be rejected under -same-domain")
	}
}

func TestExtractDomain(t *testing.T) {
	c := &Crawler{}

	tests := []struct {
		in   string
		want string
	}{
		{"https://example.com/path", "example.com"},
		{"https://example.com:8080/path", "example.com"},
		{"http://sub.example.com", "sub.example.com"},
		{"://not a url", ""},
	}

	for _, tt := range tests {
		if got := c.extractDomain(tt.in); got != tt.want {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
