package services

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/models"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

type Service struct {
	User interface {
		Follow(ctx context.Context, followerID int64, followedID int64) error
		Unfollow(ctx context.Context, unfollowerID int64, unfollowedID int64) error
		Activate(ctx context.Context, token string) error
		GetByUsername(ctx context.Context, username string) (*models.User, error)
		GetCached(ctx context.Context, userID int64) (*models.User, error)
	}
	Post interface {
		Create(ctx context.Context, user *models.User, payload *models.CreatePostPayload, file []byte) (*models.Post, error)
		GetWithUser(ctx context.Context, postID int64) (*models.Post, error)
		Delete(ctx context.Context, postID int64) error
		Update(ctx context.Context, post *models.Post, payload *models.UpdatePostPayload) error
	}
	Auth interface {
		RegisterUser(ctx context.Context, payload *models.RegisterUserPayload) (*models.UserInvitation, error)
		CreateUserToken(ctx context.Context, payload *models.CreateUserTokenPayload) (string, error)
		ValidateToken(token string) (*jwt.Token, error)
	}
	Session interface {
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
			serviceCfg.Authenticator,
			serviceCfg.Logger,
		},
		Session: &SessionService{
			store:  serviceCfg.Store,
			cfg:    serviceCfg.Cfg,
			logger: serviceCfg.Logger,
		},
	}
}
