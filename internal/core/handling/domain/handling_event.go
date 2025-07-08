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

	// Validate event type
	validEventTypes := []HandlingEventType{
		HandlingEventTypeReceive,
		HandlingEventTypeLoad,
		HandlingEventTypeUnload,
		HandlingEventTypeClaim,
		HandlingEventTypeCustoms,
	}
	isValidEventType := false
	for _, validType := range validEventTypes {
		if eventType == validType {
			isValidEventType = true
			break
		}
	}
	if !isValidEventType {
		return HandlingEvent{}, NewDomainValidationError("invalid event type", nil)
	}

	// Validate business rules
	if completionTime.After(time.Now()) {
		return HandlingEvent{}, NewDomainValidationError("completion time cannot be in the future", nil)
	}

	// Validate completion time is not too far in the past (e.g., more than 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if completionTime.Before(thirtyDaysAgo) {
		return HandlingEvent{}, NewDomainValidationError("completion time cannot be more than 30 days in the past", nil)
	}

	// Validate voyage number is required for LOAD and UNLOAD events
	if (eventType == HandlingEventTypeLoad || eventType == HandlingEventTypeUnload) && voyageNumber == "" {
		return HandlingEvent{}, NewDomainValidationError("voyage number is required for LOAD and UNLOAD events", nil)
	}

	// Validate voyage number should not be provided for RECEIVE and CLAIM events
	if (eventType == HandlingEventTypeReceive || eventType == HandlingEventTypeClaim) && voyageNumber != "" {
		return HandlingEvent{}, NewDomainValidationError("voyage number should not be provided for RECEIVE and CLAIM events", nil)
	}

	// Validate location format (should be UN/LOCODE - 5 characters)
	if len(location) != 5 {
		return HandlingEvent{}, NewDomainValidationError("location must be a valid 5-character UN/LOCODE", nil)
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
	return &h.Events[len(h.Events)-1]
}

// GetEventCount returns the number of events in the history
func (h HandlingHistory) GetEventCount() int {
	return len(h.Events)
}

// HasEventType checks if the history contains an event of the specified type
func (h HandlingHistory) HasEventType(eventType HandlingEventType) bool {
	for _, event := range h.Events {
		if event.GetEventType() == eventType {
			return true
		}
	}
	return false
}

// GetLastEventAtLocation returns the most recent event at the specified location
func (h HandlingHistory) GetLastEventAtLocation(location string) *HandlingEvent {
	for i := len(h.Events) - 1; i >= 0; i-- {
		if h.Events[i].GetLocation() == location {
			return &h.Events[i]
		}
	}
	return nil
}

// IsValidSequence validates that the events form a logical sequence
func (h HandlingHistory) IsValidSequence() error {
	if len(h.Events) == 0 {
		return nil
	}

	// First event must be RECEIVE
	if h.Events[0].GetEventType() != HandlingEventTypeReceive {
		return NewDomainValidationError("first handling event must be RECEIVE", nil)
	}

	// Check chronological order
	for i := 1; i < len(h.Events); i++ {
		prevEvent := h.Events[i-1]
		currentEvent := h.Events[i]

		if currentEvent.GetCompletionTime().Before(prevEvent.GetCompletionTime()) {
			return NewDomainValidationError("events must be in chronological order", nil)
		}

		// Business rule: Can't CLAIM before final UNLOAD
		if currentEvent.GetEventType() == HandlingEventTypeClaim {
			lastUnload := h.getLastUnloadEvent()
			if lastUnload == nil {
				return NewDomainValidationError("cannot claim cargo before unloading", nil)
			}
		}
	}

	return nil
}

// getLastUnloadEvent returns the most recent UNLOAD event
func (h HandlingHistory) getLastUnloadEvent() *HandlingEvent {
	for i := len(h.Events) - 1; i >= 0; i-- {
		if h.Events[i].GetEventType() == HandlingEventTypeUnload {
			return &h.Events[i]
		}
	}
	return nil
}

// IsCompleted checks if the cargo handling is completed (claimed)
func (h HandlingHistory) IsCompleted() bool {
	return h.HasEventType(HandlingEventTypeClaim)
}

// IsReceived checks if the cargo has been received
func (h HandlingHistory) IsReceived() bool {
	return h.HasEventType(HandlingEventTypeReceive)
}

// GetCurrentLocation returns the last known location from handling events
func (h HandlingHistory) GetCurrentLocation() string {
	lastEvent := h.GetMostRecentEvent()
	if lastEvent == nil {
		return ""
	}
	return lastEvent.GetLocation()
}

// GetCurrentVoyage returns the current voyage from the most recent handling event
func (h HandlingHistory) GetCurrentVoyage() string {
	lastEvent := h.GetMostRecentEvent()
	if lastEvent == nil {
		return ""
	}
	return lastEvent.GetVoyageNumber()
}

// GetEventsOfType returns all events of the specified type
func (h HandlingHistory) GetEventsOfType(eventType HandlingEventType) []HandlingEvent {
	var events []HandlingEvent
	for _, event := range h.Events {
		if event.GetEventType() == eventType {
			events = append(events, event)
		}
	}
	return events
}

// GetEventsAtLocation returns all events that occurred at the specified location
func (h HandlingHistory) GetEventsAtLocation(location string) []HandlingEvent {
	var events []HandlingEvent
	for _, event := range h.Events {
		if event.GetLocation() == location {
			events = append(events, event)
		}
	}
	return events
}
