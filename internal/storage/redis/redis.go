package redis

import (
	"context"
	"fmt"
	"task_tracker/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

const tasksTTL = 5 * time.Minute

func New(ctx context.Context, cfg config.CacheConfig) (*Cache, error) {
	const fn = "storage.redis.New"

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	return &Cache{client: client}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func teamTasksKey(teamID int64) string {
	return fmt.Sprintf("team:%d:tasks", teamID)
}
