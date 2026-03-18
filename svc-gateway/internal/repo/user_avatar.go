package repo

import (
	"context"
	"gateway/internal/model"

	"gorm.io/gorm"
)

type UserAvatarRepository interface {
	Create(ctx context.Context, avatar *model.UserAvatar) error
	FindByUserID(ctx context.Context, userID uint) ([]model.UserAvatar, error)
}

type userAvatarRepo struct {
	db *gorm.DB
}

func NewUserAvatarRepo(db *gorm.DB) UserAvatarRepository {
	return &userAvatarRepo{db: db}
}

func (r *userAvatarRepo) Create(ctx context.Context, avatar *model.UserAvatar) error {
	return gorm.G[model.UserAvatar](r.db).Create(ctx, avatar)
}

func (r *userAvatarRepo) FindByUserID(ctx context.Context, userID uint) ([]model.UserAvatar, error) {
	return gorm.G[model.UserAvatar](r.db).Where("user_id = ?", userID).Find(ctx)
}
