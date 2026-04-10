package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"

	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
)

type LabService struct {
	repo repo.LabRepository
}

func NewLabService(labRepo repo.LabRepository) *LabService {
	return &LabService{repo: labRepo}
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
		Role:   "member",
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

	if member.Role == "owner" {
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

// generateInviteCode returns an 8-character uppercase alphanumeric code.
func generateInviteCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// base32 of 5 bytes = exactly 8 characters, no padding needed
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)), nil
}
