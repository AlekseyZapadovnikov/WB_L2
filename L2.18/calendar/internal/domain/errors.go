package domain

import "errors"

var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateInvalid   = errors.New("date parameter is invalid or missing")
	ErrOwnerMismatch = errors.New("user does not own this event")
)