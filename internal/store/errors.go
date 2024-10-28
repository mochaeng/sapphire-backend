package store

import "errors"

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")

	ErrInvalidDateFormat = errors.New("invalid date format passed")
)
