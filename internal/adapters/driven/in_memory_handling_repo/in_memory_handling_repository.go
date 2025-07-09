package in_memory_handling_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/handling/domain"
	"go_hex/internal/handling/ports/secondary"
)

// InMemoryHandlingEventRepository provides an in-memory implementation of the HandlingEventRepository
type InMemoryHandlingEventRepository struct {
	events map[string]domain.HandlingEvent
	mutex  sync.RWMutex
}

// NewInMemoryHandlingEventRepository creates a new in-memory handling event repository
func NewInMemoryHandlingEventRepository() secondary.HandlingEventRepository {
	return &InMemoryHandlingEventRepository{
		events: make(map[string]domain.HandlingEvent),
	}
}

// Store saves a handling event to the repository
func (r *InMemoryHandlingEventRepository) Store(event domain.HandlingEvent) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	eventId := event.GetEventId().String()
	r.events[eventId] = event
	return nil
}

// FindById retrieves a handling event by its ID
func (r *InMemoryHandlingEventRepository) FindById(eventId domain.HandlingEventId) (domain.HandlingEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	event, exists := r.events[eventId.String()]
	if !exists {
		return domain.HandlingEvent{}, fmt.Errorf("handling event with ID %s not found", eventId.String())
	}
	return event, nil
}

// FindByTrackingId retrieves all handling events for a specific cargo
func (r *InMemoryHandlingEventRepository) FindByTrackingId(trackingId string) ([]domain.HandlingEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var events []domain.HandlingEvent
	for _, event := range r.events {
		if event.GetTrackingId() == trackingId {
			events = append(events, event)
		}
	}
	return events, nil
}

// FindAll retrieves all handling events in the repository
func (r *InMemoryHandlingEventRepository) FindAll() ([]domain.HandlingEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	events := make([]domain.HandlingEvent, 0, len(r.events))
	for _, event := range r.events {
		events = append(events, event)
	}
	return events, nil
}
