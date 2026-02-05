package main

import (
	"bytes"
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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

// S3 Client Wrapper
type StorageService struct {
	Client *s3.Client
	Bucket string
}

func NewStorageService() *StorageService {
	// 1. Configure AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("admin", "password", "")),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://s3:9000")
		o.UsePathStyle = true
	})

	svc := &StorageService{Client: client, Bucket: "avatars"}

	// 2. SELF-HEALING: Ensure Bucket Exists
	// We loop here because MinIO might take a few seconds to start up.
	// This is much better than a "sleep 5" in a shell script.
	for i := 0; i < 30; i++ {
		_, err := client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: aws.String(svc.Bucket),
		})

		if err == nil {
			log.Println(" S3 Bucket 'avatars' exists.")
			break
		}

		// If it doesn't exist (or we can't connect yet), try to create it
		_, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
			Bucket: aws.String(svc.Bucket),
		})

		if err == nil {
			log.Println(" Created missing S3 Bucket 'avatars'.")
			break
		}

		log.Printf("Waiting for S3 (MinIO) to be ready... (%d/30)", i+1)
		time.Sleep(2 * time.Second)
	}

	return svc
}
func (s *StorageService) UploadAvatar(username string) string {
	// Generate a fake image (random colored pixel)
	// In production, this would be the downloaded profile picture
	dummyImage := []byte(fmt.Sprintf("Fake image content for %s", username))
	key := fmt.Sprintf("%s.txt", username) // Using .txt for simplicity, usually .jpg

	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(dummyImage),
		ContentType: aws.String("text/plain"),
	})

	if err != nil {
		log.Printf("Failed to upload avatar to S3: %v", err)
		return ""
	}

	// Return the public URL
	return fmt.Sprintf("http://localhost:9000/%s/%s", s.Bucket, key)
}
func main() {
	// Start a tiny web server for Prometheus in the background
	go func() {
        http.Handle("/metrics", promhttp.Handler())
        if err := http.ListenAndServe(":8081", nil); err != nil {
             log.Printf("Metrics server stopped: %v", err)
        }
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
	storage := NewStorageService()
	log.Println("AWS S3 (MinIO) Client Initialized")
	ctx := context.Background()
	fmt.Println("Scraper Service Started! Generating profiles...")
	// 4. The Infinite Generation Loop
	for {
		// A. Generate a profile
		profile := generateSmartProfile()

		url := storage.UploadAvatar(profile.Username)
		profile.AvatarURL = url
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
