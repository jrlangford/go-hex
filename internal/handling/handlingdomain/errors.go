package handlingdomain

import (
	"go_hex/internal/support/errors"
)

// DomainValidationError represents handling domain validation failures
type DomainValidationError struct {
	errors.BaseError
}

// NewDomainValidationError creates a new handling domain validation error
func NewDomainValidationError(message string, cause error) DomainValidationError {
	return DomainValidationError{
		BaseError: errors.NewBaseError(message, cause),
	}
}
