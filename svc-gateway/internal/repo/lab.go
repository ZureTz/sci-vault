package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"gateway/internal/model"
)

// MemberInfo carries member fields joined from the users and user_profiles tables.
type MemberInfo struct {
	UserID    uint
	Username  string
	Role      string
	JoinedAt  time.Time
	AvatarKey *string
}

// LabWithRole carries lab fields alongside the requesting user's membership role and total member count.
type LabWithRole struct {
	ID          uint
	Name        string
	Description *string
	InviteCode  string
	OwnerID     uint
	Role        string
	MemberCount int64
}

type LabRepository interface {
	CreateWithOwner(ctx context.Context, lab *model.Lab) error
	FindByInviteCode(ctx context.Context, code string) (model.Lab, error)
	FindMember(ctx context.Context, labID, userID uint) (model.LabMember, error)
	AddMember(ctx context.Context, member *model.LabMember) error
	CountMembers(ctx context.Context, labID uint) (int64, error)
	FindLabsByUserID(ctx context.Context, userID uint) ([]LabWithRole, error)
	FindByID(ctx context.Context, labID uint) (model.Lab, error)
	FindMembersByLabID(ctx context.Context, labID uint) ([]MemberInfo, error)
	RemoveMember(ctx context.Context, labID, userID uint) error
	TransferOwnership(ctx context.Context, labID, oldOwnerID, newOwnerID uint) error
	DeleteLab(ctx context.Context, labID uint) error
	UpdateInviteCode(ctx context.Context, labID uint, newCode string) error
	UpdateLabInfo(ctx context.Context, labID uint, name string, description *string) error
}

type labRepo struct {
	db *gorm.DB
}

func NewLabRepo(db *gorm.DB) LabRepository {
	return &labRepo{db: db}
}

// CreateWithOwner creates the lab and adds the owner as a member atomically.
func (r *labRepo) CreateWithOwner(ctx context.Context, lab *model.Lab) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(lab).Error; err != nil {
			return err
		}
		return tx.Create(&model.LabMember{
			LabID:  lab.ID,
			UserID: lab.OwnerID,
			Role:   model.LabRoleOwner,
		}).Error
	})
}

func (r *labRepo) FindByInviteCode(ctx context.Context, code string) (model.Lab, error) {
	return gorm.G[model.Lab](r.db).Where("invite_code = ?", code).First(ctx)
}

func (r *labRepo) FindMember(ctx context.Context, labID, userID uint) (model.LabMember, error) {
	return gorm.G[model.LabMember](r.db).Where("lab_id = ? AND user_id = ?", labID, userID).First(ctx)
}

func (r *labRepo) AddMember(ctx context.Context, member *model.LabMember) error {
	return gorm.G[model.LabMember](r.db).Create(ctx, member)
}

func (r *labRepo) CountMembers(ctx context.Context, labID uint) (int64, error) {
	return gorm.G[model.LabMember](r.db).Where("lab_id = ?", labID).Count(ctx, "*")
}

func (r *labRepo) FindByID(ctx context.Context, labID uint) (model.Lab, error) {
	return gorm.G[model.Lab](r.db).Where("id = ?", labID).First(ctx)
}

func (r *labRepo) FindMembersByLabID(ctx context.Context, labID uint) ([]MemberInfo, error) {
	var results []MemberInfo
	err := r.db.WithContext(ctx).Raw(`
		SELECT lm.user_id AS user_id,
		       u.username AS username,
		       lm.role AS role,
		       lm.created_at AS joined_at,
		       up.avatar_key AS avatar_key
		FROM lab_members lm
		JOIN users u ON u.id = lm.user_id AND u.deleted_at IS NULL
		LEFT JOIN user_profiles up ON up.user_id = lm.user_id AND up.deleted_at IS NULL
		WHERE lm.lab_id = ? AND lm.deleted_at IS NULL
		ORDER BY lm.created_at ASC
	`, labID).Scan(&results).Error
	return results, err
}

func (r *labRepo) RemoveMember(ctx context.Context, labID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("lab_id = ? AND user_id = ?", labID, userID).
		Delete(&model.LabMember{}).Error
}

// TransferOwnership atomically updates the lab owner and swaps member roles.
func (r *labRepo) TransferOwnership(ctx context.Context, labID, oldOwnerID, newOwnerID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Lab{}).Where("id = ?", labID).Update("owner_id", newOwnerID).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.LabMember{}).
			Where("lab_id = ? AND user_id = ?", labID, oldOwnerID).
			Update("role", model.LabRoleMember).Error; err != nil {
			return err
		}
		return tx.Model(&model.LabMember{}).
			Where("lab_id = ? AND user_id = ?", labID, newOwnerID).
			Update("role", model.LabRoleOwner).Error
	})
}

// DeleteLab soft-deletes the lab and all its memberships atomically.
// Associated documents are disassociated (lab_id set to NULL, visibility reset to private).
func (r *labRepo) DeleteLab(ctx context.Context, labID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Document{}).
			Where("lab_id = ?", labID).
			Updates(map[string]any{"lab_id": nil, "visibility": model.DocVisibilityPrivate}).Error; err != nil {
			return err
		}
		if err := tx.Where("lab_id = ?", labID).Delete(&model.LabMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Lab{}, labID).Error
	})
}

func (r *labRepo) UpdateInviteCode(ctx context.Context, labID uint, newCode string) error {
	return r.db.WithContext(ctx).Model(&model.Lab{}).Where("id = ?", labID).Update("invite_code", newCode).Error
}

func (r *labRepo) UpdateLabInfo(ctx context.Context, labID uint, name string, description *string) error {
	return r.db.WithContext(ctx).Model(&model.Lab{}).Where("id = ?", labID).Updates(map[string]any{
		"name":        name,
		"description": description,
	}).Error
}

func (r *labRepo) FindLabsByUserID(ctx context.Context, userID uint) ([]LabWithRole, error) {
	var results []LabWithRole
	err := r.db.WithContext(ctx).Raw(`
		SELECT l.id, l.name, l.description, l.invite_code, l.owner_id, lm.role,
		       (SELECT COUNT(*) FROM lab_members lm2 WHERE lm2.lab_id = l.id AND lm2.deleted_at IS NULL) AS member_count
		FROM labs l
		JOIN lab_members lm ON lm.lab_id = l.id AND lm.user_id = ? AND lm.deleted_at IS NULL
		WHERE l.deleted_at IS NULL
		ORDER BY l.created_at DESC
	`, userID).Scan(&results).Error
	return results, err
}
