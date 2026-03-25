package dto

import (
	"mime/multipart"
	"time"
)

type UploadDocumentForm struct {
	File  *multipart.FileHeader `form:"file" binding:"required"`
	Title string                `form:"title" binding:"required,min=1,max=255"`
	Year  *int                  `form:"year" binding:"omitempty,min=1000,max=9999"`
	DOI   string                `form:"doi" binding:"omitempty,max=255"`
}

type DocumentIDUri struct {
	DocID uint `uri:"doc_id" binding:"required"`
}

type EnrichStatusResponse struct {
	DocID  uint   `json:"doc_id"`
	Status string `json:"status"`
}

type DocumentResponse struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	OriginalFileName string    `json:"original_file_name"`
	FileSize         int64     `json:"file_size"`
	ContentType      string    `json:"content_type"`
	Year             *int      `json:"year,omitempty"`
	DOI              string    `json:"doi,omitempty"`
	EnrichStatus     string    `json:"enrich_status"`
	Authors          []string  `json:"authors"`
	Summary          string    `json:"summary"`
	Tags             []string  `json:"tags"`
	ViewCount        uint      `json:"view_count"`
	LikeCount        uint      `json:"like_count"`
	UploadedByUserID uint      `json:"uploaded_by"`
	DownloadURL      string    `json:"download_url"`
	CreatedAt        time.Time `json:"created_at"`
}
