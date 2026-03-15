package repository

import (
	"context"

	"taskmanager/internal/models"
)

type TaskRepository interface {
	Create(ctx context.Context, userID int, input models.CreateTaskInput) (models.Task, error)
	List(ctx context.Context, userID int, filter models.TaskListFilter) ([]models.Task, error)
	GetByID(ctx context.Context, userID, id int) (models.Task, error)
	Update(ctx context.Context, userID, id int, input models.UpdateTaskInput) (models.Task, error)
	Delete(ctx context.Context, userID, id int) error
}
