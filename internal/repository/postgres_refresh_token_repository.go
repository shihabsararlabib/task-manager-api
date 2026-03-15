package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(db *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) Save(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	if _, err := r.db.Exec(ctx, query, userID, hashToken(token), expiresAt); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *PostgresRefreshTokenRepository) Validate(ctx context.Context, userID int, token string) error {
	const query = `
		SELECT id
		FROM refresh_tokens
		WHERE user_id = $1
		  AND token_hash = $2
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`
	var id int64
	err := r.db.QueryRow(ctx, query, userID, hashToken(token)).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRefreshTokenInvalid
		}
		return fmt.Errorf("validate refresh token: %w", err)
	}
	return nil
}

func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, hashToken(token))
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrRefreshTokenInvalid
	}
	return nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
