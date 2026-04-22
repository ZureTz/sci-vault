package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
	"gateway/pkg/grpc_client"
)

const (
	defaultRecommendLimit = 5
	maxRecommendLimit     = 20
)

// RecommendService surfaces document recommendations sourced from the
// recommender microservice. Currently only "similar documents" is exposed,
// but this service is the natural home for future recommendation flows
// (e.g. trending, personalised feed).
type RecommendService struct {
	labRepo           repo.LabRepository
	recommenderClient *grpc_client.RecommenderClient
}

func NewRecommendService(
	labRepo repo.LabRepository,
	recommenderClient *grpc_client.RecommenderClient,
) *RecommendService {
	return &RecommendService{
		labRepo:           labRepo,
		recommenderClient: recommenderClient,
	}
}

// RecommendSimilar returns documents most similar to docID, scoped to the
// caller's accessible library (private-owned + the given lab).
//
// The recommender enforces row-level access on the results via its SQL
// access clauses (private docs must be owned by user_id; lab docs must be
// visible in lab_id), so the gateway only needs to guard the user-supplied
// lab_id by verifying membership — otherwise a caller could pass a lab they
// don't belong to and get back that lab's documents.
func (s *RecommendService) RecommendSimilar(ctx context.Context, userID, docID uint, q dto.RecommendSimilarQuery) (*dto.RecommendSimilarResponse, error) {
	if q.LabID > 0 {
		if _, err := s.labRepo.FindMember(ctx, q.LabID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, app_error.ErrNotMember
			}
			return nil, fmt.Errorf("failed to check lab membership: %w", err)
		}
	}

	limit := uint32(q.Limit)
	if limit == 0 {
		limit = defaultRecommendLimit
	}
	if limit > maxRecommendLimit {
		limit = maxRecommendLimit
	}

	resp, err := s.recommenderClient.RecommendSimilar(ctx, uint64(docID), uint64(userID), uint64(q.LabID), limit)
	if err != nil {
		return nil, fmt.Errorf("recommend-similar RPC: %w", err)
	}

	results := make([]dto.SimilarDocumentItem, len(resp.Results))
	for i, r := range resp.Results {
		results[i] = dto.SimilarDocumentItem{
			DocID:            uint(r.DocId),
			Title:            r.Title,
			OriginalFileName: r.OriginalFileName,
			Summary:          r.Summary,
			Authors:          r.Authors,
			Tags:             r.Tags,
			Similarity:       r.Similarity,
		}
	}
	return &dto.RecommendSimilarResponse{Results: results}, nil
}
