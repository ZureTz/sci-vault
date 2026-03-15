package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/cache"
	"gateway/pkg/jwt"
	"gateway/pkg/mailer"
	"gateway/pkg/password"
)

type UserService struct {
	repo         repo.UserRepository
	jwtGenerator *jwt.JWTGenerator
	mailer       *mailer.Mailer
	cacheConn    *cache.CacheConnector
}

func NewUserService(repo repo.UserRepository, jwtGenerator *jwt.JWTGenerator, mailer *mailer.Mailer, cacheConn *cache.CacheConnector) *UserService {
	return &UserService{
		repo:         repo,
		jwtGenerator: jwtGenerator,
		mailer:       mailer,
		cacheConn:    cacheConn,
	}
}

func (s *UserService) SendEmailCode(ctx context.Context, req dto.SendEmailCodeRequest) error {
	// Generate a random 6-digit code using cryptographically secure random number generator
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return fmt.Errorf("failed to generate secure code: %w", err)
	}
	code := fmt.Sprintf("%06d", n.Int64()+100000)

	// Store the code in Redis with a short expiration (e.g. 5 minutes)
	cacheKey := fmt.Sprintf("%s:code", req.Email)
	err = s.cacheConn.Set(ctx, cacheKey, code, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	// Send the code via email asynchronously
	s.mailer.SendMail(&mailer.MailRequest{
		To:      []string{req.Email},
		Subject: "Your verification code for sci-vault",
		Body:    fmt.Sprintf("<p>Your verification code is: <strong>%s</strong></p><p>This code will expire in 5 minutes.</p>", code),
	})

	return nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by username or email
	account := req.Username
	if account == "" {
		account = req.Email
	}
	user, err := s.repo.FindByUsernameOrEmail(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check password
	if err := password.Verify(user.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Generate JWT token
	jwtToken, err := s.jwtGenerator.GenerateJWT(user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return &dto.LoginResponse{
		UserID:   fmt.Sprintf("%d", user.ID),
		Username: user.Username,
		Email:    user.Email,
		JWTToken: jwtToken,
	}, nil
}

func (s *UserService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	// Verify email code from Redis
	cacheKey := fmt.Sprintf("%s:code", req.Email)
	storedCode, err := s.cacheConn.Get(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("verification code expired or invalid")
	}
	if storedCode != req.EmailCode {
		return fmt.Errorf("verification code does not match")
	}

	// Delete verification code after successful check
	defer s.cacheConn.Del(context.Background(), cacheKey)

	// Hash new password and update in database
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdatePasswordByEmail(ctx, req.Email, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest) error {
	// Verify email code from Redis
	cacheKey := fmt.Sprintf("%s:code", req.Email)
	storedCode, err := s.cacheConn.Get(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("verification code expired or invalid")
	}
	if storedCode != req.EmailCode {
		return fmt.Errorf("verification code does not match")
	}

	// Delete verification code after successful check
	defer s.cacheConn.Del(context.Background(), cacheKey)

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

	// Send welcome email asynchronously (don't block registration flow)
	s.mailer.SendMail(&mailer.MailRequest{
		To:      []string{req.Email},
		Subject: "Welcome to sci-vault",
		Body:    fmt.Sprintf("<h1>Hello %s!</h1><p>Welcome to sci-vault!</p>", req.Username),
	})

	return nil
}
