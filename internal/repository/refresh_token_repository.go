package repository

import (
	"context"
	"errors"
	"time"
)

var ErrRefreshTokenInvalid = errors.New("refresh token invalid")

type RefreshTokenRepository interface {
	Save(ctx context.Context, userID int, token string, expiresAt time.Time) error
	Validate(ctx context.Context, userID int, token string) error
	Revoke(ctx context.Context, token string) error
}
