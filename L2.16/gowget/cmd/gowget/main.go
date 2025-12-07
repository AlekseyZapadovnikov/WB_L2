package main

import (
	"flag"
	"log"
	"time"

	"gowget/internal/crawler"
)

func main() {
	urlFlag := flag.String("url", "", "URL to download (required)")
	depthFlag := flag.Int("depth", 1, "Recursion depth")
	concurrencyFlag := flag.Int("concurrency", 5, "Max concurrent downloads")
	outputFlag := flag.String("output", "downloaded_site", "Output directory")
	timeoutFlag := flag.Duration("timeout", 10*time.Second, "Request timeout")

	flag.Parse()

	if *urlFlag == "" {
		log.Fatal("Error: --url flag is required")
	}

	cfg := crawler.Config{
		MaxDepth:       *depthFlag,
		MaxConcurrency: *concurrencyFlag,
		Timeout:        *timeoutFlag,
		OutputDir:      *outputFlag,
	}

	c := crawler.New(cfg)

	log.Printf("Starting crawl of %s with depth %d...", *urlFlag, *depthFlag)
	startTime := time.Now()

	if err := c.Start(*urlFlag); err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	log.Printf("Completed in %v", time.Since(startTime))
}