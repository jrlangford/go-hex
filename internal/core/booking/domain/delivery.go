package domain

import (
	"go_hex/internal/support/validation"
	"time"
)

// TransportStatus represents the current state of cargo's physical journey
type TransportStatus string

const (
	TransportStatusNotReceived    TransportStatus = "NOT_RECEIVED"
	TransportStatusInPort         TransportStatus = "IN_PORT"
	TransportStatusOnboardCarrier TransportStatus = "ONBOARD_CARRIER"
	TransportStatusClaimed        TransportStatus = "CLAIMED"
	TransportStatusUnknown        TransportStatus = "UNKNOWN"
)

// RoutingStatus indicates whether cargo is routed and on track
type RoutingStatus string

const (
	RoutingStatusNotRouted   RoutingStatus = "NOT_ROUTED"
	RoutingStatusRouted      RoutingStatus = "ROUTED"
	RoutingStatusMisdirected RoutingStatus = "MISDIRECTED"
)

// Delivery represents a snapshot of the cargo's current transportation status
// This is derived from the handling history and represents the "as-is" state
type Delivery struct {
	TransportStatus   TransportStatus `json:"transport_status" validate:"required"`
	RoutingStatus     RoutingStatus   `json:"routing_status" validate:"required"`
	LastKnownLocation string          `json:"last_known_location,omitempty"` // UN/LOCODE
	CurrentVoyage     string          `json:"current_voyage,omitempty"`
	IsUnloadedAtDest  bool            `json:"is_unloaded_at_dest"`
	CalculatedAt      time.Time       `json:"calculated_at" validate:"required"`
}

// NewDelivery creates a new Delivery status with validation
func NewDelivery(transportStatus TransportStatus, routingStatus RoutingStatus, lastKnownLocation, currentVoyage string, isUnloadedAtDest bool) (Delivery, error) {
	delivery := Delivery{
		TransportStatus:   transportStatus,
		RoutingStatus:     routingStatus,
		LastKnownLocation: lastKnownLocation,
		CurrentVoyage:     currentVoyage,
		IsUnloadedAtDest:  isUnloadedAtDest,
		CalculatedAt:      time.Now(),
	}

	if err := validation.Validate(delivery); err != nil {
		return Delivery{}, NewDomainValidationError("delivery validation failed", err)
	}

	return delivery, nil
}

// NewInitialDelivery creates the initial delivery status for new cargo
func NewInitialDelivery() Delivery {
	return Delivery{
		TransportStatus:   TransportStatusNotReceived,
		RoutingStatus:     RoutingStatusNotRouted,
		LastKnownLocation: "",
		CurrentVoyage:     "",
		IsUnloadedAtDest:  false,
		CalculatedAt:      time.Now(),
	}
}

// IsDelivered checks if the cargo has been successfully delivered
func (d Delivery) IsDelivered() bool {
	return d.TransportStatus == TransportStatusClaimed && d.IsUnloadedAtDest
}

// IsOnTrack checks if the cargo is following its planned route
func (d Delivery) IsOnTrack() bool {
	return d.RoutingStatus == RoutingStatusRouted
}
