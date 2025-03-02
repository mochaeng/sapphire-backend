package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/lib/pq"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"github.com/mochaeng/sapphire-backend/internal/testutils"
)

type UserStore struct {
	db *sql.DB
}

func newTestUserStore(connStr string) *UserStore {
	db := testutils.NewPostgresConnection(connStr)
	store := &UserStore{db}
	return store
}

func (s *UserStore) CreateAndInvite(ctx context.Context, userInvitation *models.UserInvitation, userProfile *models.UserProfile) error {
	return store.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.create(ctx, tx, userInvitation.User); err != nil {
			return err
		}
		if err := s.createProfile(ctx, tx, userProfile); err != nil {
			return err
		}
		if err := s.createUserInvitation(ctx, tx, userInvitation); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select
			u.id, u.first_name, u.last_name, u.email, u.username, u.password, u.created_at, u.is_active, u.role_id, r.name, r.level
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
		&user.Role.Name,
		&user.Role.Level,
	)
	if err != nil {
		return nil, errorUserTransform(err)
	}
	return &user, nil
}

func (s *UserStore) GetByActivatedEmail(ctx context.Context, email string) (*models.User, error) {
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

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select id, first_name, last_name, email, username, password, created_at, role_id
		from "user" where email = $1
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

func (s *UserStore) GetProfile(ctx context.Context, username string) (*models.UserProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		select
			u.username,
			u.first_name,
			u.last_name,
			coalesce(up.description, ''),
			coalesce(up.avatar_url, ''),
			coalesce(up.banner_url, ''),
			coalesce(up."location", ''),
			coalesce(up.user_link, ''),
			up.created_at,
			up.updated_at,
			count(distinct case when f.followed_id = u.id then f.follower_id end) as num_followers,
		    count(distinct case when f.follower_id = u.id then f.followed_id end) as num_following,
		    count(distinct p.id) as num_posts,
		    count(distinct case when p.media_url is null then p.id end) as num_media_posts
		from "user" u
		left join user_profile up on up.user_id = u.id
		left join follower f on (f.follower_id = u.id or f.followed_id = u.id)
		left join post p on p.user_id = u.id
		where username = $1
		group by u.id, up.id;
	`
	var profile models.UserProfile
	profile.User = &models.User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		username,
	).Scan(
		&profile.User.Username,
		&profile.User.FirstName,
		&profile.User.LastName,
		&profile.Description,
		&profile.AvatarURL,
		&profile.BannerURL,
		&profile.Location,
		&profile.UserLink,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.NumFollowers,
		&profile.NumFollowing,
		&profile.NumPosts,
		&profile.NumMediaPosts,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (s *UserStore) GetPosts(ctx context.Context, username string, cursor time.Time, limit int) ([]*models.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		SELECT p.id, p.tittle, p.user_id, p.content, p.media_url, p.tags, p.created_at,
			   p.updated_at, u.username, u.first_name, u.last_name
		FROM post p
		left join "user" u on p.user_id  = u.id
		WHERE username = $1 AND p.created_at < $2
		ORDER BY created_at DESC
		LIMIT $3;
	`
	rows, err := s.db.QueryContext(ctx, query, username, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		post.User = &models.User{}
		err := rows.Scan(
			&post.ID,
			&post.Tittle,
			&post.User.ID,
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
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// Deletes unconfirmed user accounts whose invitation has expired.
func (s *UserStore) CleanUpExpiredPendingAccounts(ctx context.Context) error {
	return store.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		query := `
			SELECT ui.user_id
			FROM user_invitation ui
			JOIN "user" u ON u.id = ui.user_id
			WHERE ui.expired < $1
			  AND u.is_active = false;
		`
		rows, err := tx.QueryContext(ctx, query, time.Now())
		if err != nil {
			return err
		}
		defer rows.Close()

		var userIDs []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return err
			}
			userIDs = append(userIDs, id)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		if len(userIDs) == 0 {
			return nil
		}

		queryInv := `DELETE FROM user_invitation WHERE user_id = ANY($1)`
		if _, err := tx.ExecContext(ctx, queryInv, pq.Array(userIDs)); err != nil {
			return err
		}
		queryProf := `DELETE FROM user_profile WHERE user_id = ANY($1)`
		if _, err := tx.ExecContext(ctx, queryProf, pq.Array(userIDs)); err != nil {
			return err
		}
		queryUser := `DELETE FROM "user" WHERE id = ANY($1)`
		if _, err := tx.ExecContext(ctx, queryUser, pq.Array(userIDs)); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createProfile(ctx context.Context, tx *sql.Tx, userProfile *models.UserProfile) error {
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	query := `
		insert into user_profile (user_id)
		values($1)
		returning id, created_at
	`
	var profile models.UserProfile
	err := tx.QueryRowContext(
		ctx,
		query,
		&userProfile.User.ID,
	).Scan(&profile.ID, &profile.CreatedAt)
	if err != nil {
		return errorUserTransform(err)
	}
	return nil
}

func (s *UserStore) create(ctx context.Context, tx *sql.Tx, user *models.User) error {
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
