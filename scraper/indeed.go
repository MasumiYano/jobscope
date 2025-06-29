package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"jobscope/models"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
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
	err := i.debugResponse(searchURL)
	if err != nil {
		fmt.Printf("Debug falied: %v\n", err)
	}

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
	params := url.Values{}
	params.Set("q", jobTitle)
	params.Set("l", location)
	return i.baseURL + "/jobs?" + params.Encode()
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

// Add this debugging function to your indeed.go file
func (i *IndeedScraper) debugResponse(searchURL string) error {
	fmt.Printf("🔍 Requesting URL: %s\n", searchURL)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return err
	}

	// Add more headers to look like a real browser
	req.Header.Set("User-Agent", i.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Response status: %s\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	htmlContent := string(body)

	// Save HTML to file for inspection
	err = os.WriteFile("debug_search.html", body, 0644)
	if err != nil {
		fmt.Printf("⚠️ Could not save debug file: %v\n", err)
	} else {
		fmt.Println("💾 Saved HTML response to debug_search.html")
	}

	fmt.Printf("📄 HTML content length: %d characters\n", len(htmlContent))

	// Check what JavaScript variables exist
	if strings.Contains(htmlContent, "_initialData") {
		fmt.Println("✅ Found _initialData in HTML")
	} else {
		fmt.Println("❌ _initialData not found in HTML")
	}

	if strings.Contains(htmlContent, "window.mosaic") {
		fmt.Println("✅ Found window.mosaic")
	}

	if strings.Contains(htmlContent, "data-jk") {
		fmt.Println("✅ Found data-jk attributes")
	}

	// Check if we got blocked
	if strings.Contains(htmlContent, "blocked") || strings.Contains(htmlContent, "captcha") || strings.Contains(htmlContent, "robot") {
		fmt.Println("🚫 Possibly blocked by anti-bot detection")
	}

	return nil
}

func (i *IndeedScraper) scrapeSearchResult(searchURL string) (*SearchResults, error) {
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", i.userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data, err := i.extractInitialData(string(body))
	if err != nil {
		return nil, err
	}

	jobKeys, err := i.extractJobKeysFromSearchData(data)
	if err != nil {
		return nil, err
	}

	return &SearchResults{JobKeys: jobKeys}, nil
}

func removeDuplicates(jobKeys []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, jobKey := range jobKeys {
		if !seen[jobKey] {
			seen[jobKey] = true
			result = append(result, jobKey)
		}
	}

	return result
}

func getStringField(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (i *IndeedScraper) parseJobFromData(data map[string]interface{}) (models.JobPosting, error) {
	jobInfoWrapper, ok := data["jobInfoWrapperModel"].(map[string]interface{})
	if !ok {
		return models.JobPosting{}, fmt.Errorf("jobInfoWrapperModel not found")
	}

	jobInfo, ok := jobInfoWrapper["jobInfoModel"].(map[string]interface{})
	if !ok {
		return models.JobPosting{}, fmt.Errorf("jobInfoModel not found")
	}

	job := models.JobPosting{
		Title:       getStringField(jobInfo, "jobTitle"),
		Company:     getStringField(jobInfo, "companyName"),
		Location:    getStringField(jobInfo, "formattedLocation"),
		Salary:      getStringField(jobInfo, "salary"),
		Description: getStringField(jobInfo, "sanitizedJobDescription"),
		PostedDate:  getStringField(jobInfo, "pubDate"),
		URL:         fmt.Sprintf("%s/viewjob?jk=%s", i.baseURL, getStringField(jobInfo, "jobkey")),
		ScrapedAt:   time.Now(),
	}

	return job, nil
}

func (i *IndeedScraper) extractJobKeysFromSearchData(data map[string]interface{}) ([]string, error) {
	var jobKeys []string

	fmt.Printf("Top level keys: %v\n", getKeys(data))

	if metaData, ok := data["metaData"].(map[string]interface{}); ok {
		fmt.Printf("Metadata keys: %v\n", getKeys(metaData))

		if mosaicModel, ok := metaData["mosaicProviderJobCardsModel"].(map[string]interface{}); ok {
			fmt.Printf("Mosaicmodel keys: %v\n", getKeys(mosaicModel))

			if results, ok := mosaicModel["results"].([]interface{}); ok {
				fmt.Printf("Found %d results\n", len(results))

				for _, result := range results {
					if jobCard, ok := result.(map[string]interface{}); ok {
						if jobKey, ok := jobCard["jobkey"].(string); ok {
							jobKeys = append(jobKeys, jobKey)
						}
					}
				}
			}
		}
	}

	if len(jobKeys) == 0 {
		return nil, fmt.Errorf("No job keys found in search results")
	}

	return jobKeys, nil
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func NewIndeedScraper() *IndeedScraper {
	return &IndeedScraper{
		baseURL:   "https://indeed.com",
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
}
