package routingapplication

import (
	"context"
	"go_hex/internal/routing/ports/routingprimary"
	"go_hex/internal/routing/ports/routingsecondary"
	"go_hex/internal/routing/routingdomain"
	"go_hex/internal/support/auth"
	"log/slog"
	"time"
)

// RoutingApplicationService implements the primary port for routing operations
type RoutingApplicationService struct {
	voyageRepo   routingsecondary.VoyageRepository
	locationRepo routingsecondary.LocationRepository
	logger       *slog.Logger
}

// Ensure RoutingApplicationService implements the primary port
var _ routingprimary.RouteFinder = (*RoutingApplicationService)(nil)

// NewRoutingApplicationService creates a new RoutingApplicationService
func NewRoutingApplicationService(
	voyageRepo routingsecondary.VoyageRepository,
	locationRepo routingsecondary.LocationRepository,
	logger *slog.Logger,
) *RoutingApplicationService {
	return &RoutingApplicationService{
		voyageRepo:   voyageRepo,
		locationRepo: locationRepo,
		logger:       logger,
	}
}

// FindOptimalItineraries finds the best routes that satisfy the given specification
func (s *RoutingApplicationService) FindOptimalItineraries(ctx context.Context, routeSpec routingdomain.RouteSpecification) ([]routingdomain.Itinerary, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized route planning attempt", "error", err)
		return nil, err
	}
	if err := RequireRoutingPermission(claims, auth.PermissionPlanRoutes); err != nil {
		s.logger.Warn("Unauthorized route planning attempt", "error", err)
		return nil, err
	}

	s.logger.Info("Finding optimal itineraries",
		"origin", routeSpec.Origin,
		"destination", routeSpec.Destination,
		"deadline", routeSpec.ArrivalDeadline)

	// Parse arrival deadline
	arrivalDeadline, err := time.Parse(time.RFC3339, routeSpec.ArrivalDeadline)
	if err != nil {
		s.logger.Error("Invalid arrival deadline format", "error", err)
		return nil, routingdomain.NewDomainValidationError("invalid arrival deadline format, expected RFC3339", err)
	}

	// Convert external route spec to internal format
	origin, err := routingdomain.NewUnLocode(routeSpec.Origin)
	if err != nil {
		s.logger.Error("Invalid origin UN/LOCODE", "error", err)
		return nil, err
	}

	destination, err := routingdomain.NewUnLocode(routeSpec.Destination)
	if err != nil {
		s.logger.Error("Invalid destination UN/LOCODE", "error", err)
		return nil, err
	}

	// Get all voyages for analysis
	allVoyages, err := s.voyageRepo.FindAll()
	if err != nil {
		s.logger.Error("Failed to retrieve voyages", "error", err)
		return nil, routingdomain.NewDomainValidationError("failed to retrieve voyages", err)
	}

	// Find route candidates using simplified algorithm
	candidates := s.findRoutes(allVoyages, origin, destination, arrivalDeadline)

	// Convert internal candidates to external format
	itineraries := s.convertToExternalFormat(candidates)

	s.logger.Info("Found route candidates", "count", len(itineraries))
	return itineraries, nil
}

// findRoutes implements a simplified routing algorithm
func (s *RoutingApplicationService) findRoutes(voyages []routingdomain.Voyage, origin, destination routingdomain.UnLocode, deadline time.Time) []routeCandidate {
	var candidates []routeCandidate

	// Find direct routes
	directRoutes := s.findDirectRoutes(voyages, origin, destination, deadline)
	candidates = append(candidates, directRoutes...)

	// Find routes with one connection
	oneConnectionRoutes := s.findOneConnectionRoutes(voyages, origin, destination, deadline)
	candidates = append(candidates, oneConnectionRoutes...)

	return candidates
}

// routeCandidate represents an internal route candidate
type routeCandidate struct {
	legs []routeLeg
}

// routeLeg represents an internal route leg
type routeLeg struct {
	voyageNumber   routingdomain.VoyageNumber
	loadLocation   routingdomain.UnLocode
	unloadLocation routingdomain.UnLocode
	loadTime       time.Time
	unloadTime     time.Time
}

// findDirectRoutes finds routes that can be completed with a single voyage
func (s *RoutingApplicationService) findDirectRoutes(voyages []routingdomain.Voyage, origin, destination routingdomain.UnLocode, deadline time.Time) []routeCandidate {
	var candidates []routeCandidate

	for _, voyage := range voyages {
		schedule := voyage.GetSchedule()

		// Check if this voyage connects origin to destination
		for _, movement := range schedule.Movements {
			if movement.DepartureLocation == origin &&
				movement.ArrivalLocation == destination &&
				(movement.ArrivalTime.Before(deadline) || movement.ArrivalTime.Equal(deadline)) {

				leg := routeLeg{
					voyageNumber:   voyage.GetVoyageNumber(),
					loadLocation:   movement.DepartureLocation,
					unloadLocation: movement.ArrivalLocation,
					loadTime:       movement.DepartureTime,
					unloadTime:     movement.ArrivalTime,
				}

				candidate := routeCandidate{
					legs: []routeLeg{leg},
				}
				candidates = append(candidates, candidate)
			}
		}
	}

	return candidates
}

// findOneConnectionRoutes finds routes that require exactly one connection
func (s *RoutingApplicationService) findOneConnectionRoutes(voyages []routingdomain.Voyage, origin, destination routingdomain.UnLocode, deadline time.Time) []routeCandidate {
	var candidates []routeCandidate

	// Find all intermediate locations
	intermediateLocations := s.findIntermediateLocations(voyages, origin, destination)

	for _, intermediate := range intermediateLocations {
		// Find first leg: origin to intermediate
		firstLegs := s.findLegsConnecting(voyages, origin, intermediate)

		// Find second leg: intermediate to destination
		secondLegs := s.findLegsConnecting(voyages, intermediate, destination)

		// Combine compatible legs
		for _, firstLeg := range firstLegs {
			for _, secondLeg := range secondLegs {
				// Check transshipment time and deadline
				if secondLeg.loadTime.After(firstLeg.unloadTime.Add(2*time.Hour)) &&
					(secondLeg.unloadTime.Before(deadline) || secondLeg.unloadTime.Equal(deadline)) {

					candidate := routeCandidate{
						legs: []routeLeg{firstLeg, secondLeg},
					}
					candidates = append(candidates, candidate)
				}
			}
		}
	}

	return candidates
}

// findIntermediateLocations finds potential connection points
func (s *RoutingApplicationService) findIntermediateLocations(voyages []routingdomain.Voyage, origin, destination routingdomain.UnLocode) []routingdomain.UnLocode {
	locationSet := make(map[routingdomain.UnLocode]bool)

	for _, voyage := range voyages {
		schedule := voyage.GetSchedule()
		for _, movement := range schedule.Movements {
			if movement.DepartureLocation != origin && movement.DepartureLocation != destination {
				locationSet[movement.DepartureLocation] = true
			}
			if movement.ArrivalLocation != origin && movement.ArrivalLocation != destination {
				locationSet[movement.ArrivalLocation] = true
			}
		}
	}

	var intermediates []routingdomain.UnLocode
	for location := range locationSet {
		intermediates = append(intermediates, location)
	}

	return intermediates
}

// findLegsConnecting finds voyage legs that connect two locations
func (s *RoutingApplicationService) findLegsConnecting(voyages []routingdomain.Voyage, origin, destination routingdomain.UnLocode) []routeLeg {
	var legs []routeLeg

	for _, voyage := range voyages {
		schedule := voyage.GetSchedule()
		for _, movement := range schedule.Movements {
			if movement.DepartureLocation == origin && movement.ArrivalLocation == destination {
				leg := routeLeg{
					voyageNumber:   voyage.GetVoyageNumber(),
					loadLocation:   movement.DepartureLocation,
					unloadLocation: movement.ArrivalLocation,
					loadTime:       movement.DepartureTime,
					unloadTime:     movement.ArrivalTime,
				}
				legs = append(legs, leg)
			}
		}
	}

	return legs
}

// convertToExternalFormat converts internal route candidates to external itinerary format
func (s *RoutingApplicationService) convertToExternalFormat(candidates []routeCandidate) []routingdomain.Itinerary {
	var itineraries []routingdomain.Itinerary

	for _, candidate := range candidates {
		var legs []routingdomain.Leg

		for _, leg := range candidate.legs {
			externalLeg := routingdomain.Leg{
				VoyageNumber:   leg.voyageNumber.String(),
				LoadLocation:   leg.loadLocation.String(),
				UnloadLocation: leg.unloadLocation.String(),
				LoadTime:       leg.loadTime.Format(time.RFC3339),
				UnloadTime:     leg.unloadTime.Format(time.RFC3339),
			}
			legs = append(legs, externalLeg)
		}

		itinerary := routingdomain.Itinerary{
			Legs: legs,
		}
		itineraries = append(itineraries, itinerary)
	}

	return itineraries
}

// ListAllVoyages retrieves all voyages from the repository
func (s *RoutingApplicationService) ListAllVoyages(ctx context.Context) ([]routingdomain.Voyage, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized voyages list attempt", "error", err)
		return nil, err
	}
	if err := RequireRoutingPermission(claims, auth.PermissionViewVoyages); err != nil {
		s.logger.Warn("Unauthorized voyages list attempt", "error", err)
		return nil, err
	}

	s.logger.Info("Listing all voyages")

	// Get all voyages from repository
	allVoyages, err := s.voyageRepo.FindAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all voyages", "error", err)
		return nil, err
	}

	s.logger.Info("Retrieved all voyages", "count", len(allVoyages))
	return allVoyages, nil
}

// ListAllLocations retrieves all locations from the repository
func (s *RoutingApplicationService) ListAllLocations(ctx context.Context) ([]routingdomain.Location, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized locations list attempt", "error", err)
		return nil, err
	}
	if err := RequireRoutingPermission(claims, auth.PermissionViewLocations); err != nil {
		s.logger.Warn("Unauthorized locations list attempt", "error", err)
		return nil, err
	}

	s.logger.Info("Listing all locations")

	// Get all locations from repository
	allLocations, err := s.locationRepo.FindAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all locations", "error", err)
		return nil, err
	}

	s.logger.Info("Retrieved all locations", "count", len(allLocations))
	return allLocations, nil
}
