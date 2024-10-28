package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/redis/go-redis/v9"
)

const UserExpTime = time.Minute

type UserStore struct {
	rdb *redis.Client
}

func (s *UserStore) Get(ctx context.Context, userID int64) (*models.User, error) {
	key := fmt.Sprintf("user-%v", userID)
	data, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var user models.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("user-%v", user.ID)
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.SetEx(ctx, key, json, UserExpTime).Err()
}

func (s *UserStore) Invalidate(ctx context.Context, user *models.User) {
	// s.rdb.Del(ctx, key)
}
