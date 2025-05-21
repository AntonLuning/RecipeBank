package service

import "errors"

var (
	ErrValidation    = errors.New("validation error")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAIUnsupported = errors.New("AI is not supported")
	ErrAI            = errors.New("AI error")
)
