package core

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrInvalidQueryParams  = errors.New("invalid query parameters")
	ErrMissingPathParam    = errors.New("missing path parameter")
	ErrRequestBodyTooLarge = errors.New("request body too large")
)
