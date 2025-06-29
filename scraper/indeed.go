package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"jobscope/models"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type IndeedScraper struct {
	baseURL   string
	userAgent string
}

type SearchResults struct {
	JobKeys []string
}

func (i *IndeedScraper) SearchJobs(jobTitle, location, level string, limit int) ([]models.JobPosting, error) {
	searchURL := i.buildSearchURL(jobTitle, location, level, limit)
	searchResults, err := i.scrapeSearchResult(searchURL)

	if err != nil {
		return nil, err
	}

	var jobs []models.JobPosting
	for _, jobkey := range searchResults.JobKeys {
		jobDetail, err := i.scrapeJobDetail(jobkey)

		if err != nil {
			continue
		}
		jobs = append(jobs, jobDetail)
		time.Sleep(1 * time.Second)
	}

	return jobs, nil
}

func (i *IndeedScraper) buildSearchURL(jobTitle, location, level string, limit int) string {
	searchPath := "/jobs"
	params := url.Values{}
	params.Set("q", jobTitle)
	params.Set("l", location)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	return i.baseURL + searchPath + "?" + params.Encode()
}

func (i *IndeedScraper) extractInitialData(htmlContent string) (map[string]interface{}, error) {
	re := regexp.MustCompile(`_initialData=(\{.+?\});`)
	matches := re.FindStringSubmatch(htmlContent)

	if len(matches) < 2 {
		return nil, fmt.Errorf("Cound not find _initialData in HTML")
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(matches[1]), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return data, nil
}

func (i *IndeedScraper) scrapeJobDetail(jobKey string) (models.JobPosting, error) {
	jobURL := fmt.Sprintf("%s/m/basecamp/viewjob?viewtype=embedded&jk=%s", i.baseURL, jobKey)

	req, err := http.NewRequest("GET", jobURL, nil)
	if err != nil {
		return models.JobPosting{}, err
	}
	req.Header.Set("User-Agent", i.userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.JobPosting{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.JobPosting{}, err
	}

	data, err := i.extractInitialData(string(body))
	if err != nil {
		return models.JobPosting{}, err
	}

	return i.parseJobFromData(data)
}

func (i *IndeedScraper) scrapeSearchResult(searchURL string) (*SearchResults, error) {
}

func (i *IndeedScraper) parseJobFromData(data map[string]interface{}) (models.JobPosting, error) {
}

func NewIndeedScraper() *IndeedScraper {
	return &IndeedScraper{
		baseURL:   "https://indeed.com",
		userAgent: "JobScope/1.0",
	}
}
