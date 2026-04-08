package dto

import (
	"mime/multipart"
	"time"
)

type UploadDocumentForm struct {
	File  *multipart.FileHeader `form:"file" binding:"required"`
	Title *string               `form:"title" binding:"omitempty,max=255"`
	Year  *int                  `form:"year" binding:"omitempty,min=1000,max=9999"`
	DOI   *string               `form:"doi" binding:"omitempty,max=255"`
}

type DocumentIDUri struct {
	DocID uint `uri:"doc_id" binding:"required"`
}

type ListMyDocumentsQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// DocumentListItem is a lightweight summary used in list views.
// It intentionally omits fields that require expensive operations (e.g. signed download URLs).
type DocumentListItem struct {
	ID               uint      `json:"id"`
	Title            *string   `json:"title"`
	OriginalFileName string    `json:"original_file_name"`
	FileSize         int64     `json:"file_size"`
	EnrichStatus     string    `json:"enrich_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type ListDocumentsResponse struct {
	Documents []DocumentListItem `json:"documents"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

type EnrichStatusResponse struct {
	DocID  uint   `json:"doc_id"`
	Status string `json:"status"`
}

type DocumentResponse struct {
	ID               uint      `json:"id"`
	Title            *string   `json:"title"`
	OriginalFileName string    `json:"original_file_name"`
	FileSize         int64     `json:"file_size"`
	ContentType      string    `json:"content_type"`
	Year             *int      `json:"year"`
	DOI              *string   `json:"doi"`
	EnrichStatus     string    `json:"enrich_status"`
	Authors          []string  `json:"authors"`
	Summary          *string   `json:"summary"`
	Tags             []string  `json:"tags"`
	ViewCount        uint      `json:"view_count"`
	LikeCount        uint      `json:"like_count"`
	UploadedByUserID uint      `json:"uploaded_by"`
	DownloadURL      string    `json:"download_url"`
	CreatedAt        time.Time `json:"created_at"`
}
