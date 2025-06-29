package models

import "time"

type JobPosting struct {
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	Salary      string    `json:"salary"`
	Description string    `json:"description"`
	PostedDate  string    `json:"posted_date"`
	URL         string    `json:"url"`
	ScrapedAt   time.Time `json:"scraped_at"`
}
