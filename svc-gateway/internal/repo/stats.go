package repo

import (
	"context"
	"gateway/internal/model"

	"gorm.io/gorm"
)

// EnrichStatusCount holds a single row from a GROUP BY enrich_status aggregation.
type EnrichStatusCount struct {
	EnrichStatus string `gorm:"column:enrich_status"`
	Count        int64  `gorm:"column:count"`
}

type StatsRepository interface {
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CountByStatusForUser(ctx context.Context, userID uint) ([]EnrichStatusCount, error)
	TotalStorageByUser(ctx context.Context, userID uint) (int64, error)
	RecentByUserID(ctx context.Context, userID uint, limit int) ([]model.Document, error)
	TotalViewsByUser(ctx context.Context, userID uint) (int64, error)
}

type statsRepo struct {
	db *gorm.DB
}

func NewStatsRepo(db *gorm.DB) StatsRepository {
	return &statsRepo{db: db}
}

func (r *statsRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	return gorm.G[model.Document](r.db).Where("uploaded_by_user_id = ?", userID).Count(ctx, "*")
}

func (r *statsRepo) CountByStatusForUser(ctx context.Context, userID uint) ([]EnrichStatusCount, error) {
	var results []EnrichStatusCount
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("enrich_status, COUNT(*) as count").
		Where("uploaded_by_user_id = ?", userID).
		Group("enrich_status").
		Scan(&results).Error
	return results, err
}

func (r *statsRepo) TotalStorageByUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(file_size), 0)").
		Where("uploaded_by_user_id = ?", userID).
		Scan(&total).Error
	return total, err
}

func (r *statsRepo) RecentByUserID(ctx context.Context, userID uint, limit int) ([]model.Document, error) {
	return gorm.G[model.Document](r.db).
		Where("uploaded_by_user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(ctx)
}

func (r *statsRepo) TotalViewsByUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(view_count), 0)").
		Where("uploaded_by_user_id = ?", userID).
		Scan(&total).Error
	return total, err
}
