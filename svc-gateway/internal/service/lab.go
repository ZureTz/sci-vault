package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
	"gateway/pkg/cache"
	"gateway/pkg/mailer"
)

type LabService struct {
	repo      repo.LabRepository
	userRepo  repo.UserRepository
	cacheConn *cache.CacheConnector
	mailer    *mailer.Mailer
}

func NewLabService(labRepo repo.LabRepository, userRepo repo.UserRepository, cacheConn *cache.CacheConnector, mailer *mailer.Mailer) *LabService {
	return &LabService{
		repo:      labRepo,
		userRepo:  userRepo,
		cacheConn: cacheConn,
		mailer:    mailer,
	}
}

func (s *LabService) CreateLab(ctx context.Context, ownerID uint, req dto.CreateLabRequest) (*dto.JoinLabResponse, error) {
	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, err
	}

	lab := &model.Lab{
		Name:        req.Name,
		Description: req.Description,
		InviteCode:  inviteCode,
		OwnerID:     ownerID,
	}
	if err := s.repo.CreateWithOwner(ctx, lab); err != nil {
		return nil, err
	}

	return &dto.JoinLabResponse{
		ID:          lab.ID,
		Name:        lab.Name,
		Description: lab.Description,
		InviteCode:  lab.InviteCode,
		OwnerID:     lab.OwnerID,
		MemberCount: 1,
	}, nil
}

func (s *LabService) JoinLabByCode(ctx context.Context, userID uint, req dto.JoinLabByCodeRequest) (*dto.JoinLabResponse, error) {
	lab, err := s.repo.FindByInviteCode(ctx, req.InviteCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_error.ErrInvalidInviteCode
		}
		return nil, err
	}

	_, err = s.repo.FindMember(ctx, lab.ID, userID)
	if err == nil {
		return nil, app_error.ErrAlreadyMember
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.repo.AddMember(ctx, &model.LabMember{
		LabID:  lab.ID,
		UserID: userID,
		Role:   model.LabRoleMember,
	}); err != nil {
		return nil, err
	}

	count, err := s.repo.CountMembers(ctx, lab.ID)
	if err != nil {
		return nil, err
	}

	return &dto.JoinLabResponse{
		ID:          lab.ID,
		Name:        lab.Name,
		Description: lab.Description,
		InviteCode:  lab.InviteCode,
		OwnerID:     lab.OwnerID,
		MemberCount: count,
	}, nil
}

func (s *LabService) GetLab(ctx context.Context, labID, userID uint) (*dto.LabDetailResponse, error) {
	member, err := s.repo.FindMember(ctx, labID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_error.ErrNotMember
		}
		return nil, err
	}

	lab, err := s.repo.FindByID(ctx, labID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_error.ErrLabNotFound
		}
		return nil, err
	}

	count, err := s.repo.CountMembers(ctx, labID)
	if err != nil {
		return nil, err
	}

	return &dto.LabDetailResponse{
		ID:          lab.ID,
		Name:        lab.Name,
		Description: lab.Description,
		InviteCode:  lab.InviteCode,
		OwnerID:     lab.OwnerID,
		MemberCount: count,
		MyRole:      member.Role,
	}, nil
}

func (s *LabService) GetMembers(ctx context.Context, labID, userID uint) ([]dto.LabMemberInfo, error) {
	_, err := s.repo.FindMember(ctx, labID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_error.ErrNotMember
		}
		return nil, err
	}

	members, err := s.repo.FindMembersByLabID(ctx, labID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.LabMemberInfo, len(members))
	for i, m := range members {
		items[i] = dto.LabMemberInfo{
			UserID:   m.UserID,
			Username: m.Username,
			Role:     m.Role,
			JoinedAt: m.JoinedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
	}
	return items, nil
}

func (s *LabService) LeaveLab(ctx context.Context, labID, userID uint) error {
	member, err := s.repo.FindMember(ctx, labID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrNotMember
		}
		return err
	}

	if member.Role == model.LabRoleOwner {
		return app_error.ErrOwnerCannotLeave
	}

	return s.repo.RemoveMember(ctx, labID, userID)
}

func (s *LabService) GetMyLabs(ctx context.Context, userID uint) ([]dto.LabListItem, error) {
	labs, err := s.repo.FindLabsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.LabListItem, len(labs))
	for i, l := range labs {
		items[i] = dto.LabListItem{
			ID:          l.ID,
			Name:        l.Name,
			Description: l.Description,
			OwnerID:     l.OwnerID,
			MemberCount: l.MemberCount,
			Role:        l.Role,
		}
	}
	return items, nil
}

func (s *LabService) KickMember(ctx context.Context, labID, requesterID, targetUserID uint) error {
	if requesterID == targetUserID {
		return app_error.ErrCannotKickSelf
	}

	requester, err := s.repo.FindMember(ctx, labID, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrNotMember
		}
		return err
	}
	if requester.Role != model.LabRoleOwner {
		return app_error.ErrNotOwner
	}

	target, err := s.repo.FindMember(ctx, labID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrTargetNotMember
		}
		return err
	}
	if target.Role == model.LabRoleOwner {
		return app_error.ErrCannotKickOwner
	}

	return s.repo.RemoveMember(ctx, labID, targetUserID)
}

func (s *LabService) TransferOwnership(ctx context.Context, labID, requesterID, targetUserID uint) error {
	requester, err := s.repo.FindMember(ctx, labID, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrNotMember
		}
		return err
	}
	if requester.Role != model.LabRoleOwner {
		return app_error.ErrNotOwner
	}

	_, err = s.repo.FindMember(ctx, labID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrTargetNotMember
		}
		return err
	}

	return s.repo.TransferOwnership(ctx, labID, requesterID, targetUserID)
}

func (s *LabService) RequestDeleteLab(ctx context.Context, labID, requesterID uint) error {
	requester, err := s.repo.FindMember(ctx, labID, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrNotMember
		}
		return err
	}
	if requester.Role != model.LabRoleOwner {
		return app_error.ErrNotOwner
	}

	lab, err := s.repo.FindByID(ctx, labID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrLabNotFound
		}
		return err
	}

	owner, err := s.userRepo.FindByID(ctx, requesterID)
	if err != nil {
		return fmt.Errorf("failed to find owner: %w", err)
	}

	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return fmt.Errorf("failed to generate secure code: %w", err)
	}
	code := fmt.Sprintf("%06d", n.Int64()+100000)

	cacheKey := fmt.Sprintf("lab_delete:%d:code", labID)
	if err := s.cacheConn.Set(ctx, cacheKey, code, 5*time.Minute); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	s.mailer.SendMail(&mailer.MailRequest{
		To:      []string{owner.Email},
		Subject: fmt.Sprintf("Confirm deletion of lab \"%s\"", lab.Name),
		Body: fmt.Sprintf(
			"<p>You requested to delete the lab <strong>%s</strong>.</p>"+
				"<p>Your confirmation code is: <strong>%s</strong></p>"+
				"<p>This code will expire in 5 minutes. If you did not request this, please ignore this email.</p>",
			lab.Name, code,
		),
	})

	return nil
}

func (s *LabService) DeleteLab(ctx context.Context, labID, requesterID uint, confirmName, emailCode string) error {
	requester, err := s.repo.FindMember(ctx, labID, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrNotMember
		}
		return err
	}
	if requester.Role != model.LabRoleOwner {
		return app_error.ErrNotOwner
	}

	lab, err := s.repo.FindByID(ctx, labID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrLabNotFound
		}
		return err
	}
	if lab.Name != confirmName {
		return app_error.ErrLabNameMismatch
	}

	cacheKey := fmt.Sprintf("lab_delete:%d:code", labID)
	storedCode, err := s.cacheConn.Get(ctx, cacheKey)
	if err != nil {
		return app_error.ErrEmailCodeExpired
	}
	if storedCode != emailCode {
		return app_error.ErrEmailCodeMismatch
	}
	defer s.cacheConn.Del(context.Background(), cacheKey)

	return s.repo.DeleteLab(ctx, labID)
}

func (s *LabService) ResetInviteCode(ctx context.Context, labID, requesterID uint) (string, error) {
	requester, err := s.repo.FindMember(ctx, labID, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", app_error.ErrNotMember
		}
		return "", err
	}
	if requester.Role != model.LabRoleOwner {
		return "", app_error.ErrNotOwner
	}

	newCode, err := generateInviteCode()
	if err != nil {
		return "", err
	}

	if err := s.repo.UpdateInviteCode(ctx, labID, newCode); err != nil {
		return "", err
	}
	return newCode, nil
}

// generateInviteCode returns an 8-character uppercase alphanumeric code.
func generateInviteCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// base32 of 5 bytes = exactly 8 characters, no padding needed
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)), nil
}
