package scraper

import "jobscope/models"

type Scraper interface {
	SearchJobs(jobTitle, location, level string, limit int) ([]models.JobPosting, error)
}
