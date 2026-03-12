package service

import (
	"context"

	"gateway/internal/dto"
	"gateway/internal/repo"
)

type UserService struct {
	repo repo.UserRepository
}

func NewUserService(repo repo.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// TODO: implement login logic
	return nil, nil
}

func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest) error {
	// TODO: implement register logic
	return nil
}
