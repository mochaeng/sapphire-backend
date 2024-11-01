package postgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type PostStore struct {
	db *sql.DB
}

func newTestPostStore(connStr string) *PostStore {
	db := createDB(connStr)
	store := &PostStore{db}
	return store
}

func (s *PostStore) Create(ctx context.Context, post *models.Post) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		INSERT INTO post (content, tittle, user_id, media_url, tags)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Tittle,
		post.User.ID,
		post.Media,
		pq.Array(post.Tags),
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return errorPostTransform(err)
	}
	return nil
}

func (s *PostStore) GetByID(ctx context.Context, postID int64) (*models.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		SELECT id, user_id, tittle, content, media_url, tags, created_at, updated_at
		FROM post
		WHERE id = $1
	`
	var post models.Post
	post.User = &models.User{}
	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.User.ID,
		&post.Tittle,
		&post.Content,
		&post.Media,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, errorPostTransform(err)
	}
	return &post, nil
}

func (s *PostStore) GetByIDWithUser(ctx context.Context, postID int64) (*models.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		SELECT p.id, p.user_id, p.tittle, p.content, p.media_url, p.tags, p.created_at, p.updated_at, u.username, u.first_name , u.last_name
		FROM post p
		join "user" u on p.user_id = u.id
		WHERE p.id = $1;
	`
	var post models.Post
	post.User = &models.User{}
	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.User.ID,
		&post.Tittle,
		&post.Content,
		&post.Media,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.User.Username,
		&post.User.FirstName,
		&post.User.LastName,
	)
	if err != nil {
		return nil, errorPostTransform(err)
	}
	return &post, nil
}

func (s *PostStore) GetAllByUsername(ctx context.Context, username string) ([]models.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select p.id, p.created_at, p.tittle, p."content", p.tags, p.media_url, u.first_name, u.last_name, u.username
		from "user" u join post p on u.id = p.user_id
		where username = $1;
	`
	rows, err := s.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		post := models.Post{}
		post.User = &models.User{}
		err := rows.Scan(
			&post.ID,
			&post.CreatedAt,
			&post.Tittle,
			&post.Content,
			&post.Tags,
			&post.Media,
			&post.User.FirstName,
			&post.User.LastName,
			&post.User.Username,
		)
		if err != nil {
			return posts, errorUserTransform(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	return posts, nil
}

func (s *PostStore) DeleteByID(ctx context.Context, postID int64) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		delete from post
		where id = $1
	`
	result, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return errorPostTransform(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *PostStore) UpdateByID(ctx context.Context, post *models.Post) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		update "post"
		set tittle = $2, "content" = $3
		where id = $1
		returning post.tittle, post."content", post.updated_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.ID,
		post.Tittle,
		post.Content,
	).Scan(
		&post.Tittle,
		&post.Content,
		&post.UpdatedAt,
	)
	if err != nil {
		return errorPostTransform(err)
	}
	return nil
}
