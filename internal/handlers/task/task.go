package service

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"task_tracker/internal/domain"
	mw "task_tracker/internal/middleware"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Service interface {
	CreateTask(ctx context.Context, task domain.TaskRequest) error
	GetFilteredTasks(ctx context.Context, filter domain.TaskFilter) ([]domain.TaskResponse, error)
	UpdateTask(ctx context.Context, update domain.TaskUpdate) error
	GetHistory(ctx context.Context, task_id int64) ([]domain.TaskHistory, error)
	GetTeamTasks(ctx context.Context, teamID int64) ([]domain.TaskResponse, error)
}

type Handler struct {
	log     *slog.Logger
	service Service
}

func New(log *slog.Logger, service Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	const fn = "handler.task.CreateTask"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	var task domain.TaskRequest
	if err := render.DecodeJSON(r.Body, &task); err != nil {
		log.Error("Failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Failed to decode request body",
		})
		return
	}

	if task.Title == "" {
		log.Error("Empty task title")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Empty task title",
		})
		return
	}

	creator, ok := r.Context().Value(mw.UserIDKey).(int64)
	if !ok {
		log.Error("failed to get userID")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to get userID",
		})
		return
	}
	task.Creator = creator

	err := h.service.CreateTask(r.Context(), task)
	if err != nil {
		log.Error("Failed to create task", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to create task",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetFilteredTasks(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.tasks.GetFilteredTasks"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	filter := domain.TaskFilter{
		Limit: 10,
		Page:  1,
	}

	query := r.URL.Query()

	if v := query.Get("team_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Error("Failed to parse team_id", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"error":   "Bad request",
				"message": "Failed to parse team_id",
			})
			return
		}
		filter.Team_id = &id
	}
	if v := query.Get("assignee_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Error("Failed to parse assignee_id", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"error":   "Bad request",
				"message": "Failed to parse assignee_id",
			})
			return
		}
		filter.Assignee = &id
	}
	if v := query.Get("status"); v != "" {
		filter.Status = &v
	}
	if v := query.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Failed to parse limit", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"error":   "Bad request",
				"message": "Failed to parse limit",
			})
			return
		}
		filter.Limit = limit
	}
	if v := query.Get("page"); v != "" {
		page, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Failed to parse page", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"error":   "Bad request",
				"message": "Failed to parse page",
			})
			return
		}
		filter.Page = page
	}

	tasks, err := h.service.GetFilteredTasks(r.Context(), filter)
	if err != nil {
		log.Error("Failed to get tasks", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to get tasks",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"tasks": tasks,
	})
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.task.UpdateTask"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid id",
		})
		return
	}

	userID, ok := r.Context().Value(mw.UserIDKey).(int64)
	if !ok {
		log.Error("failed to get userID")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to get userID",
		})
		return
	}

	var status domain.TaskUpdateRequest

	if err := render.DecodeJSON(r.Body, &status); err != nil {
		log.Error("Failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Failed to decode request body",
		})
		return
	}

	update := domain.TaskUpdate{
		Assignee: userID,
		Id:       id,
		Status:   status.Status,
	}

	err = h.service.UpdateTask(r.Context(), update)
	if err != nil {
		log.Error("Failed to update task", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to update task",
		})
		return
	}

	log.Info("Updated successfully")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.task.GetHistory"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid id",
		})
		return
	}

	history, err := h.service.GetHistory(r.Context(), id)
	if err != nil {
		log.Error("Failed to get history", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "failed to get history",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"history": history,
	})
}

func (h *Handler) GetTeamTasks(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.task.GetTeamTasks"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	var task domain.TaskResponse
	if err := render.DecodeJSON(r.Body, &task); err != nil {
		log.Error("Failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Failed to decode request body",
		})
		return
	}

	tasks, err := h.service.GetTeamTasks(r.Context(), task.Team_id)
	if err != nil {
		log.Error("Failed to get team tasks", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "failed to get team tasks",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"tasks": tasks,
	})
}
