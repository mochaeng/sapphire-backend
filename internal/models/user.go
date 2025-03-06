package models

import (
	"time"
)

type User struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
	Email     string
	Password  password `json:"-"`
	CreatedAt string
	IsActive  bool
	Role      Role
}

type UserInvitation struct {
	User    *User         `json:"user"`
	Token   string        `json:"token"`
	Expired time.Duration `json:"expired"`
}

type UserProfile struct {
	User          *User
	ID            int64
	Description   string
	AvatarURL     string
	BannerURL     string
	Location      string
	UserLink      string
	NumFollowing  int
	NumFollowers  int
	NumPosts      int
	NumMediaPosts int
	CreatedAt     string
	UpdatedAt     string
}

func ValidateUsername(username string) error {
	return Validate.Struct(struct {
		Username string `validate:"required,username"`
	}{Username: username})
}
