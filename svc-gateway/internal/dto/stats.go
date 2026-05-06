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

// DayCount is one bucket of a daily time series. Date is YYYY-MM-DD in UTC.
type DayCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// FormatBucket is one slice of the format-distribution donut.
type FormatBucket struct {
	ContentType string `json:"content_type"`
	Count       int64  `json:"count"`
}

// TopDocument is a row in the "top viewed documents" list shown on the user dashboard.
type TopDocument struct {
	ID               uint    `json:"id"`
	Title            *string `json:"title"`
	OriginalFileName string  `json:"original_file_name"`
	ViewCount        int64   `json:"view_count"`
	LikeCount        int64   `json:"like_count"`
}

// Contributor is a row in the "top contributors" list shown on the lab dashboard.
type Contributor struct {
	UserID    uint    `json:"user_id"`
	Username  string  `json:"username"`
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
	DocCount  int64   `json:"doc_count"`
}

type MyDashboardStatsResponse struct {
	TotalDocuments     int64            `json:"total_documents"`
	TotalStorage       int64            `json:"total_storage"`
	TotalViews         int64            `json:"total_views"`
	TotalLikes         int64            `json:"total_likes"`
	StatusBreakdown    StatusBreakdown  `json:"status_breakdown"`
	RecentDocuments    []RecentDocument `json:"recent_documents"`
	UploadsByDay       []DayCount       `json:"uploads_by_day"`
	ViewsByDay         []DayCount       `json:"views_by_day"`
	LikesByDay         []DayCount       `json:"likes_by_day"`
	FormatDistribution []FormatBucket   `json:"format_distribution"`
	TopViewed          []TopDocument    `json:"top_viewed"`
}

type LabDashboardStatsResponse struct {
	TotalDocuments     int64            `json:"total_documents"`
	TotalStorage       int64            `json:"total_storage"`
	TotalViews         int64            `json:"total_views"`
	TotalLikes         int64            `json:"total_likes"`
	MemberCount        int64            `json:"member_count"`
	StatusBreakdown    StatusBreakdown  `json:"status_breakdown"`
	RecentDocuments    []RecentDocument `json:"recent_documents"`
	UploadsByDay       []DayCount       `json:"uploads_by_day"`
	ViewsByDay         []DayCount       `json:"views_by_day"`
	LikesByDay         []DayCount       `json:"likes_by_day"`
	FormatDistribution []FormatBucket   `json:"format_distribution"`
	TopContributors    []Contributor    `json:"top_contributors"`
}
