package handlers

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

type Handler struct {
	log     *slog.Logger
	service Service
}

type Service interface {
	MakeATeam(ctx context.Context, team domain.TeamRequest, owner int64) error
	GetUsersTeams(ctx context.Context, userID int64) ([]domain.TeamResponse, error)
	InviteUser(ctx context.Context, userID, team_id, owner int64) error
	TeamInfo(ctx context.Context) ([]domain.TeamInfo, error)
	BestUsers(ctx context.Context) ([]domain.TopUsers, error)
	ExternalUser(ctx context.Context) ([]domain.NotInTeam, error)
}

func New(log *slog.Logger, service Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) MakeATeam(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.team.MakeATeam"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	var team domain.TeamRequest

	if err := render.DecodeJSON(r.Body, &team); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Failed to decode request body",
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

	err := h.service.MakeATeam(r.Context(), team, userID)
	if err != nil {
		log.Error("failed to get userID", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to make a team",
		})
		return
	}

	log.Info("Team created successfully")
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetUsersTeams(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.teams.GetAllTeams"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

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

	teams, err := h.service.GetUsersTeams(r.Context(), userID)
	if err != nil {
		log.Error("failed to get userID", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to make a team",
		})
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"userID": userID,
		"teams":  teams,
	})
}

func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {
	const fn = "handler.team.InviteUser"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	team_idStr := chi.URLParam(r, "id")
	if team_idStr == "" {
		log.Error("Empty id")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Empty id",
		})
		return
	}

	team_id, err := strconv.ParseInt(team_idStr, 10, 64)
	if err != nil {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid id",
		})
		return
	}

	var user_id domain.TeamInviteRequest

	if err := render.DecodeJSON(r.Body, &user_id); err != nil {
		log.Error("Failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Failed to decode request body",
		})
		return
	}

	owner, ok := r.Context().Value(mw.UserIDKey).(int64)
	if !ok {
		log.Error("failed to get userID")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to get userID",
		})
		return
	}

	err = h.service.InviteUser(r.Context(), user_id.User_id, team_id, owner)
	if err != nil {
		log.Error("failed to get userID")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to invite user",
		})
		return
	}
}

func (h *Handler) TeamsInfo(w http.ResponseWriter, r *http.Request) {
	const fn = "handler.team.TeamsInfo"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	info, err := h.service.TeamInfo(r.Context())
	if err != nil {
		log.Error("failed to get team info")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to get teams info",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"teams": info,
	})
}

func (h *Handler) BestUsers(w http.ResponseWriter, r *http.Request) {
	const fn = "handler.team.BestUsers"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	users, err := h.service.BestUsers(r.Context())
	if err != nil {
		log.Error("failed to get best user")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to get best users",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"users": users,
	})
}

func (h *Handler) ExternalUser(w http.ResponseWriter, r *http.Request) {
	const fn = "handler.team.ExternalUser"

	log := slog.With(
		slog.String("fn", fn),
		slog.String("requestID", middleware.GetReqID(r.Context())),
	)

	users, err := h.service.ExternalUser(r.Context())
	if err != nil {
		log.Error("failed to get external users")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal error",
			"message": "Failed to get external users",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"users": users,
	})
}
