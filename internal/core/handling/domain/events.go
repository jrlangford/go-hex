package domain

import (
	"time"
)

// HandlingEventRegisteredEvent represents the domain event when a handling event is registered
type HandlingEventRegisteredEvent struct {
	EventId        HandlingEventId   `json:"event_id"`
	TrackingId     string            `json:"tracking_id"`
	EventType      HandlingEventType `json:"event_type"`
	Location       string            `json:"location"`
	VoyageNumber   string            `json:"voyage_number,omitempty"`
	CompletionTime time.Time         `json:"completion_time"`
	OccurredOn     time.Time         `json:"occurred_on"`
}

// NewHandlingEventRegisteredEvent creates a new HandlingEventRegisteredEvent
func NewHandlingEventRegisteredEvent(
	eventId HandlingEventId,
	trackingId string,
	eventType HandlingEventType,
	location string,
	voyageNumber string,
	completionTime time.Time,
) HandlingEventRegisteredEvent {
	return HandlingEventRegisteredEvent{
		EventId:        eventId,
		TrackingId:     trackingId,
		EventType:      eventType,
		Location:       location,
		VoyageNumber:   voyageNumber,
		CompletionTime: completionTime,
		OccurredOn:     time.Now(),
	}
}

// EventName returns the name of this event
func (e HandlingEventRegisteredEvent) EventName() string {
	return "HandlingEventRegistered"
}

// OccurredAt returns when this event occurred
func (e HandlingEventRegisteredEvent) OccurredAt() time.Time {
	return e.OccurredOn
}
