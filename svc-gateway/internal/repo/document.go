package repo

import (
	"context"
	"time"

	"gateway/internal/model"

	"gorm.io/gorm"
)

// ListMyDocumentsFilter drives the filter/sort/pagination behaviour of
// DocumentRepository.FindByUserID. All fields are optional; empty values mean
// "no filter" (and for sort fields, use the default: created_at DESC).
type ListMyDocumentsFilter struct {
	Search     string // case-insensitive substring match against title / original_file_name
	Status     string // exact enrich_status match
	Visibility string // exact visibility match
	LabID      *uint  // only meaningful when Visibility == "lab"
	SortBy     string // one of: created_at, title, file_size, view_count
	SortOrder  string // asc | desc (default desc)
	Offset     int
	Limit      int
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	CreateBatch(ctx context.Context, docs []*model.Document) error
	FindByID(ctx context.Context, id uint) (model.Document, error)
	FindByUserID(ctx context.Context, userID uint, filter ListMyDocumentsFilter) ([]model.Document, int64, error)
	FindByUserIDAndStatus(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Document, int64, error)
	FindExistingByHash(ctx context.Context, visibility string, userID uint, labID *uint, sha256 string) (model.Document, error)
	FindExistingHashesInSet(ctx context.Context, visibility string, userID uint, labID *uint, hashes []string) ([]string, error)
	FindStaleNotStarted(ctx context.Context, olderThan time.Time, limit int) ([]model.Document, error)
	IncrementViewCount(ctx context.Context, id uint) error
	IncrementLikeCount(ctx context.Context, id uint) error
	UpdateVisibility(ctx context.Context, docID, ownerID uint, visibility string, labID *uint) error
	BatchUpdateVisibility(ctx context.Context, docIDs []uint, ownerID uint, visibility string, labID *uint) (int64, error)
	UpdateMetadata(ctx context.Context, docID, ownerID uint, fields map[string]any) error
	DeleteByID(ctx context.Context, docID, ownerID uint) (model.Document, error)
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

func (r *documentRepo) FindByUserID(ctx context.Context, userID uint, filter ListMyDocumentsFilter) ([]model.Document, int64, error) {
	var docs []model.Document
	var count int64

	tx := r.db.WithContext(ctx).Model(&model.Document{}).Where("uploaded_by_user_id = ?", userID)
	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		tx = tx.Where("(title ILIKE ? OR original_file_name ILIKE ?)", pattern, pattern)
	}
	if filter.Status != "" {
		tx = tx.Where("enrich_status = ?", filter.Status)
	}
	if filter.Visibility != "" {
		tx = tx.Where("visibility = ?", filter.Visibility)
		if filter.Visibility == model.DocVisibilityLab && filter.LabID != nil {
			tx = tx.Where("lab_id = ?", *filter.LabID)
		}
	}
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	orderBy := "created_at DESC, id DESC"
	ascending := filter.SortOrder == "asc"
	switch filter.SortBy {
	case "title":
		if ascending {
			orderBy = "title ASC, id ASC"
		} else {
			orderBy = "title DESC, id DESC"
		}
	case "file_size":
		if ascending {
			orderBy = "file_size ASC, id ASC"
		} else {
			orderBy = "file_size DESC, id DESC"
		}
	case "view_count":
		if ascending {
			orderBy = "view_count ASC, id ASC"
		} else {
			orderBy = "view_count DESC, id DESC"
		}
	case "created_at":
		if ascending {
			orderBy = "created_at ASC, id ASC"
		} else {
			orderBy = "created_at DESC, id DESC"
		}
	}

	err := tx.Preload("Lab").Order(orderBy).Offset(filter.Offset).Limit(filter.Limit).Find(&docs).Error
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

// FindExistingByHash looks up an existing document within the same dedup scope
// (user's private docs when visibility=private; the lab's docs when visibility=lab).
// Returns gorm.ErrRecordNotFound when none exists.
func (r *documentRepo) FindExistingByHash(ctx context.Context, visibility string, userID uint, labID *uint, sha256 string) (model.Document, error) {
	var doc model.Document
	tx := r.db.WithContext(ctx).Where("content_sha256 = ? AND visibility = ?", sha256, visibility)
	switch visibility {
	case model.DocVisibilityPrivate:
		tx = tx.Where("uploaded_by_user_id = ?", userID)
	case model.DocVisibilityLab:
		if labID == nil {
			return doc, gorm.ErrRecordNotFound
		}
		tx = tx.Where("lab_id = ?", *labID)
	default:
		return doc, gorm.ErrRecordNotFound
	}
	err := tx.First(&doc).Error
	return doc, err
}

// FindExistingHashesInSet returns the subset of the given hashes that already
// exist within the same dedup scope (see FindExistingByHash). Used for batch dedup pre-check.
func (r *documentRepo) FindExistingHashesInSet(ctx context.Context, visibility string, userID uint, labID *uint, hashes []string) ([]string, error) {
	if len(hashes) == 0 {
		return nil, nil
	}
	tx := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Where("visibility = ? AND content_sha256 IN ?", visibility, hashes)
	switch visibility {
	case model.DocVisibilityPrivate:
		tx = tx.Where("uploaded_by_user_id = ?", userID)
	case model.DocVisibilityLab:
		if labID == nil {
			return nil, nil
		}
		tx = tx.Where("lab_id = ?", *labID)
	default:
		return nil, nil
	}
	var existing []string
	err := tx.Pluck("content_sha256", &existing).Error
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
	err := tx.Preload("Lab").Order("created_at DESC, id DESC").Offset(offset).Limit(limit).Find(&docs).Error
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
		Updates(map[string]any{
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

// UpdateMetadata patches user-editable metadata fields (title, year, doi) on a
// document the caller owns. The fields map contains only the keys that should
// change — callers must restrict it to safe columns.
func (r *documentRepo) UpdateMetadata(ctx context.Context, docID, ownerID uint, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("id = ? AND uploaded_by_user_id = ?", docID, ownerID).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteByID soft-deletes a document the caller owns and returns the deleted
// row so callers can clean up side effects (e.g. remove the S3 object).
// Returns gorm.ErrRecordNotFound when the caller does not own the document
// or it does not exist.
func (r *documentRepo) DeleteByID(ctx context.Context, docID, ownerID uint) (model.Document, error) {
	var doc model.Document
	if err := r.db.WithContext(ctx).
		Where("id = ? AND uploaded_by_user_id = ?", docID, ownerID).
		First(&doc).Error; err != nil {
		return doc, err
	}
	if err := r.db.WithContext(ctx).Delete(&doc).Error; err != nil {
		return doc, err
	}
	return doc, nil
}

// BatchUpdateVisibility atomically updates multiple documents' visibility and lab_id.
// Returns the number of rows updated. Caller should compare against len(docIDs) to detect partial ownership.
func (r *documentRepo) BatchUpdateVisibility(ctx context.Context, docIDs []uint, ownerID uint, visibility string, labID *uint) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("id IN ? AND uploaded_by_user_id = ?", docIDs, ownerID).
		Updates(map[string]any{
			"visibility": visibility,
			"lab_id":     labID,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
