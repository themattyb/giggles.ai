package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// HashRecord represents a file hash record
type HashRecord struct {
	Filename    string    `json:"filename"`
	Hash        string    `json:"hash"`
	CreatedAt   time.Time `json:"created_at"`
	FileModTime time.Time `json:"file_mod_time"`
}

// HashDatabase stores all hash records
type HashDatabase struct {
	Records []HashRecord `json:"records"`
}

// Deduplicator handles file deduplication
type Deduplicator struct {
	imageDir    string
	hashDBFile  string
	database    HashDatabase
}

// NewDeduplicator creates a new deduplicator instance
func NewDeduplicator(imageDir string) *Deduplicator {
	hashDBFile := filepath.Join(imageDir, ".hashdb.json")
	return &Deduplicator{
		imageDir:   imageDir,
		hashDBFile: hashDBFile,
		database:   HashDatabase{Records: []HashRecord{}},
	}
}

// LoadDatabase loads the hash database from file
func (d *Deduplicator) LoadDatabase() error {
	data, err := os.ReadFile(d.hashDBFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Database doesn't exist yet, start with empty database
			return nil
		}
		return fmt.Errorf("failed to read hash database: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	err = json.Unmarshal(data, &d.database)
	if err != nil {
		return fmt.Errorf("failed to parse hash database: %w", err)
	}

	return nil
}

// SaveDatabase saves the hash database to file
func (d *Deduplicator) SaveDatabase() error {
	data, err := json.MarshalIndent(d.database, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal hash database: %w", err)
	}

	err = os.WriteFile(d.hashDBFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write hash database: %w", err)
	}

	return nil
}

// CalculateHash calculates SHA256 hash of a file
func (d *Deduplicator) CalculateHash(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetFileModTime gets the modification time of a file
func (d *Deduplicator) GetFileModTime(filepath string) (time.Time, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// ProcessFiles processes all files in the image directory
func (d *Deduplicator) ProcessFiles() error {
	log.Printf("Scanning directory: %s", d.imageDir)

	// Get all files in the directory
	files, err := os.ReadDir(d.imageDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Hash map to track duplicates: hash -> list of files with that hash
	hashMap := make(map[string][]HashRecord)
	newRecords := []HashRecord{}

	// Process each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Skip the hash database file itself
		if file.Name() == ".hashdb.json" {
			continue
		}

		filePath := filepath.Join(d.imageDir, file.Name())
		
		log.Printf("Processing: %s", file.Name())

		// Calculate hash
		hash, err := d.CalculateHash(filePath)
		if err != nil {
			log.Printf("Error hashing %s: %v", file.Name(), err)
			continue
		}

		// Get file modification time
		modTime, err := d.GetFileModTime(filePath)
		if err != nil {
			log.Printf("Error getting mod time for %s: %v", file.Name(), err)
			modTime = time.Now()
		}

		record := HashRecord{
			Filename:    file.Name(),
			Hash:        hash,
			CreatedAt:   time.Now(),
			FileModTime: modTime,
		}

		// Check if this hash already exists in our database
		hashMap[hash] = append(hashMap[hash], record)
		newRecords = append(newRecords, record)
	}

	// Merge with existing database
	existingHashMap := make(map[string]HashRecord)
	for _, record := range d.database.Records {
		// Check if file still exists
		filePath := filepath.Join(d.imageDir, record.Filename)
		if _, err := os.Stat(filePath); err == nil {
			existingHashMap[record.Hash] = record
		}
	}

	// Find duplicates and remove newest files
	duplicatesRemoved := 0
	for hash, records := range hashMap {
		if len(records) > 1 {
			// Sort by modification time (oldest first)
			sort.Slice(records, func(i, j int) bool {
				return records[i].FileModTime.Before(records[j].FileModTime)
			})

			// Keep the oldest, remove the rest
			keepRecord := records[0]
			for i := 1; i < len(records); i++ {
				filePath := filepath.Join(d.imageDir, records[i].Filename)
				log.Printf("Removing duplicate (newer): %s (hash: %s)", records[i].Filename, hash[:16])
				if err := os.Remove(filePath); err != nil {
					log.Printf("Error removing duplicate file %s: %v", records[i].Filename, err)
				} else {
					duplicatesRemoved++
				}
			}

			// Update hashMap to only contain the kept record
			hashMap[hash] = []HashRecord{keepRecord}
		}
	}

	// Check for duplicates between new files and existing database
	for hash, newRecord := range hashMap {
		if len(newRecord) == 0 {
			continue
		}
		record := newRecord[0]

		if existingRecord, exists := existingHashMap[hash]; exists {
			// Duplicate found with existing file
			// Keep the one with older modification time
			if record.FileModTime.Before(existingRecord.FileModTime) {
				// New file is older, remove existing
				filePath := filepath.Join(d.imageDir, existingRecord.Filename)
				log.Printf("Removing duplicate (newer existing): %s (hash: %s)", existingRecord.Filename, hash[:16])
				if err := os.Remove(filePath); err != nil {
					log.Printf("Error removing duplicate file %s: %v", existingRecord.Filename, err)
				} else {
					duplicatesRemoved++
				}
				// Update database with new record
				existingHashMap[hash] = record
			} else {
				// Existing file is older, remove new one
				filePath := filepath.Join(d.imageDir, record.Filename)
				log.Printf("Removing duplicate (newer new): %s (hash: %s)", record.Filename, hash[:16])
				if err := os.Remove(filePath); err != nil {
					log.Printf("Error removing duplicate file %s: %v", record.Filename, err)
				} else {
					duplicatesRemoved++
				}
				// Keep existing record
			}
		} else {
			// New unique file, add to database
			existingHashMap[hash] = record
		}
	}

	// Rebuild database from existingHashMap
	d.database.Records = []HashRecord{}
	for _, record := range existingHashMap {
		// Verify file still exists
		filePath := filepath.Join(d.imageDir, record.Filename)
		if _, err := os.Stat(filePath); err == nil {
			d.database.Records = append(d.database.Records, record)
		}
	}

	// Save updated database
	if err := d.SaveDatabase(); err != nil {
		return fmt.Errorf("failed to save database: %w", err)
	}

	log.Printf("Deduplication complete:")
	log.Printf("  Files processed: %d", len(newRecords))
	log.Printf("  Duplicates removed: %d", duplicatesRemoved)
	log.Printf("  Unique files: %d", len(d.database.Records))

	return nil
}

// RunDeduplication runs the deduplication process
func RunDeduplication(imageDir string) error {
	dedupe := NewDeduplicator(imageDir)

	// Load existing database
	if err := dedupe.LoadDatabase(); err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	// Process files
	if err := dedupe.ProcessFiles(); err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	return nil
}

