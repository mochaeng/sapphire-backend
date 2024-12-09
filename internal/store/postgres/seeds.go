package postgres

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

// / this function should only be called during seed
func (s *UserStore) CreateAndActivate(ctx context.Context, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		insert into "user"(username, first_name, last_name, email, "password", role_id, is_active)
		values($1, $2, $3, $4, $5, $6, true)
		returning id, created_at
	`
	err := s.db.QueryRowContext(
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

// / this function should only be called during seed
func (s *UserStore) CreateProfileFull(ctx context.Context, userProfile *models.UserProfile) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		insert into public.user_profile
		(user_id, description, avatar_url, banner_url, "location", user_link)
		values($1, $2, $3, $4, $5, $6)
		returning id, created_at
	`
	var profile models.UserProfile
	err := s.db.QueryRowContext(
		ctx,
		query,
		&userProfile.User.ID,
		&userProfile.Description,
		&userProfile.AvatarURL,
		&userProfile.BannerURL,
		&userProfile.Location,
		&userProfile.UserLink,
	).Scan(&profile.ID, &profile.CreatedAt)
	if err != nil {
		return errorUserTransform(err)
	}
	return nil
}
