package repository

import (
	"context"

	pb "github.com/hammo/influScope/gen/analytics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcAnalyticsClient struct {
	conn   *grpc.ClientConn
	client pb.AnalyticsServiceClient
}

func NewGRPCAnalyticsClient(target string) (*grpcAnalyticsClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &grpcAnalyticsClient{
		conn:   conn,
		client: pb.NewAnalyticsServiceClient(conn),
	}, nil
}

func (g *grpcAnalyticsClient) GetEngagement(ctx context.Context, username string, followers int, platform string) (float64, error) {
	resp, err := g.client.CalculateEngagement(ctx, &pb.EngagementRequest{
		Username:  username,
		Followers: int64(followers),
		Platform:  platform,
	})
	if err != nil {
		return 0, err
	}
	return resp.EngagementRate, nil
}

func (g *grpcAnalyticsClient) Close() error {
	return g.conn.Close()
}
