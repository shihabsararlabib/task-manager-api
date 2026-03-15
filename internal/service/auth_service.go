package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"taskmanager/internal/auth"
	"taskmanager/internal/models"
	"taskmanager/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUserName    = errors.New("name is required")
	ErrInvalidUserEmail   = errors.New("valid email is required")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
)

type AuthService struct {
	users         repository.UserRepository
	refreshTokens repository.RefreshTokenRepository
	tokens        *auth.JWTManager
}

func NewAuthService(users repository.UserRepository, refreshTokens repository.RefreshTokenRepository, tokens *auth.JWTManager) *AuthService {
	return &AuthService{users: users, refreshTokens: refreshTokens, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, input models.RegisterInput) (models.AuthResponse, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Name == "" {
		return models.AuthResponse{}, ErrInvalidUserName
	}
	if !strings.Contains(input.Email, "@") {
		return models.AuthResponse{}, ErrInvalidUserEmail
	}
	if len(input.Password) < 8 {
		return models.AuthResponse{}, ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, input, string(hash))
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("register user: %w", err)
	}

	accessToken, refreshToken, err := s.tokens.GeneratePair(user.ID, user.Email, user.Role)
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("generate token: %w", err)
	}
	refreshClaims, err := s.tokens.Parse(refreshToken)
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("parse refresh token: %w", err)
	}
	if err := s.refreshTokens.Save(ctx, user.ID, refreshToken, refreshClaims.ExpiresAt.Time); err != nil {
		return models.AuthResponse{}, fmt.Errorf("save refresh token: %w", err)
	}

	return models.AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: sanitizeUser(user)}, nil
}

func (s *AuthService) Login(ctx context.Context, input models.LoginInput) (models.AuthResponse, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if !strings.Contains(input.Email, "@") {
		return models.AuthResponse{}, ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return models.AuthResponse{}, ErrInvalidCredentials
		}
		return models.AuthResponse{}, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return models.AuthResponse{}, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.tokens.GeneratePair(user.ID, user.Email, user.Role)
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("generate token: %w", err)
	}
	refreshClaims, err := s.tokens.Parse(refreshToken)
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("parse refresh token: %w", err)
	}
	if err := s.refreshTokens.Save(ctx, user.ID, refreshToken, refreshClaims.ExpiresAt.Time); err != nil {
		return models.AuthResponse{}, fmt.Errorf("save refresh token: %w", err)
	}

	return models.AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: sanitizeUser(user)}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (models.AuthResponse, error) {
	claims, err := s.tokens.Parse(refreshToken)
	if err != nil || claims.TokenType != "refresh" {
		return models.AuthResponse{}, ErrInvalidCredentials
	}
	if err := s.refreshTokens.Validate(ctx, claims.UserID, refreshToken); err != nil {
		if errors.Is(err, repository.ErrRefreshTokenInvalid) {
			return models.AuthResponse{}, ErrInvalidCredentials
		}
		return models.AuthResponse{}, fmt.Errorf("validate refresh token: %w", err)
	}

	user, err := s.users.GetByEmail(ctx, claims.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return models.AuthResponse{}, ErrInvalidCredentials
		}
		return models.AuthResponse{}, fmt.Errorf("get user: %w", err)
	}

	accessToken, newRefreshToken, err := s.tokens.GeneratePair(user.ID, user.Email, user.Role)
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("generate token pair: %w", err)
	}
	if err := s.refreshTokens.Revoke(ctx, refreshToken); err != nil {
		if !errors.Is(err, repository.ErrRefreshTokenInvalid) {
			return models.AuthResponse{}, fmt.Errorf("revoke old refresh token: %w", err)
		}
	}
	newRefreshClaims, err := s.tokens.Parse(newRefreshToken)
	if err != nil {
		return models.AuthResponse{}, fmt.Errorf("parse new refresh token: %w", err)
	}
	if err := s.refreshTokens.Save(ctx, user.ID, newRefreshToken, newRefreshClaims.ExpiresAt.Time); err != nil {
		return models.AuthResponse{}, fmt.Errorf("save new refresh token: %w", err)
	}

	return models.AuthResponse{AccessToken: accessToken, RefreshToken: newRefreshToken, User: sanitizeUser(user)}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return ErrInvalidCredentials
	}
	if err := s.refreshTokens.Revoke(ctx, refreshToken); err != nil {
		if errors.Is(err, repository.ErrRefreshTokenInvalid) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("logout revoke refresh token: %w", err)
	}
	return nil
}

func (s *AuthService) ListUsers(ctx context.Context, page, limit int) ([]models.User, error) {
	users, err := s.users.List(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	for i := range users {
		users[i] = sanitizeUser(users[i])
	}
	return users, nil
}

func sanitizeUser(user models.User) models.User {
	user.PasswordHash = ""
	return user
}
