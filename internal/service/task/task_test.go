package service_test

import (
	"context"
	"errors"
	"log/slog"
	"task_tracker/internal/domain"
	service "task_tracker/internal/service/task"
	"task_tracker/internal/service/task/mocks"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/stretchr/testify/mock"
)

func logger() *slog.Logger {
	return slog.Default()
}

func TestCreateTaskSuccess(t *testing.T) {
	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	task := domain.TaskRequest{
		Title:   "Fix bug",
		Team_id: 10,
	}

	storage.EXPECT().
		InsertTask(mock.Anything, task).
		Return(nil)

	cache.EXPECT().
		InvalidateCache(mock.Anything, int64(10)).
		Return(nil)

	err := service.CreateTask(context.Background(), task)

	assert.NoError(t, err)
}

func TestCreateTaskInsertError(t *testing.T) {
	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	task := domain.TaskRequest{
		Title: "Fix bug",
	}

	storage.EXPECT().
		InsertTask(mock.Anything, task).
		Return(errors.New("db error"))

	err := service.CreateTask(context.Background(), task)

	assert.Error(t, err)
}

func TestGetFilteredTasksSuccess(t *testing.T) {
	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	filter := domain.TaskFilter{}

	expected := []domain.TaskResponse{
		{},
	}

	storage.EXPECT().
		SelectTasks(mock.Anything, filter).
		Return(expected, nil)

	tasks, err := service.GetFilteredTasks(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expected, tasks)
}

func TestGetFilteredTasksError(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	filter := domain.TaskFilter{}

	storage.EXPECT().
		SelectTasks(mock.Anything, filter).
		Return(nil, errors.New("db"))

	_, err := service.GetFilteredTasks(context.Background(), filter)

	assert.Error(t, err)
}

func TestUpdateTaskSuccess(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	update := domain.TaskUpdate{}

	storage.EXPECT().
		TaskUpdate(mock.Anything, update).
		Return(nil)

	err := service.UpdateTask(context.Background(), update)

	assert.NoError(t, err)
}

func TestUpdateTaskError(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	update := domain.TaskUpdate{}

	storage.EXPECT().
		TaskUpdate(mock.Anything, update).
		Return(errors.New("db"))

	err := service.UpdateTask(context.Background(), update)

	assert.Error(t, err)
}

func TestGetHistorySuccess(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	expected := []domain.TaskHistory{
		{},
	}

	storage.EXPECT().
		SelectHistory(mock.Anything, int64(1)).
		Return(expected, nil)

	history, err := service.GetHistory(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, history)
}

func TestGetHistoryError(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	storage.EXPECT().
		SelectHistory(mock.Anything, int64(1)).
		Return(nil, errors.New("db"))

	_, err := service.GetHistory(context.Background(), 1)

	assert.Error(t, err)
}

func TestGetTeamTasksCacheHit(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	cache.EXPECT().
		GetTeamTasks(mock.Anything, int64(1), mock.Anything).
		Run(func(ctx context.Context, id int64, dest any) {
			ptr := dest.(*[]domain.TaskResponse)

			*ptr = []domain.TaskResponse{
				{},
			}
		}).
		Return(true, nil)

	tasks, err := service.GetTeamTasks(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, tasks, 1)

	storage.AssertNotCalled(t, "GetTeamTasks")
}

func TestGetTeamTasksCacheMiss(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	expected := []domain.TaskResponse{
		{},
	}

	cache.EXPECT().
		GetTeamTasks(mock.Anything, int64(1), mock.Anything).
		Return(false, nil)

	storage.EXPECT().
		GetTeamTasks(mock.Anything, int64(1)).
		Return(expected, nil)

	cache.EXPECT().
		SetTeamTasks(mock.Anything, int64(1), expected).
		Return(nil)

	tasks, err := service.GetTeamTasks(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, tasks)
}

func TestGetTeamTasksCacheError(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	cache.EXPECT().
		GetTeamTasks(mock.Anything, int64(1), mock.Anything).
		Return(false, errors.New("redis"))

	_, err := service.GetTeamTasks(context.Background(), 1)

	assert.Error(t, err)
}

func TestGetTeamTasksStorageError(t *testing.T) {

	storage := mocks.NewStorage(t)
	cache := mocks.NewCache(t)

	service := service.New(logger(), storage, cache)

	cache.EXPECT().
		GetTeamTasks(mock.Anything, int64(1), mock.Anything).
		Return(false, nil)

	storage.EXPECT().
		GetTeamTasks(mock.Anything, int64(1)).
		Return(nil, errors.New("db"))

	_, err := service.GetTeamTasks(context.Background(), 1)

	assert.Error(t, err)
}
