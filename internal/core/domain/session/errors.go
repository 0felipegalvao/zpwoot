package session

import "errors"

var (
	ErrNotConnected  = errors.New("session not connected")
	ErrInvalidStatus = errors.New("invalid session status")
)

var (
	ErrInvalidQRCode = errors.New("invalid QR code")
	ErrQRCodeExpired = errors.New("QR code expired")
)
