package grpc

import (
	"context"
	"log"
	"net"

	"github.com/hammo/influScope/analytics/internal/domain"
	pb "github.com/hammo/influScope/gen/analytics"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedAnalyticsServiceServer
	calculator domain.EngagementCalculator
	metrics    domain.MetricsTracker
}

func NewServer(calc domain.EngagementCalculator, metrics domain.MetricsTracker) *Server {
	return &Server{
		calculator: calc,
		metrics:    metrics,
	}
}

func (s *Server) CalculateEngagement(ctx context.Context, req *pb.EngagementRequest) (*pb.EngagementResponse, error) {
	// 1. Start Metrics
	stopTimer := s.metrics.StartTimer()
	defer stopTimer()
	s.metrics.IncEngagementRequest(req.Platform)

	// 2. Delegate to Business Logic
	rate := s.calculator.Calculate(ctx, req.Platform, req.Followers)
	log.Printf("Engagement Rate for %s on %s is %.2f\n", req.Username, req.Platform, rate)

	// 3. Return Protobuf Response
	return &pb.EngagementResponse{
		EngagementRate: rate,
	}, nil
}

func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAnalyticsServiceServer(grpcServer, s)

	log.Printf("Analytics Service (gRPC) running on %s", port)
	return grpcServer.Serve(lis)
}
