package service

import "errors"

var (
	ErrValidation   = errors.New("validation error")
	ErrInvalidInput = errors.New("invalid input")
)
