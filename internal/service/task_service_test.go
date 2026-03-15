package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"taskmanager/internal/models"
)

type mockRepo struct {
	createFn func(ctx context.Context, userID int, input models.CreateTaskInput) (models.Task, error)
	listFn   func(ctx context.Context, userID int, filter models.TaskListFilter) ([]models.Task, error)
	getFn    func(ctx context.Context, userID, id int) (models.Task, error)
	updateFn func(ctx context.Context, userID, id int, input models.UpdateTaskInput) (models.Task, error)
	deleteFn func(ctx context.Context, userID, id int) error
}

func (m mockRepo) Create(ctx context.Context, userID int, input models.CreateTaskInput) (models.Task, error) {
	return m.createFn(ctx, userID, input)
}
func (m mockRepo) List(ctx context.Context, userID int, filter models.TaskListFilter) ([]models.Task, error) {
	return m.listFn(ctx, userID, filter)
}
func (m mockRepo) GetByID(ctx context.Context, userID, id int) (models.Task, error) {
	return m.getFn(ctx, userID, id)
}
func (m mockRepo) Update(ctx context.Context, userID, id int, input models.UpdateTaskInput) (models.Task, error) {
	return m.updateFn(ctx, userID, id, input)
}
func (m mockRepo) Delete(ctx context.Context, userID, id int) error {
	return m.deleteFn(ctx, userID, id)
}

func TestCreateTask_ValidInput(t *testing.T) {
	repo := mockRepo{createFn: func(_ context.Context, userID int, input models.CreateTaskInput) (models.Task, error) {
		return models.Task{ID: 1, UserID: userID, Title: input.Title, Description: input.Description, Status: models.StatusTodo, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
	}}

	svc := NewTaskService(repo)
	task, err := svc.CreateTask(context.Background(), 10, models.CreateTaskInput{Title: "  test task  ", Description: " desc "})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Title != "test task" {
		t.Fatalf("expected trimmed title, got %q", task.Title)
	}
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	repo := mockRepo{}
	svc := NewTaskService(repo)

	_, err := svc.CreateTask(context.Background(), 1, models.CreateTaskInput{Title: "   "})
	if !errors.Is(err, ErrInvalidTaskTitle) {
		t.Fatalf("expected ErrInvalidTaskTitle, got %v", err)
	}
}

func TestUpdateTask_InvalidStatus(t *testing.T) {
	repo := mockRepo{}
	svc := NewTaskService(repo)

	_, err := svc.UpdateTask(context.Background(), 1, 1, models.UpdateTaskInput{Title: "ok", Status: "bad-status"})
	if !errors.Is(err, ErrInvalidTaskStatus) {
		t.Fatalf("expected ErrInvalidTaskStatus, got %v", err)
	}
}
