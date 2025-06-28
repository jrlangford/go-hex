package secondary

import "go_hex/internal/core/domain/common"

// EventPublisher defines the secondary port for publishing domain events.
type EventPublisher interface {
	Publish(event common.DomainEvent) error
}
