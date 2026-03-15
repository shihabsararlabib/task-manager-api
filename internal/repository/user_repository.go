package repository

import (
	"context"
	"errors"

	"taskmanager/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	Create(ctx context.Context, input models.RegisterInput, passwordHash string) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	List(ctx context.Context, page, limit int) ([]models.User, error)
}
