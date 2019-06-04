package apperror

import "github.com/pkg/errors"

var (
	// ErrorNotFound is helper error for API to show HTTP 404 error
	ErrorNotFound = errors.New("NotFound")
)
