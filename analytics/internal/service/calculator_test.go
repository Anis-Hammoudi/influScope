package service

import (
	"context"
	"testing"
)

func TestCalculateEngagement(t *testing.T) {
	tests := []struct {
		name      string
		platform  string
		followers int64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "TikTok user over 1M followers",
			platform:  "TikTok",
			followers: 2000000,
			wantMin:   3.0,
			wantMax:   8.0,
		},
		{
			name:      "TikTok user under 1M followers",
			platform:  "TikTok",
			followers: 100000,
			wantMin:   6.0,
			wantMax:   8.0,
		},
		{
			name:      "Instagram user over 1M followers",
			platform:  "Instagram",
			followers: 1500000,
			wantMin:   1.5,
			wantMax:   5.0,
		},
		{
			name:      "Instagram user below 1M followers",
			platform:  "Instagram",
			followers: 500000,
			wantMin:   3.0,
			wantMax:   5.0,
		},
		{
			name:      "Edge case with zero followers",
			platform:  "YouTube",
			followers: 0,
			wantMin:   3.0,
			wantMax:   5.0,
		},
		{
			name:      "Unsupported platform",
			platform:  "Unknown",
			followers: 1000,
			wantMin:   3.0,
			wantMax:   5.0,
		},
	}

	calc := NewAnalyticsCalculator()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := calc.Calculate(ctx, tt.platform, tt.followers)

			if rate < tt.wantMin || rate > tt.wantMax {
				t.Errorf("Calculate() EngagementRate = %v, want between %v and %v", rate, tt.wantMin, tt.wantMax)
			}
		})
	}
}
