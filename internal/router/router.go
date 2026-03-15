//go:build ignore
// +build ignore

package router
package router

import (


























}	return r	taskHandler.RegisterRoutes(r)	})		_, _ = w.Write([]byte("ok"))		w.WriteHeader(http.StatusOK)	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {	r.Use(middleware.Logging)	r.Use(middleware.Recovery)	r.Use(chimiddleware.NoCache)	r.Use(chimiddleware.RealIP)	r.Use(chimiddleware.RequestID)	r := chi.NewRouter()func New(taskHandler *handlers.TaskHandler) http.Handler {)	"taskmanager/internal/middleware"	"taskmanager/internal/handlers"	chimiddleware "github.com/go-chi/chi/v5/middleware"	"github.com/go-chi/chi/v5"	"net/http"