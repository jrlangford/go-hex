package basedomain

// EventPublisher defines the secondary port for publishing domain events.
type EventPublisher interface {
	Publish(event DomainEvent) error
}
