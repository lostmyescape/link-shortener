package redis

import (
	"context"
	"fmt"

	"github.com/lostmyescape/link-shortener/sso/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("Failed to connect to Redis: %v", err)
		return nil, err
	}

	return rdb, nil
}
