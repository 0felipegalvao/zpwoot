package webhook

import "errors"

var (
	ErrInvalidURL    = errors.New("invalid webhook URL")
	ErrInvalidSecret = errors.New("invalid webhook secret")
	ErrInvalidEvent  = errors.New("invalid webhook event")
)
