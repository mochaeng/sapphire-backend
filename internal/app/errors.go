package app

import "errors"

var (
	ErrAuthorizationHeaderMissing   = errors.New("authorization header is missing")
	ErrAuthorizationHeaderMalformed = errors.New("authorization header is malformed")
	ErrInvalidCredentials           = errors.New("invalid credentials")

	ErrInvalidOrigin = errors.New("non get request from invalid origin")

	ErrInvalidUserSession      = errors.New("user session is not valid")
	ErrMissingOrEmptyAuthToken = errors.New("auth token is missing or is empty")
	ErrSessionContextNotFound  = errors.New("session was not found on context")
	ErrUserContextNotFound     = errors.New("user was not found on context")
	ErrPostContextNotFound     = errors.New("post was not found on context")
)
