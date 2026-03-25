package model

import (
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Document struct {
	gorm.Model

	// Metadata — filled by Go on upload
	Title            string `gorm:"not null"`
	OriginalFileName string `gorm:"not null"`             // original filename for download
	FileKey          string `gorm:"not null;uniqueIndex"` // S3 object key: documents/{time}/{hash}
	FileSize         int64  `gorm:"not null"`
	ContentType      string `gorm:"not null"`
	Year             *int   // publication year, optional at upload time
	DOI              string // digital Object Identifier, optional at upload time

	// Uploader info
	UploadedByUserID uint `gorm:"not null;index"`
	UploadedBy       User `gorm:"foreignKey:UploadedByUserID"`

	// Enrichment pipeline status: not_started | pending | processing | done | failed
	EnrichStatus string `gorm:"not null;default:'not_started'"`

	// Enrichment — filled by Python microservice via LLM / embedding model
	Authors   pq.StringArray `gorm:"type:text[]"`
	Summary   string
	Tags      pq.StringArray  `gorm:"type:text[]"`
	Embedding pgvector.Vector `gorm:"type:vector(1536)"`

	// User interactions
	ViewCount uint `gorm:"default:0;not null"`
	LikeCount uint `gorm:"default:0;not null"`
}
