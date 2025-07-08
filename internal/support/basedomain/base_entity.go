package basedomain

import (
	"time"
)

// DomainEvent represents a base interface for all domain events.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// EntityID represents a domain-specific identifier that wraps a UUID.
type EntityID interface {
	String() string
}

// BaseEntity provides common functionality for domain entities, such as ID generation and validation, and tracking of creation and modification timestamps.
type BaseEntity[T EntityID] struct {
	Id        T         `json:"id" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required,gtefield=CreatedAt"`

	// Event queue for domain events (not serialized)
	events []DomainEvent `json:"-"`
}

func NewBaseEntity[T EntityID](id T) BaseEntity[T] {
	now := time.Now()
	return BaseEntity[T]{
		Id:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (b *BaseEntity[T]) ID() T {
	return b.Id
}

func (b *BaseEntity[T]) Touch() {
	b.UpdatedAt = time.Now()
}

func (b *BaseEntity[T]) AddEvent(event DomainEvent) {
	if b.events == nil {
		b.events = make([]DomainEvent, 0)
	}
	b.events = append(b.events, event)
}

func (b *BaseEntity[T]) GetEvents() []DomainEvent {
	return b.events
}

func (b *BaseEntity[T]) ClearEvents() {
	b.events = nil
}
