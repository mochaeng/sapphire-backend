package services

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type FeedService struct {
	store *store.Store
	// cfg    *config.Cfg
	// logger *zap.SugaredLogger
}

func (s *FeedService) Get(ctx context.Context, userID int64, feedQuery *models.PaginateFeedQuery) ([]*models.PostWithMetadata, error) {
	if err := models.Validate.Struct(feedQuery); err != nil {
		return nil, ErrInvalidPayload
	}
	return s.store.Feed.Get(ctx, userID, *feedQuery)
}
