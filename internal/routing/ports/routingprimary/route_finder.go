package routingprimary

import (
	"context"
	"go_hex/internal/routing/routingdomain"
)

// RouteFinder defines the primary port for route calculation services
type RouteFinder interface {
	// FindOptimalItineraries finds the best routes that satisfy the given specification
	FindOptimalItineraries(ctx context.Context, routeSpec routingdomain.RouteSpecification) ([]routingdomain.Itinerary, error)
	ListAllVoyages(ctx context.Context) ([]routingdomain.Voyage, error)
	ListAllLocations(ctx context.Context) ([]routingdomain.Location, error)
}
