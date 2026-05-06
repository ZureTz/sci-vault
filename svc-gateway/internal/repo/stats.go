package repo

import (
	"context"
	"time"

	"gateway/internal/model"

	"gorm.io/gorm"
)

// EnrichStatusCount holds a single row from a GROUP BY enrich_status aggregation.
type EnrichStatusCount struct {
	EnrichStatus string `gorm:"column:enrich_status"`
	Count        int64  `gorm:"column:count"`
}

// DayCount holds a single bucket from a GROUP BY day aggregation. The day is
// truncated to UTC midnight in SQL so the time-series is timezone-stable.
type DayCount struct {
	Day   time.Time `gorm:"column:day"`
	Count int64     `gorm:"column:count"`
}

// FormatCount holds a single row from a GROUP BY content_type aggregation.
type FormatCount struct {
	ContentType string `gorm:"column:content_type"`
	Count       int64  `gorm:"column:count"`
}

// ContributorCount holds one row of "top contributors" aggregation for a lab:
// who uploaded the most lab-visible documents.
type ContributorCount struct {
	UserID    uint    `gorm:"column:user_id"`
	Username  string  `gorm:"column:username"`
	Nickname  *string `gorm:"column:nickname"`
	AvatarKey *string `gorm:"column:avatar_key"`
	Count     int64   `gorm:"column:count"`
}

type StatsRepository interface {
	// User-scoped: filtered by documents.uploaded_by_user_id.
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CountByStatusForUser(ctx context.Context, userID uint) ([]EnrichStatusCount, error)
	TotalStorageByUser(ctx context.Context, userID uint) (int64, error)
	TotalViewsByUser(ctx context.Context, userID uint) (int64, error)
	TotalLikesByUser(ctx context.Context, userID uint) (int64, error)
	RecentByUserID(ctx context.Context, userID uint, limit int) ([]model.Document, error)
	UploadsByDayForUser(ctx context.Context, userID uint, days int) ([]DayCount, error)
	ViewsByDayForUser(ctx context.Context, userID uint, days int) ([]DayCount, error)
	LikesByDayForUser(ctx context.Context, userID uint, days int) ([]DayCount, error)
	FormatDistributionForUser(ctx context.Context, userID uint) ([]FormatCount, error)
	TopViewedByUser(ctx context.Context, userID uint, limit int) ([]model.Document, error)

	// Lab-scoped: filtered by documents.lab_id AND visibility = 'lab'.
	CountByLab(ctx context.Context, labID uint) (int64, error)
	CountByStatusForLab(ctx context.Context, labID uint) ([]EnrichStatusCount, error)
	TotalStorageByLab(ctx context.Context, labID uint) (int64, error)
	TotalViewsByLab(ctx context.Context, labID uint) (int64, error)
	TotalLikesByLab(ctx context.Context, labID uint) (int64, error)
	RecentByLab(ctx context.Context, labID uint, limit int) ([]model.Document, error)
	UploadsByDayForLab(ctx context.Context, labID uint, days int) ([]DayCount, error)
	ViewsByDayForLab(ctx context.Context, labID uint, days int) ([]DayCount, error)
	LikesByDayForLab(ctx context.Context, labID uint, days int) ([]DayCount, error)
	FormatDistributionForLab(ctx context.Context, labID uint) ([]FormatCount, error)
	TopContributorsByLab(ctx context.Context, labID uint, limit int) ([]ContributorCount, error)
}

type statsRepo struct {
	db *gorm.DB
}

func NewStatsRepo(db *gorm.DB) StatsRepository {
	return &statsRepo{db: db}
}

// ---------------- User-scoped ----------------

func (r *statsRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	return gorm.G[model.Document](r.db).Where("uploaded_by_user_id = ?", userID).Count(ctx, "*")
}

func (r *statsRepo) CountByStatusForUser(ctx context.Context, userID uint) ([]EnrichStatusCount, error) {
	var results []EnrichStatusCount
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("enrich_status, COUNT(*) as count").
		Where("uploaded_by_user_id = ?", userID).
		Group("enrich_status").
		Scan(&results).Error
	return results, err
}

func (r *statsRepo) TotalStorageByUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(file_size), 0)").
		Where("uploaded_by_user_id = ?", userID).
		Scan(&total).Error
	return total, err
}

func (r *statsRepo) TotalViewsByUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(view_count), 0)").
		Where("uploaded_by_user_id = ?", userID).
		Scan(&total).Error
	return total, err
}

func (r *statsRepo) TotalLikesByUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(like_count), 0)").
		Where("uploaded_by_user_id = ?", userID).
		Scan(&total).Error
	return total, err
}

func (r *statsRepo) RecentByUserID(ctx context.Context, userID uint, limit int) ([]model.Document, error) {
	return gorm.G[model.Document](r.db).
		Where("uploaded_by_user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(ctx)
}

func (r *statsRepo) UploadsByDayForUser(ctx context.Context, userID uint, days int) ([]DayCount, error) {
	var rows []DayCount
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("date_trunc('day', created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Where("uploaded_by_user_id = ? AND created_at >= NOW() - make_interval(days => ?)", userID, days).
		Group("day").
		Order("day").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) ViewsByDayForUser(ctx context.Context, userID uint, days int) ([]DayCount, error) {
	var rows []DayCount
	err := r.db.WithContext(ctx).
		Table("document_views v").
		Select("date_trunc('day', v.created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Joins("JOIN documents d ON d.id = v.document_id AND d.deleted_at IS NULL").
		Where("v.deleted_at IS NULL AND d.uploaded_by_user_id = ? AND v.created_at >= NOW() - make_interval(days => ?)", userID, days).
		Group("day").
		Order("day").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) LikesByDayForUser(ctx context.Context, userID uint, days int) ([]DayCount, error) {
	var rows []DayCount
	err := r.db.WithContext(ctx).
		Table("document_likes l").
		Select("date_trunc('day', l.created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Joins("JOIN documents d ON d.id = l.document_id AND d.deleted_at IS NULL").
		Where("l.deleted_at IS NULL AND d.uploaded_by_user_id = ? AND l.created_at >= NOW() - make_interval(days => ?)", userID, days).
		Group("day").
		Order("day").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) FormatDistributionForUser(ctx context.Context, userID uint) ([]FormatCount, error) {
	var rows []FormatCount
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("content_type, COUNT(*) AS count").
		Where("uploaded_by_user_id = ?", userID).
		Group("content_type").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) TopViewedByUser(ctx context.Context, userID uint, limit int) ([]model.Document, error) {
	return gorm.G[model.Document](r.db).
		Where("uploaded_by_user_id = ?", userID).
		Order("view_count DESC, id DESC").
		Limit(limit).
		Find(ctx)
}

// ---------------- Lab-scoped ----------------

func (r *statsRepo) CountByLab(ctx context.Context, labID uint) (int64, error) {
	return gorm.G[model.Document](r.db).
		Where("lab_id = ? AND visibility = ?", labID, model.DocVisibilityLab).
		Count(ctx, "*")
}

func (r *statsRepo) CountByStatusForLab(ctx context.Context, labID uint) ([]EnrichStatusCount, error) {
	var results []EnrichStatusCount
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("enrich_status, COUNT(*) as count").
		Where("lab_id = ? AND visibility = ?", labID, model.DocVisibilityLab).
		Group("enrich_status").
		Scan(&results).Error
	return results, err
}

func (r *statsRepo) TotalStorageByLab(ctx context.Context, labID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(file_size), 0)").
		Where("lab_id = ? AND visibility = ?", labID, model.DocVisibilityLab).
		Scan(&total).Error
	return total, err
}

func (r *statsRepo) TotalViewsByLab(ctx context.Context, labID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(view_count), 0)").
		Where("lab_id = ? AND visibility = ?", labID, model.DocVisibilityLab).
		Scan(&total).Error
	return total, err
}

func (r *statsRepo) TotalLikesByLab(ctx context.Context, labID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("COALESCE(SUM(like_count), 0)").
		Where("lab_id = ? AND visibility = ?", labID, model.DocVisibilityLab).
		Scan(&total).Error
	return total, err
}

func (r *statsRepo) RecentByLab(ctx context.Context, labID uint, limit int) ([]model.Document, error) {
	return gorm.G[model.Document](r.db).
		Where("lab_id = ? AND visibility = ?", labID, model.DocVisibilityLab).
		Order("created_at DESC").
		Limit(limit).
		Find(ctx)
}

func (r *statsRepo) UploadsByDayForLab(ctx context.Context, labID uint, days int) ([]DayCount, error) {
	var rows []DayCount
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("date_trunc('day', created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Where("lab_id = ? AND visibility = ? AND created_at >= NOW() - make_interval(days => ?)",
			labID, model.DocVisibilityLab, days).
		Group("day").
		Order("day").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) ViewsByDayForLab(ctx context.Context, labID uint, days int) ([]DayCount, error) {
	var rows []DayCount
	err := r.db.WithContext(ctx).
		Table("document_views v").
		Select("date_trunc('day', v.created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Joins("JOIN documents d ON d.id = v.document_id AND d.deleted_at IS NULL").
		Where("v.deleted_at IS NULL AND d.lab_id = ? AND d.visibility = ? AND v.created_at >= NOW() - make_interval(days => ?)",
			labID, model.DocVisibilityLab, days).
		Group("day").
		Order("day").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) LikesByDayForLab(ctx context.Context, labID uint, days int) ([]DayCount, error) {
	var rows []DayCount
	err := r.db.WithContext(ctx).
		Table("document_likes l").
		Select("date_trunc('day', l.created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Joins("JOIN documents d ON d.id = l.document_id AND d.deleted_at IS NULL").
		Where("l.deleted_at IS NULL AND d.lab_id = ? AND d.visibility = ? AND l.created_at >= NOW() - make_interval(days => ?)",
			labID, model.DocVisibilityLab, days).
		Group("day").
		Order("day").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) FormatDistributionForLab(ctx context.Context, labID uint) ([]FormatCount, error) {
	var rows []FormatCount
	err := r.db.WithContext(ctx).
		Model(&model.Document{}).
		Select("content_type, COUNT(*) AS count").
		Where("lab_id = ? AND visibility = ?", labID, model.DocVisibilityLab).
		Group("content_type").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *statsRepo) TopContributorsByLab(ctx context.Context, labID uint, limit int) ([]ContributorCount, error) {
	var rows []ContributorCount
	err := r.db.WithContext(ctx).
		Table("documents d").
		Select(`d.uploaded_by_user_id AS user_id,
			u.username AS username,
			p.nickname AS nickname,
			p.avatar_key AS avatar_key,
			COUNT(*) AS count`).
		Joins("JOIN users u ON u.id = d.uploaded_by_user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN user_profiles p ON p.user_id = d.uploaded_by_user_id AND p.deleted_at IS NULL").
		Where("d.deleted_at IS NULL AND d.lab_id = ? AND d.visibility = ?", labID, model.DocVisibilityLab).
		Group("d.uploaded_by_user_id, u.username, p.nickname, p.avatar_key").
		Order("count DESC, d.uploaded_by_user_id ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}
