package repo

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"gateway/internal/model"
)

// HistoryItem is a flat row joining an interaction (view / like) with the
// minimum fields needed to render it in the UI. The repo emits this directly
// so the service layer doesn't need to N+1 fetch each document.
type HistoryItem struct {
	InteractionID uint      // ID of the view/like row, useful for keyed lists
	InteractedAt  time.Time // view's updated_at, like's created_at
	DocID         uint
	Title         *string
	OriginalFile  string
	Visibility    string
	LabID         *uint
	LabName       *string
	EnrichStatus  string
}

type DocumentInteractionRepository interface {
	// RecordView either inserts a new view row (and bumps documents.view_count
	// in the same transaction) or, if the same user has viewed this doc within
	// the throttle window, refreshes the existing row's updated_at without
	// double-counting. Returns true when a new row was inserted.
	RecordView(ctx context.Context, userID, docID uint, throttle time.Duration) (inserted bool, err error)

	// SetLike inserts a like row and bumps documents.like_count atomically.
	// If the user already has an active like for this doc the call is a no-op
	// (returns alreadyLiked=true). Soft-deleted prior likes are revived.
	// likeCount is the document's like_count after the call.
	SetLike(ctx context.Context, userID, docID uint) (alreadyLiked bool, likeCount uint, err error)

	// ClearLike soft-deletes the user's active like and decrements documents.like_count.
	// If no active like existed the call is a no-op (returns notLiked=true).
	// likeCount is the document's like_count after the call.
	ClearLike(ctx context.Context, userID, docID uint) (notLiked bool, likeCount uint, err error)

	// IsLikedBy reports whether the user currently has an active like for the doc.
	IsLikedBy(ctx context.Context, userID, docID uint) (bool, error)

	// AreLikedBy reports, for each docID, whether the user currently likes it.
	// The returned map only contains entries the user has liked.
	AreLikedBy(ctx context.Context, userID uint, docIDs []uint) (map[uint]bool, error)

	// ListViewHistory / ListLikeHistory return the user's interaction history
	// joined with the document's display fields. Soft-deleted documents are
	// excluded — if the underlying doc is gone, the entry disappears too.
	ListViewHistory(ctx context.Context, userID uint, offset, limit int) ([]HistoryItem, int64, error)
	ListLikeHistory(ctx context.Context, userID uint, offset, limit int) ([]HistoryItem, int64, error)
}

type documentInteractionRepo struct {
	db *gorm.DB
}

func NewDocumentInteractionRepo(db *gorm.DB) DocumentInteractionRepository {
	return &documentInteractionRepo{db: db}
}

func (r *documentInteractionRepo) RecordView(ctx context.Context, userID, docID uint, throttle time.Duration) (bool, error) {
	var inserted bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If a recent view exists within the window, just refresh updated_at.
		var recent model.DocumentView
		err := tx.
			Where("user_id = ? AND document_id = ? AND updated_at > ?", userID, docID, time.Now().Add(-throttle)).
			Order("updated_at DESC").
			First(&recent).Error
		if err == nil {
			return tx.Model(&recent).UpdateColumn("updated_at", time.Now()).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		v := &model.DocumentView{UserID: userID, DocumentID: docID}
		if err := tx.Create(v).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Document{}).
			Where("id = ?", docID).
			UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
			return err
		}
		inserted = true
		return nil
	})
	return inserted, err
}

func (r *documentInteractionRepo) SetLike(ctx context.Context, userID, docID uint) (bool, uint, error) {
	var (
		alreadyLiked bool
		likeCount    uint
	)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Active like? No-op (still load count to return).
		var active model.DocumentLike
		err := tx.Where("user_id = ? AND document_id = ?", userID, docID).First(&active).Error
		if err == nil {
			alreadyLiked = true
			return tx.Model(&model.Document{}).
				Where("id = ?", docID).
				Select("like_count").
				Scan(&likeCount).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Revive a soft-deleted prior like if one exists, otherwise insert fresh.
		// Either path uses the partial unique index — which excludes soft-deleted
		// rows — so insertion never collides with a tombstone.
		var prior model.DocumentLike
		err = tx.Unscoped().
			Where("user_id = ? AND document_id = ? AND deleted_at IS NOT NULL", userID, docID).
			First(&prior).Error
		switch {
		case err == nil:
			if err := tx.Unscoped().Model(&prior).Updates(map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
				"created_at": time.Now(),
			}).Error; err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			like := &model.DocumentLike{UserID: userID, DocumentID: docID}
			if err := tx.Create(like).Error; err != nil {
				return err
			}
		default:
			return err
		}

		if err := tx.Model(&model.Document{}).
			Where("id = ?", docID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			return err
		}
		return tx.Model(&model.Document{}).
			Where("id = ?", docID).
			Select("like_count").
			Scan(&likeCount).Error
	})
	return alreadyLiked, likeCount, err
}

func (r *documentInteractionRepo) ClearLike(ctx context.Context, userID, docID uint) (bool, uint, error) {
	var (
		notLiked  bool
		likeCount uint
	)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND document_id = ?", userID, docID).
			Delete(&model.DocumentLike{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			notLiked = true
			return tx.Model(&model.Document{}).
				Where("id = ?", docID).
				Select("like_count").
				Scan(&likeCount).Error
		}
		// GREATEST keeps the column non-negative even if a race ever lets a
		// decrement run before its matching increment.
		if err := tx.Model(&model.Document{}).
			Where("id = ?", docID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count, 1) - 1")).Error; err != nil {
			return err
		}
		return tx.Model(&model.Document{}).
			Where("id = ?", docID).
			Select("like_count").
			Scan(&likeCount).Error
	})
	return notLiked, likeCount, err
}

func (r *documentInteractionRepo) IsLikedBy(ctx context.Context, userID, docID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.DocumentLike{}).
		Where("user_id = ? AND document_id = ?", userID, docID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *documentInteractionRepo) AreLikedBy(ctx context.Context, userID uint, docIDs []uint) (map[uint]bool, error) {
	out := map[uint]bool{}
	if len(docIDs) == 0 {
		return out, nil
	}
	var liked []uint
	if err := r.db.WithContext(ctx).
		Model(&model.DocumentLike{}).
		Where("user_id = ? AND document_id IN ?", userID, docIDs).
		Pluck("document_id", &liked).Error; err != nil {
		return nil, err
	}
	for _, id := range liked {
		out[id] = true
	}
	return out, nil
}

// docToHistoryItem flattens a Preload-loaded Document (with Lab) onto the row
// shape the service exposes. The interaction's id/timestamp are passed in by
// the caller because they live on the view/like row, not the document.
func docToHistoryItem(interactionID uint, interactedAt time.Time, doc *model.Document) HistoryItem {
	var labName *string
	if doc.Lab != nil {
		name := doc.Lab.Name
		labName = &name
	}
	return HistoryItem{
		InteractionID: interactionID,
		InteractedAt:  interactedAt,
		DocID:         doc.ID,
		Title:         doc.Title,
		OriginalFile:  doc.OriginalFileName,
		Visibility:    doc.Visibility,
		LabID:         doc.LabID,
		LabName:       labName,
		EnrichStatus:  doc.EnrichStatus,
	}
}

func (r *documentInteractionRepo) ListViewHistory(ctx context.Context, userID uint, offset, limit int) ([]HistoryItem, int64, error) {
	tx := r.db.WithContext(ctx).
		Model(&model.DocumentView{}).
		Joins("JOIN documents d ON d.id = document_views.document_id AND d.deleted_at IS NULL").
		Where("document_views.user_id = ?", userID)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var views []model.DocumentView
	if err := tx.Preload("Document.Lab").
		Order("document_views.updated_at DESC, document_views.id DESC").
		Offset(offset).Limit(limit).
		Find(&views).Error; err != nil {
		return nil, 0, err
	}

	items := make([]HistoryItem, len(views))
	for i := range views {
		items[i] = docToHistoryItem(views[i].ID, views[i].UpdatedAt, &views[i].Document)
	}
	return items, total, nil
}

func (r *documentInteractionRepo) ListLikeHistory(ctx context.Context, userID uint, offset, limit int) ([]HistoryItem, int64, error) {
	tx := r.db.WithContext(ctx).
		Model(&model.DocumentLike{}).
		Joins("JOIN documents d ON d.id = document_likes.document_id AND d.deleted_at IS NULL").
		Where("document_likes.user_id = ?", userID)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var likes []model.DocumentLike
	if err := tx.Preload("Document.Lab").
		Order("document_likes.created_at DESC, document_likes.id DESC").
		Offset(offset).Limit(limit).
		Find(&likes).Error; err != nil {
		return nil, 0, err
	}

	items := make([]HistoryItem, len(likes))
	for i := range likes {
		items[i] = docToHistoryItem(likes[i].ID, likes[i].CreatedAt, &likes[i].Document)
	}
	return items, total, nil
}
