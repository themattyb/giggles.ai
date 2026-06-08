package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name, content string, modTime time.Time) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if !modTime.IsZero() {
		if err := os.Chtimes(p, modTime, modTime); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func TestCalculateHash(t *testing.T) {
	dir := t.TempDir()
	d := NewDeduplicator(dir)

	a := writeFile(t, dir, "a.txt", "same content", time.Time{})
	b := writeFile(t, dir, "b.txt", "same content", time.Time{})
	c := writeFile(t, dir, "c.txt", "different", time.Time{})

	ha, err := d.CalculateHash(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := d.CalculateHash(b)
	if err != nil {
		t.Fatal(err)
	}
	hc, err := d.CalculateHash(c)
	if err != nil {
		t.Fatal(err)
	}

	if ha != hb {
		t.Errorf("identical files produced different hashes: %s vs %s", ha, hb)
	}
	if ha == hc {
		t.Errorf("different files produced the same hash: %s", ha)
	}
}

func TestLoadSaveDatabaseRoundTrip(t *testing.T) {
	dir := t.TempDir()
	d := NewDeduplicator(dir)

	// LoadDatabase on a missing file is not an error.
	if err := d.LoadDatabase(); err != nil {
		t.Fatalf("LoadDatabase on missing file returned error: %v", err)
	}

	d.database.Records = []HashRecord{
		{Filename: "x.jpg", Hash: "abc", CreatedAt: time.Now(), FileModTime: time.Now()},
	}
	if err := d.SaveDatabase(); err != nil {
		t.Fatal(err)
	}

	d2 := NewDeduplicator(dir)
	if err := d2.LoadDatabase(); err != nil {
		t.Fatal(err)
	}
	if len(d2.database.Records) != 1 || d2.database.Records[0].Filename != "x.jpg" || d2.database.Records[0].Hash != "abc" {
		t.Errorf("round-trip mismatch: %+v", d2.database.Records)
	}
}

func TestProcessFilesRemovesNewerDuplicate(t *testing.T) {
	dir := t.TempDir()

	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)
	writeFile(t, dir, "original.jpg", "duplicate content", older)
	writeFile(t, dir, "copy.jpg", "duplicate content", newer)
	writeFile(t, dir, "unique.jpg", "unique content", newer)

	d := NewDeduplicator(dir)
	if err := d.LoadDatabase(); err != nil {
		t.Fatal(err)
	}
	if err := d.ProcessFiles(); err != nil {
		t.Fatal(err)
	}

	// The older file of the duplicate pair is kept; the newer one is removed.
	if _, err := os.Stat(filepath.Join(dir, "original.jpg")); err != nil {
		t.Errorf("expected original.jpg to be kept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "copy.jpg")); !os.IsNotExist(err) {
		t.Errorf("expected copy.jpg to be removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "unique.jpg")); err != nil {
		t.Errorf("expected unique.jpg to be kept: %v", err)
	}

	// A valid hash database is written.
	if _, err := os.Stat(filepath.Join(dir, ".hashdb.json")); err != nil {
		t.Errorf("expected .hashdb.json to be written: %v", err)
	}
}

func TestProcessFilesNoDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.jpg", "content a", time.Time{})
	writeFile(t, dir, "b.jpg", "content b", time.Time{})

	d := NewDeduplicator(dir)
	if err := d.LoadDatabase(); err != nil {
		t.Fatal(err)
	}
	if err := d.ProcessFiles(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"a.jpg", "b.jpg"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to be kept: %v", name, err)
		}
	}
	if len(d.database.Records) != 2 {
		t.Errorf("expected 2 unique records, got %d", len(d.database.Records))
	}
}
