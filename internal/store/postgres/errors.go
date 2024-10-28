package postgres

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

const (
	ForeignKeyViolation pq.ErrorCode = "23503"
	UniqueViolation     pq.ErrorCode = "23505"
)

const (
	DuplicateEmailMsg    = `pq: duplicate key value violates unique constraint "user_email_key"`
	DuplicateUsernameMsg = `pq: duplicate key value violates unique constraint "user_username_key"`
)

func errorUserTransform(err error) error {
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == UniqueViolation {
			switch pqErr.Error() {
			case DuplicateEmailMsg:
				return store.ErrDuplicateEmail
			case DuplicateUsernameMsg:
				return store.ErrDuplicateUsername
			}
			return store.ErrConflict
		} else if ok && pqErr.Code == ForeignKeyViolation {
			return store.ForeignKeyViolation
		}
		if err == sql.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}
	return nil
}

func errorRoleTransforme(err error) error {
	if err != nil {
		if err == sql.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}
	return nil
}
