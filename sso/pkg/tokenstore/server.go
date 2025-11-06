package tokenstore

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *TokenStore {
	return &TokenStore{rdb: rdb}
}

func New(addr, password string) *TokenStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &TokenStore{rdb: client}
}

func (s *TokenStore) SaveToken(ctx context.Context, userID int64, token string, ttl time.Duration) error {
	if s.rdb == nil {
		panic("REDIS CLIENT IS NIL")
	}
	key := getKey(userID)
	return s.rdb.Set(ctx, key, token, ttl).Err()
}

func (s *TokenStore) GetToken(ctx context.Context, userID int64) (string, error) {
	key := getKey(userID)
	return s.rdb.Get(ctx, key).Result()
}

func (s *TokenStore) DeleteToken(ctx context.Context, userID int64) error {
	key := getKey(userID)
	return s.rdb.Del(ctx, key).Err()
}

func getKey(userID int64) string {
	return "user_token:" + fmt.Sprint(userID)
}
