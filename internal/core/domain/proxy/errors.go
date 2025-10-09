package proxy

import "errors"

var (
	ErrInvalidHost     = errors.New("invalid proxy host")
	ErrInvalidPort     = errors.New("invalid proxy port")
	ErrInvalidProtocol = errors.New("invalid proxy protocol")
)
