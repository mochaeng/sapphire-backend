package services

import (
	"context"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/markbates/goth"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/models/payloads"
)

var (
	Validate      *validator.Validate
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
	Validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})
}

type Service struct {
	User interface {
		Follow(ctx context.Context, followerID int64, followedID int64) error
		Unfollow(ctx context.Context, unfollowerID int64, unfollowedID int64) error
		Activate(ctx context.Context, token string) error
		GetByUsername(ctx context.Context, username string) (*models.User, error)
		GetCached(ctx context.Context, userID int64) (*models.User, error)
		GetProfile(ctx context.Context, username string) (*models.UserProfile, error)
		GetPosts(ctx context.Context, username string, cursor time.Time, limit int) ([]*models.Post, error)
		LinkOrCreateUserFromOAuth(ctx context.Context, gothUser *goth.User) error
	}
	Post interface {
		Create(ctx context.Context, user *models.User, payload *payloads.CreatePostDataValuesPayload, file []byte) (*models.Post, error)
		GetWithUser(ctx context.Context, postID int64) (*models.Post, error)
		Delete(ctx context.Context, postID int64) error
		Update(ctx context.Context, post *models.Post, payload *payloads.UpdatePostPayload) error
	}
	Auth interface {
		RegisterUser(ctx context.Context, payload *payloads.RegisterUserPayload) (*models.UserInvitation, error)
		Authenticate(ctx context.Context, payload *payloads.SigninPayload) (*models.User, error)
		GenerateSessionToken() (string, error)
		CreateSession(token string, userID int64) (*models.Session, error)
		ValidateSessionToken(token string) (*models.Session, error)
		InvalidateSession(sessionID string) error
	}
	Feed interface {
		Get(ctx context.Context, userID int64, feedQuery *models.PaginateFeedQuery) ([]*models.PostWithMetadata, error)
	}
}

func NewServices(serviceCfg *config.ServiceCfg) *Service {
	return &Service{
		User: &UserService{
			serviceCfg.Store,
			serviceCfg.Cfg,
			serviceCfg.Logger,
			serviceCfg.CacheStore,
		},
		Post: &PostService{
			serviceCfg.Store,
			serviceCfg.Cfg,
			serviceCfg.Logger,
		},
		Auth: &AuthService{
			serviceCfg.Store,
			serviceCfg.Cfg,
			serviceCfg.Mailer,
			serviceCfg.Logger,
		},
	}
}
