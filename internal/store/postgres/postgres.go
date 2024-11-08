package postgres

import (
	"database/sql"

	"github.com/mochaeng/sapphire-backend/internal/store"
)

func NewPostgresStore(db *sql.DB) *store.Store {
	return &store.Store{
		Post:    &PostStore{db: db},
		User:    &UserStore{db: db},
		Comment: &CommentStore{db: db},
		Feed:    &FeedStore{db: db},
	}
}

func NewPostgresUserStore(db *sql.DB) *UserStore {
	return &UserStore{db}
}

func NewPostgresPostStore(db *sql.DB) *UserStore {
	return &UserStore{db}
}
