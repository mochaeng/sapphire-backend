package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"github.com/mochaeng/sapphire-backend/internal/store/cache"
	"go.uber.org/zap"
)

const (
	MaxUsernameSize = 18
	MinUsernameSize = 2
)

type UserService struct {
	store      *store.Store
	cfg        *config.Cfg
	logger     *zap.SugaredLogger
	cacheStore *cache.Store
}

func (s *UserService) LinkOrCreateUserFromOAuth(ctx context.Context, gothUser *goth.User) error {
	if gothUser.Provider == "" || gothUser.Email == "" || gothUser.UserID == "" {
		return fmt.Errorf("empty field from OAuth provider")
	}

	existingUser, err := s.store.User.GetByEmail(ctx, gothUser.Email)
	if err != nil && err != store.ErrNotFound {
		return err
	}

	oauthAccount := models.OAuthAccount{
		ProviderID:     gothUser.Provider,
		ProviderUserID: gothUser.UserID,
	}

	if existingUser != nil {
		id, err := s.store.OAuth.GetUserID(ctx, gothUser.Provider, gothUser.UserID)
		if err != nil && err != store.ErrNotFound {
			return err
		}
		if id == nil {
			oauthAccount.UserID = existingUser.ID
			err := s.store.OAuth.CreateWithUserActivation(ctx, &oauthAccount, existingUser)
			if err != nil {
				return err
			}
		}
	} else {
		user := models.User{
			Username:  uuid.NewString(),
			FirstName: gothUser.FirstName,
			LastName:  gothUser.LastName,
			Email:     gothUser.Email,
			IsActive:  true,
			Role: models.Role{
				ID: config.Roles["user"].ID,
			},
		}
		s.logger.Infow("creating user and oatuh", "user", user, "oauth", oauthAccount)
		if err := s.store.OAuth.CreateWithUser(ctx, &oauthAccount, &user); err != nil {
			return err
		}
	}

	return nil
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

func (s *UserService) GetProfile(ctx context.Context, username string) (*models.UserProfile, error) {
	if len(username) < MinUsernameSize || len(username) > MaxUsernameSize {
		return nil, ErrInvalidPayload
	}
	return s.store.User.GetProfile(ctx, username)
}

func (s *UserService) Follow(ctx context.Context, followerID int64, followedID int64) error {
	if followerID == followedID {
		return ErrOperationNotAllowed
	}
	return s.store.User.Follow(ctx, followerID, followedID)
}

func (s *UserService) Unfollow(ctx context.Context, unfollowerID int64, unfollowedID int64) error {
	if unfollowerID == unfollowedID {
		return ErrOperationNotAllowed
	}
	return s.store.User.Unfollow(ctx, unfollowerID, unfollowedID)
}

func (s *UserService) Activate(ctx context.Context, token string) error {
	return s.store.User.Activate(ctx, token)
}

func (s *UserService) GetPosts(ctx context.Context, username string, cursor time.Time, limit int) ([]*models.Post, error) {
	return s.store.User.GetPosts(ctx, username, cursor, limit)
}
