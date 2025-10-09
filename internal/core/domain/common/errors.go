package common

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternalError = errors.New("internal server error")
)

var (
	ErrInvalidJID          = errors.New("invalid JID format")
	ErrInvalidRecipient    = errors.New("invalid recipient")
	ErrInvalidMessageType  = errors.New("invalid message type")
	ErrEmptyMessageContent = errors.New("message content cannot be empty")
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrContactNotFound = errors.New("contact not found")
)

type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e DomainError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e DomainError) Unwrap() error {
	return e.Cause
}

func NewDomainError(code, message string, cause error) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
