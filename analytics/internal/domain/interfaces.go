package domain

import "context"

// MetricsTracker defines how we track analytics observability
type MetricsTracker interface {
	StartTimer() func() // Returns a function to defer stopping the timer
	IncEngagementRequest(platform string)
}

// EngagementCalculator defines the pure business logic contract
type EngagementCalculator interface {
	Calculate(ctx context.Context, platform string, followers int64) float64
}
