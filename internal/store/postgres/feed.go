package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/models/pagination"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type FeedStore struct {
	db *sql.DB
}

func (s *FeedStore) Get(ctx context.Context, userID int64, paginateQuery pagination.PaginateFeedQuery) ([]*models.PostWithMetadata, string, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	query := `
		select
			p.id, p.user_id, p."content", p.created_at, p.media_url , u.username,
	 		u.first_name, u.last_name
		from post p
		left join "user" u on p.user_id = u.id
		left join follower f on f.followed_id = p.user_id or p.user_id = $1
		where f.follower_id = $1 and p.created_at < coalesce($2::timestamp, now())
		order by p.created_at desc, p.id desc
		limit $3;
	`
	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
		paginateQuery.Cursor,
		paginateQuery.Limit+1,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var posts []*models.PostWithMetadata
	for rows.Next() {
		post := &models.PostWithMetadata{}
		post.User = &models.User{}
		err := rows.Scan(
			&post.ID,
			&post.User.ID,
			&post.Content,
			&post.CreatedAt,
			&post.Media,
			&post.User.Username,
			&post.User.FirstName,
			&post.User.LastName,
		)
		if err != nil {
			return nil, "", err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(posts) > paginateQuery.Limit {
		nextCursor = posts[paginateQuery.Limit-1].CreatedAt.Format(time.RFC3339Nano)
		posts = posts[:paginateQuery.Limit]
	}

	return posts, nextCursor, nil
}
