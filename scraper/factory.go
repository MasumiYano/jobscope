package scraper

func CreateScraper(siteName string) Scraper {
	switch siteName {
	case "indeed":
		return NewIndeedScraper()
	default:
		return NewIndeedScraper()
	}
}
