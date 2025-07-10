package bookingprimary

import (
	"context"
	"go_hex/internal/booking/bookingdomain"
)

// BookingService defines the primary port for cargo booking operations
type BookingService interface {
	// BookNewCargo initiates the creation of a new cargo based on customer's request
	BookNewCargo(ctx context.Context, origin, destination string, arrivalDeadline string) (bookingdomain.Cargo, error)

	// AssignRouteToCargo assigns a chosen itinerary to an existing cargo
	AssignRouteToCargo(ctx context.Context, trackingId bookingdomain.TrackingId, itinerary bookingdomain.Itinerary) error

	// GetCargoDetails retrieves the full state of a cargo for tracking
	GetCargoDetails(ctx context.Context, trackingId bookingdomain.TrackingId) (bookingdomain.Cargo, error)

	// ListUnroutedCargo gets all cargo that require route assignment
	ListUnroutedCargo(ctx context.Context) ([]bookingdomain.Cargo, error)

	// RequestRouteCandidates gets possible itineraries for a cargo
	RequestRouteCandidates(ctx context.Context, trackingId bookingdomain.TrackingId) ([]bookingdomain.Itinerary, error)

	// UpdateCargoDelivery updates the delivery status of a cargo
	UpdateCargoDelivery(ctx context.Context, trackingId bookingdomain.TrackingId, handlingHistory []bookingdomain.HandlingEventSummary) error
}

// CargoTracker defines the primary port for cargo tracking queries
type CargoTracker interface {
	// TrackCargo returns the current status of cargo by tracking ID
	TrackCargo(ctx context.Context, trackingId bookingdomain.TrackingId) (bookingdomain.Cargo, error)
}
