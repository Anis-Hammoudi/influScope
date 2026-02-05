package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	pb "github.com/hammo/influScope/gen/analytics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// --- METRICS DEFINITIONS ---
var (
	engagementRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_engagement_requests_total",
			Help: "Total number of engagement calculation requests received",
		},
		[]string{"platform"}, // Label by platform (TikTok, Instagram, etc.)
	)

	calculationDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analytics_calculation_duration_seconds",
			Help:    "Time taken to calculate engagement rate",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(engagementRequests)
	prometheus.MustRegister(calculationDuration)
}

type server struct {
	pb.UnimplementedAnalyticsServiceServer
}

func (s *server) CalculateEngagement(ctx context.Context, req *pb.EngagementRequest) (*pb.EngagementResponse, error) {
	// Start timer
	timer := prometheus.NewTimer(calculationDuration)
	defer timer.ObserveDuration()

	// Increment request counter (labeled by platform)
	engagementRequests.WithLabelValues(req.Platform).Inc()

	// Simulate complex logic: TikTok usually has higher engagement than Instagram
	baseRate := 3.0
	if req.Platform == "TikTok" {
		baseRate = 6.0
	}

	// Simulate a calculation based on followers
	followerFactor := 1.0
	if req.Followers > 1000000 {
		followerFactor = 0.5 // Big accounts have lower engagement
	}

	// Add some randomness
	finalRate := (baseRate * followerFactor) + (rand.Float64() * 2.0)
	log.Printf("Engagement Rate for %s on %s is %.2f\n", req.Username, req.Platform, finalRate)

	return &pb.EngagementResponse{
		EngagementRate: finalRate,
	}, nil
}

func main() {
	// 1. Start Prometheus HTTP Server (Background Routine)
	// We use port 8084 inside the container, but exposing /metrics on a specific path
	go func() {
		// Note: Since gRPC uses 50051, we can use a different port or the same if multiplexing (easier to use different port for simplicity)
		// However, looking at your docker-compose, 'analytics' maps 8084:8084.
		// But your code listens on 50051 for gRPC.
		// Let's expose metrics on :8084 so it matches your Docker Compose port mapping if that was intended for HTTP.
		metricsPort := ":8084"
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server listening on %s", metricsPort)
		if err := http.ListenAndServe(metricsPort, nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// 2. Start gRPC Server
	grpcPort := ":50051"
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAnalyticsServiceServer(s, &server{})

	log.Printf("Analytics Service (gRPC) running on %s", grpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
