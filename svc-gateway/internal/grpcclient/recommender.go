package grpcclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "gateway/internal/pb/recommender"
)

// RecommenderClient wraps the gRPC connection to svc-recommender.
type RecommenderClient struct {
	conn   *grpc.ClientConn
	client pb.RecommenderServiceClient
}

// NewRecommenderClient dials the recommender service at addr (e.g. "localhost:50051")
// and returns a ready-to-use client. Call Close when done.
func NewRecommenderClient(addr string) (*RecommenderClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial recommender at %s: %w", addr, err)
	}
	return &RecommenderClient{
		conn:   conn,
		client: pb.NewRecommenderServiceClient(conn),
	}, nil
}

// Health calls the Health RPC on the recommender service.
// A short deadline is applied so a slow peer does not block the gateway's own health route.
func (r *RecommenderClient) Health(ctx context.Context) (*pb.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return r.client.Health(ctx, &pb.HealthRequest{})
}

// Close releases the underlying gRPC connection.
func (r *RecommenderClient) Close() error {
	return r.conn.Close()
}
