package shared

import "go_hex/internal/support/errors"

// DomainValidationError represents domain validation failures
type DomainValidationError struct {
	errors.BaseError
}

func NewDomainValidationError(message string, cause error) DomainValidationError {
	return DomainValidationError{
		BaseError: errors.NewBaseError(message, cause),
	}
}
