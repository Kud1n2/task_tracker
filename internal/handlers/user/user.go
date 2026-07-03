package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"task_tracker/internal/domain"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type User interface {
	Register(ctx context.Context, user domain.UserRequest) error
	Login(ctx context.Context, user domain.UserRequest) (string, error)
}

type Handler struct {
	log     *slog.Logger
	service User
}

func New(log *slog.Logger, service User) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.Register"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var user domain.UserRequest

	if err := render.DecodeJSON(r.Body, &user); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "failed do decode request body",
		})
		return
	}

	if user.Name == "" {
		log.Error("empty name")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "user must have a name",
		})
		return
	}
	if user.Password == "" {
		log.Error("empty password")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "user must have a password",
		})
		return
	}

	err := h.service.Register(r.Context(), user)
	if err != nil {
		log.Error("failed to register user", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal Error",
			"message": "Failed to register user",
		})
		return
	}

	log.Info("User registered")
	render.JSON(w, r, map[string]string{
		"message": "user registered successfully",
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.Login"

	log := slog.With(
		slog.String("requestID", middleware.GetReqID(r.Context())),
		slog.String("fn", fn),
	)

	var user domain.UserRequest

	if err := render.DecodeJSON(r.Body, &user); err != nil {
		log.Error("failed to decode request body", slog.Any("err", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "failed to decode request body",
		})
		return
	}

	if user.Name == "" {
		log.Error("empty name")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "user must have a name",
		})
		return
	}
	if user.Password == "" {
		log.Error("empty password")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "user must have a password",
		})
		return
	}

	token, err := h.service.Login(r.Context(), user)
	if err != nil {
		log.Error("Failed to login", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "failed to login user",
		})
		return
	}

	log.Info("User logged in")
	render.JSON(w, r, map[string]string{
		"token":  token,
		"status": "user logged in successfully",
	})
}
