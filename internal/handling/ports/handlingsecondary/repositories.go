package handlingsecondary

import (
	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/support/basedomain"
)

// HandlingEventRepository defines the secondary port for handling event persistence
type HandlingEventRepository interface {
	// Store persists a handling event
	Store(event handlingdomain.HandlingEvent) error

	// FindById retrieves a handling event by its ID
	FindById(eventId handlingdomain.HandlingEventId) (handlingdomain.HandlingEvent, error)

	// FindByTrackingId retrieves all handling events for a specific cargo
	FindByTrackingId(trackingId string) ([]handlingdomain.HandlingEvent, error)

	// FindAll retrieves all handling events (mainly for administrative purposes)
	FindAll() ([]handlingdomain.HandlingEvent, error)
}

// EventPublisher defines the secondary port for publishing domain events
type EventPublisher interface {
	// Publish publishes a domain event
	Publish(event basedomain.DomainEvent) error
}
