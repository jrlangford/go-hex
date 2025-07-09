package primary

import (
	"context"
	"go_hex/internal/routing/domain"
)

// RouteFinder defines the primary port for route calculation services
type RouteFinder interface {
	// FindOptimalItineraries finds the best routes that satisfy the given specification
	FindOptimalItineraries(ctx context.Context, routeSpec domain.RouteSpecification) ([]domain.Itinerary, error)
	ListAllVoyages(ctx context.Context) ([]domain.Voyage, error)
	ListAllLocations(ctx context.Context) ([]domain.Location, error)
}
