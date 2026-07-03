package service

import (
	"context"
	"log/slog"
	"task_tracker/internal/domain"
)

type Team struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	CreateTeam(ctx context.Context, name string, owner int64) error
	GetTeams(ctx context.Context, userID int64) ([]domain.TeamResponse, error)
	AddUser(ctx context.Context, team_id, userID, owner int64) error
	GetTeamsInfo(ctx context.Context) ([]domain.TeamInfo, error)
	TopUsers(ctx context.Context) ([]domain.TopUsers, error)
	NotInTeam(ctx context.Context) ([]domain.NotInTeam, error)
}

func New(log *slog.Logger, storage Storage) *Team {
	return &Team{
		log:     log,
		storage: storage,
	}
}

func (t *Team) MakeATeam(ctx context.Context, team domain.TeamRequest, owner int64) error {
	const fn = "service.team.MakeATeam"

	log := slog.With(
		slog.String("fn", fn),
	)

	err := t.storage.CreateTeam(ctx, team.Name, owner)
	if err != nil {
		log.Error("failed to create team", slog.Any("error", err))
		return err
	}

	return nil
}

func (t *Team) GetUsersTeams(ctx context.Context, userID int64) ([]domain.TeamResponse, error) {
	const fn = "service.team.GetAllTeams"

	log := slog.With(
		slog.String("fn", fn),
	)

	teams, err := t.storage.GetTeams(ctx, userID)
	if err != nil {
		log.Error("failed to get teams", slog.Any("error", err))
		return nil, err
	}

	return teams, nil
}

func (t *Team) InviteUser(ctx context.Context, userID, team_id, owner int64) error {
	const fn = "service.team.InviteUser"

	log := slog.With(
		slog.String("fn", fn),
	)

	err := t.storage.AddUser(ctx, team_id, userID, owner)
	if err != nil {
		log.Error("Failed to add user to the team", slog.Any("error", err))
		return err
	}

	return nil
}

func (t *Team) TeamInfo(ctx context.Context) ([]domain.TeamInfo, error) {
	const fn = "service.team.TeamInfo"

	log := slog.With(
		slog.String("fn", fn),
	)

	info, err := t.storage.GetTeamsInfo(ctx)
	if err != nil {
		log.Error("Failed to get teams info", slog.Any("error", err))
		return nil, err
	}

	return info, nil
}

func (t *Team) BestUsers(ctx context.Context) ([]domain.TopUsers, error) {
	const fn = "service.team.BestUsers"

	log := slog.With(
		slog.String("fn", fn),
	)

	users, err := t.storage.TopUsers(ctx)
	if err != nil {
		log.Error("Failed to get top users", slog.Any("error", err))
		return nil, err
	}

	return users, nil
}

func (t *Team) ExternalUser(ctx context.Context) ([]domain.NotInTeam, error) {
	const fn = "service.team.ExternalUser"

	log := slog.With(
		slog.String("fn", fn),
	)

	users, err := t.storage.NotInTeam(ctx)
	if err != nil {
		log.Error("Failed to get external users", slog.Any("error", err))
		return nil, err
	}

	return users, nil
}
