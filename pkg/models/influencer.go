package models

// Shared struct used by Scraper (Writer) and Indexer (Reader)
type Influencer struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	Platform       string  `json:"platform"`
	Followers      int     `json:"followers"`
	Category       string  `json:"category"`
	Bio            string  `json:"bio"`
	EngagementRate float64 `json:"engagement_rate"`
}
