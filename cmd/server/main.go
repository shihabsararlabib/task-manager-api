package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskmanager/internal/auth"
	"taskmanager/internal/config"
	"taskmanager/internal/database"
	"taskmanager/internal/handlers"
	"taskmanager/internal/repository"
	"taskmanager/internal/router"
	"taskmanager/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	defer dbPool.Close()

	taskRepo := repository.NewPostgresTaskRepository(dbPool)
	userRepo := repository.NewPostgresUserRepository(dbPool)
	refreshRepo := repository.NewPostgresRefreshTokenRepository(dbPool)
	taskService := service.NewTaskService(taskRepo)
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		time.Duration(cfg.AccessTokenTTLMin)*time.Minute,
		time.Duration(cfg.RefreshTokenTTLHr)*time.Hour,
	)
	authService := service.NewAuthService(userRepo, refreshRepo, jwtManager)
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(authService)
	taskHandler := handlers.NewTaskHandler(taskService)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router.New(jwtManager, authHandler, adminHandler, taskHandler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("task manager API running on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
