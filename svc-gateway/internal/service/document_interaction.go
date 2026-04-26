package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
)

type DocumentInteractionService struct {
	repo            repo.DocumentRepository
	interactionRepo repo.DocumentInteractionRepository
	labRepo         repo.LabRepository
}

func NewDocumentInteractionService(
	docRepo repo.DocumentRepository,
	interactionRepo repo.DocumentInteractionRepository,
	labRepo repo.LabRepository,
) *DocumentInteractionService {
	return &DocumentInteractionService{
		repo:            docRepo,
		interactionRepo: interactionRepo,
		labRepo:         labRepo,
	}
}

// authorizeRead enforces the same access rules as DocumentService.canAccessDocument:
// uploader, or member of the doc's lab when visibility=lab. Like/unlike must
// require read access — otherwise non-members could probe doc IDs by liking.
func (s *DocumentInteractionService) authorizeRead(ctx context.Context, userID, docID uint) error {
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrInteractionDocNotFound
		}
		return err
	}
	if doc.UploadedByUserID == userID {
		return nil
	}
	if doc.Visibility == model.DocVisibilityLab && doc.LabID != nil {
		if _, err := s.labRepo.FindMember(ctx, *doc.LabID, userID); err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return app_error.ErrInteractionDocNotFound
}

func (s *DocumentInteractionService) Like(ctx context.Context, userID, docID uint) (*dto.LikeStateResponse, error) {
	if err := s.authorizeRead(ctx, userID, docID); err != nil {
		return nil, err
	}
	_, likeCount, err := s.interactionRepo.SetLike(ctx, userID, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to set like: %w", err)
	}
	return &dto.LikeStateResponse{DocID: docID, Liked: true, LikeCount: likeCount}, nil
}

func (s *DocumentInteractionService) Unlike(ctx context.Context, userID, docID uint) (*dto.LikeStateResponse, error) {
	if err := s.authorizeRead(ctx, userID, docID); err != nil {
		return nil, err
	}
	_, likeCount, err := s.interactionRepo.ClearLike(ctx, userID, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear like: %w", err)
	}
	return &dto.LikeStateResponse{DocID: docID, Liked: false, LikeCount: likeCount}, nil
}

func (s *DocumentInteractionService) ListViewHistory(ctx context.Context, userID uint, q dto.ListHistoryQuery) (*dto.ListHistoryResponse, error) {
	page, pageSize := pageDefaults(q.Page, q.PageSize)
	items, total, err := s.interactionRepo.ListViewHistory(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list view history: %w", err)
	}
	return historyResponse(items, total, page, pageSize), nil
}

func (s *DocumentInteractionService) ListLikeHistory(ctx context.Context, userID uint, q dto.ListHistoryQuery) (*dto.ListHistoryResponse, error) {
	page, pageSize := pageDefaults(q.Page, q.PageSize)
	items, total, err := s.interactionRepo.ListLikeHistory(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list like history: %w", err)
	}
	return historyResponse(items, total, page, pageSize), nil
}

func pageDefaults(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	return page, size
}

func historyResponse(items []repo.HistoryItem, total int64, page, pageSize int) *dto.ListHistoryResponse {
	out := &dto.ListHistoryResponse{
		Items:    make([]dto.HistoryItem, 0, len(items)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, it := range items {
		out.Items = append(out.Items, dto.HistoryItem{
			InteractionID: it.InteractionID,
			InteractedAt:  it.InteractedAt,
			DocID:         it.DocID,
			Title:         it.Title,
			OriginalFile:  it.OriginalFile,
			Visibility:    it.Visibility,
			LabID:         it.LabID,
			LabName:       it.LabName,
			EnrichStatus:  it.EnrichStatus,
		})
	}
	return out
}
