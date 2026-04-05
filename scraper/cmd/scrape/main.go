package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/hammo/influScope/scraper/internal/repository"
	"github.com/hammo/influScope/scraper/internal/service"
)

func main() {
	// Note: rand.Seed is deprecated in Go 1.20+, but perfectly fine to leave for now
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Setup Metrics
	profilesDiscovered := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "influencers_discovered_total",
		Help: "Total number of influencer profiles generated",
	})
	prometheus.MustRegister(profilesDiscovered)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	// 2. Initialize Repositories

	// Pull the endpoint from Docker Compose environment variables
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		// Fallback for when you run it locally outside of Docker
		endpoint = "http://localhost:9000"
	}

	// Initialize S3 with the dynamic endpoint (Named 'storage' to match your service)
	storage, err := repository.NewS3Storage(
		ctx,
		endpoint,
		"avatars",
		"admin",
		"password",
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3: %v", err)
	}
	log.Println("Successfully connected to S3!")

	publisher, err := repository.NewRabbitMQPublisher("influencer-events", 10)
	if err != nil {
		log.Fatalf("Failed to init broker: %v", err)
	}
	defer publisher.Close()

	// 3. Initialize & Run Service
	scraperService := service.NewScraperService(storage, publisher, profilesDiscovered)

	// Handle graceful shutdown
	go scraperService.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down gracefully...")
}
