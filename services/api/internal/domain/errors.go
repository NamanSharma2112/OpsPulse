package domain

import "errors"

// Sentinel errors the transport layer maps onto HTTP status codes.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
)
