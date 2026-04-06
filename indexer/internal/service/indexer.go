package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hammo/influScope/indexer/internal/domain"

	"github.com/hammo/influScope/pkg/models"
)

type IndexerService struct {
	consumer  domain.MessageConsumer
	analytics domain.AnalyticsClient
	search    domain.SearchRepository
	metrics   domain.MetricsTracker
}

func NewIndexerService(c domain.MessageConsumer, a domain.AnalyticsClient, s domain.SearchRepository, m domain.MetricsTracker) *IndexerService {
	return &IndexerService{
		consumer:  c,
		analytics: a,
		search:    s,
		metrics:   m,
	}
}

func (s *IndexerService) Start(ctx context.Context) {
	log.Println("Indexer listening for profiles...")

	for {
		msg, err := s.consumer.Next(ctx)
		if err != nil {
			log.Printf("Consumer error: %v", err)
			continue
		}

		var influencer models.Influencer
		if err := json.Unmarshal(msg.Body(), &influencer); err != nil {
			log.Printf("JSON Decode Error: %v", err)
			msg.Ack() // Discard bad messages
			continue
		}

		// 1. gRPC Enrichment
		grpcCtx, cancel := context.WithTimeout(ctx, time.Second)
		rate, err := s.analytics.GetEngagement(grpcCtx, influencer.Username, influencer.Followers, influencer.Platform)
		cancel()

		if err != nil {
			log.Printf("Analytics Service failed: %v", err)
			influencer.EngagementRate = 0.0
		} else {
			influencer.EngagementRate = rate
		}

		// 2. Index to Elasticsearch
		if err := s.search.IndexProfile(ctx, &influencer); err != nil {
			log.Printf("Elastic Error: %v", err)
			s.metrics.IncError()
			// Deliberately NOT acking here to allow broker requeue/DLX strategies
			continue
		}

		// 3. Complete and Metrics
		if err := msg.Ack(); err != nil {
			log.Printf("Failed to ACK message: %v", err)
		} else {
			s.metrics.IncIndexed()
			fmt.Print(".")
		}
	}
}
