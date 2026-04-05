package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/hammo/influScope/pkg/models"
	"github.com/hammo/influScope/scraper/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	categories  = []string{"Tech", "Fashion", "Travel", "Food", "Gaming"}
	bioKeywords = map[string][]string{
		"Tech":    {"gadgets", "coding", "AI", "golang", "developer"},
		"Fashion": {"style", "OOTD", "luxury", "vogue", "streetwear"},
		"Travel":  {"wanderlust", "adventure", "nomad", "exploring"},
		"Food":    {"vegan", "tasty", "recipes", "organic", "chef"},
		"Gaming":  {"esports", "twitch", "fortnite", "streamer"},
	}
)

type ScraperService struct {
	storage   domain.AvatarStorage
	publisher domain.EventPublisher
	metric    prometheus.Counter
}

func NewScraperService(storage domain.AvatarStorage, publisher domain.EventPublisher, metric prometheus.Counter) *ScraperService {
	return &ScraperService{
		storage:   storage,
		publisher: publisher,
		metric:    metric,
	}
}

// GenerateSmartProfile is now exported so it can be tested easily
func (s *ScraperService) GenerateSmartProfile() models.Influencer {
	category := categories[rand.Intn(len(categories))]
	keywords := bioKeywords[category]
	keyword := keywords[rand.Intn(len(keywords))]

	return models.Influencer{
		ID:             gofakeit.UUID(),
		Username:       gofakeit.Username(),
		Platform:       gofakeit.RandomString([]string{"Instagram", "TikTok", "YouTube"}),
		Followers:      gofakeit.Number(1000, 5000000),
		Category:       category,
		Bio:            fmt.Sprintf("%s | Loves %s | #%s", gofakeit.JobDescriptor(), keyword, category),
		EngagementRate: float64(gofakeit.Number(10, 80)) / 10.0,
	}
}

// Run executes the continuous scraping loop
func (s *ScraperService) Run(ctx context.Context) {
	log.Println("Scraper Service Started! Generating profiles...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Scraper service stopping...")
			return
		default:
			profile := s.GenerateSmartProfile()
			dummyImage := []byte(fmt.Sprintf("Fake image content for %s", profile.Username))

			url, err := s.storage.UploadAvatar(ctx, profile.Username, dummyImage)
			if err != nil {
				log.Printf("Avatar upload failed: %v", err)
				continue
			}
			profile.AvatarURL = url

			if err := s.publisher.PublishProfile(ctx, profile); err != nil {
				log.Printf("Failed to publish profile: %v", err)
			} else {
				log.Printf("Discovered: %-15s | %s", profile.Username, profile.Category)
				s.metric.Inc()
			}

			time.Sleep(1 * time.Second)
		}
	}
}
