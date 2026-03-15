package repo

import (
	"context"
	"gateway/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (model.User, error)
	FindByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (model.User, error)
	UpdatePasswordByEmail(ctx context.Context, email string, passwordHash string) error
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

func (r *userRepo) FindByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (model.User, error) {
	return gorm.G[model.User](r.db).Where("username = ?", usernameOrEmail).Or("email = ?", usernameOrEmail).First(ctx)
}

func (r *userRepo) UpdatePasswordByEmail(ctx context.Context, email string, passwordHash string) error {
	_, err := gorm.G[model.User](r.db).Where("email = ?", email).Update(ctx, "password_hash", passwordHash)
	return err
}
