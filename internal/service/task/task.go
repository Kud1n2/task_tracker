package service

import (
	"context"
	"log/slog"
	"task_tracker/internal/domain"
)

type Storage interface {
	InsertTask(ctx context.Context, task domain.TaskRequest) error
	SelectTasks(ctx context.Context, filter domain.TaskFilter) ([]domain.TaskResponse, error)
	TaskUpdate(ctx context.Context, update domain.TaskUpdate) error
	SelectHistory(ctx context.Context, task_id int64) ([]domain.TaskHistory, error)
	GetTeamTasks(ctx context.Context, teamID int64) ([]domain.TaskResponse, error)
}

type Cache interface {
	GetTeamTasks(ctx context.Context, teamID int64, dest any) (bool, error)
	SetTeamTasks(ctx context.Context, teamID int64, tasks any) error
	InvalidateCache(ctx context.Context, teamID int64) error
}

type Task struct {
	log     *slog.Logger
	storage Storage
	cache   Cache
}

func New(log *slog.Logger, storage Storage, cache Cache) *Task {
	return &Task{
		log:     log,
		storage: storage,
		cache:   cache,
	}
}

func (t *Task) CreateTask(ctx context.Context, task domain.TaskRequest) error {
	const fn = "service.task.CreateTask"

	log := slog.With(
		slog.String("fn", fn),
	)

	err := t.storage.InsertTask(ctx, task)
	if err != nil {
		log.Error("Failed to create task", slog.Any("error", err))
		return err
	}

	if err := t.cache.InvalidateCache(ctx, task.Team_id); err != nil {
		log.Error("Failed to invalidate cache", slog.Any("error", err))
		return err
	}

	return nil
}

func (t *Task) GetFilteredTasks(ctx context.Context, filter domain.TaskFilter) ([]domain.TaskResponse, error) {
	const fn = "service.task.GetFilteredTasks"

	log := slog.With(
		slog.String("fn", fn),
	)

	tasks, err := t.storage.SelectTasks(ctx, filter)
	if err != nil {
		log.Error("Failed to get tasks", slog.Any("error", err))
		return nil, err
	}

	return tasks, nil
}

func (t *Task) UpdateTask(ctx context.Context, update domain.TaskUpdate) error {
	const fn = "service.task.UpdateTask"

	log := slog.With(
		slog.String("fn", fn),
	)

	err := t.storage.TaskUpdate(ctx, update)
	if err != nil {
		log.Error("Failed to update task", slog.Any("error", err))
		return err
	}

	return nil
}

func (t *Task) GetHistory(ctx context.Context, task_id int64) ([]domain.TaskHistory, error) {
	const fn = "service.tasks.GetHistory"

	log := slog.With(
		slog.String("fn", fn),
	)

	history, err := t.storage.SelectHistory(ctx, task_id)
	if err != nil {
		log.Error("Failed to get history", slog.Any("error", err))
		return nil, err
	}

	return history, nil
}

func (t *Task) GetTeamTasks(ctx context.Context, teamID int64) ([]domain.TaskResponse, error) {
	const fn = "service.task.GetTeamTasks"

	log := slog.With(
		slog.String("fn", fn),
	)

	var tasks []domain.TaskResponse

	hit, err := t.cache.GetTeamTasks(ctx, teamID, &tasks)
	if err != nil {
		log.Error("Failed to get team tasks", slog.Any("error", err))
		return nil, err
	}
	if hit {
		return tasks, nil
	}

	tasks, err = t.storage.GetTeamTasks(ctx, teamID)
	if err != nil {
		log.Error("Failed to get team tasks", slog.Any("error", err))
		return nil, err
	}

	if err := t.cache.SetTeamTasks(ctx, teamID, tasks); err != nil {
		log.Error("Failed to set team tasks", slog.Any("error", err))
		return nil, err
	}

	return tasks, nil
}
