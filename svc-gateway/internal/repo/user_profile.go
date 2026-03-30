package repo

import (
	"context"
	"gateway/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserProfileRepository interface {
	// UpsertAvatar creates the profile if it does not exist, or updates only avatar_key if it does.
	UpsertAvatar(ctx context.Context, profile *model.UserProfile) error
	// UpsertProfile creates the profile if it does not exist, or updates all non-avatar fields if it does.
	UpsertProfile(ctx context.Context, profile *model.UserProfile) error
	FindByUserID(ctx context.Context, userID uint) (model.UserProfile, error)
}

type userProfileRepo struct {
	db *gorm.DB
}

func NewUserProfileRepo(db *gorm.DB) UserProfileRepository {
	return &userProfileRepo{db: db}
}

func (r *userProfileRepo) UpsertAvatar(ctx context.Context, profile *model.UserProfile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"avatar_key", "updated_at"}),
	}).Create(profile).Error
}

func (r *userProfileRepo) UpsertProfile(ctx context.Context, profile *model.UserProfile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nickname", "bio", "website", "location", "updated_at"}),
	}).Create(profile).Error
}

func (r *userProfileRepo) FindByUserID(ctx context.Context, userID uint) (model.UserProfile, error) {
	return gorm.G[model.UserProfile](r.db).Where("user_id = ?", userID).First(ctx)
}
