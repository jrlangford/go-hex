package domain

import (
	"time"
)

// CargoBookedEvent represents the domain event when cargo is booked
type CargoBookedEvent struct {
	TrackingId         TrackingId         `json:"tracking_id"`
	RouteSpecification RouteSpecification `json:"route_specification"`
	OccurredOn         time.Time          `json:"occurred_on"`
}

// NewCargoBookedEvent creates a new CargoBookedEvent
func NewCargoBookedEvent(trackingId TrackingId, routeSpec RouteSpecification) CargoBookedEvent {
	return CargoBookedEvent{
		TrackingId:         trackingId,
		RouteSpecification: routeSpec,
		OccurredOn:         time.Now(),
	}
}

// EventName returns the name of this event
func (e CargoBookedEvent) EventName() string {
	return "CargoBooked"
}

// OccurredAt returns when this event occurred
func (e CargoBookedEvent) OccurredAt() time.Time {
	return e.OccurredOn
}

// CargoRoutedEvent represents the domain event when cargo is assigned to a route
type CargoRoutedEvent struct {
	TrackingId TrackingId `json:"tracking_id"`
	Itinerary  Itinerary  `json:"itinerary"`
	OccurredOn time.Time  `json:"occurred_on"`
}

// NewCargoRoutedEvent creates a new CargoRoutedEvent
func NewCargoRoutedEvent(trackingId TrackingId, itinerary Itinerary) CargoRoutedEvent {
	return CargoRoutedEvent{
		TrackingId: trackingId,
		Itinerary:  itinerary,
		OccurredOn: time.Now(),
	}
}

// EventName returns the name of this event
func (e CargoRoutedEvent) EventName() string {
	return "CargoRouted"
}

// OccurredAt returns when this event occurred
func (e CargoRoutedEvent) OccurredAt() time.Time {
	return e.OccurredOn
}

// CargoDeliveryUpdatedEvent represents the domain event when cargo delivery status is updated
type CargoDeliveryUpdatedEvent struct {
	TrackingId TrackingId `json:"tracking_id"`
	Delivery   Delivery   `json:"delivery"`
	OccurredOn time.Time  `json:"occurred_on"`
}

// NewCargoDeliveryUpdatedEvent creates a new CargoDeliveryUpdatedEvent
func NewCargoDeliveryUpdatedEvent(trackingId TrackingId, delivery Delivery) CargoDeliveryUpdatedEvent {
	return CargoDeliveryUpdatedEvent{
		TrackingId: trackingId,
		Delivery:   delivery,
		OccurredOn: time.Now(),
	}
}

// EventName returns the name of this event
func (e CargoDeliveryUpdatedEvent) EventName() string {
	return "CargoDeliveryUpdated"
}

// OccurredAt returns when this event occurred
func (e CargoDeliveryUpdatedEvent) OccurredAt() time.Time {
	return e.OccurredOn
}
