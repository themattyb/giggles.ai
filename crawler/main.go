package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/giggles-ai/crawler/internal/crawler"
	"github.com/giggles-ai/crawler/internal/s3"
)

func main() {
	// Command line flags
	var (
		workers      = flag.Int("workers", 5, "Number of concurrent workers")
		delay        = flag.Duration("delay", 2*time.Second, "Delay between requests")
		maxPages     = flag.Int("max-pages", 100, "Maximum number of pages to crawl")
		s3Bucket     = flag.String("s3-bucket", "", "S3 bucket name for storing images")
		s3Region     = flag.String("s3-region", "us-east-1", "AWS region for S3 bucket")
		startURLs    = flag.String("start-urls", "", "Comma-separated list of starting URLs to crawl from")
		startURL     = flag.String("start-url", "", "Starting URL to crawl from (deprecated, use -start-urls)")
		userAgent    = flag.String("user-agent", "giggles-ai-crawler/1.0", "User agent string")
		localDir     = flag.String("local-dir", "found-images", "Local directory to save images")
		insecure     = flag.Bool("insecure", false, "Skip TLS certificate verification (use only for testing)")
		dedupe       = flag.Bool("dedupe", false, "Run deduplication on found-images directory (exits after deduplication)")
		dedupeDir    = flag.String("dedupe-dir", "found-images", "Directory to deduplicate (used with -dedupe)")
	)
	flag.Parse()

	// If dedupe flag is set, run deduplication and exit
	if *dedupe {
		log.Printf("Running deduplication on directory: %s", *dedupeDir)
		if err := RunDeduplication(*dedupeDir); err != nil {
			log.Fatalf("Deduplication failed: %v", err)
		}
		log.Println("Deduplication completed successfully")
		return
	}

	// Parse start URLs
	var startURLsList []string
	if *startURLs != "" {
		// Parse comma-separated URLs
		urls := strings.Split(*startURLs, ",")
		for _, u := range urls {
			u = strings.TrimSpace(u)
			if u != "" {
				startURLsList = append(startURLsList, u)
			}
		}
	} else if *startURL != "" {
		// Support deprecated -start-url flag for backward compatibility
		startURLsList = []string{*startURL}
	}

	// Validate required flags (only if not running deduplication)
	if len(startURLsList) == 0 {
		log.Fatal("Error: -start-urls is required (comma-separated list of URLs), or use -dedupe to run deduplication")
	}

	if *s3Bucket == "" {
		log.Printf("Info: -s3-bucket not specified, images will be saved locally to: %s", *localDir)
	}

	// Initialize S3 client if bucket is provided
	var s3Client *s3.Client
	if *s3Bucket != "" {
		var err error
		s3Client, err = s3.NewClient(*s3Bucket, *s3Region)
		if err != nil {
			log.Fatalf("Error initializing S3 client: %v", err)
		}
		log.Printf("S3 client initialized for bucket: %s (region: %s)", *s3Bucket, *s3Region)
	}

	// Create crawler configuration
	config := crawler.Config{
		Workers:            *workers,
		Delay:              *delay,
		MaxPages:           *maxPages,
		StartURLs:          startURLsList,
		UserAgent:          *userAgent,
		S3Client:           s3Client,
		LocalDir:           *localDir,
		InsecureSkipVerify: *insecure,
	}

	// Create and run crawler
	c, err := crawler.New(config)
	if err != nil {
		log.Fatalf("Error creating crawler: %v", err)
	}

	log.Printf("Starting crawler with %d workers, %v delay, max %d pages", 
		*workers, *delay, *maxPages)
	log.Printf("Starting URLs (%d):", len(startURLsList))
	for i, url := range startURLsList {
		log.Printf("  %d. %s", i+1, url)
	}

	// Run the crawler
	stats, err := c.Run()
	if err != nil {
		log.Fatalf("Error running crawler: %v", err)
	}

	// Print statistics
	fmt.Println("\n=== Crawler Statistics ===")
	fmt.Printf("Pages crawled: %d\n", stats.PagesCrawled)
	fmt.Printf("Images found: %d\n", stats.ImagesFound)
	fmt.Printf("Images downloaded: %d\n", stats.ImagesDownloaded)
	fmt.Printf("Images uploaded to S3: %d\n", stats.ImagesUploaded)
	fmt.Printf("Errors: %d\n", stats.Errors)
	fmt.Printf("Duration: %v\n", stats.Duration)
}


