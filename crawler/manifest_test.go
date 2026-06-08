package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateManifest(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"ai-fails-1.jpg", "robot.png", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	out := filepath.Join(t.TempDir(), "memes.json")
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	n, err := GenerateManifest(dir, out, "https://cdn.example.com/memes/", now)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 image entries (txt excluded), got %d", n)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}
	if len(m.Memes) != 2 {
		t.Fatalf("expected 2 memes in manifest, got %d", len(m.Memes))
	}
	if !m.GeneratedAt.Equal(now) {
		t.Errorf("generatedAt = %v, want %v", m.GeneratedAt, now)
	}

	byID := map[string]ManifestEntry{}
	for _, e := range m.Memes {
		byID[e.ID] = e
	}

	jpg, ok := byID["ai-fails-1.jpg"]
	if !ok {
		t.Fatal("expected ai-fails-1.jpg in manifest")
	}
	if jpg.URL != "https://cdn.example.com/memes/ai-fails-1.jpg" {
		t.Errorf("url = %q, want base URL joined to filename", jpg.URL)
	}
	if jpg.Title != "Ai Fails 1" {
		t.Errorf("title = %q, want %q", jpg.Title, "Ai Fails 1")
	}
	if _, ok := byID["notes.txt"]; ok {
		t.Error("non-image notes.txt should not be in the manifest")
	}
}

func TestGenerateManifestEmptyBaseURL(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.gif"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "memes.json")

	if _, err := GenerateManifest(dir, out, "", time.Now()); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(out)
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Memes) != 1 || m.Memes[0].URL != "a.gif" {
		t.Errorf("with empty base URL the bare filename should be used, got %+v", m.Memes)
	}
}
