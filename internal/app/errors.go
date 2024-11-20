package app

import "errors"

var (
	ErrAuthorizationHeaderMissing   = errors.New("authorization header is missing")
	ErrAuthorizationHeaderMalformed = errors.New("authorization header is malformed")
	ErrInvalidCredentials           = errors.New("invalid credentials")

	ErrInvalidOrigin = errors.New("non get request from invalid origin")
)
