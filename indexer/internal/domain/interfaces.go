package domain

import (
	"context"

	"github.com/hammo/influScope/pkg/models"
)

// Message abstracts the underlying broker's delivery mechanism
type Message interface {
	Body() []byte
	Ack() error
}

// MessageConsumer handles pulling messages from the broker
type MessageConsumer interface {
	Next(ctx context.Context) (Message, error)
	Close() error
}

// AnalyticsClient handles gRPC requests
type AnalyticsClient interface {
	GetEngagement(ctx context.Context, username string, followers int, platform string) (float64, error)
	Close() error
}

// SearchRepository handles saving to Elasticsearch
type SearchRepository interface {
	IndexProfile(ctx context.Context, profile *models.Influencer) error
}

// MetricsTracker handles Prometheus counters
type MetricsTracker interface {
	IncIndexed()
	IncError()
}
