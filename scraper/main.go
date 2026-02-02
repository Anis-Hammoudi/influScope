package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"

	// The Upfluence Stack
	"github.com/upfluence/amqp"
	"github.com/upfluence/amqp/amqputil"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"

	"github.com/hammo/influScope/pkg/models"
)

// Configuration for "Smart" Data
var categories = []string{"Tech", "Fashion", "Travel", "Food", "Gaming"}
var bioKeywords = map[string][]string{
	"Tech":    {"gadgets", "coding", "AI", "golang", "developer"},
	"Fashion": {"style", "OOTD", "luxury", "vogue", "streetwear"},
	"Travel":  {"wanderlust", "adventure", "nomad", "exploring"},
	"Food":    {"vegan", "tasty", "recipes", "organic", "chef"},
	"Gaming":  {"esports", "twitch", "fortnite", "streamer"},
}
var (
	profilesDiscovered = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "influencers_discovered_total",
			Help: "Total number of influencer profiles generated",
		},
	)
)

func init() {
	// Register it so Prometheus can see it
	prometheus.MustRegister(profilesDiscovered)
}

func main() {
	// Start a tiny web server for Prometheus in the background
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
	// 1. Initialize Random Seed
	gofakeit.Seed(time.Now().UnixNano())
	const exchangeName = "influencer-events"
	// We loop until we connect or run out of attempts.
	var broker amqp.Broker
	var err error
	maxRetries := 10

	for i := 1; i <= maxRetries; i++ {
		fmt.Printf("Connecting to RabbitMQ (Attempt %d/%d)...\n", i, maxRetries)

		broker = amqputil.Open()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = broker.DeclareExchange(ctx, "influencer-events", amqp.Fanout, amqp.DeclareExchangeOptions{Durable: true})
		cancel()

		if err == nil {
			fmt.Println(" Connected to RabbitMQ successfully!")
			break
		}

		fmt.Printf(" Connection failed: %v. Retrying in 3 seconds...\n", err)
		time.Sleep(3 * time.Second)
	}

	// If we exhausted all retries and still failed, we crash.
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ after %d attempts. Exiting.", maxRetries)
	}
	defer broker.Close()

	ctx := context.Background()
	fmt.Println("Scraper Service Started! Generating profiles...")
	// 4. The Infinite Generation Loop
	for {
		// A. Generate a realistic profile
		profile := generateSmartProfile()

		// B. Convert to JSON
		body, err := json.Marshal(profile)
		if err != nil {
			log.Printf("Error marshalling JSON: %v", err)
			continue
		}

		// C. Publish using Upfluence Library
		err = broker.Publish(
			ctx,
			exchangeName,
			"",
			amqp.Message{
				Body:        body,
				ContentType: "application/json",
				Headers: map[string]interface{}{
					"service": "scraper-v1",
					"version": "1.0.0",
				},
			},
			amqp.PublishOptions{},
		)

		if err != nil {
			log.Printf(" Failed to publish: %v", err)
		} else {
			log.Printf("Discovered: %-15s | %s", profile.Username, profile.Category)
			profilesDiscovered.Inc()
		}

		// D. Wait (Simulate network latency)
		time.Sleep(1 * time.Second)
	}
}

// generateSmartProfile creates consistent data (e.g., Tech category = Tech bio)
func generateSmartProfile() models.Influencer {
	// Pick random category
	category := categories[rand.Intn(len(categories))]

	// Pick random keyword associated with that category
	keywords := bioKeywords[category]
	keyword := keywords[rand.Intn(len(keywords))]

	return models.Influencer{
		ID:             gofakeit.UUID(),
		Username:       gofakeit.Username(),
		Platform:       gofakeit.RandomString([]string{"Instagram", "TikTok", "YouTube"}),
		Followers:      gofakeit.Number(1000, 5000000),
		Category:       category,
		Bio:            fmt.Sprintf("%s | Loves %s | #%s", gofakeit.JobDescriptor(), keyword, category),
		EngagementRate: float64(gofakeit.Number(10, 80)) / 10.0, // e.g. 4.5
	}
}
