package repo

import (
	"context"
	"gateway/internal/model"

	"gorm.io/gorm"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	FindByID(ctx context.Context, id uint) (model.Document, error)
	IncrementViewCount(ctx context.Context, id uint) error
	IncrementLikeCount(ctx context.Context, id uint) error
}

type documentRepo struct {
	db *gorm.DB
}

func NewDocumentRepo(db *gorm.DB) DocumentRepository {
	return &documentRepo{db: db}
}

func (r *documentRepo) Create(ctx context.Context, doc *model.Document) error {
	return gorm.G[model.Document](r.db).Create(ctx, doc)
}

func (r *documentRepo) FindByID(ctx context.Context, id uint) (model.Document, error) {
	return gorm.G[model.Document](r.db).Where("id = ?", id).First(ctx)
}

func (r *documentRepo) IncrementViewCount(ctx context.Context, id uint) error {
	_, err := gorm.G[model.Document](r.db).Where("id = ?", id).Update(ctx, "view_count", gorm.Expr("view_count + 1"))
	return err
}

func (r *documentRepo) IncrementLikeCount(ctx context.Context, id uint) error {
	_, err := gorm.G[model.Document](r.db).Where("id = ?", id).Update(ctx, "like_count", gorm.Expr("like_count + 1"))
	return err
}
