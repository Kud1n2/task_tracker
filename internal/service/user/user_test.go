package service_test

import (
	"context"
	"errors"
	"log/slog"
	"task_tracker/internal/domain"
	service "task_tracker/internal/service/user"
	"task_tracker/internal/service/user/mocks"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func logger() *slog.Logger {
	return slog.Default()
}

func TestRegisterSuccess(t *testing.T) {

	storage := mocks.NewUser(t)

	storage.
		EXPECT().
		CreateUser(
			mock.Anything,
			mock.MatchedBy(func(u domain.UserRequest) bool {
				return u.Name == "admin" &&
					u.Password != "" &&
					u.Password != "123456"
			}),
		).
		Return(nil)

	service := service.New(
		logger(),
		storage,
		service.NewJWT("secret"),
	)

	err := service.Register(context.Background(), domain.UserRequest{
		Name:     "admin",
		Password: "123456",
	})

	assert.NoError(t, err)
}

func TestRegisterStorageError(t *testing.T) {

	storage := mocks.NewUser(t)

	storage.
		EXPECT().
		CreateUser(mock.Anything, mock.Anything).
		Return(errors.New("db error"))

	service := service.New(
		logger(),
		storage,
		service.NewJWT("secret"),
	)

	err := service.Register(context.Background(), domain.UserRequest{
		Name:     "admin",
		Password: "123456",
	})

	assert.Error(t, err)
}

func TestLoginSuccess(t *testing.T) {

	hash, _ := bcrypt.GenerateFromPassword(
		[]byte("123456"),
		bcrypt.DefaultCost,
	)

	storage := mocks.NewUser(t)

	storage.
		EXPECT().
		GetUser(mock.Anything, "admin").
		Return(domain.UserResponse{
			ID:       1,
			Name:     "admin",
			Password: string(hash),
		}, nil)

	service := service.New(
		logger(),
		storage,
		service.NewJWT("secret"),
	)

	token, err := service.Login(context.Background(), domain.UserRequest{
		Name:     "admin",
		Password: "123456",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLoginUserNotFound(t *testing.T) {

	storage := mocks.NewUser(t)

	storage.
		EXPECT().
		GetUser(mock.Anything, "admin").
		Return(domain.UserResponse{}, errors.New("not found"))

	service := service.New(
		logger(),
		storage,
		service.NewJWT("secret"),
	)

	_, err := service.Login(context.Background(), domain.UserRequest{
		Name:     "admin",
		Password: "123456",
	})

	assert.Error(t, err)
}

func TestLoginInvalidPassword(t *testing.T) {

	hash, _ := bcrypt.GenerateFromPassword(
		[]byte("qwerty"),
		bcrypt.DefaultCost,
	)

	storage := mocks.NewUser(t)

	storage.
		EXPECT().
		GetUser(mock.Anything, "admin").
		Return(domain.UserResponse{
			ID:       1,
			Name:     "admin",
			Password: string(hash),
		}, nil)

	service := service.New(
		logger(),
		storage,
		service.NewJWT("secret"),
	)

	_, err := service.Login(context.Background(), domain.UserRequest{
		Name:     "admin",
		Password: "123456",
	})

	assert.EqualError(t, err, "invalid credentials")
}
