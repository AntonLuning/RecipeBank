package storage

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidID     = errors.New("invalid ID format")
	ErrDatabaseError = errors.New("database error")
)
