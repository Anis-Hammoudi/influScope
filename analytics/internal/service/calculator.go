package service

import (
	"context"
	"math/rand"
)

type AnalyticsCalculator struct{}

func NewAnalyticsCalculator() *AnalyticsCalculator {
	return &AnalyticsCalculator{}
}

func (s *AnalyticsCalculator) Calculate(ctx context.Context, platform string, followers int64) float64 {
	baseRate := 3.0
	if platform == "TikTok" {
		baseRate = 6.0
	}

	followerFactor := 1.0
	if followers > 1000000 {
		followerFactor = 0.5 // Big accounts have lower engagement
	}

	// Add some randomness
	return (baseRate * followerFactor) + (rand.Float64() * 2.0)
}
