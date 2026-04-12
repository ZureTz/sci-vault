package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/cache"
)

const dashboardStatsTTL = 1 * time.Hour

func dashboardStatsKey(userID uint) string {
	return fmt.Sprintf("stats:dashboard:%d", userID)
}

type StatsService struct {
	repo      repo.StatsRepository
	cacheConn *cache.CacheConnector
}

func NewStatsService(repo repo.StatsRepository, cacheConn *cache.CacheConnector) *StatsService {
	return &StatsService{repo: repo, cacheConn: cacheConn}
}

func (s *StatsService) GetMyDashboardStats(ctx context.Context, userID uint) (*dto.MyDashboardStatsResponse, error) {
	// Try cache first
	key := dashboardStatsKey(userID)
	if cached, err := s.cacheConn.Get(ctx, key); err == nil {
		var resp dto.MyDashboardStatsResponse
		if json.Unmarshal([]byte(cached), &resp) == nil {
			return &resp, nil
		}
	}

	// Query from DB
	totalDocs, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	statusCounts, err := s.repo.CountByStatusForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}

	totalStorage, err := s.repo.TotalStorageByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to sum storage: %w", err)
	}

	totalViews, err := s.repo.TotalViewsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to sum views: %w", err)
	}

	recentDocs, err := s.repo.RecentByUserID(ctx, userID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent docs: %w", err)
	}

	var breakdown dto.StatusBreakdown
	for _, sc := range statusCounts {
		switch sc.EnrichStatus {
		case model.EnrichStatusNotStarted:
			breakdown.NotStarted = sc.Count
		case model.EnrichStatusPending:
			breakdown.Pending = sc.Count
		case model.EnrichStatusProcessing:
			breakdown.Processing = sc.Count
		case model.EnrichStatusDone:
			breakdown.Done = sc.Count
		case model.EnrichStatusFailed:
			breakdown.Failed = sc.Count
		}
	}

	recent := make([]dto.RecentDocument, 0, len(recentDocs))
	for _, doc := range recentDocs {
		recent = append(recent, dto.RecentDocument{
			ID:               doc.ID,
			Title:            doc.Title,
			OriginalFileName: doc.OriginalFileName,
			FileSize:         doc.FileSize,
			EnrichStatus:     doc.EnrichStatus,
			CreatedAt:        doc.CreatedAt,
		})
	}

	resp := &dto.MyDashboardStatsResponse{
		TotalDocuments:  totalDocs,
		TotalStorage:    totalStorage,
		TotalViews:      totalViews,
		StatusBreakdown: breakdown,
		RecentDocuments: recent,
	}

	// Cache the result
	if data, err := json.Marshal(resp); err == nil {
		if err := s.cacheConn.Set(ctx, key, string(data), dashboardStatsTTL); err != nil {
			slog.Warn("failed to cache dashboard stats", "userID", userID, "err", err)
		}
	}

	return resp, nil
}
