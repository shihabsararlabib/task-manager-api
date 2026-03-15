package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"taskmanager/internal/models"
	"taskmanager/internal/repository"
	"taskmanager/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterRoutes(mux interface {
	Post(string, http.HandlerFunc)
}) {
	mux.Post("/register", h.Register)
	mux.Post("/login", h.Login)
	mux.Post("/refresh", h.Refresh)
	mux.Post("/logout", h.Logout)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.authService.Register(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserName),
			errors.Is(err, service.ErrInvalidUserEmail),
			errors.Is(err, service.ErrInvalidPassword):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, repository.ErrUserAlreadyExists):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.authService.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to login"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var input models.RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.authService.Refresh(r.Context(), input.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to refresh token"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var input models.RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.authService.Logout(r.Context(), input.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to logout"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
