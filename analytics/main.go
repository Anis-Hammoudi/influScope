package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	pb "github.com/hammo/influScope/gen/analytics"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedAnalyticsServiceServer
}

func (s *server) CalculateEngagement(ctx context.Context, req *pb.EngagementRequest) (*pb.EngagementResponse, error) {
	rand.Seed(time.Now().UnixNano())

	// Simulate complex logic: TikTok usually has higher engagement than Instagram
	baseRate := 3.0
	if req.Platform == "TikTok" {
		baseRate = 6.0
	}

	// Simulate a calculation based on followers (fewer followers = often higher engagement)
	followerFactor := 1.0
	if req.Followers > 1000000 {
		followerFactor = 0.5 // Big accounts have lower engagement
	}

	// Add some randomness to simulate real data
	finalRate := (baseRate * followerFactor) + (rand.Float64() * 2.0)
	log.Printf("Engagement Rate for %s on %s is %.2f\n", req.Username, req.Platform, finalRate)

	return &pb.EngagementResponse{
		EngagementRate: finalRate,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAnalyticsServiceServer(s, &server{})

	log.Println("Analytics Service (gRPC) running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
