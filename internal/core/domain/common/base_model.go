// Common domain infrastructure provides shared functionality across domain models.
package common

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

// BaseModel provides common functionality for domain models, such as ID generation and validation, and tracking of creation and modification timestamps.
type BaseModel[T EntityID] struct {
	Id        T         `json:"id" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required,gtfield=CreatedAt"`

	// Event queue for domain events (not serialized)
	events []DomainEvent `json:"-"`
}

func NewBaseModel[T EntityID](id T) BaseModel[T] {
	now := time.Now()
	return BaseModel[T]{
		Id:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (b *BaseModel[T]) ID() T {
	return b.Id
}

func (b *BaseModel[T]) Touch() {
	b.UpdatedAt = time.Now()
}

// AddEvent adds a domain event to the internal event queue
func (b *BaseModel[T]) AddEvent(event DomainEvent) {
	if b.events == nil {
		b.events = make([]DomainEvent, 0)
	}
	b.events = append(b.events, event)
}

// GetEvents returns all queued domain events
func (b *BaseModel[T]) GetEvents() []DomainEvent {
	return b.events
}

// ClearEvents removes all queued domain events
func (b *BaseModel[T]) ClearEvents() {
	b.events = nil
}
