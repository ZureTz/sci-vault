package repo

import (
	"context"
	"gateway/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	return gorm.G[model.User](r.db).Create(ctx, user)
}

func (r *userRepo) FindByID(ctx context.Context, id uint) (model.User, error) {
	return gorm.G[model.User](r.db).Where("id = ?", id).First(ctx)
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (model.User, error) {
	return gorm.G[model.User](r.db).Where("username = ?", username).First(ctx)
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (model.User, error) {
	return gorm.G[model.User](r.db).Where("email = ?", email).First(ctx)
}
