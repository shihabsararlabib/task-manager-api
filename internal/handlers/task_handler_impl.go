package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"taskmanager/internal/auth"
	"taskmanager/internal/models"
	"taskmanager/internal/repository"
	"taskmanager/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tasks", func(r chi.Router) {
		r.Post("/", h.CreateTask)
		r.Get("/", h.ListTasks)
		r.Get("/{id}", h.GetTask)
		r.Put("/{id}", h.UpdateTask)
		r.Delete("/{id}", h.DeleteTask)
	})
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input models.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	task, err := h.taskService.CreateTask(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTaskTitle) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create task"})
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	filter := models.TaskListFilter{
		Status: models.TaskStatus(r.URL.Query().Get("status")),
		Page:   parsePositiveInt(r.URL.Query().Get("page"), 1),
		Limit:  parsePositiveInt(r.URL.Query().Get("limit"), 20),
	}

	tasks, err := h.taskService.ListTasks(r.Context(), userID, filter)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTaskStatus) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list tasks"})
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	task, err := h.taskService.GetTaskByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get task"})
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	var input models.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	task, err := h.taskService.UpdateTask(r.Context(), userID, id, input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrTaskNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		case errors.Is(err, service.ErrInvalidTaskTitle), errors.Is(err, service.ErrInvalidTaskStatus):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update task"})
		}
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	if err := h.taskService.DeleteTask(r.Context(), userID, id); err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete task"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDParam(w http.ResponseWriter, r *http.Request) (int, bool) {
	idText := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid task id"})
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func currentUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return 0, false
	}
	return userID, true
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	v, err := strconv.Atoi(value)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
