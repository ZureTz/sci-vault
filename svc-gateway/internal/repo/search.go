package repo

import (
	"context"

	"gorm.io/gorm"

	"gateway/internal/model"
)

type SearchRepository interface {
	CreateHistory(ctx context.Context, h *model.SearchHistory) error
	FindHistoryByUserID(ctx context.Context, userID uint, limit int) ([]model.SearchHistory, error)
	DeleteHistoryByUserID(ctx context.Context, userID uint) (int64, error)
}

type searchRepo struct {
	db *gorm.DB
}

func NewSearchRepo(db *gorm.DB) SearchRepository {
	return &searchRepo{db: db}
}

func (r *searchRepo) CreateHistory(ctx context.Context, h *model.SearchHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *searchRepo) FindHistoryByUserID(ctx context.Context, userID uint, limit int) ([]model.SearchHistory, error) {
	var items []model.SearchHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *searchRepo) DeleteHistoryByUserID(ctx context.Context, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.SearchHistory{})
	return res.RowsAffected, res.Error
}
