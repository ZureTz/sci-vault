package model

import (
	"time"

	"gorm.io/gorm"
)

// DocumentView records a single read of a document by a user. Append-only,
// throttled to one row per (user, doc, calendar day) by an application-level
// guard so the table stays bounded. The denormalised counter on documents is
// only bumped when a new row lands here, keeping the count and history aligned.
type DocumentView struct {
	gorm.Model
	UserID     uint `gorm:"not null;index:idx_doc_views_user_created,priority:1,sort:desc"`
	DocumentID uint `gorm:"not null;index"`
	Document   Document
}

// DocumentLike records that a user currently likes a document. Toggle-shaped:
// at most one live row per (user_id, document_id) — guarded by a partial unique
// index on the non-soft-deleted set. CreatedAt is the "liked at" timestamp; an
// unlike soft-deletes the row.
type DocumentLike struct {
	gorm.Model
	UserID     uint `gorm:"not null;index:idx_doc_likes_user_created,priority:1,sort:desc"`
	DocumentID uint `gorm:"not null;index"`
	Document   Document
}

func (*DocumentLike) CustomIndexes() []string {
	return []string{
		// Toggle uniqueness: a given user has at most one *live* like per document.
		// Soft-deleted rows are excluded so a user can re-like after unliking.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_doc_likes_user_doc_active
			ON document_likes (user_id, document_id)
			WHERE deleted_at IS NULL`,
	}
}

// ViewThrottleWindow defines how often a single user's repeat views of the
// same document produce new history rows. Anything finer than this collapses
// into the most recent row's updated_at bump.
const ViewThrottleWindow = 15 * time.Minute
