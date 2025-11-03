package redis

import (
	"context"
	"log"

	"github.com/lostmyescape/sso/internal/config"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func NewClient(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return rdb
}
