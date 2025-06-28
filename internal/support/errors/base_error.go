package errors

// BaseError represents a common error structure that can be reused across layers
type BaseError struct {
	Message string
	Cause   error
}

func (e BaseError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e BaseError) Unwrap() error {
	return e.Cause
}

// NewBaseError creates a new base error with optional cause
func NewBaseError(message string, cause error) BaseError {
	return BaseError{
		Message: message,
		Cause:   cause,
	}
}
