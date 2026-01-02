package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/giggles-ai/crawler/internal/crawler"
	"github.com/giggles-ai/crawler/internal/s3"
)

func main() {
	// Command line flags
	var (
		workers     = flag.Int("workers", 5, "Number of concurrent workers")
		delay       = flag.Duration("delay", 2*time.Second, "Delay between requests")
		maxPages    = flag.Int("max-pages", 100, "Maximum number of pages to crawl")
		s3Bucket    = flag.String("s3-bucket", "", "S3 bucket name for storing images")
		s3Region    = flag.String("s3-region", "us-east-1", "AWS region for S3 bucket")
		startURL    = flag.String("start-url", "", "Starting URL to crawl from")
		userAgent   = flag.String("user-agent", "giggles-ai-crawler/1.0", "User agent string")
		configFile  = flag.String("config", "", "Path to configuration file (optional)")
	)
	flag.Parse()

	// Validate required flags
	if *startURL == "" {
		log.Fatal("Error: -start-url is required")
	}

	if *s3Bucket == "" {
		log.Println("Warning: -s3-bucket not specified, images will not be uploaded to S3")
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
		Workers:     *workers,
		Delay:       *delay,
		MaxPages:    *maxPages,
		StartURL:    *startURL,
		UserAgent:   *userAgent,
		S3Client:    s3Client,
	}

	// Create and run crawler
	c, err := crawler.New(config)
	if err != nil {
		log.Fatalf("Error creating crawler: %v", err)
	}

	log.Printf("Starting crawler with %d workers, %v delay, max %d pages", 
		*workers, *delay, *maxPages)
	log.Printf("Starting URL: %s", *startURL)

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

