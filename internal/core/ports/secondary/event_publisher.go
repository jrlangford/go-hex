package secondary

import "go_hex/internal/core/domain/shared"

// EventPublisher defines the secondary port for publishing domain events.
type EventPublisher interface {
	Publish(event shared.DomainEvent) error
}
