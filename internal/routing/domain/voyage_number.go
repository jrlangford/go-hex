package domain

import (
	"go_hex/internal/support/validation"

	"github.com/google/uuid"
)

// VoyageNumber represents the unique identifier for a voyage
type VoyageNumber struct {
	uuid.UUID
}

// NewVoyageNumber creates a new VoyageNumber with a generated UUID
func NewVoyageNumber() VoyageNumber {
	return VoyageNumber{
		UUID: uuid.New(),
	}
}

// VoyageNumberFromString creates a VoyageNumber from a string representation
func VoyageNumberFromString(id string) (VoyageNumber, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return VoyageNumber{}, NewDomainValidationError("invalid voyage number format", err)
	}
	return VoyageNumber{UUID: parsedUUID}, nil
}

// String returns the string representation of the VoyageNumber
func (v VoyageNumber) String() string {
	return v.UUID.String()
}

// Validate ensures the VoyageNumber is valid
func (v VoyageNumber) Validate() error {
	if v.UUID == uuid.Nil {
		return NewDomainValidationError("voyage number cannot be empty", nil)
	}
	return validation.Validate(v)
}
