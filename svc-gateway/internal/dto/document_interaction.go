package dto

import "time"

type ListHistoryQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// HistoryItem is a row in the user's view/like history list.
type HistoryItem struct {
	InteractionID uint      `json:"interaction_id"`
	InteractedAt  time.Time `json:"interacted_at"`
	DocID         uint      `json:"doc_id"`
	Title         *string   `json:"title"`
	OriginalFile  string    `json:"original_file_name"`
	Visibility    string    `json:"visibility"`
	LabID         *uint     `json:"lab_id"`
	LabName       *string   `json:"lab_name"`
	EnrichStatus  string    `json:"enrich_status"`
}

type ListHistoryResponse struct {
	Items    []HistoryItem `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type LikeStateResponse struct {
	DocID     uint `json:"doc_id"`
	Liked     bool `json:"liked"`
	LikeCount uint `json:"like_count"`
}
