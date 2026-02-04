package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/hammo/influScope/gen/analytics"
)

func TestCalculateEngagement(t *testing.T) {
	tests := []struct {
		name       string
		request    *pb.EngagementRequest
		expectFunc func(resp *pb.EngagementResponse, err error) bool
	}{
		{
			name: "Standard platform with less than 1M followers",
			request: &pb.EngagementRequest{
				Platform:  "Instagram",
				Username:  "user1",
				Followers: 500000,
			},
			expectFunc: func(resp *pb.EngagementResponse, err error) bool {
				return err == nil && resp.EngagementRate >= 3.0 && resp.EngagementRate < 5.0
			},
		},
		{
			name: "TikTok platform with less than 1M followers",
			request: &pb.EngagementRequest{
				Platform:  "TikTok",
				Username:  "user1",
				Followers: 500000,
			},
			expectFunc: func(resp *pb.EngagementResponse, err error) bool {
				return err == nil && resp.EngagementRate >= 6.0 && resp.EngagementRate < 8.0
			},
		},
		{
			name: "High followers on standard platform",
			request: &pb.EngagementRequest{
				Platform:  "Instagram",
				Username:  "user2",
				Followers: 2000000,
			},
			expectFunc: func(resp *pb.EngagementResponse, err error) bool {
				return err == nil && resp.EngagementRate >= 1.5 && resp.EngagementRate < 3.5
			},
		},
		{
			name: "High followers on TikTok platform",
			request: &pb.EngagementRequest{
				Platform:  "TikTok",
				Username:  "user2",
				Followers: 2000000,
			},
			expectFunc: func(resp *pb.EngagementResponse, err error) bool {
				return err == nil && resp.EngagementRate >= 3.0 && resp.EngagementRate < 5.0
			},
		},
		{
			name: "Zero followers",
			request: &pb.EngagementRequest{
				Platform:  "Instagram",
				Username:  "user3",
				Followers: 0,
			},
			expectFunc: func(resp *pb.EngagementResponse, err error) bool {
				baseRate := 3.0
				return err == nil && resp.EngagementRate >= baseRate && resp.EngagementRate < (baseRate+2.0)
			},
		},
		{
			name: "Unknown platform",
			request: &pb.EngagementRequest{
				Platform:  "UnknownPlatform",
				Username:  "user4",
				Followers: 10000,
			},
			expectFunc: func(resp *pb.EngagementResponse, err error) bool {
				baseRate := 3.0
				return err == nil && resp.EngagementRate >= baseRate && resp.EngagementRate < (baseRate+2.0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulating the server and context setup
			s := &server{}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := s.CalculateEngagement(ctx, tt.request)

			if !tt.expectFunc(resp, err) {
				t.Errorf("test %q failed: got %v, error: %v", tt.name, resp, err)
			}
		})
	}
}
