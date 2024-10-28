package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		insert into "user"(username, first_name, last_name, email, "password", role_id)
		values($1, $2, $3, $4, $5, $6)
		returning id, created_at
	`
	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.Hash,
		user.Role.ID,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return errorUserTransform(err)
	}
	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select
			u.id, u.first_name, u.last_name, u.email, u.username, u.password, u.created_at, u.is_active, u.role_id, r.level
		from "user" u
		join "role" r on (u.role_id = r.id)
		where u.id = $1 and u.is_active = true;
	`
	var user models.User
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Username,
		&user.Password.Hash,
		&user.CreatedAt,
		&user.IsActive,
		&user.Role.ID,
		&user.Role.Level,
	)
	if err != nil {
		return nil, errorUserTransform(err)
	}
	return &user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select id, first_name, last_name, email, username, password, created_at, role_id
		from "user" where email = $1 and is_active = true
	`
	var user models.User
	err := s.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Username,
		&user.Password.Hash,
		&user.CreatedAt,
		&user.Role.ID,
	)
	if err != nil {
		return nil, errorUserTransform(err)
	}
	return &user, nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select id, first_name, last_name, email, username, created_at, role_id
		from "user" where username = $1 and is_active = true
	`
	var user models.User
	err := s.db.QueryRowContext(
		ctx,
		query,
		username,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.Role.ID,
	)
	if err != nil {
		return nil, errorUserTransform(err)
	}
	return &user, nil
}

func (s *UserStore) Follow(ctx context.Context, followerID int64, followedID int64) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		insert into "follower"(follower_id, followed_id)
		values ($1, $2)
	`
	result, err := s.db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		return errorUserTransform(err)
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

func (s *UserStore) Unfollow(ctx context.Context, followerID int64, followedID int64) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		delete from "follower"
		where follower_id = $1 and followed_id = $2
	`
	result, err := s.db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		return errorUserTransform(err)
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

func (s *UserStore) Activate(ctx context.Context, plainToken string) error {
	return store.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitationToken(ctx, tx, plainToken)
		if err != nil {
			return err
		}
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}
		if err := s.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	return store.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.deleteUserInvitation(ctx, tx, userID); err != nil {
			return err
		}
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		delete from "user" where id = $1
	`
	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return errorUserTransform(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return errorUserTransform(err)
	}
	if count == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		update "user" u set username = $2, email = $3, first_name = $4, last_name = $5, is_active = $6
		where u.id = $1
	`
	result, err := tx.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.IsActive,
	)
	if err != nil {
		return errorUserTransform(err)
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

func (s *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		delete from "user_invitation" where user_id = $1
	`
	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return errorUserTransform(err)
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

func (s *UserStore) getUserFromInvitationToken(ctx context.Context, tx *sql.Tx, plainToken string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select u.id, u.username, u.email, u.first_name, u.last_name, u.is_active from "user" u
		join user_invitation ui on u.id = ui.user_id
		where ui.token = $1 and ui.expired > $2
	`
	hash256 := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash256[:])
	var user models.User
	err := tx.QueryRowContext(
		ctx,
		query,
		hashToken,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
	)
	if err != nil {
		return nil, errorUserTransform(err)
	}
	return &user, nil
}
