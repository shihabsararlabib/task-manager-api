package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"taskmanager/internal/auth"
	"taskmanager/internal/models"
	"taskmanager/internal/repository"
)

type mockUserRepo struct {
	createFn func(ctx context.Context, input models.RegisterInput, passwordHash string) (models.User, error)
	getFn    func(ctx context.Context, email string) (models.User, error)
	listFn   func(ctx context.Context, page, limit int) ([]models.User, error)
}

func (m mockUserRepo) Create(ctx context.Context, input models.RegisterInput, passwordHash string) (models.User, error) {
	return m.createFn(ctx, input, passwordHash)
}

func (m mockUserRepo) GetByEmail(ctx context.Context, email string) (models.User, error) {
	return m.getFn(ctx, email)
}

func (m mockUserRepo) List(ctx context.Context, page, limit int) ([]models.User, error) {
	if m.listFn == nil {
		return []models.User{}, nil
	}
	return m.listFn(ctx, page, limit)
}

type mockRefreshRepo struct{}

func (mockRefreshRepo) Save(context.Context, int, string, time.Time) error { return nil }
func (mockRefreshRepo) Validate(context.Context, int, string) error        { return nil }
func (mockRefreshRepo) Revoke(context.Context, string) error               { return nil }

func TestRegister_InvalidPassword(t *testing.T) {
	svc := NewAuthService(mockUserRepo{}, mockRefreshRepo{}, auth.NewJWTManager("secret", time.Hour, 24*time.Hour))
	_, err := svc.Register(context.Background(), models.RegisterInput{Name: "A", Email: "a@b.com", Password: "123"})
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := mockUserRepo{getFn: func(context.Context, string) (models.User, error) {
		return models.User{}, repository.ErrUserNotFound
	}}
	svc := NewAuthService(repo, mockRefreshRepo{}, auth.NewJWTManager("secret", time.Hour, 24*time.Hour))
	_, err := svc.Login(context.Background(), models.LoginInput{Email: "x@y.com", Password: "password"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
