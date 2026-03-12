package service

import (
	"context"
	"fmt"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/password"
)

type UserService struct {
	repo repo.UserRepository
}

func NewUserService(repo repo.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by username or email
	user, err := s.repo.FindByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check password
	if err := password.Verify(user.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	return &dto.LoginResponse{
		UserID:   int64(user.ID),
		Username: user.Username,
		JWTToken: "sample-jwt-token", // TODO: Implement JWT token generation
	}, nil
}

func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest) error {
	// Create new user in the database
	hashedPassword, err := password.Hash(req.Password) // Implement password hashing
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}
	err = s.repo.Create(ctx, newUser)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
