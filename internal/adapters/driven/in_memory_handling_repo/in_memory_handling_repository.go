package in_memory_handling_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/handling/ports/handlingsecondary"
)

// InMemoryHandlingEventRepository provides an in-memory implementation of the HandlingEventRepository
type InMemoryHandlingEventRepository struct {
	events map[string]handlingdomain.HandlingEvent
	mutex  sync.RWMutex
}

// NewInMemoryHandlingEventRepository creates a new in-memory handling event repository
func NewInMemoryHandlingEventRepository() handlingsecondary.HandlingEventRepository {
	return &InMemoryHandlingEventRepository{
		events: make(map[string]handlingdomain.HandlingEvent),
	}
}

// Store saves a handling event to the repository
func (r *InMemoryHandlingEventRepository) Store(event handlingdomain.HandlingEvent) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	eventId := event.GetEventId().String()
	r.events[eventId] = event
	return nil
}

// FindById retrieves a handling event by its ID
func (r *InMemoryHandlingEventRepository) FindById(eventId handlingdomain.HandlingEventId) (handlingdomain.HandlingEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	event, exists := r.events[eventId.String()]
	if !exists {
		return handlingdomain.HandlingEvent{}, fmt.Errorf("handling event with ID %s not found", eventId.String())
	}
	return event, nil
}

// FindByTrackingId retrieves all handling events for a specific cargo
func (r *InMemoryHandlingEventRepository) FindByTrackingId(trackingId string) ([]handlingdomain.HandlingEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var events []handlingdomain.HandlingEvent
	for _, event := range r.events {
		if event.GetTrackingId() == trackingId {
			events = append(events, event)
		}
	}
	return events, nil
}

// FindAll retrieves all handling events in the repository
func (r *InMemoryHandlingEventRepository) FindAll() ([]handlingdomain.HandlingEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	events := make([]handlingdomain.HandlingEvent, 0, len(r.events))
	for _, event := range r.events {
		events = append(events, event)
	}
	return events, nil
}
