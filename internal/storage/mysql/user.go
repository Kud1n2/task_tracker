package mysql

import (
	"context"
	"fmt"
	"task_tracker/internal/domain"
)

func (s *Storage) CreateUser(ctx context.Context, user domain.UserRequest) error {
	const fn = "storage.user.CreateUser"

	_, err := s.db.Exec(`INSERT INTO users(name, password_hash) VALUES (?,?)`, user.Name, user.Password)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, name string) (domain.UserResponse, error) {
	const fn = "storage.user.GetUser"

	var user domain.UserResponse

	err := s.db.QueryRowContext(ctx, `SELECT id, name, password_hash FROM users WHERE name = ?`, name).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		return domain.UserResponse{}, fmt.Errorf("%s:%w", fn, err)
	}
	return user, nil
}
