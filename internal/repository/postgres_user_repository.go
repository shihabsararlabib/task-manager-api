package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskmanager/internal/models"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, input models.RegisterInput, passwordHash string) (models.User, error) {
	const query = `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, 'user')
		RETURNING id, name, email, role, password_hash, created_at
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, input.Name, strings.ToLower(input.Email), passwordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, ErrUserAlreadyExists
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	const query = `
		SELECT id, name, email, role, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) List(ctx context.Context, page, limit int) ([]models.User, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	const query = `
		SELECT id, name, email, role, created_at
		FROM users
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users query: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users rows: %w", err)
	}

	return users, nil
}
