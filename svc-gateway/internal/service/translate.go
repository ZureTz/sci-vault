package service

import (
	"context"
	"io"

	"gateway/pkg/grpc_client"
)

type translateService struct {
	recommenderClient *grpc_client.RecommenderClient
}

func NewTranslateService(client *grpc_client.RecommenderClient) *translateService {
	return &translateService{recommenderClient: client}
}

func (s *translateService) TranslateStream(ctx context.Context, text, targetLang string, onChunk func(chunk string) error) error {
	stream, err := s.recommenderClient.TranslateTextStream(ctx, text, targetLang)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if cbErr := onChunk(resp.Chunk); cbErr != nil {
			return cbErr
		}
	}

	return nil
}
