package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	job := flag.String("job", "", "Your job title (required)")
	location := flag.String("location", "vancouver", "Job location")
	level := flag.String("level", "junior", "Experience level")
	limit := flag.Int("limit", 20, "Maximum number of results")
	output := flag.String("output", "jobs.json", "Output JSON file")
	flag.Parse()

	if *job == "" {
		fmt.Println("Error: --job is required")
		flag.Usage()
		os.Exit(1)
	}
}
