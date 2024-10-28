package postgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type FeedStore struct {
	db *sql.DB
}

func (s *FeedStore) Get(ctx context.Context, userID int64, paginateQuery store.PaginateFeedQuery) ([]*models.PostWithMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select
				p.id, p.user_id, p.tittle, p."content", p.created_at, p.tags, u.username, u.first_name, u.last_name, count(c.id) as comment_count
		from post p
		left join "comment" c on c.post_id = p.id
		left join "user" u on p.user_id = u.id
		join follower f on f.followed_id = p.user_id or p.user_id = $1
		where
				f.follower_id = $1 and
				(p.tittle ilike '%' || $4 || '%' or p."content" ilike '%' || $4 || '%') and
				(
					p.tags @> $5::varchar[] or array_length($5::varchar[], 1) is null
				)
		group by p.id, u.username, u.first_name, u.last_name
		order by p.created_at ` + paginateQuery.Sort + `
		limit $2 offset $3
	`
	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
		paginateQuery.Limit,
		paginateQuery.Offset,
		paginateQuery.Search,
		pq.Array(paginateQuery.Tags),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feed []*models.PostWithMetadata
	for rows.Next() {
		var post models.PostWithMetadata
		post.User = &models.User{}
		err := rows.Scan(
			&post.ID,
			&post.User.ID,
			&post.Tittle,
			&post.Content,
			&post.CreatedAt,
			pq.Array(&post.Tags),
			&post.User.Username,
			&post.User.FirstName,
			&post.User.LastName,
			&post.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, &post)
	}
	return feed, nil
}
