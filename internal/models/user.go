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

type UserResponse struct {
	ID        int64  `json:"id,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type GetUserResponse struct {
	*UserResponse
}

type CreateUserPayload struct{}

type UpdateUserPayload struct{}

type DeleteUserPayload struct{}

type RegisterUserPayload struct {
	Username  string `json:"username" validate:"required,max=16,min=3"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=72"`
	FirstName string `json:"first_name" validate:"required,min=2,max=30"`
}

type RegisterUserResponse struct {
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	IsActive  bool   `json:"is_active"`
	Token     string `json:"token"`
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}
