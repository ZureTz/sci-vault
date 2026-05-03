package dto

import "time"

type SearchDocumentsQuery struct {
	Query string `form:"query" binding:"required,min=1,max=500"`
	LabID uint   `form:"lab_id"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=50"`
}

type SearchResultItem struct {
	DocID            uint     `json:"doc_id"`
	Title            string   `json:"title"`
	OriginalFileName string   `json:"original_file_name"`
	Summary          string   `json:"summary"`
	Authors          []string `json:"authors"`
	Tags             []string `json:"tags"`
	Similarity       float64  `json:"similarity"`
	MatchType        int32    `json:"match_type"`
}

type SearchDocumentsResponse struct {
	Results []SearchResultItem `json:"results"`
}

type ListSearchHistoryQuery struct {
	LabID uint `form:"lab_id"`
	Limit int  `form:"limit" binding:"omitempty,min=1,max=100"`
}

type SearchHistoryItem struct {
	ID          uint      `json:"id"`
	Query       string    `json:"query"`
	LabID       *uint     `json:"lab_id,omitempty"`
	ResultCount int       `json:"result_count"`
	LastUsedAt  time.Time `json:"last_used_at"`
}

type ListSearchHistoryResponse struct {
	Items []SearchHistoryItem `json:"items"`
}

type DeleteSearchHistoryResponse struct {
	Deleted int64 `json:"deleted"`
}
