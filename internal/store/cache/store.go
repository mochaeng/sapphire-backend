package cache

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/models"
)

type Store struct {
	User interface {
		Get(ctx context.Context, userID int64) (*models.User, error)
		Set(ctx context.Context, user *models.User) error
	}
}
