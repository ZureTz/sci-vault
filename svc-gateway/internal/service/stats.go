package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
	"gateway/pkg/cache"
	"gateway/pkg/storage"
)

const (
	dashboardStatsTTL = 1 * time.Hour
	dashboardDays     = 30
	topListLimit      = 5
)

func dashboardStatsKey(userID uint) string {
	return fmt.Sprintf("stats:dashboard:%d", userID)
}

func labDashboardStatsKey(labID uint) string {
	return fmt.Sprintf("stats:lab:dashboard:%d", labID)
}

type StatsService struct {
	repo          repo.StatsRepository
	labRepo       repo.LabRepository
	cacheConn     *cache.CacheConnector
	storageClient *storage.Client
}

func NewStatsService(
	repo repo.StatsRepository,
	labRepo repo.LabRepository,
	cacheConn *cache.CacheConnector,
	storageClient *storage.Client,
) *StatsService {
	return &StatsService{
		repo:          repo,
		labRepo:       labRepo,
		cacheConn:     cacheConn,
		storageClient: storageClient,
	}
}

func (s *StatsService) GetMyDashboardStats(ctx context.Context, userID uint) (*dto.MyDashboardStatsResponse, error) {
	key := dashboardStatsKey(userID)
	if cached, err := s.cacheConn.Get(ctx, key); err == nil {
		var resp dto.MyDashboardStatsResponse
		if json.Unmarshal([]byte(cached), &resp) == nil {
			return &resp, nil
		}
	}

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

	totalLikes, err := s.repo.TotalLikesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to sum likes: %w", err)
	}

	recentDocs, err := s.repo.RecentByUserID(ctx, userID, topListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent docs: %w", err)
	}

	uploadsByDay, err := s.repo.UploadsByDayForUser(ctx, userID, dashboardDays)
	if err != nil {
		return nil, fmt.Errorf("failed to bucket uploads: %w", err)
	}

	viewsByDay, err := s.repo.ViewsByDayForUser(ctx, userID, dashboardDays)
	if err != nil {
		return nil, fmt.Errorf("failed to bucket views: %w", err)
	}

	likesByDay, err := s.repo.LikesByDayForUser(ctx, userID, dashboardDays)
	if err != nil {
		return nil, fmt.Errorf("failed to bucket likes: %w", err)
	}

	formats, err := s.repo.FormatDistributionForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute format distribution: %w", err)
	}

	topViewed, err := s.repo.TopViewedByUser(ctx, userID, topListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top viewed: %w", err)
	}

	resp := &dto.MyDashboardStatsResponse{
		TotalDocuments:     totalDocs,
		TotalStorage:       totalStorage,
		TotalViews:         totalViews,
		TotalLikes:         totalLikes,
		StatusBreakdown:    toStatusBreakdown(statusCounts),
		RecentDocuments:    toRecentDocuments(recentDocs),
		UploadsByDay:       zeroFillDays(uploadsByDay, dashboardDays),
		ViewsByDay:         zeroFillDays(viewsByDay, dashboardDays),
		LikesByDay:         zeroFillDays(likesByDay, dashboardDays),
		FormatDistribution: toFormatBuckets(formats),
		TopViewed:          toTopDocuments(topViewed),
	}

	if data, err := json.Marshal(resp); err == nil {
		if err := s.cacheConn.Set(ctx, key, string(data), dashboardStatsTTL); err != nil {
			slog.Warn("failed to cache dashboard stats", "userID", userID, "err", err)
		}
	}

	return resp, nil
}

func (s *StatsService) GetLabDashboardStats(ctx context.Context, userID, labID uint) (*dto.LabDashboardStatsResponse, error) {
	if _, err := s.labRepo.FindMember(ctx, labID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_error.ErrNotMember
		}
		return nil, fmt.Errorf("failed to verify lab membership: %w", err)
	}

	key := labDashboardStatsKey(labID)
	if cached, err := s.cacheConn.Get(ctx, key); err == nil {
		var resp dto.LabDashboardStatsResponse
		if json.Unmarshal([]byte(cached), &resp) == nil {
			return &resp, nil
		}
	}

	totalDocs, err := s.repo.CountByLab(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to count lab documents: %w", err)
	}

	statusCounts, err := s.repo.CountByStatusForLab(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to count lab status: %w", err)
	}

	totalStorage, err := s.repo.TotalStorageByLab(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to sum lab storage: %w", err)
	}

	totalViews, err := s.repo.TotalViewsByLab(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to sum lab views: %w", err)
	}

	totalLikes, err := s.repo.TotalLikesByLab(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to sum lab likes: %w", err)
	}

	memberCount, err := s.labRepo.CountMembers(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to count lab members: %w", err)
	}

	recentDocs, err := s.repo.RecentByLab(ctx, labID, topListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent lab docs: %w", err)
	}

	uploadsByDay, err := s.repo.UploadsByDayForLab(ctx, labID, dashboardDays)
	if err != nil {
		return nil, fmt.Errorf("failed to bucket lab uploads: %w", err)
	}

	viewsByDay, err := s.repo.ViewsByDayForLab(ctx, labID, dashboardDays)
	if err != nil {
		return nil, fmt.Errorf("failed to bucket lab views: %w", err)
	}

	likesByDay, err := s.repo.LikesByDayForLab(ctx, labID, dashboardDays)
	if err != nil {
		return nil, fmt.Errorf("failed to bucket lab likes: %w", err)
	}

	formats, err := s.repo.FormatDistributionForLab(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute lab format distribution: %w", err)
	}

	contributors, err := s.repo.TopContributorsByLab(ctx, labID, topListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top contributors: %w", err)
	}

	resp := &dto.LabDashboardStatsResponse{
		TotalDocuments:     totalDocs,
		TotalStorage:       totalStorage,
		TotalViews:         totalViews,
		TotalLikes:         totalLikes,
		MemberCount:        memberCount,
		StatusBreakdown:    toStatusBreakdown(statusCounts),
		RecentDocuments:    toRecentDocuments(recentDocs),
		UploadsByDay:       zeroFillDays(uploadsByDay, dashboardDays),
		ViewsByDay:         zeroFillDays(viewsByDay, dashboardDays),
		LikesByDay:         zeroFillDays(likesByDay, dashboardDays),
		FormatDistribution: toFormatBuckets(formats),
		TopContributors:    s.toContributors(contributors),
	}

	if data, err := json.Marshal(resp); err == nil {
		if err := s.cacheConn.Set(ctx, key, string(data), dashboardStatsTTL); err != nil {
			slog.Warn("failed to cache lab dashboard stats", "labID", labID, "err", err)
		}
	}

	return resp, nil
}

func toStatusBreakdown(rows []repo.EnrichStatusCount) dto.StatusBreakdown {
	var b dto.StatusBreakdown
	for _, sc := range rows {
		switch sc.EnrichStatus {
		case model.EnrichStatusNotStarted:
			b.NotStarted = sc.Count
		case model.EnrichStatusPending:
			b.Pending = sc.Count
		case model.EnrichStatusProcessing:
			b.Processing = sc.Count
		case model.EnrichStatusDone:
			b.Done = sc.Count
		case model.EnrichStatusFailed:
			b.Failed = sc.Count
		}
	}
	return b
}

func toRecentDocuments(docs []model.Document) []dto.RecentDocument {
	out := make([]dto.RecentDocument, 0, len(docs))
	for _, d := range docs {
		out = append(out, dto.RecentDocument{
			ID:               d.ID,
			Title:            d.Title,
			OriginalFileName: d.OriginalFileName,
			FileSize:         d.FileSize,
			EnrichStatus:     d.EnrichStatus,
			CreatedAt:        d.CreatedAt,
		})
	}
	return out
}

func toFormatBuckets(rows []repo.FormatCount) []dto.FormatBucket {
	out := make([]dto.FormatBucket, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.FormatBucket{ContentType: r.ContentType, Count: r.Count})
	}
	return out
}

func toTopDocuments(docs []model.Document) []dto.TopDocument {
	out := make([]dto.TopDocument, 0, len(docs))
	for _, d := range docs {
		out = append(out, dto.TopDocument{
			ID:               d.ID,
			Title:            d.Title,
			OriginalFileName: d.OriginalFileName,
			ViewCount:        int64(d.ViewCount),
			LikeCount:        int64(d.LikeCount),
		})
	}
	return out
}

func (s *StatsService) toContributors(rows []repo.ContributorCount) []dto.Contributor {
	out := make([]dto.Contributor, 0, len(rows))
	for _, r := range rows {
		var avatarURL *string
		if r.AvatarKey != nil {
			url := s.storageClient.PublicObjectURL(*r.AvatarKey)
			avatarURL = &url
		}
		out = append(out, dto.Contributor{
			UserID:    r.UserID,
			Username:  r.Username,
			Nickname:  r.Nickname,
			AvatarURL: avatarURL,
			DocCount:  r.Count,
		})
	}
	return out
}

// zeroFillDays returns a dense series with one entry per UTC day for the last
// `days` calendar days (oldest first, today last). Postgres returns only days
// that had activity; this fills the gaps with zero so the chart x-axis is
// continuous regardless of the data.
func zeroFillDays(rows []repo.DayCount, days int) []dto.DayCount {
	byDay := make(map[string]int64, len(rows))
	for _, r := range rows {
		byDay[r.Day.UTC().Format("2006-01-02")] = r.Count
	}
	out := make([]dto.DayCount, 0, days)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := days - 1; i >= 0; i-- {
		d := today.Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		out = append(out, dto.DayCount{Date: d, Count: byDay[d]})
	}
	return out
}
