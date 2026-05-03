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

// ListLabDocumentsFilter drives FindByLabID. Visibility and lab_id are
// implicit (always visibility='lab', always the supplied labID), so the
// filter only carries the user-controllable knobs.
type ListLabDocumentsFilter struct {
	Search    string
	Status    string
	SortBy    string
	SortOrder string
	Offset    int
	Limit     int
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	CreateBatch(ctx context.Context, docs []*model.Document) error
	FindByID(ctx context.Context, id uint) (model.Document, error)
	FindByUserID(ctx context.Context, userID uint, filter ListMyDocumentsFilter) ([]model.Document, int64, error)
	FindByLabID(ctx context.Context, labID uint, filter ListLabDocumentsFilter) ([]model.Document, int64, error)
	FindByUserIDAndStatus(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Document, int64, error)
	FindExistingByHash(ctx context.Context, visibility string, userID uint, labID *uint, sha256 string) (model.Document, error)
	FindExistingHashesInSet(ctx context.Context, visibility string, userID uint, labID *uint, hashes []string) ([]string, error)
	FindStaleNotStarted(ctx context.Context, olderThan time.Time, limit int) ([]model.Document, error)
	UpdateVisibility(ctx context.Context, docID, ownerID uint, visibility string, labID *uint) error
	BatchUpdateVisibility(ctx context.Context, docIDs []uint, ownerID uint, visibility string, labID *uint) (int64, error)
	UpdateMetadata(ctx context.Context, docID uint, patch DocumentMetadataPatch) error
	DeleteByID(ctx context.Context, docID uint) error
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
	err := r.db.WithContext(ctx).Preload("Lab").Preload("UploadedBy").Where("id = ?", id).First(&doc).Error
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

// FindByLabID lists every lab-visible document in a lab, regardless of uploader.
// Used by the lab-scope document management page (owner-only). The visibility
// and lab_id predicates are baked in — no caller can widen them.
func (r *documentRepo) FindByLabID(ctx context.Context, labID uint, filter ListLabDocumentsFilter) ([]model.Document, int64, error) {
	var docs []model.Document
	var count int64

	tx := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("visibility = ? AND lab_id = ?", model.DocVisibilityLab, labID)
	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		tx = tx.Where("(title ILIKE ? OR original_file_name ILIKE ?)", pattern, pattern)
	}
	if filter.Status != "" {
		tx = tx.Where("enrich_status = ?", filter.Status)
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

	err := tx.Preload("Lab").Preload("UploadedBy").
		Order(orderBy).Offset(filter.Offset).Limit(filter.Limit).Find(&docs).Error
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

// DocumentMetadataPatch enumerates the columns UpdateMetadata is allowed to
// touch. A nil pointer means "leave this column as-is"; a non-nil pointer
// (including an empty string / zero) is applied as-is. Keeping this as a typed
// struct owned by the repo means column names never cross the service ↔ repo
// boundary as strings, which rules out column-name injection by construction.
type DocumentMetadataPatch struct {
	Title *string
	Year  *int
	DOI   *string
}

// UpdateMetadata patches user-editable metadata fields on a document.
// Only fields set on the patch are written. Authorisation (uploader OR lab
// owner) is enforced by the service layer before this is called.
func (r *documentRepo) UpdateMetadata(ctx context.Context, docID uint, patch DocumentMetadataPatch) error {
	fields := map[string]any{}
	if patch.Title != nil {
		fields["title"] = *patch.Title
	}
	if patch.Year != nil {
		fields["year"] = *patch.Year
	}
	if patch.DOI != nil {
		fields["doi"] = *patch.DOI
	}
	if len(fields) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("id = ?", docID).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteByID soft-deletes a document. Authorisation is enforced by the service
// layer before this is called.
func (r *documentRepo) DeleteByID(ctx context.Context, docID uint) error {
	res := r.db.WithContext(ctx).Where("id = ?", docID).Delete(&model.Document{})
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
		Updates(map[string]any{
			"visibility": visibility,
			"lab_id":     labID,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
