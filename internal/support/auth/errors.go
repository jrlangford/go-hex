package auth

import "go_hex/internal/support/errors"

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	errors.BaseError
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) error {
	return AuthenticationError{
		BaseError: errors.NewBaseError(message, nil),
	}
}

// AuthorizationError represents authorization failures
type AuthorizationError struct {
	errors.BaseError
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) error {
	return AuthorizationError{
		BaseError: errors.NewBaseError(message, nil),
	}
}
