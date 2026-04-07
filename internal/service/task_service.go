package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"taskmanager/internal/models"
	"taskmanager/internal/repository"
)

var (
	ErrInvalidTaskTitle  = errors.New("task title is required")
	ErrInvalidTaskStatus = errors.New("invalid task status")
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, userID int, input models.CreateTaskInput) (models.Task, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return models.Task{}, ErrInvalidTaskTitle
	}
	input.Description = strings.TrimSpace(input.Description)

	task, err := s.repo.Create(ctx, userID, input)
	if err != nil {
		return models.Task{}, fmt.Errorf("service create task: %w", err)
	}
	return task, nil
}

func (s *TaskService) ListTasks(ctx context.Context, userID int, filter models.TaskListFilter) ([]models.Task, error) {
	if filter.Status != "" && !isValidStatus(filter.Status) {
		return nil, ErrInvalidTaskStatus
	}
	tasks, err := s.repo.List(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("service list tasks: %w", err)
	}
	return tasks, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, userID, id int) (models.Task, error) {
	task, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return models.Task{}, fmt.Errorf("service get task by id: %w", err)
	}
	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, userID, id int, input models.UpdateTaskInput) (models.Task, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return models.Task{}, ErrInvalidTaskTitle
	}
	if !isValidStatus(input.Status) {
		return models.Task{}, ErrInvalidTaskStatus
	}

	task, err := s.repo.Update(ctx, userID, id, input)
	if err != nil {
		return models.Task{}, fmt.Errorf("service update task: %w", err)
	}
	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, userID, id int) error {
	if err := s.repo.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("service delete task: %w", err)
	}
	return nil
}

func isValidStatus(status models.TaskStatus) bool {
	switch status {
	case models.StatusTodo, models.StatusInProgress, models.StatusDone:
		return true
	default:
		return false
	}
}
