package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAthenticator struct {
	secret string
	aud    string
	exp    time.Duration
	iat    time.Duration
	iss    string
	nbf    time.Duration
	sub    int64
}

func NewJWTAuthenticator(secret, iss, aud string, exp time.Duration) *JWTAthenticator {
	return &JWTAthenticator{
		secret: secret,
		exp:    exp,
		iss:    iss,
		aud:    aud,
	}
}

func (a *JWTAthenticator) GenerateToken(subject int64) (string, error) {
	claims := jwt.MapClaims{
		"aud": a.aud,
		"exp": time.Now().Add(a.exp).Unix(),
		"iat": time.Now().Unix(),
		"iss": a.iss,
		"nbf": time.Now().Unix(),
		"sub": subject,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}
	return tokenString, err
}

func (a *JWTAthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(
		token,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
			}
			return []byte(a.secret), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.aud),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
