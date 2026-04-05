package service

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestGenerateSmartProfile(t *testing.T) {
	// Initialize service with nil dependencies since we only test generation logic
	metric := prometheus.NewCounter(prometheus.CounterOpts{Name: "test"})
	svc := NewScraperService(nil, nil, metric)

	tests := []struct {
		name          string
		iterations    int
		checkPlatform bool
		checkCategory bool
		checkBounds   bool
	}{
		{
			name:          "Validates Struct Integrity and Platform",
			iterations:    50,
			checkPlatform: true,
		},
		{
			name:          "Validates Category and Bio Correlation",
			iterations:    50,
			checkCategory: true,
		},
		{
			name:        "Validates Follower and Engagement Bounds",
			iterations:  50,
			checkBounds: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.iterations; i++ {
				profile := svc.GenerateSmartProfile()

				if tt.checkPlatform {
					if profile.ID == "" || profile.Username == "" || profile.Bio == "" {
						t.Error("Expected required fields to be populated")
					}
					valid := map[string]bool{"Instagram": true, "TikTok": true, "YouTube": true}
					if !valid[profile.Platform] {
						t.Errorf("Invalid platform generated: %s", profile.Platform)
					}
				}

				if tt.checkCategory {
					expectedHashtag := "#" + profile.Category
					if !strings.Contains(profile.Bio, expectedHashtag) {
						t.Errorf("Logic Fail: Bio '%s' missing hashtag '%s'", profile.Bio, expectedHashtag)
					}
				}

				if tt.checkBounds {
					if profile.Followers < 1000 || profile.Followers > 5000000 {
						t.Errorf("Followers %d out of bounds", profile.Followers)
					}
					if profile.EngagementRate < 1.0 || profile.EngagementRate > 8.0 {
						t.Errorf("Engagement Rate %.2f out of bounds", profile.EngagementRate)
					}
				}
			}
		})
	}
}
