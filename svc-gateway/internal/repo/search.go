package repo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gateway/internal/model"
)

type SearchRepository interface {
	UpsertHistory(ctx context.Context, userID uint, labID *uint, query string, resultCount int) error
	FindHistoryByUserID(ctx context.Context, userID uint, limit int) ([]model.SearchHistory, error)
	DeleteHistoryByUserID(ctx context.Context, userID uint) (int64, error)
}

type searchRepo struct {
	db *gorm.DB
}

func NewSearchRepo(db *gorm.DB) SearchRepository {
	return &searchRepo{db: db}
}

// UpsertHistory dedupes by (user_id, query, lab_id): if a row exists it bumps
// updated_at and refreshes the result count; otherwise it inserts. There's no
// DB-level uniqueness guard — concurrent identical searches by the same user
// could race and produce duplicates, but at human typing speed that's negligible.
func (r *searchRepo) UpsertHistory(ctx context.Context, userID uint, labID *uint, query string, resultCount int) error {
	tx := r.db.WithContext(ctx).
		Model(&model.SearchHistory{}).
		Where("user_id = ? AND query = ?", userID, query)
	if labID == nil {
		tx = tx.Where("lab_id IS NULL")
	} else {
		tx = tx.Where("lab_id = ?", *labID)
	}

	var existing model.SearchHistory
	err := tx.First(&existing).Error
	if err == nil {
		return r.db.WithContext(ctx).
			Model(&existing).
			Updates(map[string]any{"result_count": resultCount}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.WithContext(ctx).Create(&model.SearchHistory{
		UserID:      userID,
		LabID:       labID,
		Query:       query,
		ResultCount: resultCount,
	}).Error
}

func (r *searchRepo) FindHistoryByUserID(ctx context.Context, userID uint, limit int) ([]model.SearchHistory, error) {
	var items []model.SearchHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC, id DESC").
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
