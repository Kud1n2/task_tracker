package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (c *Cache) GetTeamTasks(ctx context.Context, teamID int64, dest any) (bool, error) {
	const fn = "storage.redis.task.GetTeamTasks"

	data, err := c.client.Get(ctx, teamTasksKey(teamID)).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%s:%w", fn, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("Failed to unmarshal tasks:%s:%w", fn, err)
	}

	return true, nil
}

func (c *Cache) SetTeamTasks(ctx context.Context, teamID int64, tasks any) error {
	const fn = "storage.redis.SetTeamTasks"

	data, err := json.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("Failed to marshal tasks:%s:%w", fn, err)
	}

	return c.client.Set(ctx, teamTasksKey(teamID), data, tasksTTL).Err()
}

func (c *Cache) InvalidateCache(ctx context.Context, teamID int64) error {
	return c.client.Del(ctx, teamTasksKey(teamID)).Err()
}
