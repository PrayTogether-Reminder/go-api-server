package domain

import "errors"

// Additional domain errors not in auth.go
var (
	// Authentication errors
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// General errors
	ErrUnauthorized = errors.New("unauthorized")
)
