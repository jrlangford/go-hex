package domain

import (
	"go_hex/internal/support/basedomain"
	"go_hex/internal/support/validation"
	"time"

	"github.com/google/uuid"
)

// HandlingEventType represents the type of handling that occurred
type HandlingEventType string

const (
	HandlingEventTypeReceive HandlingEventType = "RECEIVE"
	HandlingEventTypeLoad    HandlingEventType = "LOAD"
	HandlingEventTypeUnload  HandlingEventType = "UNLOAD"
	HandlingEventTypeClaim   HandlingEventType = "CLAIM"
	HandlingEventTypeCustoms HandlingEventType = "CUSTOMS"
)

// HandlingEventId represents the unique identifier for a handling event
type HandlingEventId struct {
	uuid.UUID
}

// NewHandlingEventId creates a new HandlingEventId with a generated UUID
func NewHandlingEventId() HandlingEventId {
	return HandlingEventId{
		UUID: uuid.New(),
	}
}

// HandlingEventIdFromString creates a HandlingEventId from a string representation
func HandlingEventIdFromString(id string) (HandlingEventId, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return HandlingEventId{}, NewDomainValidationError("invalid handling event id format", err)
	}
	return HandlingEventId{UUID: parsedUUID}, nil
}

// String returns the string representation of the HandlingEventId
func (h HandlingEventId) String() string {
	return h.UUID.String()
}

// HandlingEvent represents an immutable record of a cargo handling event
// This is an aggregate root with its own transactional boundary
type HandlingEvent struct {
	basedomain.BaseEntity[HandlingEventId] `json:",inline"`

	// The handling event data
	Data HandlingEventData `json:"data"`
}

// HandlingEventData represents the value object containing event's business data
type HandlingEventData struct {
	EventId          HandlingEventId   `json:"event_id"`
	TrackingId       string            `json:"tracking_id" validate:"required"` // References cargo from booking context
	EventType        HandlingEventType `json:"event_type" validate:"required"`
	Location         string            `json:"location" validate:"required"`          // UN/LOCODE
	VoyageNumber     string            `json:"voyage_number,omitempty"`               // Optional for some event types
	CompletionTime   time.Time         `json:"completion_time" validate:"required"`   // When the physical event occurred
	RegistrationTime time.Time         `json:"registration_time" validate:"required"` // When the event was recorded in the system
}

// NewHandlingEvent creates a new HandlingEvent with validation
func NewHandlingEvent(
	trackingId string,
	eventType HandlingEventType,
	location string,
	voyageNumber string,
	completionTime time.Time,
) (HandlingEvent, error) {
	eventId := NewHandlingEventId()

	data := HandlingEventData{
		EventId:          eventId,
		TrackingId:       trackingId,
		EventType:        eventType,
		Location:         location,
		VoyageNumber:     voyageNumber,
		CompletionTime:   completionTime,
		RegistrationTime: time.Now(),
	}

	// Validate business rules
	if completionTime.After(time.Now()) {
		return HandlingEvent{}, NewDomainValidationError("completion time cannot be in the future", nil)
	}

	// Validate voyage number is required for LOAD and UNLOAD events
	if (eventType == HandlingEventTypeLoad || eventType == HandlingEventTypeUnload) && voyageNumber == "" {
		return HandlingEvent{}, NewDomainValidationError("voyage number is required for LOAD and UNLOAD events", nil)
	}

	if err := validation.Validate(data); err != nil {
		return HandlingEvent{}, NewDomainValidationError("handling event data validation failed", err)
	}

	event := HandlingEvent{
		BaseEntity: basedomain.NewBaseEntity(eventId),
		Data:       data,
	}

	// Raise domain event for handling event creation
	event.AddEvent(NewHandlingEventRegisteredEvent(eventId, trackingId, eventType, location, voyageNumber, completionTime))

	return event, nil
}

// GetEventId returns the event's identifier
func (h HandlingEvent) GetEventId() HandlingEventId {
	return h.Data.EventId
}

// GetTrackingId returns the cargo tracking ID
func (h HandlingEvent) GetTrackingId() string {
	return h.Data.TrackingId
}

// GetEventType returns the type of handling event
func (h HandlingEvent) GetEventType() HandlingEventType {
	return h.Data.EventType
}

// GetLocation returns the location where the event occurred
func (h HandlingEvent) GetLocation() string {
	return h.Data.Location
}

// GetVoyageNumber returns the voyage number (if applicable)
func (h HandlingEvent) GetVoyageNumber() string {
	return h.Data.VoyageNumber
}

// GetCompletionTime returns when the physical event occurred
func (h HandlingEvent) GetCompletionTime() time.Time {
	return h.Data.CompletionTime
}

// GetRegistrationTime returns when the event was recorded in the system
func (h HandlingEvent) GetRegistrationTime() time.Time {
	return h.Data.RegistrationTime
}

// HandlingHistory represents an unmodifiable, ordered collection of handling events
type HandlingHistory struct {
	TrackingId string          `json:"tracking_id" validate:"required"`
	Events     []HandlingEvent `json:"events" validate:"dive"`
}

// NewHandlingHistory creates a new HandlingHistory
func NewHandlingHistory(trackingId string, events []HandlingEvent) (HandlingHistory, error) {
	history := HandlingHistory{
		TrackingId: trackingId,
		Events:     events,
	}

	if err := validation.Validate(history); err != nil {
		return HandlingHistory{}, NewDomainValidationError("handling history validation failed", err)
	}

	return history, nil
}

// GetMostRecentEvent returns the most recent handling event, or nil if no events
func (h HandlingHistory) GetMostRecentEvent() *HandlingEvent {
	if len(h.Events) == 0 {
		return nil
	}

	// Find the most recent event by completion time
	mostRecent := &h.Events[0]
	for i := 1; i < len(h.Events); i++ {
		if h.Events[i].GetCompletionTime().After(mostRecent.GetCompletionTime()) {
			mostRecent = &h.Events[i]
		}
	}

	return mostRecent
}

// GetEventCount returns the number of events in the history
func (h HandlingHistory) GetEventCount() int {
	return len(h.Events)
}
