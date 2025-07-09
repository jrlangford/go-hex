package integration

import (
	"context"
	"time"

	bookingDomain "go_hex/internal/booking/domain"
	bookingSecondary "go_hex/internal/booking/ports/secondary"
	routingDomain "go_hex/internal/routing/domain"
	routingPorts "go_hex/internal/routing/ports/primary"
)

// RoutingServiceAdapter adapts the Routing context's application service
// to the interface expected by the Booking context (Anti-Corruption Layer)
type RoutingServiceAdapter struct {
	routingService routingPorts.RouteFinder
}

// NewRoutingServiceAdapter creates a new adapter for the routing service
func NewRoutingServiceAdapter(routingService routingPorts.RouteFinder) bookingSecondary.RoutingService {
	return &RoutingServiceAdapter{
		routingService: routingService,
	}
}

// FindOptimalItineraries adapts the routing service's interface to the booking context's needs
func (a *RoutingServiceAdapter) FindOptimalItineraries(ctx context.Context, routeSpec bookingDomain.RouteSpecification) ([]bookingDomain.Itinerary, error) {
	// Convert Booking domain RouteSpecification to Routing domain format (Anti-Corruption Layer)
	routingRouteSpec := routingDomain.RouteSpecification{
		Origin:          routeSpec.Origin,
		Destination:     routeSpec.Destination,
		ArrivalDeadline: routeSpec.ArrivalDeadline.Format(time.RFC3339), // Convert to string for routing service
	}

	// Call the routing service
	routingItineraries, err := a.routingService.FindOptimalItineraries(ctx, routingRouteSpec)
	if err != nil {
		return nil, err
	}

	// Convert Routing domain Itineraries to Booking domain format (Anti-Corruption Layer)
	bookingItineraries := make([]bookingDomain.Itinerary, len(routingItineraries))
	for i, routingItinerary := range routingItineraries {
		bookingLegs := make([]bookingDomain.Leg, len(routingItinerary.Legs))
		for j, routingLeg := range routingItinerary.Legs {
			// Parse time strings from routing context
			loadTime, err := time.Parse(time.RFC3339, routingLeg.LoadTime)
			if err != nil {
				return nil, err
			}
			unloadTime, err := time.Parse(time.RFC3339, routingLeg.UnloadTime)
			if err != nil {
				return nil, err
			}

			bookingLeg, err := bookingDomain.NewLeg(
				routingLeg.VoyageNumber,
				routingLeg.LoadLocation,
				routingLeg.UnloadLocation,
				loadTime,
				unloadTime,
			)
			if err != nil {
				return nil, err
			}

			bookingLegs[j] = bookingLeg
		}

		bookingItinerary, err := bookingDomain.NewItinerary(bookingLegs)
		if err != nil {
			return nil, err
		}

		bookingItineraries[i] = bookingItinerary
	}

	return bookingItineraries, nil
}
