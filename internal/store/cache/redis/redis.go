package redis

import (
	"github.com/mochaeng/sapphire-backend/internal/store/cache"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func NewRedisStore(rdb *redis.Client) cache.Store {
	return cache.Store{
		User: &UserStore{rdb: rdb},
	}
}
