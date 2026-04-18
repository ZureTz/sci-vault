package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
	"gateway/pkg/grpc_client"
)

const (
	defaultSearchLimit  = 10
	defaultHistoryLimit = 20
)

type SearchService struct {
	repo              repo.SearchRepository
	labRepo           repo.LabRepository
	recommenderClient *grpc_client.RecommenderClient
}

func NewSearchService(
	repo repo.SearchRepository,
	labRepo repo.LabRepository,
	recommenderClient *grpc_client.RecommenderClient,
) *SearchService {
	return &SearchService{
		repo:              repo,
		labRepo:           labRepo,
		recommenderClient: recommenderClient,
	}
}

func (s *SearchService) SearchDocuments(ctx context.Context, userID uint, q dto.SearchDocumentsQuery) (*dto.SearchDocumentsResponse, error) {
	limit := uint32(q.Limit)
	if limit == 0 {
		limit = defaultSearchLimit
	}

	// If a lab_id is provided, the caller must be a member of that lab;
	// otherwise they could read any lab's shared documents by guessing IDs.
	if q.LabID > 0 {
		if _, err := s.labRepo.FindMember(ctx, q.LabID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, app_error.ErrNotMember
			}
			return nil, fmt.Errorf("failed to check lab membership: %w", err)
		}
	}

	resp, err := s.recommenderClient.SemanticSearch(ctx, q.Query, uint64(userID), uint64(q.LabID), limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search RPC: %w", err)
	}

	results := make([]dto.SearchResultItem, len(resp.Results))
	for i, r := range resp.Results {
		results[i] = dto.SearchResultItem{
			DocID:            uint(r.DocId),
			Title:            r.Title,
			OriginalFileName: r.OriginalFileName,
			Summary:          r.Summary,
			Authors:          r.Authors,
			Tags:             r.Tags,
			Similarity:       r.Similarity,
			MatchType:        int32(r.GetMatchType()),
		}
	}

	// Record history for successful searches only. A failure to persist history
	// must not fail the request — it's a non-essential side effect.
	history := &model.SearchHistory{
		UserID:      userID,
		Query:       strings.TrimSpace(q.Query),
		ResultCount: len(results),
	}
	if q.LabID > 0 {
		labID := q.LabID
		history.LabID = &labID
	}
	if err := s.repo.CreateHistory(ctx, history); err != nil {
		slog.Warn("failed to record search history", "userID", userID, "err", err)
	}

	return &dto.SearchDocumentsResponse{Results: results}, nil
}

func (s *SearchService) ListMyHistory(ctx context.Context, userID uint, limit int) (*dto.ListSearchHistoryResponse, error) {
	if limit <= 0 {
		limit = defaultHistoryLimit
	}
	rows, err := s.repo.FindHistoryByUserID(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list search history: %w", err)
	}
	items := make([]dto.SearchHistoryItem, len(rows))
	for i, r := range rows {
		items[i] = dto.SearchHistoryItem{
			ID:          r.ID,
			Query:       r.Query,
			LabID:       r.LabID,
			ResultCount: r.ResultCount,
			CreatedAt:   r.CreatedAt,
		}
	}
	return &dto.ListSearchHistoryResponse{Items: items}, nil
}

func (s *SearchService) ClearMyHistory(ctx context.Context, userID uint) (int64, error) {
	deleted, err := s.repo.DeleteHistoryByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to clear search history: %w", err)
	}
	return deleted, nil
}
