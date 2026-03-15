package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskmanager/internal/models"
)

var ErrTaskNotFound = errors.New("task not found")

type PostgresTaskRepository struct {
	db *pgxpool.Pool
}

func NewPostgresTaskRepository(db *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, userID int, input models.CreateTaskInput) (models.Task, error) {
	const query = `
		INSERT INTO tasks (user_id, title, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, description, status, created_at, updated_at
	`

	var task models.Task
	err := r.db.QueryRow(ctx, query, userID, input.Title, input.Description, models.StatusTodo).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return models.Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func (r *PostgresTaskRepository) List(ctx context.Context, userID int, filter models.TaskListFilter) ([]models.Task, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
	`
	args := []any{userID}
	argIdx := 2
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks query: %w", err)
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tasks rows: %w", err)
	}

	return tasks, nil
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, userID, id int) (models.Task, error) {
	const query = `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2
	`

	var task models.Task
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Task{}, ErrTaskNotFound
		}
		return models.Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (r *PostgresTaskRepository) Update(ctx context.Context, userID, id int, input models.UpdateTaskInput) (models.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = NOW()
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, title, description, status, created_at, updated_at
	`

	var task models.Task
	err := r.db.QueryRow(ctx, query, input.Title, input.Description, input.Status, id, userID).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Task{}, ErrTaskNotFound
		}
		return models.Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, userID, id int) error {
	const query = `DELETE FROM tasks WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}
