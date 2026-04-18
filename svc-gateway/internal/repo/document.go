package repo

import (
	"context"
	"time"

	"gateway/internal/model"

	"gorm.io/gorm"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	CreateBatch(ctx context.Context, docs []*model.Document) error
	FindByID(ctx context.Context, id uint) (model.Document, error)
	FindByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Document, int64, error)
	FindByUserIDAndStatus(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Document, int64, error)
	FindPrivateByUserIDAndHash(ctx context.Context, userID uint, sha256 string) (model.Document, error)
	FindPrivateHashesInSet(ctx context.Context, userID uint, hashes []string) ([]string, error)
	FindStaleNotStarted(ctx context.Context, olderThan time.Time, limit int) ([]model.Document, error)
	IncrementViewCount(ctx context.Context, id uint) error
	IncrementLikeCount(ctx context.Context, id uint) error
	UpdateVisibility(ctx context.Context, docID, ownerID uint, visibility string, labID *uint) error
	BatchUpdateVisibility(ctx context.Context, docIDs []uint, ownerID uint, visibility string, labID *uint) (int64, error)
}

type documentRepo struct {
	db *gorm.DB
}

func NewDocumentRepo(db *gorm.DB) DocumentRepository {
	return &documentRepo{db: db}
}

func (r *documentRepo) Create(ctx context.Context, doc *model.Document) error {
	return gorm.G[model.Document](r.db).Create(ctx, doc)
}

func (r *documentRepo) FindByID(ctx context.Context, id uint) (model.Document, error) {
	var doc model.Document
	err := r.db.WithContext(ctx).Preload("Lab").Where("id = ?", id).First(&doc).Error
	return doc, err
}

func (r *documentRepo) FindByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Document, int64, error) {
	var docs []model.Document
	var count int64

	tx := r.db.WithContext(ctx).Model(&model.Document{}).Where("uploaded_by_user_id = ?", userID)
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := tx.Preload("Lab").Order("created_at DESC").Offset(offset).Limit(limit).Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}
	return docs, count, nil
}

// CreateBatch inserts multiple documents in a single statement (GORM batch insert).
// IDs are populated on the input structs on success.
func (r *documentRepo) CreateBatch(ctx context.Context, docs []*model.Document) error {
	if len(docs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&docs).Error
}

// FindPrivateByUserIDAndHash looks up an existing private document owned by userID
// with the given sha256. Returns gorm.ErrRecordNotFound when none exists.
func (r *documentRepo) FindPrivateByUserIDAndHash(ctx context.Context, userID uint, sha256 string) (model.Document, error) {
	var doc model.Document
	err := r.db.WithContext(ctx).
		Where("uploaded_by_user_id = ? AND content_sha256 = ? AND visibility = ?", userID, sha256, model.DocVisibilityPrivate).
		First(&doc).Error
	return doc, err
}

// FindPrivateHashesInSet returns the subset of the given hashes that already
// exist as private documents for this user. Used for batch dedup pre-check.
func (r *documentRepo) FindPrivateHashesInSet(ctx context.Context, userID uint, hashes []string) ([]string, error) {
	if len(hashes) == 0 {
		return nil, nil
	}
	var existing []string
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Where("uploaded_by_user_id = ? AND visibility = ? AND content_sha256 IN ?", userID, model.DocVisibilityPrivate, hashes).
		Pluck("content_sha256", &existing).Error
	return existing, err
}

// FindStaleNotStarted returns docs still in enrich_status="not_started" older than olderThan,
// capped to limit. Used by the re-enrich background job to reschedule presumed-failed enrichments.
func (r *documentRepo) FindStaleNotStarted(ctx context.Context, olderThan time.Time, limit int) ([]model.Document, error) {
	var docs []model.Document
	err := r.db.WithContext(ctx).
		Where("enrich_status = ? AND created_at < ?", model.EnrichStatusNotStarted, olderThan).
		Order("created_at ASC").
		Limit(limit).
		Find(&docs).Error
	return docs, err
}

func (r *documentRepo) FindByUserIDAndStatus(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Document, int64, error) {
	var docs []model.Document
	var count int64

	tx := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("uploaded_by_user_id = ? AND enrich_status = ?", userID, status)
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := tx.Preload("Lab").Order("created_at DESC").Offset(offset).Limit(limit).Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}
	return docs, count, nil
}

func (r *documentRepo) IncrementViewCount(ctx context.Context, id uint) error {
	_, err := gorm.G[model.Document](r.db).Where("id = ?", id).Update(ctx, "view_count", gorm.Expr("view_count + 1"))
	return err
}

func (r *documentRepo) IncrementLikeCount(ctx context.Context, id uint) error {
	_, err := gorm.G[model.Document](r.db).Where("id = ?", id).Update(ctx, "like_count", gorm.Expr("like_count + 1"))
	return err
}

// UpdateVisibility updates a single document's visibility and lab_id.
// The document must be owned by ownerID; if not (or if it doesn't exist), returns gorm.ErrRecordNotFound.
func (r *documentRepo) UpdateVisibility(ctx context.Context, docID, ownerID uint, visibility string, labID *uint) error {
	res := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("id = ? AND uploaded_by_user_id = ?", docID, ownerID).
		Updates(map[string]interface{}{
			"visibility": visibility,
			"lab_id":     labID,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// BatchUpdateVisibility atomically updates multiple documents' visibility and lab_id.
// Returns the number of rows updated. Caller should compare against len(docIDs) to detect partial ownership.
func (r *documentRepo) BatchUpdateVisibility(ctx context.Context, docIDs []uint, ownerID uint, visibility string, labID *uint) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("id IN ? AND uploaded_by_user_id = ?", docIDs, ownerID).
		Updates(map[string]interface{}{
			"visibility": visibility,
			"lab_id":     labID,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
