package main

import (
	"context"
	"log"

	"github.com/hammo/influScope/indexer/internal/metrics"
	"github.com/hammo/influScope/indexer/internal/repository"
	"github.com/hammo/influScope/indexer/internal/service"
)

const (
	exchangeName = "influencer-events"
	queueName    = "indexer-queue"
	indexName    = "influencers"
)

func main() {
	ctx := context.Background()

	// 1. Initialize Metrics
	metricsSvc := metrics.NewPrometheusMetrics()
	go metricsSvc.StartServer(":8082")

	// 2. Initialize Repositories
	esRepo, err := repository.NewESRepository("http://elasticsearch:9200", indexName)
	if err != nil {
		log.Fatalf("Error connecting to ES: %v", err)
	}
	log.Println("Connected to Elasticsearch!")

	grpcRepo, err := repository.NewGRPCAnalyticsClient("analytics:50051")
	if err != nil {
		log.Fatalf("Error connecting to gRPC: %v", err)
	}
	defer grpcRepo.Close()
	log.Println("Connected to Analytics gRPC Service")

	rmqRepo, err := repository.NewRabbitMQConsumer(ctx, exchangeName, queueName)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer rmqRepo.Close()
	log.Println("Connected to RabbitMQ")

	// 3. Initialize & Start Core Service
	indexerSvc := service.NewIndexerService(rmqRepo, grpcRepo, esRepo, metricsSvc)
	indexerSvc.Start(ctx)
}
