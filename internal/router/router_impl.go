package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"taskmanager/internal/auth"
	"taskmanager/internal/handlers"
	"taskmanager/internal/middleware"
)

func New(jwtManager *auth.JWTManager, authHandler *handlers.AuthHandler, adminHandler *handlers.AdminHandler, taskHandler *handlers.TaskHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.NoCache)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
	}))
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/auth", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtManager))
		taskHandler.RegisterRoutes(r)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.Auth(jwtManager))
		r.Use(middleware.RequireRole("admin"))
		adminHandler.RegisterRoutes(r)
	})

	return r
}
