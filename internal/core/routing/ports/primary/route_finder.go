package primary

import (
	"context"
)

// RouteSpecification represents routing requirements from external contexts
// This is separate from the booking domain's RouteSpecification to maintain bounded context independence
type RouteSpecification struct {
	Origin          string `json:"origin"`           // UN/LOCODE
	Destination     string `json:"destination"`      // UN/LOCODE
	ArrivalDeadline string `json:"arrival_deadline"` // RFC3339 format
}

// Leg represents a single step in a route for external contexts
type Leg struct {
	VoyageNumber   string `json:"voyage_number"`
	LoadLocation   string `json:"load_location"`   // UN/LOCODE
	UnloadLocation string `json:"unload_location"` // UN/LOCODE
	LoadTime       string `json:"load_time"`       // RFC3339 format
	UnloadTime     string `json:"unload_time"`     // RFC3339 format
}

// Itinerary represents a complete route for external contexts
type Itinerary struct {
	Legs []Leg `json:"legs"`
}

// RouteFinder defines the primary port for route calculation services
type RouteFinder interface {
	// FindOptimalItineraries finds the best routes that satisfy the given specification
	FindOptimalItineraries(ctx context.Context, routeSpec RouteSpecification) ([]Itinerary, error)
}
