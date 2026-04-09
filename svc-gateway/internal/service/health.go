package service

import (
	"context"

	"gateway/pkg/grpc_client"
)

type healthService struct {
	recommenderClient *grpc_client.RecommenderClient
}

func NewHealthService(client *grpc_client.RecommenderClient) *healthService {
	return &healthService{recommenderClient: client}
}

func (s *healthService) CheckRecommender(ctx context.Context) (string, string, error) {
	resp, err := s.recommenderClient.Health(ctx)
	if err != nil {
		return "", "", err
	}
	return resp.GetStatus(), resp.GetService(), nil
}
