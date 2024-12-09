package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, userInvitation *models.UserInvitation) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		insert into "user_invitation"(token, user_id, expired)
		values ($1, $2, $3)
	`
	_, err := tx.ExecContext(
		ctx,
		query,
		userInvitation.Token,
		userInvitation.User.ID,
		time.Now().Add(userInvitation.Expired),
	)
	if err != nil {
		return errorUserTransform(err)
	}
	return nil
}
