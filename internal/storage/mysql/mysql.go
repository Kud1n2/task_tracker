package mysql

import (
	"context"
	"fmt"
	"task_tracker/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func New(ctx context.Context, cfg config.StorageConfig) (*Storage, error) {
	const fn = "storage.mysql.New"

	db, err := sqlx.Connect("mysql", cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}
