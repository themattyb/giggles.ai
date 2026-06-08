package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// ManifestEntry is one meme in the GUI manifest (memes.json).
type ManifestEntry struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	Source     string    `json:"source"`
	UploadedAt time.Time `json:"uploadedAt"`
}

// Manifest is the document the GUI fetches and renders.
type Manifest struct {
	GeneratedAt time.Time       `json:"generatedAt"`
	Memes       []ManifestEntry `json:"memes"`
}

var manifestImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// GenerateManifest scans imageDir for image files and writes a memes.json
// manifest to outPath. baseURL is prepended to each filename to form the image
// URL (e.g. an S3 or CloudFront prefix); if empty, the bare filename is used,
// which is correct when the GUI is served from the same directory as the
// images. Returns the number of memes written.
func GenerateManifest(imageDir, outPath, baseURL string, now time.Time) (int, error) {
	entries, err := os.ReadDir(imageDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read image directory: %w", err)
	}

	manifest := Manifest{GeneratedAt: now, Memes: []ManifestEntry{}}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !manifestImageExts[strings.ToLower(filepath.Ext(name))] {
			continue
		}

		var modTime time.Time
		if info, err := entry.Info(); err == nil {
			modTime = info.ModTime()
		}

		manifest.Memes = append(manifest.Memes, ManifestEntry{
			ID:         name,
			URL:        joinURL(baseURL, name),
			Title:      titleFromFilename(name),
			Source:     "crawler",
			UploadedAt: modTime,
		})
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write manifest: %w", err)
	}
	return len(manifest.Memes), nil
}

// joinURL prepends a base URL to a filename, tolerating a trailing slash.
func joinURL(baseURL, name string) string {
	if baseURL == "" {
		return name
	}
	return strings.TrimRight(baseURL, "/") + "/" + name
}

// titleFromFilename turns "ai-fails-107_700.jpg" into "Ai Fails 107 700".
func titleFromFilename(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	base = strings.NewReplacer("_", " ", "-", " ", ".", " ").Replace(base)

	fields := strings.Fields(base)
	for i, f := range fields {
		fields[i] = capitalize(f)
	}
	title := strings.Join(fields, " ")
	if title == "" {
		return "Untitled"
	}
	return title
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
