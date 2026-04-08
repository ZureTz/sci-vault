package dto

import "time"

type StatusBreakdown struct {
	NotStarted int64 `json:"not_started"`
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Done       int64 `json:"done"`
	Failed     int64 `json:"failed"`
}

type RecentDocument struct {
	ID               uint      `json:"id"`
	Title            *string   `json:"title"`
	OriginalFileName string    `json:"original_file_name"`
	FileSize         int64     `json:"file_size"`
	EnrichStatus     string    `json:"enrich_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type DashboardStatsResponse struct {
	TotalDocuments  int64            `json:"total_documents"`
	TotalStorage    int64            `json:"total_storage"`
	TotalViews      int64            `json:"total_views"`
	StatusBreakdown StatusBreakdown  `json:"status_breakdown"`
	RecentDocuments []RecentDocument `json:"recent_documents"`
}
