package postgres

import (
	"context"
	"database/sql"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) (*[]models.Comment, error) {
	query := `
		select c.id, c.post_id, c.user_id, c."content", c.created_at, u.username, u.first_name, u.last_name from "comment" c join "user" u on u.id  = c.user_id
		where c.post_id = $1
		order by c.created_at desc;
	`
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, errorUserTransform(err)
	}
	defer rows.Close()
	comments := []models.Comment{}
	for rows.Next() {
		var comment models.Comment
		comment.User = models.UserComment{}
		err := rows.Scan(
			&comment.ID,
			&comment.PostId,
			&comment.UserId,
			&comment.Content,
			&comment.CreatedAt,
			&comment.User.Username,
			&comment.User.FirstName,
			&comment.User.LastName,
		)
		if err != nil {
			return nil, errorUserTransform(err)
		}
		comments = append(comments, comment)
	}
	return &comments, nil
}

func (s *CommentStore) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		insert into comment (post_id, user_id, content)
		values ($1, $2, $3)
		returning id, created_at
	`
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx,
		query, comment.PostId,
		comment.UserId,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	if err != nil {
		return errorUserTransform(err)
	}
	return nil
}
