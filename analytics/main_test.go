package main

import (
	"context"
	"testing"

	pb "github.com/hammo/influScope/gen/analytics"
)

func TestCalculateEngagement(t *testing.T) {
	tests := []struct {
		name      string
		request   *pb.EngagementRequest
		wantMin   float64
		wantMax   float64
		expectErr bool
	}{
		{
			name: "TikTok user over 1M followers",
			request: &pb.EngagementRequest{
				Platform:  "TikTok",
				Username:  "test_user",
				Followers: 2000000,
			},
			wantMin:   3.0,
			wantMax:   8.0,
			expectErr: false,
		},
		{
			name: "TikTok user under 1M followers",
			request: &pb.EngagementRequest{
				Platform:  "TikTok",
				Username:  "small_user",
				Followers: 100000,
			},
			wantMin:   6.0,
			wantMax:   8.0,
			expectErr: false,
		},
		{
			name: "Instagram user over 1M followers",
			request: &pb.EngagementRequest{
				Platform:  "Instagram",
				Username:  "influencer1",
				Followers: 1500000,
			},
			wantMin:   1.5,
			wantMax:   5.0,
			expectErr: false,
		},
		{
			name: "Instagram user below 1M followers",
			request: &pb.EngagementRequest{
				Platform:  "Instagram",
				Username:  "creator2",
				Followers: 500000,
			},
			wantMin:   3.0,
			wantMax:   5.0,
			expectErr: false,
		},
		{
			name: "Edge case with zero followers",
			request: &pb.EngagementRequest{
				Platform:  "YouTube",
				Username:  "new_user",
				Followers: 0,
			},
			wantMin:   3.0,
			wantMax:   5.0,
			expectErr: false,
		},
		{
			name: "Unsupported platform",
			request: &pb.EngagementRequest{
				Platform:  "Unknown",
				Username:  "test_user",
				Followers: 1000,
			},
			wantMin:   3.0,
			wantMax:   5.0,
			expectErr: false,
		},
	}

	s := &server{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.CalculateEngagement(context.Background(), tt.request)

			if (err != nil) != tt.expectErr {
				t.Errorf("CalculateEngagement() error = %v, expected error = %v", err, tt.expectErr)
				return
			}

			if resp != nil {
				if resp.EngagementRate < tt.wantMin || resp.EngagementRate > tt.wantMax {
					t.Errorf("CalculateEngagement() EngagementRate = %v, want between %v and %v", resp.EngagementRate, tt.wantMin, tt.wantMax)
				}
			}
		})
	}
}
