package bookingapplication

import (
	"context"
	"go_hex/internal/booking/bookingdomain"
	"go_hex/internal/booking/ports/bookingprimary"
	"go_hex/internal/booking/ports/bookingsecondary"
	"go_hex/internal/support/auth"
	"log/slog"
	"time"
)

// BookingApplicationService implements the primary ports for booking operations
type BookingApplicationService struct {
	cargoRepo      bookingsecondary.CargoRepository
	routingService bookingsecondary.RoutingService
	eventPublisher bookingsecondary.EventPublisher
	logger         *slog.Logger
}

// Ensure BookingApplicationService implements the primary ports
var _ bookingprimary.BookingService = (*BookingApplicationService)(nil)

// NewBookingApplicationService creates a new BookingApplicationService
func NewBookingApplicationService(
	cargoRepo bookingsecondary.CargoRepository,
	routingService bookingsecondary.RoutingService,
	eventPublisher bookingsecondary.EventPublisher,
	logger *slog.Logger,
) *BookingApplicationService {
	return &BookingApplicationService{
		cargoRepo:      cargoRepo,
		routingService: routingService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// BookNewCargo initiates the creation of a new cargo based on customer's request
func (s *BookingApplicationService) BookNewCargo(ctx context.Context, origin, destination string, arrivalDeadlineStr string) (bookingdomain.Cargo, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized cargo booking attempt", "error", err)
		return bookingdomain.Cargo{}, err
	}
	if err := RequireBookingPermission(claims, auth.PermissionBookCargo); err != nil {
		s.logger.Warn("Unauthorized cargo booking attempt", "error", err)
		return bookingdomain.Cargo{}, err
	}

	s.logger.Info("Booking new cargo",
		"origin", origin,
		"destination", destination,
		"arrivalDeadline", arrivalDeadlineStr)

	// Parse arrival deadline
	arrivalDeadline, err := time.Parse(time.RFC3339, arrivalDeadlineStr)
	if err != nil {
		s.logger.Error("Invalid arrival deadline format", "error", err)
		return bookingdomain.Cargo{}, bookingdomain.NewDomainValidationError("invalid arrival deadline format, expected RFC3339", err)
	}

	// Create new cargo
	cargo, err := bookingdomain.NewCargo(origin, destination, arrivalDeadline)
	if err != nil {
		s.logger.Error("Failed to create new cargo", "error", err)
		return bookingdomain.Cargo{}, err
	}

	// Store cargo
	if err := s.cargoRepo.Store(cargo); err != nil {
		s.logger.Error("Failed to store cargo", "trackingId", cargo.GetTrackingId(), "error", err)
		return bookingdomain.Cargo{}, err
	}

	// Publish domain events
	s.publishCargoEvents(cargo)

	s.logger.Info("Cargo booked successfully", "trackingId", cargo.GetTrackingId())
	return cargo, nil
}

// AssignRouteToCargo assigns a chosen itinerary to an existing cargo
func (s *BookingApplicationService) AssignRouteToCargo(ctx context.Context, trackingId bookingdomain.TrackingId, itinerary bookingdomain.Itinerary) error {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized route assignment attempt", "trackingId", trackingId, "error", err)
		return err
	}
	if err := RequireBookingPermission(claims, auth.PermissionAssignRoute); err != nil {
		s.logger.Warn("Unauthorized route assignment attempt", "trackingId", trackingId, "error", err)
		return err
	}

	s.logger.Info("Assigning route to cargo", "trackingId", trackingId)

	// Find cargo
	cargo, err := s.cargoRepo.FindByTrackingId(trackingId)
	if err != nil {
		s.logger.Error("Cargo not found", "trackingId", trackingId, "error", err)
		return err
	}

	// Assign route
	if err := cargo.AssignToRoute(itinerary); err != nil {
		s.logger.Error("Failed to assign route", "trackingId", trackingId, "error", err)
		return err
	}

	// Update cargo
	if err := s.cargoRepo.Update(cargo); err != nil {
		s.logger.Error("Failed to update cargo", "trackingId", trackingId, "error", err)
		return err
	}

	// Publish domain events
	s.publishCargoEvents(cargo)

	s.logger.Info("Route assigned successfully", "trackingId", trackingId)
	return nil
}

// GetCargoDetails retrieves the full state of a cargo for tracking
func (s *BookingApplicationService) GetCargoDetails(ctx context.Context, trackingId bookingdomain.TrackingId) (bookingdomain.Cargo, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized cargo view attempt", "trackingId", trackingId, "error", err)
		return bookingdomain.Cargo{}, err
	}
	if err := RequireBookingPermission(claims, auth.PermissionViewCargo); err != nil {
		s.logger.Warn("Unauthorized cargo view attempt", "trackingId", trackingId, "error", err)
		return bookingdomain.Cargo{}, err
	}

	s.logger.Debug("Getting cargo details", "trackingId", trackingId)

	cargo, err := s.cargoRepo.FindByTrackingId(trackingId)
	if err != nil {
		s.logger.Error("Cargo not found", "trackingId", trackingId, "error", err)
		return bookingdomain.Cargo{}, err
	}

	return cargo, nil
}

// TrackCargo returns the current status of cargo by tracking ID (implements CargoTracker)
func (s *BookingApplicationService) TrackCargo(ctx context.Context, trackingId bookingdomain.TrackingId) (bookingdomain.Cargo, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized cargo tracking attempt", "trackingId", trackingId, "error", err)
		return bookingdomain.Cargo{}, err
	}
	if err := RequireBookingPermission(claims, auth.PermissionTrackCargo); err != nil {
		s.logger.Warn("Unauthorized cargo tracking attempt", "trackingId", trackingId, "error", err)
		return bookingdomain.Cargo{}, err
	}

	return s.GetCargoDetails(ctx, trackingId)
}

// ListUnroutedCargo gets all cargo that require route assignment
func (s *BookingApplicationService) ListUnroutedCargo(ctx context.Context) ([]bookingdomain.Cargo, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized unrouted cargo list attempt", "error", err)
		return nil, err
	}
	if err := RequireBookingPermission(claims, auth.PermissionViewCargo); err != nil {
		s.logger.Warn("Unauthorized unrouted cargo list attempt", "error", err)
		return nil, err
	}

	s.logger.Debug("Listing unrouted cargo")

	cargo, err := s.cargoRepo.FindUnrouted()
	if err != nil {
		s.logger.Error("Failed to list unrouted cargo", "error", err)
		return nil, err
	}

	s.logger.Debug("Found unrouted cargo", "count", len(cargo))
	return cargo, nil
}

// RequestRouteCandidates gets possible itineraries for a cargo
func (s *BookingApplicationService) RequestRouteCandidates(ctx context.Context, trackingId bookingdomain.TrackingId) ([]bookingdomain.Itinerary, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized route candidates request", "trackingId", trackingId, "error", err)
		return nil, err
	}
	if err := RequireBookingPermission(claims, auth.PermissionAssignRoute); err != nil {
		s.logger.Warn("Unauthorized route candidates request", "trackingId", trackingId, "error", err)
		return nil, err
	}

	s.logger.Info("Requesting route candidates", "trackingId", trackingId)

	// Find cargo
	cargo, err := s.cargoRepo.FindByTrackingId(trackingId)
	if err != nil {
		s.logger.Error("Cargo not found", "trackingId", trackingId, "error", err)
		return nil, err
	}

	// Request route candidates from routing service
	routeSpec := cargo.GetRouteSpecification()
	candidates, err := s.routingService.FindOptimalItineraries(ctx, routeSpec)
	if err != nil {
		s.logger.Error("Failed to find route candidates", "trackingId", trackingId, "error", err)
		return nil, err
	}

	s.logger.Info("Found route candidates", "trackingId", trackingId, "count", len(candidates))
	return candidates, nil
}

// UpdateCargoDelivery updates cargo delivery status based on handling events
func (s *BookingApplicationService) UpdateCargoDelivery(ctx context.Context, trackingId bookingdomain.TrackingId, handlingHistory []bookingdomain.HandlingEventSummary) error {
	s.logger.Info("Updating cargo delivery status", "trackingId", trackingId)

	// Find cargo
	cargo, err := s.cargoRepo.FindByTrackingId(trackingId)
	if err != nil {
		s.logger.Error("Cargo not found", "trackingId", trackingId, "error", err)
		return err
	}

	// Update delivery progress
	if err := cargo.DeriveDeliveryProgress(handlingHistory); err != nil {
		s.logger.Error("Failed to derive delivery progress", "trackingId", trackingId, "error", err)
		return err
	}

	// Update cargo
	if err := s.cargoRepo.Update(cargo); err != nil {
		s.logger.Error("Failed to update cargo", "trackingId", trackingId, "error", err)
		return err
	}

	// Publish domain events
	s.publishCargoEvents(cargo)

	s.logger.Info("Cargo delivery status updated", "trackingId", trackingId)
	return nil
}

// ListAllCargo retrieves all cargo from the repository
func (s *BookingApplicationService) ListAllCargo(ctx context.Context) ([]bookingdomain.Cargo, error) {
	// Check permissions
	claims, err := auth.ExtractClaims(ctx)
	if err != nil {
		s.logger.Warn("Unauthorized cargo list attempt", "error", err)
		return nil, err
	}
	if err := RequireBookingPermission(claims, auth.PermissionViewCargo); err != nil {
		s.logger.Warn("Unauthorized cargo list attempt", "error", err)
		return nil, err
	}

	s.logger.Info("Listing all cargo")

	// Get all cargo from repository
	allCargo, err := s.cargoRepo.FindAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all cargo", "error", err)
		return nil, err
	}

	s.logger.Info("Retrieved all cargo", "count", len(allCargo))
	return allCargo, nil
}

// publishCargoEvents publishes all pending events from the cargo aggregate
func (s *BookingApplicationService) publishCargoEvents(cargo bookingdomain.Cargo) {
	events := cargo.GetEvents()
	for _, event := range events {
		if err := s.eventPublisher.Publish(event); err != nil {
			s.logger.Error("Failed to publish event",
				"eventName", event.EventName(),
				"error", err)
		}
	}
	cargo.ClearEvents()
}
