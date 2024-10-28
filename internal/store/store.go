package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
)

var (
	ErrDuplicateEmail    = errors.New("e-mail already taken")
	ErrDuplicateUsername = errors.New("username already taken")
	ForeignKeyViolation  = errors.New("no user was found")
)

const QueryTimeoutDuration time.Duration = 5 * time.Second

type Store struct {
	Post interface {
		Create(context.Context, *models.Post) error
		GetByID(context.Context, int64) (*models.Post, error)
		GetByIDWithUser(context.Context, int64) (*models.Post, error)
		DeleteByID(context.Context, int64) error
		UpdateByID(context.Context, *models.Post) error
		GetAllByUsername(context.Context, string) ([]models.Post, error)
	}
	User interface {
		Create(ctx context.Context, tx *sql.Tx, user *models.User) error
		GetByID(context.Context, int64) (*models.User, error)
		GetByUsername(ctx context.Context, username string) (*models.User, error)
		GetByEmail(ctx context.Context, email string) (*models.User, error)
		Follow(ctx context.Context, followerID int64, followedID int64) error
		Unfollow(ctx context.Context, followerID int64, followedID int64) error
		CreateAndInvite(ctx context.Context, userInvitation *models.UserInvitation) error
		Activate(ctx context.Context, plainToken string) error
		Delete(ctx context.Context, userID int64) error
	}
	Feed interface {
		Get(ctx context.Context, userID int64, paginateQuery PaginateFeedQuery) ([]*models.PostWithMetadata, error)
	}
	Comment interface {
		GetByPostID(context.Context, int64) (*[]models.Comment, error)
		Create(context.Context, *models.Comment) error
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
