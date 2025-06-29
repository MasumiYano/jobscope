package main

import (
	"flag"
	"fmt"
	"jobscope/scraper"
	"os"
)

func main() {
	job := flag.String("job", "", "Your job title (required)")
	location := flag.String("location", "vancouver", "Job location")
	level := flag.String("level", "junior", "Experience level")
	limit := flag.Int("limit", 20, "Maximum number of results")
	// output := flag.String("output", "jobs.json", "Output JSON file")
	flag.Parse()

	if *job == "" {
		fmt.Println("Error: --job is required")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Running Indeed scraper")
	scraper := scraper.NewIndeedScraper()
	jobs, err := scraper.SearchJobs(*job, *location, *level, *limit)

	if err != nil {
		fmt.Printf("Error scraping jobs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d jobs:\n", len(jobs))
	for _, job := range jobs {
		fmt.Printf("- %s at %s\n", job.Title, job.Company)
	}
}
