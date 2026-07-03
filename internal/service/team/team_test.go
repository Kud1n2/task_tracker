package service_test

import (
	"context"
	"errors"
	"log/slog"
	"task_tracker/internal/domain"
	email "task_tracker/internal/email"
	service "task_tracker/internal/service/team"
	"task_tracker/internal/service/team/mocks"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/stretchr/testify/mock"
)

func logger() *slog.Logger {
	return slog.Default()
}

func TestMakeATeamSuccess(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		CreateTeam(
			mock.Anything,
			"Backend",
			int64(1),
		).
		Return(nil)

	service := service.New(logger(), storage, email)

	err := service.MakeATeam(
		context.Background(),
		domain.TeamRequest{
			Name: "Backend",
		},
		1,
	)

	assert.NoError(t, err)
}

func TestMakeATeamError(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		CreateTeam(mock.Anything, "Backend", int64(1)).
		Return(errors.New("db error"))

	service := service.New(logger(), storage, email)

	err := service.MakeATeam(
		context.Background(),
		domain.TeamRequest{
			Name: "Backend",
		},
		1,
	)

	assert.Error(t, err)
}

func TestGetUsersTeamsSuccess(t *testing.T) {

	expected := []domain.TeamResponse{
		{
			Team_name: "Backend",
			Owner:     1,
		},
	}

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		GetTeams(mock.Anything, int64(1)).
		Return(expected, nil)

	service := service.New(logger(), storage, email)

	teams, err := service.GetUsersTeams(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, teams)
}

func TestGetUsersTeamsError(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		GetTeams(mock.Anything, int64(1)).
		Return(nil, errors.New("db"))

	service := service.New(logger(), storage, email)

	_, err := service.GetUsersTeams(context.Background(), 1)

	assert.Error(t, err)
}

func TestInviteUserSuccess(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		AddUser(
			mock.Anything,
			int64(10),
			int64(5),
			int64(1),
		).
		Return(nil)

	service := service.New(logger(), storage, email)

	err := service.InviteUser(
		context.Background(),
		5,
		10,
		1,
	)

	assert.NoError(t, err)
}

func TestInviteUserError(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		AddUser(mock.Anything, int64(10), int64(5), int64(1)).
		Return(errors.New("db"))

	service := service.New(logger(), storage, email)

	err := service.InviteUser(
		context.Background(),
		5,
		10,
		1,
	)

	assert.Error(t, err)
}

func TestTeamInfoSuccess(t *testing.T) {

	expected := []domain.TeamInfo{
		{},
	}

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		GetTeamsInfo(mock.Anything).
		Return(expected, nil)

	service := service.New(logger(), storage, email)

	info, err := service.TeamInfo(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expected, info)
}

func TestTeamInfoError(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		GetTeamsInfo(mock.Anything).
		Return(nil, errors.New("db"))

	service := service.New(logger(), storage, email)

	_, err := service.TeamInfo(context.Background())

	assert.Error(t, err)
}

func TestBestUsersSuccess(t *testing.T) {

	expected := []domain.TopUsers{
		{},
	}

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		TopUsers(mock.Anything).
		Return(expected, nil)

	service := service.New(logger(), storage, email)

	users, err := service.BestUsers(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expected, users)
}

func TestBestUsersError(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		TopUsers(mock.Anything).
		Return(nil, errors.New("db"))

	service := service.New(logger(), storage, email)

	_, err := service.BestUsers(context.Background())

	assert.Error(t, err)
}

func TestExternalUserSuccess(t *testing.T) {

	expected := []domain.NotInTeam{
		{},
	}

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		NotInTeam(mock.Anything).
		Return(expected, nil)

	service := service.New(logger(), storage, email)

	users, err := service.ExternalUser(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expected, users)
}

func TestExternalUserError(t *testing.T) {

	storage := mocks.NewStorage(t)
	email := &email.Mock{}

	storage.EXPECT().
		NotInTeam(mock.Anything).
		Return(nil, errors.New("db"))

	service := service.New(logger(), storage, email)

	_, err := service.ExternalUser(context.Background())

	assert.Error(t, err)
}
