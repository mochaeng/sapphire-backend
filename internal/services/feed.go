package service

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"go.uber.org/zap"
)

type FeedService struct {
	store  *store.Store
	cfg    *config.Cfg
	logger *zap.SugaredLogger
}

func (s *PostService) Get(ctx context.Context, userID int64, feedQuery *models.PaginateFeedQuery) ([]*models.PostWithMetadata, error) {
	if err := models.Validate.Struct(feedQuery); err != nil {
		return nil, ErrInvalidPayload
	}
	return s.store.Feed.Get(ctx, userID, *feedQuery)
}
