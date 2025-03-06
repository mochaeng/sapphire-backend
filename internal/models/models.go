package models

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	MaxUsernameSize = 16
	MinUsernameSize = 3
)

var (
	Validate      *validator.Validate
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	Validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		usernameSize := len(fl.Field().String())
		isCorrectSize := usernameSize <= MaxUsernameSize && usernameSize >= MinUsernameSize
		return usernameRegex.MatchString(fl.Field().String()) && isCorrectSize
	})
}

type Follower struct {
	FollowerID int64  `json:"follower_id"`
	FollowedID int64  `json:"followed_id"`
	CreatedAt  string `json:"created_at"`
}

type UserComment struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Comment struct {
	ID        int64       `json:"id"`
	PostId    int64       `json:"post_id"`
	UserId    int64       `json:"user_id"`
	Content   string      `json:"content"`
	CreatedAt string      `json:"created_at"`
	User      UserComment `json:"user"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Role struct {
	ID          int
	Name        string
	Level       int
	Description string
}

type OAuthAccount struct {
	ProviderID     string
	ProviderUserID string
	UserID         int64
	CreatedAt      string
}
