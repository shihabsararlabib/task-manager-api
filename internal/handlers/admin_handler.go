package handlers

import (
	"net/http"

	"taskmanager/internal/service"
)

type AdminHandler struct {
	authService *service.AuthService
}

func NewAdminHandler(authService *service.AuthService) *AdminHandler {
	return &AdminHandler{authService: authService}
}

func (h *AdminHandler) RegisterRoutes(mux interface {
	Get(string, http.HandlerFunc)
}) {
	mux.Get("/users", h.ListUsers)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)

	users, err := h.authService.ListUsers(r.Context(), page, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	writeJSON(w, http.StatusOK, users)
}
