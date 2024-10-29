package service

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"github.com/mochaeng/sapphire-backend/internal/store/cache"
	"go.uber.org/zap"
)

const (
	MaxUsernameSize = 16
	MinUsernameSize = 3
)

type UserService struct {
	store      *store.Store
	cfg        *config.Cfg
	logger     *zap.SugaredLogger
	cacheStore *cache.Store
}

func (s *UserService) GetCached(ctx context.Context, userID int64) (*models.User, error) {
	if !s.cfg.Cacher.IsEnable {
		return s.store.User.GetByID(ctx, userID)
	}
	s.logger.Infow("cache hit", "userID", userID)
	user, err := s.cacheStore.User.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		s.logger.Infow("fetching from the database", "id", userID)
		user, err = s.store.User.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if err := s.cacheStore.User.Set(ctx, user); err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if len(username) < MinUsernameSize || len(username) > MaxUsernameSize {
		return nil, ErrInvalidPayload
	}
	return s.store.User.GetByUsername(ctx, username)
}

func (s *UserService) Follow(ctx context.Context, followerID int64, followedID int64) error {
	return s.store.User.Follow(ctx, followerID, followedID)
}

func (s *UserService) Unfollow(ctx context.Context, unfollowerID int64, unfollowedID int64) error {
	return s.store.User.Unfollow(ctx, unfollowerID, unfollowedID)
}

func (s *UserService) Activate(ctx context.Context, token string) error {
	return s.store.User.Activate(ctx, token)
}
