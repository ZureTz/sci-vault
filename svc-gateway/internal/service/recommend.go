package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
	"gateway/pkg/grpc_client"
	"gateway/pkg/storage"
)

const (
	defaultRecommendSimilarLimit = 5
	maxRecommendSimilarLimit     = 20

	defaultRecommendForUserLimit = 20
	maxRecommendForUserLimit     = 50

	defaultRecommendCollaboratorsLimit = 10
	maxRecommendCollaboratorsLimit     = 50

	// Signal-collection budgets. The recommender further weights items by
	// recency within each list, so generous bounds are fine — these are just
	// the absolute caps on payload size per gRPC call.
	recommendForUserLikeBudget   = 20
	recommendForUserViewBudget   = 30
	recommendForUserSearchBudget = 10
)

// RecommendService surfaces document and user recommendations sourced from
// the recommender microservice. Wraps the per-document "similar to this"
// flow, the personalized feed, and the collaborator-suggestion flow.
type RecommendService struct {
	labRepo           repo.LabRepository
	interactionRepo   repo.DocumentInteractionRepository
	searchRepo        repo.SearchRepository
	recommenderClient *grpc_client.RecommenderClient
	storageClient     *storage.Client
}

func NewRecommendService(
	labRepo repo.LabRepository,
	interactionRepo repo.DocumentInteractionRepository,
	searchRepo repo.SearchRepository,
	recommenderClient *grpc_client.RecommenderClient,
	storageClient *storage.Client,
) *RecommendService {
	return &RecommendService{
		labRepo:           labRepo,
		interactionRepo:   interactionRepo,
		searchRepo:        searchRepo,
		recommenderClient: recommenderClient,
		storageClient:     storageClient,
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
		limit = defaultRecommendSimilarLimit
	}
	if limit > maxRecommendSimilarLimit {
		limit = maxRecommendSimilarLimit
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

// RecommendForUser returns a personalized ranked feed for the caller, scoped
// to private-owned + the given lab. Signals are collected from the gateway's
// own DB (likes, views, search history) and forwarded to the recommender,
// which builds a weighted profile vector and runs a single nearest-neighbor
// query. Same lab-membership guard as RecommendSimilar — the recommender
// trusts the gateway to validate user-supplied lab_id.
func (s *RecommendService) RecommendForUser(ctx context.Context, userID uint, q dto.RecommendForUserQuery) (*dto.RecommendForUserResponse, error) {
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
		limit = defaultRecommendForUserLimit
	}
	if limit > maxRecommendForUserLimit {
		limit = maxRecommendForUserLimit
	}

	likedIDs, viewedIDs, recentQueries, err := s.gatherUserSignals(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp, err := s.recommenderClient.RecommendForUser(
		ctx,
		uint64(userID),
		uint64(q.LabID),
		limit,
		likedIDs,
		viewedIDs,
		recentQueries,
	)
	if err != nil {
		return nil, fmt.Errorf("recommend-for-user RPC: %w", err)
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
	return &dto.RecommendForUserResponse{Results: results}, nil
}

// RecommendCollaborators returns lab members whose interest profile is closest
// to the caller's. lab_id is required and must be a lab the caller belongs
// to. Caller-side signals match RecommendForUser exactly (likes + views +
// recent search queries); the recommender builds each candidate's centroid
// from their likes/views directly in SQL.
func (s *RecommendService) RecommendCollaborators(ctx context.Context, userID uint, q dto.RecommendCollaboratorsQuery) (*dto.RecommendCollaboratorsResponse, error) {
	if _, err := s.labRepo.FindMember(ctx, q.LabID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_error.ErrNotMember
		}
		return nil, fmt.Errorf("failed to check lab membership: %w", err)
	}

	limit := uint32(q.Limit)
	if limit == 0 {
		limit = defaultRecommendCollaboratorsLimit
	}
	if limit > maxRecommendCollaboratorsLimit {
		limit = maxRecommendCollaboratorsLimit
	}

	likedIDs, viewedIDs, recentQueries, err := s.gatherUserSignals(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp, err := s.recommenderClient.RecommendCollaborators(
		ctx,
		uint64(userID),
		uint64(q.LabID),
		limit,
		likedIDs,
		viewedIDs,
		recentQueries,
	)
	if err != nil {
		return nil, fmt.Errorf("recommend-collaborators RPC: %w", err)
	}

	results := make([]dto.CollaboratorItem, len(resp.Results))
	for i, r := range resp.Results {
		var avatarURL *string
		if r.AvatarKey != "" {
			url := s.storageClient.PublicObjectURL(r.AvatarKey)
			avatarURL = &url
		}
		results[i] = dto.CollaboratorItem{
			UserID:      uint(r.UserId),
			Username:    r.Username,
			Nickname:    r.Nickname,
			AvatarURL:   avatarURL,
			Similarity:  r.Similarity,
			SignalCount: r.SignalCount,
		}
	}
	return &dto.RecommendCollaboratorsResponse{Results: results}, nil
}

// gatherUserSignals pulls the caller's most-recent likes, views, and search
// queries in parallel — they're independent reads. Shared by RecommendForUser
// and RecommendCollaborators so the two flows always see the same signal
// shape; if you change a budget, both flows pick it up.
func (s *RecommendService) gatherUserSignals(ctx context.Context, userID uint) (likedIDs, viewedIDs []uint64, recentQueries []string, err error) {
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		items, _, err := s.interactionRepo.ListLikeHistory(gctx, userID, 0, recommendForUserLikeBudget)
		if err != nil {
			return fmt.Errorf("list like history: %w", err)
		}
		likedIDs = make([]uint64, len(items))
		for i, it := range items {
			likedIDs[i] = uint64(it.DocID)
		}
		return nil
	})
	g.Go(func() error {
		items, _, err := s.interactionRepo.ListViewHistory(gctx, userID, 0, recommendForUserViewBudget)
		if err != nil {
			return fmt.Errorf("list view history: %w", err)
		}
		viewedIDs = make([]uint64, len(items))
		for i, it := range items {
			viewedIDs[i] = uint64(it.DocID)
		}
		return nil
	})
	g.Go(func() error {
		items, err := s.searchRepo.FindHistoryByUserID(gctx, userID, recommendForUserSearchBudget)
		if err != nil {
			return fmt.Errorf("list search history: %w", err)
		}
		recentQueries = make([]string, 0, len(items))
		for _, it := range items {
			if it.Query != "" {
				recentQueries = append(recentQueries, it.Query)
			}
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}
	return likedIDs, viewedIDs, recentQueries, nil
}
