package bookingdomain

import (
	"go_hex/internal/support/validation"
	"time"
)

// RouteSpecification defines the customer's immutable transportation requirement
type RouteSpecification struct {
	Origin          string    `json:"origin" validate:"required,min=3,max=5"`      // UN/LOCODE
	Destination     string    `json:"destination" validate:"required,min=3,max=5"` // UN/LOCODE
	ArrivalDeadline time.Time `json:"arrival_deadline" validate:"required"`        // Must arrive by this date
}

// NewRouteSpecification creates a new RouteSpecification with validation
func NewRouteSpecification(origin, destination string, arrivalDeadline time.Time) (RouteSpecification, error) {
	spec := RouteSpecification{
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: arrivalDeadline,
	}

	// Validate business rules
	if origin == destination {
		return RouteSpecification{}, NewDomainValidationError("origin and destination cannot be the same", nil)
	}

	if arrivalDeadline.Before(time.Now()) {
		return RouteSpecification{}, NewDomainValidationError("arrival deadline must be in the future", nil)
	}

	if err := validation.Validate(spec); err != nil {
		return RouteSpecification{}, NewDomainValidationError("route specification validation failed", err)
	}

	return spec, nil
}

// Leg represents a single step in an itinerary
type Leg struct {
	VoyageNumber   string    `json:"voyage_number" validate:"required"`
	LoadLocation   string    `json:"load_location" validate:"required,min=3,max=5"`   // UN/LOCODE
	UnloadLocation string    `json:"unload_location" validate:"required,min=3,max=5"` // UN/LOCODE
	LoadTime       time.Time `json:"load_time" validate:"required"`
	UnloadTime     time.Time `json:"unload_time" validate:"required"`
}

// NewLeg creates a new Leg with validation
func NewLeg(voyageNumber, loadLocation, unloadLocation string, loadTime, unloadTime time.Time) (Leg, error) {
	leg := Leg{
		VoyageNumber:   voyageNumber,
		LoadLocation:   loadLocation,
		UnloadLocation: unloadLocation,
		LoadTime:       loadTime,
		UnloadTime:     unloadTime,
	}

	// Validate business rules
	if loadLocation == unloadLocation {
		return Leg{}, NewDomainValidationError("load and unload locations cannot be the same", nil)
	}

	if !unloadTime.After(loadTime) {
		return Leg{}, NewDomainValidationError("unload time must be after load time", nil)
	}

	if err := validation.Validate(leg); err != nil {
		return Leg{}, NewDomainValidationError("leg validation failed", err)
	}

	return leg, nil
}

// Itinerary represents a planned shipping route consisting of one or more legs
type Itinerary struct {
	Legs []Leg `json:"legs" validate:"required,min=1,dive"`
}

// NewItinerary creates a new Itinerary with validation
func NewItinerary(legs []Leg) (Itinerary, error) {
	if len(legs) == 0 {
		return Itinerary{}, NewDomainValidationError("itinerary must contain at least one leg", nil)
	}

	itinerary := Itinerary{Legs: legs}

	// Validate leg connectivity - each leg must connect to the next
	for i := 0; i < len(legs)-1; i++ {
		currentLeg := legs[i]
		nextLeg := legs[i+1]

		if currentLeg.UnloadLocation != nextLeg.LoadLocation {
			return Itinerary{}, NewDomainValidationError("legs must be connected - unload location must match next leg's load location", nil)
		}

		if !nextLeg.LoadTime.After(currentLeg.UnloadTime) {
			return Itinerary{}, NewDomainValidationError("insufficient time between legs for transshipment", nil)
		}
	}

	if err := validation.Validate(itinerary); err != nil {
		return Itinerary{}, NewDomainValidationError("itinerary validation failed", err)
	}

	return itinerary, nil
}

// SatisfiesSpecification checks if this itinerary satisfies the given route specification
func (i Itinerary) SatisfiesSpecification(spec RouteSpecification) bool {
	if len(i.Legs) == 0 {
		return false
	}

	firstLeg := i.Legs[0]
	lastLeg := i.Legs[len(i.Legs)-1]

	// Check origin and destination match
	if firstLeg.LoadLocation != spec.Origin || lastLeg.UnloadLocation != spec.Destination {
		return false
	}

	// Check arrival deadline is met
	if lastLeg.UnloadTime.After(spec.ArrivalDeadline) {
		return false
	}

	return true
}

// FinalArrivalTime returns the time when cargo will arrive at its final destination
func (i Itinerary) FinalArrivalTime() time.Time {
	if len(i.Legs) == 0 {
		return time.Time{}
	}
	return i.Legs[len(i.Legs)-1].UnloadTime
}

// InitialDepartureTime returns the time when cargo departs from its origin
func (i Itinerary) InitialDepartureTime() time.Time {
	if len(i.Legs) == 0 {
		return time.Time{}
	}
	return i.Legs[0].LoadTime
}

// IsOnTrack checks if the given location and voyage are part of this itinerary
func (i Itinerary) IsOnTrack(location, voyageNumber string) bool {
	for _, leg := range i.Legs {
		if leg.VoyageNumber == voyageNumber &&
			(leg.LoadLocation == location || leg.UnloadLocation == location) {
			return true
		}
	}
	return false
}
