package application

import "go_hex/internal/support/errors"

// AppWorkflowError represents application workflow failures
type AppWorkflowError struct {
	errors.BaseError
}

// NewAppWorkflowError creates a new workflow error with optional cause
func NewAppWorkflowError(message string, cause error) AppWorkflowError {
	return AppWorkflowError{
		BaseError: errors.NewBaseError(message, cause),
	}
}

// Standard application errors
var (
	ErrUserNotFound      = NewAppWorkflowError("user not found", nil)
	ErrInvalidUserData   = NewAppWorkflowError("invalid user data", nil)
	ErrUserAlreadyExists = NewAppWorkflowError("user already exists", nil)
)
