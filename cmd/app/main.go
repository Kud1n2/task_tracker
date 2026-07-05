package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"task_tracker/internal/config"
	"task_tracker/internal/email"
	task_handler "task_tracker/internal/handlers/task"
	team_handler "task_tracker/internal/handlers/team"
	user_handler "task_tracker/internal/handlers/user"
	mw "task_tracker/internal/middleware"
	task_service "task_tracker/internal/service/task"
	team_service "task_tracker/internal/service/team"
	user_service "task_tracker/internal/service/user"
	"task_tracker/internal/storage/mysql"
	"task_tracker/internal/storage/redis"
	metrics "task_tracker/pkg/metrics"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application")

	storageCtx, storageCancel := context.WithTimeout(context.Background(), cfg.DatabaseTimeout)
	defer storageCancel()
	redisCtx, redisCancel := context.WithTimeout(context.Background(), cfg.DatabaseTimeout)
	defer redisCancel()

	storage, err := mysql.New(storageCtx, cfg.StorageConfig)
	if err != nil {
		log.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	cache, err := redis.New(redisCtx, cfg.CacheConfig)
	if err != nil {
		log.Error("failed to init cache", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer cache.Close()

	rateLimiter := mw.NewRateLimiter(cfg.CacheConfig)
	metrics.Register()

	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("failed to close storage", slog.String("error", err.Error()))
		}
		log.Info("storage closed")
	}()

	router := chi.NewRouter()

	router.Use(mw.Metrics)
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	newJWT := user_service.NewJWT(cfg.JWT.Secret)
	userService := user_service.New(log, storage, newJWT)
	userHandler := user_handler.New(log, userService)

	emailService := &email.Mock{}
	teamService := team_service.New(log, storage, emailService)
	teamHandler := team_handler.New(log, teamService)

	taskService := task_service.New(log, storage, cache)
	taskHandler := task_handler.New(log, taskService)

	router.Route("/api/v1", func(chi.Router) {
		router.Handle("/metrics", promhttp.Handler())
		router.Post("/register", userHandler.Register)
		router.Post("/login", userHandler.Login)

		router.Group(func(router chi.Router) {
			router.Use(mw.Auth([]byte(cfg.JWT.Secret)))
			router.Use(rateLimiter.Middleware)

			router.Post("/teams", teamHandler.MakeATeam)
			router.Get("/teams", teamHandler.GetUsersTeams)
			router.Post("/teams/{id}/invite", teamHandler.InviteUser)
			router.Post("/tasks", taskHandler.CreateTask)
			router.Put("/tasks/{id}", taskHandler.UpdateTask)
			router.Get("/tasks/{id}/history", taskHandler.GetHistory)
			router.Get("/tasks", taskHandler.GetFilteredTasks)
			router.Get("/teams/info", teamHandler.TeamsInfo)
			router.Get("/teams/top", teamHandler.BestUsers)
			router.Get("/teams/external", teamHandler.ExternalUser)
			router.Get("/tasks/teams", taskHandler.GetTeamTasks)
		})
	})

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		log.Info("Starting server", slog.String("Address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", slog.String("error", err.Error()))
	}

	log.Info("Server stopped gracefully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
