package main

import (
	"log"

	"github.com/hammo/influScope/analytics/internal/metrics"
	"github.com/hammo/influScope/analytics/internal/service"
	transport "github.com/hammo/influScope/analytics/internal/transport/grpc"
)

func main() {
	// 1. Initialize Metrics
	metricsSvc := metrics.NewPrometheusMetrics()
	go metricsSvc.StartServer(":8084")

	// 2. Initialize Business Logic
	calculatorSvc := service.NewAnalyticsCalculator()

	// 3. Initialize and Start gRPC Server
	grpcServer := transport.NewServer(calculatorSvc, metricsSvc)

	if err := grpcServer.Start(":50051"); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
