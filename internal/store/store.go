package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
)

var (
	ErrDuplicateEmail      = errors.New("e-mail already taken")
	ErrDuplicateUsername   = errors.New("username already taken")
	ErrForeignKeyViolation = errors.New("no user was found")
)

const QueryTimeoutDuration time.Duration = 5 * time.Second

type Store struct {
	Post interface {
		Create(context.Context, *models.Post) error
		GetByID(context.Context, int64) (*models.Post, error)
		GetByIDWithUser(context.Context, int64) (*models.Post, error)
		DeleteByID(context.Context, int64) error
		UpdateByID(context.Context, *models.Post) error
	}
	User interface {
		GetByID(context.Context, int64) (*models.User, error)

		// GetByUsername returns a user by their username
		GetByUsername(ctx context.Context, username string) (*models.User, error)

		// GetActivatedByUsername returns a user by their username if their are
		// activated in the system
		GetActivatedByUsername(ctx context.Context, username string) (*models.User, error)

		GetByEmail(ctx context.Context, email string) (*models.User, error)
		GetByActivatedEmail(ctx context.Context, email string) (*models.User, error)
		Follow(ctx context.Context, followerID int64, followedID int64) error
		Unfollow(ctx context.Context, followerID int64, followedID int64) error
		CreateAndInvite(ctx context.Context, userInvitation *models.UserInvitation, userProfile *models.UserProfile) error
		Activate(ctx context.Context, plainToken string) error
		Delete(ctx context.Context, userID int64) error
		GetProfile(ctx context.Context, username string) (*models.UserProfile, error)
		GetPostsFrom(ctx context.Context, userPosts *models.UserPosts) ([]*models.Post, error)
		CleanUpExpiredPendingAccounts(ctx context.Context) error

		// this function should only be called during seed
		CreateProfileFull(ctx context.Context, userProfile *models.UserProfile) error
		// this function should only be called during seed
		CreateAndActivate(ctx context.Context, user *models.User) error
	}
	Feed interface {
		Get(ctx context.Context, userID int64, paginateQuery models.PaginateFeedQuery) ([]*models.PostWithMetadata, error)
	}
	Session interface {
		Create(ctx context.Context, session *models.Session) error
		Get(ctx context.Context, sessionID string) (*models.Session, error)
		UpdateExpires(ctx context.Context, session *models.Session) error
		Delete(ctx context.Context, sessionID string) error
	}
	Comment interface {
		GetByPostID(context.Context, int64) (*[]models.Comment, error)
		Create(context.Context, *models.Comment) error
	}
	OAuth interface {
		CreateWithUserActivation(ctx context.Context, oauthAccount *models.OAuthAccount, user *models.User) error
		CreateWithUser(ctx context.Context, oauthAccount *models.OAuthAccount, user *models.User, userProfile *models.UserProfile) error
		GetUserID(ctx context.Context, provider, providerUserID string) (*int64, error)
	}
}

func WithTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
