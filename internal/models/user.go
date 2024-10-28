package models

import "golang.org/x/crypto/bcrypt"

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

type password struct {
	Hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.Hash = hash
	return nil
}
