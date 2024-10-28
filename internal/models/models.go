package models

import (
	"time"
)

type Follower struct {
	FollowerID int64  `json:"follower_id"`
	FollowedID int64  `json:"followed_id"`
	CreatedAt  string `json:"created_at"`
}

type UserInvitation struct {
	User    *User         `json:"user"`
	Token   string        `json:"token"`
	Expired time.Duration `json:"expired"`
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

type Role struct {
	ID          int
	Name        string
	Level       int
	Description string
}
