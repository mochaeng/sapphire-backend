package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/cryptoutils"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"github.com/mochaeng/sapphire-backend/internal/store/cache"
	"go.uber.org/zap"
)

type UserService struct {
	store      *store.Store
	cfg        *config.Cfg
	logger     *zap.SugaredLogger
	cacheStore *cache.Store
}

func (s *UserService) LinkOrCreateUserFromOAuth(ctx context.Context, gothUser *goth.User) (*models.User, error) {
	if gothUser.Provider == "" || gothUser.Email == "" || gothUser.UserID == "" {
		return nil, fmt.Errorf("empty field from OAuth provider")
	}

	var user *models.User

	existingUser, err := s.store.User.GetByEmail(ctx, gothUser.Email)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	oauthAccount := models.OAuthAccount{
		ProviderID:     gothUser.Provider,
		ProviderUserID: gothUser.UserID,
	}

	if existingUser != nil {
		id, err := s.store.OAuth.GetUserID(ctx, gothUser.Provider, gothUser.UserID)
		if err != nil && err != store.ErrNotFound {
			return nil, err
		}
		if id == nil {
			oauthAccount.UserID = existingUser.ID
			err := s.store.OAuth.CreateWithUserActivation(ctx, &oauthAccount, existingUser)
			if err != nil {
				return nil, err
			}
		}
		user = existingUser
	} else {
		randomUsername, err := s.generateUniqueUsername(ctx)
		if err != nil {
			return nil, err
		}

		newUser := models.User{
			Username:  randomUsername,
			FirstName: gothUser.FirstName,
			LastName:  gothUser.LastName,
			Email:     gothUser.Email,
			IsActive:  true,
			Role: models.Role{
				ID: config.Roles["user"].ID,
			},
		}
		userProfile := models.UserProfile{
			User:      &newUser,
			AvatarURL: gothUser.AvatarURL,
		}
		err = s.store.OAuth.CreateWithUser(ctx, &oauthAccount, &newUser, &userProfile)
		if err != nil {
			return nil, err
		}
		user = &newUser
	}

	return user, nil
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

// GenerateUniqueUsername generates a unique username, if somehow there's a collision
// it defaults to a UUID following the usernames rules
func (s *UserService) generateUniqueUsername(ctx context.Context) (string, error) {
	base := "user_"
	maxAttempts := 2
	for range maxAttempts {
		size := models.MaxUsernameSize - len(base)
		suffix, err := cryptoutils.GenerateRandomSuffix(size)
		if err != nil {
			break
		}
		candidate := base + suffix
		_, err = s.store.User.GetByUsername(ctx, candidate)
		if err == store.ErrNotFound {
			return candidate, models.ValidateUsername(candidate)
		}
	}

	finalCandidate := strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	return finalCandidate, models.ValidateUsername(finalCandidate)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if err := models.ValidateUsername(username); err != nil {
		return nil, ErrInvalidPayload
	}

	return s.store.User.GetByUsername(ctx, username)
}

func (s *UserService) GetProfile(ctx context.Context, username string) (*models.UserProfile, error) {
	if err := models.ValidateUsername(username); err != nil {
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

func (s *UserService) GetPostsFromUsername(ctx context.Context, username string, userPosts *models.UserPosts) ([]*models.Post, error) {
	user, err := s.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidPayload
	}

	userPosts.UserID = user.ID
	userPosts.Username = user.Username
	userPosts.FirstName = user.FirstName
	userPosts.LastName = user.LastName

	posts, nextCursor, err := s.store.User.GetPostsFrom(ctx, userPosts)
	if err != nil {
		return nil, err
	}

	userPosts.NextCursor = nextCursor

	return posts, nil
}
