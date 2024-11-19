package postgres

import (
	"context"
	"database/sql"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type SessionStore struct {
	db *sql.DB
}

func (s *SessionStore) Create(ctx context.Context, session *models.Session) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		insert into "user_session" (id, user_id, expires_at)
		values($1, $2, $3)
	`
	result, err := s.db.ExecContext(ctx, query, session.ID, session.UserID, session.ExpiresAt)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return err
	}
	return nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select user_session.id, user_session.user_id, user_session.expires_at
		from user_session
		inner join "user" on "user".id = user_session.user_id
		where user_session.id = $1
	`
	var session models.Session
	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionStore) UpdateExpires(ctx context.Context, session *models.Session) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		update "user_session" set expires_at = $1 where id = $2
		returning "user_session".user_id
	`
	err := s.db.QueryRowContext(ctx, query, session.ExpiresAt, session.ID).Scan(&session.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		delete from user_session where id = $1
	`
	result, err := s.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return err
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
