package services

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/models/pagination"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type FeedService struct {
	store *store.Store
}

func (s *FeedService) Get(ctx context.Context, userID int64, feedQuery *pagination.PaginateFeedQuery) ([]*models.PostWithMetadata, error) {
	posts, nextCursor, err := s.store.Feed.Get(ctx, userID, *feedQuery)
	if err != nil {
		return nil, err
	}

	feedQuery.NextCursor = nextCursor

	return posts, nil
}
