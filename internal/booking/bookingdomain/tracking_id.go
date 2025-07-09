package bookingdomain

import (
	"go_hex/internal/support/validation"

	"github.com/google/uuid"
)

// TrackingId is the unique identifier for a piece of cargo
type TrackingId struct {
	uuid.UUID `json:"id" validate:"required,uuid4"`
}

// NewTrackingId creates a new TrackingId with a generated UUID
func NewTrackingId() TrackingId {
	return TrackingId{uuid.New()}
}

// TrackingIdFromString creates a TrackingId from a string representation
func TrackingIdFromString(id string) (TrackingId, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return TrackingId{}, NewDomainValidationError("invalid tracking ID format", err)
	}

	trackingId := TrackingId{parsed}
	if err := validation.Validate(trackingId); err != nil {
		return TrackingId{}, NewDomainValidationError("tracking ID validation failed", err)
	}

	return trackingId, nil
}

// String returns the string representation of the TrackingId
func (t TrackingId) String() string {
	return t.UUID.String()
}
