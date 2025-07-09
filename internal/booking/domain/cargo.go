package domain

import (
	"go_hex/internal/support/basedomain"
	"go_hex/internal/support/validation"
	"time"
)

// Cargo represents a customer's shipment request and tracks its entire lifecycle
// This is the main aggregate root for the Booking Context
type Cargo struct {
	basedomain.BaseEntity[TrackingId] `json:",inline"`

	// The cargo's data and current state
	Data CargoData `json:"data"`
}

// CargoData represents the value object containing cargo's business data
type CargoData struct {
	RouteSpecification RouteSpecification `json:"route_specification"`
	Itinerary          *Itinerary         `json:"itinerary,omitempty"` // nil if not yet routed
	Delivery           Delivery           `json:"delivery"`
}

// NewCargo creates a new Cargo aggregate with the specified route specification
func NewCargo(origin, destination string, arrivalDeadline time.Time) (Cargo, error) {
	trackingId := NewTrackingId()

	routeSpec, err := NewRouteSpecification(origin, destination, arrivalDeadline)
	if err != nil {
		return Cargo{}, err
	}

	data := CargoData{
		RouteSpecification: routeSpec,
		Itinerary:          nil, // Not yet routed
		Delivery:           NewInitialDelivery(),
	}

	if err := validation.Validate(data); err != nil {
		return Cargo{}, NewDomainValidationError("cargo data validation failed", err)
	}

	cargo := Cargo{
		BaseEntity: basedomain.NewBaseEntity(trackingId),
		Data:       data,
	}

	// Raise domain event for cargo booking
	cargo.AddEvent(NewCargoBookedEvent(trackingId, routeSpec))

	return cargo, nil
}

// NewCargoFromExisting creates a cargo from existing data (for repository loading)
func NewCargoFromExisting(trackingId TrackingId, routeSpec RouteSpecification, itinerary *Itinerary, delivery Delivery) (Cargo, error) {
	data := CargoData{
		RouteSpecification: routeSpec,
		Itinerary:          itinerary,
		Delivery:           delivery,
	}

	if err := validation.Validate(data); err != nil {
		return Cargo{}, NewDomainValidationError("cargo data validation failed", err)
	}

	return Cargo{
		BaseEntity: basedomain.NewBaseEntity(trackingId),
		Data:       data,
	}, nil
}

// GetTrackingId returns the cargo's tracking identifier
func (c Cargo) GetTrackingId() TrackingId {
	return c.Id
}

// GetRouteSpecification returns the customer's original routing requirement
func (c Cargo) GetRouteSpecification() RouteSpecification {
	return c.Data.RouteSpecification
}

// GetItinerary returns the assigned itinerary (nil if not yet routed)
func (c Cargo) GetItinerary() *Itinerary {
	return c.Data.Itinerary
}

// GetDelivery returns the current delivery status
func (c Cargo) GetDelivery() Delivery {
	return c.Data.Delivery
}

// IsRouted checks if the cargo has been assigned an itinerary
func (c Cargo) IsRouted() bool {
	return c.Data.Itinerary != nil
}

// AssignToRoute assigns an itinerary to the cargo
func (c *Cargo) AssignToRoute(itinerary Itinerary) error {
	// Cannot reassign route if already delivered
	if c.Data.Delivery.IsDelivered() {
		return NewDomainValidationError("cannot reassign route to already delivered cargo", nil)
	}

	// Validate that the itinerary satisfies the route specification
	if !itinerary.SatisfiesSpecification(c.Data.RouteSpecification) {
		return NewDomainValidationError("itinerary does not satisfy route specification", nil)
	}

	// Check if itinerary arrival deadline would be missed
	if itinerary.FinalArrivalTime().After(c.Data.RouteSpecification.ArrivalDeadline) {
		return NewDomainValidationError("itinerary arrival time exceeds deadline", nil)
	}

	c.Data.Itinerary = &itinerary

	// Update routing status
	newDelivery, err := NewDelivery(
		c.Data.Delivery.TransportStatus,
		RoutingStatusRouted,
		c.Data.Delivery.LastKnownLocation,
		c.Data.Delivery.CurrentVoyage,
		c.Data.Delivery.IsUnloadedAtDest,
	)
	if err != nil {
		return err
	}

	c.Data.Delivery = newDelivery

	// Raise domain event for route assignment
	c.AddEvent(NewCargoRoutedEvent(c.Id, itinerary))

	return nil
}

// DeriveDeliveryProgress updates delivery status based on handling history
// This is called when handling events are received
func (c *Cargo) DeriveDeliveryProgress(handlingHistory []HandlingEventSummary) error {
	if len(handlingHistory) == 0 {
		return nil // No changes needed
	}

	// Get the most recent handling event
	lastEvent := handlingHistory[len(handlingHistory)-1]

	// Calculate new transport status based on the latest event
	transportStatus := c.calculateTransportStatus(lastEvent)

	// Calculate routing status based on itinerary and current progress
	routingStatus := c.calculateRoutingStatus(lastEvent)

	// Create new delivery status
	newDelivery, err := NewDelivery(
		transportStatus,
		routingStatus,
		lastEvent.Location,
		lastEvent.VoyageNumber,
		c.isUnloadedAtDestination(lastEvent),
	)
	if err != nil {
		return err
	}

	c.Data.Delivery = newDelivery

	// Raise domain event for delivery progress update
	c.AddEvent(NewCargoDeliveryUpdatedEvent(c.Id, newDelivery))

	return nil
}

// calculateTransportStatus determines transport status from handling event
func (c *Cargo) calculateTransportStatus(lastEvent HandlingEventSummary) TransportStatus {
	switch lastEvent.Type {
	case "RECEIVE":
		return TransportStatusInPort
	case "LOAD":
		return TransportStatusOnboardCarrier
	case "UNLOAD":
		return TransportStatusInPort
	case "CLAIM":
		return TransportStatusClaimed
	default:
		return TransportStatusUnknown
	}
}

// calculateRoutingStatus determines if cargo is on track
func (c *Cargo) calculateRoutingStatus(lastEvent HandlingEventSummary) RoutingStatus {
	if c.Data.Itinerary == nil {
		return RoutingStatusNotRouted
	}

	// Check if the event location and voyage match the expected itinerary
	if c.Data.Itinerary.IsOnTrack(lastEvent.Location, lastEvent.VoyageNumber) {
		return RoutingStatusRouted
	}

	return RoutingStatusMisdirected
}

// isUnloadedAtDestination checks if cargo has been unloaded at final destination
func (c *Cargo) isUnloadedAtDestination(lastEvent HandlingEventSummary) bool {
	return lastEvent.Type == "UNLOAD" &&
		lastEvent.Location == c.Data.RouteSpecification.Destination
}

// CanBeRerouted checks if cargo can be assigned a new route
func (c Cargo) CanBeRerouted() bool {
	return !c.Data.Delivery.IsDelivered() &&
		c.Data.Delivery.TransportStatus != TransportStatusClaimed
}

// IsReadyForPickup checks if cargo is ready for pickup at origin
func (c Cargo) IsReadyForPickup() bool {
	return c.IsRouted() &&
		c.Data.Delivery.TransportStatus == TransportStatusNotReceived &&
		time.Now().Before(c.Data.Itinerary.InitialDepartureTime())
}

// IsOverdue checks if cargo delivery is overdue
func (c Cargo) IsOverdue() bool {
	return time.Now().After(c.Data.RouteSpecification.ArrivalDeadline) &&
		!c.Data.Delivery.IsDelivered()
}

// GetEstimatedTimeOfArrival returns the ETA based on current itinerary
func (c Cargo) GetEstimatedTimeOfArrival() *time.Time {
	if c.Data.Itinerary == nil {
		return nil
	}
	eta := c.Data.Itinerary.FinalArrivalTime()
	return &eta
}

// HandlingEventSummary represents key data from a handling event
// This is used for delivery progress calculation without importing handling domain
type HandlingEventSummary struct {
	Type         string    `json:"type"`
	Location     string    `json:"location"`
	VoyageNumber string    `json:"voyage_number"`
	Timestamp    time.Time `json:"timestamp"`
}
