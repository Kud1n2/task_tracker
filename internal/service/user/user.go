package service

import (
	"context"
	"errors"
	"log/slog"
	"task_tracker/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	log     *slog.Logger
	storage User
	jwt     *JWTManager
}

type User interface {
	CreateUser(ctx context.Context, user domain.UserRequest) error
	GetUser(ctx context.Context, name string) (domain.UserResponse, error)
}

func New(log *slog.Logger, storage User, jwt *JWTManager) *Service {
	return &Service{
		log:     log,
		storage: storage,
		jwt:     jwt,
	}
}

func (s *Service) Register(ctx context.Context, user domain.UserRequest) error {
	const fn = "service.user.Register"

	log := s.log.With(
		slog.String("fn", fn),
	)

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate hash", slog.Any("error", err))
		return err
	}

	user_hash := domain.UserRequest{
		Name:     user.Name,
		Password: string(hash),
	}

	err = s.storage.CreateUser(ctx, user_hash)
	if err != nil {
		log.Error("failed to create user", slog.Any("error", err))
		return err
	}

	return nil
}

func (s *Service) Login(ctx context.Context, user domain.UserRequest) (string, error) {
	const fn = "service.user.Login"

	log := slog.With(
		slog.String("fn", fn),
	)

	user_hash, err := s.storage.GetUser(ctx, user.Name)
	if err != nil {
		log.Error("Failed to get user", slog.Any("error", err))
		return "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user_hash.Password),
		[]byte(user.Password),
	)
	if err != nil {
		log.Error("invalid credentials", slog.Any("error", err))
		return "", errors.New("invalid credentials")
	}

	return s.jwt.Generate(int64(user_hash.ID))
}
