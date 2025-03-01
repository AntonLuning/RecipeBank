package core

import "errors"

var (
	ErrInvalidQueryParams  = errors.New("invalid query parameters")
	ErrMissingPathParam    = errors.New("missing path parameter")
	ErrRequestBodyTooLarge = errors.New("request body too large")
	ErrJSONDecode          = errors.New("invalid JSON")
)
