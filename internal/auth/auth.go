package auth

import "github.com/golang-jwt/jwt/v5"

type Authenticator interface {
	GenerateToken(subject int64) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}
