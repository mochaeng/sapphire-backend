package postgres

import (
	"context"
	"database/sql"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type OAuthStore struct {
	db        *sql.DB
	userStore *UserStore
}

func (s *OAuthStore) CreateWithUserActivation(ctx context.Context, oauthAccount *models.OAuthAccount, user *models.User) error {
	return store.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.create(ctx, tx, oauthAccount); err != nil {
			return err
		}
		if !user.IsActive {
			user.IsActive = true
			if err := s.userStore.update(ctx, tx, user); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *OAuthStore) CreateWithUser(ctx context.Context, oauthAccount *models.OAuthAccount, user *models.User, userProfile *models.UserProfile) error {
	return store.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.userStore.create(ctx, tx, user); err != nil {
			return err
		}
		if err := s.userStore.createProfile(ctx, tx, userProfile); err != nil {
			return err
		}
		oauthAccount.UserID = user.ID
		if err := s.create(ctx, tx, oauthAccount); err != nil {
			return err
		}
		return nil
	})
}

func (s *OAuthStore) GetUserID(ctx context.Context, provider, providerUserID string) (*int64, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	query := `
		select user_id
		from oauth_account oc
		where oc.provider_id = $1 and oc.provider_user_id = $2
	`
	var userID int64
	err := s.db.QueryRowContext(
		ctx,
		query,
		provider,
		providerUserID,
	).Scan(&userID)
	if err != nil {
		return nil, errorUserTransform(err)
	}
	return &userID, nil
}

func (s *OAuthStore) create(ctx context.Context, tx *sql.Tx, oauthAccount *models.OAuthAccount) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	query := `
		insert into oauth_account (provider_id, provider_user_id, user_id)
		values($1, $2, $3)
		returning created_at
	`
	err := tx.QueryRowContext(
		ctx,
		query,
		oauthAccount.ProviderID,
		oauthAccount.ProviderUserID,
		oauthAccount.UserID,
	).Scan(&oauthAccount.CreatedAt)

	return errorOAuthTransform(err)
}
