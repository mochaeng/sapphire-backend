package postgres

import (
	"database/sql"

	"github.com/mochaeng/sapphire-backend/internal/store"
)

func NewPostgresStore(db *sql.DB) *store.Store {
	userStore := &UserStore{db: db}
	return &store.Store{
		Post:    &PostStore{db: db},
		User:    userStore,
		Comment: &CommentStore{db: db},
		Feed:    &FeedStore{db: db},
		Session: &SessionStore{db: db},
		OAuth:   &OAuthStore{db: db, userStore: userStore},
	}
}

func NewPostgresUserStore(db *sql.DB) *UserStore {
	return &UserStore{db}
}

func NewPostgresPostStore(db *sql.DB) *UserStore {
	return &UserStore{db}
}
