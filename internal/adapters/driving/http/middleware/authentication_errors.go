package middleware

import "go_hex/internal/support/errors"

// AuthError represents authentication/authorization failures
type AuthError struct {
	errors.BaseError
}

func NewAuthError(message string, cause error) AuthError {
	return AuthError{
		BaseError: errors.NewBaseError(message, cause),
	}
}

var (
	ErrUnauthorized     = NewAuthError("unauthorized", nil)
	ErrInvalidToken     = NewAuthError("invalid or expired token", nil)
	ErrMissingToken     = NewAuthError("authentication token required", nil)
	ErrInsufficientRole = NewAuthError("insufficient role permissions", nil)
)
