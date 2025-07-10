package httpadapter

import (
	"time"

	"go_hex/internal/booking/bookingdomain"
	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/routing/routingdomain"
)

// BookCargoRequest represents the request payload for booking cargo
type BookCargoRequest struct {
	Origin          string `json:"origin" validate:"required,min=2,max=10"`
	Destination     string `json:"destination" validate:"required,min=2,max=10"`
	ArrivalDeadline string `json:"arrivalDeadline" validate:"required"`
}

// BookCargoResponse represents the response payload for successful cargo booking
type BookCargoResponse struct {
	TrackingId      string `json:"trackingId"`
	Origin          string `json:"origin"`
	Destination     string `json:"destination"`
	ArrivalDeadline string `json:"arrivalDeadline"`
	RoutingStatus   string `json:"routingStatus"`
	DeliveryStatus  string `json:"deliveryStatus"`
}

// CargoDetailsResponse represents detailed cargo information for tracking
type CargoDetailsResponse struct {
	TrackingId          string        `json:"trackingId"`
	Origin              string        `json:"origin"`
	Destination         string        `json:"destination"`
	ArrivalDeadline     string        `json:"arrivalDeadline"`
	RoutingStatus       string        `json:"routingStatus"`
	TransportStatus     string        `json:"transportStatus"`
	IsOnTrack           bool          `json:"isOnTrack"`
	IsMisdirected       bool          `json:"isMisdirected"`
	IsUnloadedAtDest    bool          `json:"isUnloadedAtDest"`
	LastKnownLocation   *string       `json:"lastKnownLocation,omitempty"`
	CurrentVoyageNumber *string       `json:"currentVoyageNumber,omitempty"`
	Itinerary           *ItineraryDTO `json:"itinerary,omitempty"`
}

// ItineraryDTO represents an itinerary for API responses
type ItineraryDTO struct {
	Legs []LegDTO `json:"legs"`
}

// LegDTO represents a single leg of an itinerary
type LegDTO struct {
	VoyageNumber   string `json:"voyageNumber"`
	LoadLocation   string `json:"loadLocation"`
	UnloadLocation string `json:"unloadLocation"`
	LoadTime       string `json:"loadTime"`
	UnloadTime     string `json:"unloadTime"`
}

// RouteCandidatesResponse represents available route options
type RouteCandidatesResponse struct {
	TrackingId string         `json:"trackingId"`
	Routes     []ItineraryDTO `json:"routes"`
}

// AssignRouteRequest represents the request to assign a route to cargo
type AssignRouteRequest struct {
	Legs []LegDTO `json:"legs" validate:"required,min=1,dive"`
}

// HandlingEventRequest represents a request to register a handling event
type HandlingEventRequest struct {
	TrackingId     string `json:"trackingId" validate:"required"`
	EventType      string `json:"eventType" validate:"required,oneof=RECEIVE LOAD UNLOAD CLAIM CUSTOMS"`
	Location       string `json:"location" validate:"required,min=2,max=10"`
	VoyageNumber   string `json:"voyageNumber,omitempty"`
	CompletionTime string `json:"completionTime" validate:"required"`
}

// HandlingEventResponse represents the response after registering a handling event
type HandlingEventResponse struct {
	EventId        string `json:"eventId"`
	TrackingId     string `json:"trackingId"`
	EventType      string `json:"eventType"`
	Location       string `json:"location"`
	VoyageNumber   string `json:"voyageNumber,omitempty"`
	CompletionTime string `json:"completionTime"`
	RegisteredAt   string `json:"registeredAt"`
}

// HandlingEventDTO represents a handling event for API responses
type HandlingEventDTO struct {
	EventId        string `json:"eventId"`
	EventType      string `json:"eventType"`
	Location       string `json:"location"`
	VoyageNumber   string `json:"voyageNumber,omitempty"`
	CompletionTime string `json:"completionTime"`
}

// VoyageRequest represents a request to create a new voyage
type VoyageRequest struct {
	VoyageNumber string   `json:"voyageNumber" validate:"required"`
	Schedule     []LegDTO `json:"schedule" validate:"required,min=1"`
}

// VoyageResponse represents a voyage in API responses
type VoyageResponse struct {
	VoyageNumber string   `json:"voyageNumber"`
	Schedule     []LegDTO `json:"schedule"`
}

// LocationResponse represents a shipping location
type LocationResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// RouteRequest represents a request to find routes
type RouteRequest struct {
	Origin            string  `json:"origin" validate:"required"`
	Destination       string  `json:"destination" validate:"required"`
	EarliestDeparture *string `json:"earliestDeparture,omitempty"`
}

// Helper functions to convert domain objects to DTOs

func CargoToResponse(cargo bookingdomain.Cargo) CargoDetailsResponse {
	delivery := cargo.GetDelivery()

	response := CargoDetailsResponse{
		TrackingId:       cargo.GetTrackingId().String(),
		Origin:           cargo.GetRouteSpecification().Origin,
		Destination:      cargo.GetRouteSpecification().Destination,
		ArrivalDeadline:  cargo.GetRouteSpecification().ArrivalDeadline.Format(time.RFC3339),
		RoutingStatus:    string(delivery.RoutingStatus),
		TransportStatus:  string(delivery.TransportStatus),
		IsOnTrack:        delivery.IsOnTrack(),
		IsMisdirected:    delivery.RoutingStatus == bookingdomain.RoutingStatusMisdirected,
		IsUnloadedAtDest: delivery.IsUnloadedAtDest,
	}

	if delivery.LastKnownLocation != "" {
		response.LastKnownLocation = &delivery.LastKnownLocation
	}

	if delivery.CurrentVoyage != "" {
		response.CurrentVoyageNumber = &delivery.CurrentVoyage
	}

	if cargo.GetItinerary() != nil {
		response.Itinerary = ItineraryToDTO(*cargo.GetItinerary())
	}

	return response
}

func ItineraryToDTO(itinerary bookingdomain.Itinerary) *ItineraryDTO {
	dto := &ItineraryDTO{
		Legs: make([]LegDTO, len(itinerary.Legs)),
	}

	for i, leg := range itinerary.Legs {
		dto.Legs[i] = LegDTO{
			VoyageNumber:   leg.VoyageNumber,
			LoadLocation:   leg.LoadLocation,
			UnloadLocation: leg.UnloadLocation,
			LoadTime:       leg.LoadTime.Format(time.RFC3339),
			UnloadTime:     leg.UnloadTime.Format(time.RFC3339),
		}
	}

	return dto
}

func BookCargoToResponse(cargo bookingdomain.Cargo) BookCargoResponse {
	return BookCargoResponse{
		TrackingId:      cargo.GetTrackingId().String(),
		Origin:          cargo.GetRouteSpecification().Origin,
		Destination:     cargo.GetRouteSpecification().Destination,
		ArrivalDeadline: cargo.GetRouteSpecification().ArrivalDeadline.Format(time.RFC3339),
		RoutingStatus:   string(cargo.GetDelivery().RoutingStatus),
		DeliveryStatus:  string(cargo.GetDelivery().TransportStatus),
	}
}

// Additional DTO conversion helper functions

func HandlingEventToDTO(event handlingdomain.HandlingEvent) HandlingEventDTO {
	return HandlingEventDTO{
		EventId:        event.GetEventId().String(),
		EventType:      string(event.GetEventType()),
		Location:       event.GetLocation(),
		VoyageNumber:   event.GetVoyageNumber(),
		CompletionTime: event.GetCompletionTime().Format(time.RFC3339),
	}
}

func VoyageToResponseFromDomain(voyage routingdomain.Voyage) VoyageResponse {
	schedule := voyage.GetSchedule()
	legs := make([]LegDTO, len(schedule.Movements))

	for i, movement := range schedule.Movements {
		legs[i] = LegDTO{
			VoyageNumber:   voyage.GetVoyageNumber().String(),
			LoadLocation:   movement.DepartureLocation.String(),
			UnloadLocation: movement.ArrivalLocation.String(),
			LoadTime:       movement.DepartureTime.Format(time.RFC3339),
			UnloadTime:     movement.ArrivalTime.Format(time.RFC3339),
		}
	}

	return VoyageResponse{
		VoyageNumber: voyage.GetVoyageNumber().String(),
		Schedule:     legs,
	}
}

func LocationToResponseFromDomain(location routingdomain.Location) LocationResponse {
	return LocationResponse{
		Code: location.GetUnLocode().String(),
		Name: location.GetName(),
	}
}
