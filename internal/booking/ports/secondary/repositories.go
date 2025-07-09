package secondary

import (
	"context"
	"go_hex/internal/booking/domain"
	"go_hex/internal/support/basedomain"
)

// CargoRepository defines the secondary port for cargo persistence
type CargoRepository interface {
	// Store persists a cargo aggregate
	Store(cargo domain.Cargo) error

	// FindByTrackingId retrieves a cargo by its tracking ID
	FindByTrackingId(trackingId domain.TrackingId) (domain.Cargo, error)

	// FindUnrouted retrieves all cargo that don't have an itinerary assigned
	FindUnrouted() ([]domain.Cargo, error)

	// FindAll retrieves all cargo (mainly for administrative purposes)
	FindAll() ([]domain.Cargo, error)

	// Update updates an existing cargo
	Update(cargo domain.Cargo) error
}

// RoutingService defines the secondary port for route calculation
type RoutingService interface {
	// FindOptimalItineraries requests route candidates from the routing context
	FindOptimalItineraries(ctx context.Context, routeSpec domain.RouteSpecification) ([]domain.Itinerary, error)
}

// EventPublisher defines the secondary port for publishing domain events
type EventPublisher interface {
	// Publish publishes a domain event
	Publish(event basedomain.DomainEvent) error
}
