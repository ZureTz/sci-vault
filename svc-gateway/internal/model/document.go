package model

import (
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// DB enrich_status values (source of truth for persistent state).
// Fine-grained transient states (pending, processing, failed) live in Redis only,
// managed entirely by the Python microservice.
const (
	EnrichStatusNotStarted = "not_started"
	EnrichStatusPending    = "pending"
	EnrichStatusProcessing = "processing"
	EnrichStatusFailed     = "failed"
	EnrichStatusDone       = "done"
)

// Document visibility values.
const (
	DocVisibilityPrivate = "private"
	DocVisibilityLab     = "lab"
)

type Document struct {
	gorm.Model

	// Metadata — filled by Go on upload
	Title            *string // optional; nil means no title provided
	OriginalFileName string  `gorm:"not null"`             // original filename for download
	FileKey          string  `gorm:"not null;uniqueIndex"` // S3 object key: documents/{time}/{hash}
	FileSize         int64   `gorm:"not null"`
	ContentType      string  `gorm:"not null"`
	ContentSHA256    string  `gorm:"type:char(64);index"` // hex sha256 of file bytes; used for dedup
	Year             *int    // publication year, optional at upload time
	DOI              *string // digital Object Identifier, optional at upload time

	// Uploader info
	UploadedByUserID uint `gorm:"not null;index"`
	UploadedBy       User `gorm:"foreignKey:UploadedByUserID"`

	// Lab association
	LabID      *uint  `gorm:"index"`
	Lab        *Lab   `gorm:"foreignKey:LabID"`
	Visibility string `gorm:"not null;default:'private'"` // private | lab

	// Enrichment pipeline status: not_started | pending | processing | done | failed
	EnrichStatus string `gorm:"not null;default:'not_started'"`

	// Enrichment — filled by Python microservice via LLM / embedding model
	Authors   pq.StringArray `gorm:"type:text[]"`
	Summary   *string
	Tags      pq.StringArray   `gorm:"type:text[]"`
	Embedding *pgvector.Vector `gorm:"type:vector(768)"`

	// User interactions
	ViewCount uint `gorm:"default:0;not null"`
	LikeCount uint `gorm:"default:0;not null"`
}

// CustomIndexes returns raw-SQL index definitions for the documents table that
// AutoMigrate cannot express (pgvector HNSW, partial unique indexes). The
// database package discovers this via the HasCustomIndexes interface and runs
// each statement after AutoMigrate. Statements must be idempotent.
func (*Document) CustomIndexes() []string {
	return []string{
		// Vector similarity search over enrichment embeddings.
		`CREATE INDEX IF NOT EXISTS idx_documents_embedding_hnsw
			ON documents USING hnsw (embedding vector_cosine_ops)`,

		// Dedup guard: a user cannot upload the same bytes twice as a private document.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_private_user_sha
			ON documents (uploaded_by_user_id, content_sha256)
			WHERE visibility = 'private' AND deleted_at IS NULL AND content_sha256 <> ''`,

		// Dedup guard: a lab cannot contain two copies of the same bytes, regardless of uploader.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_lab_sha
			ON documents (lab_id, content_sha256)
			WHERE visibility = 'lab' AND lab_id IS NOT NULL AND deleted_at IS NULL AND content_sha256 <> ''`,
	}
}
