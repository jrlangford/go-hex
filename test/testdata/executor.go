// Package testdata provides test scenario execution utilities
package testdata

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	bookingDomain "go_hex/internal/core/booking/domain"
	handlingDomain "go_hex/internal/core/handling/domain"
	handlingPrimary "go_hex/internal/core/handling/ports/primary"
)

// ScenarioExecutor runs test scenarios against the application services
type ScenarioExecutor struct {
	env    *TestEnvironment
	logger *slog.Logger
}

// NewScenarioExecutor creates a new scenario executor
func NewScenarioExecutor(env *TestEnvironment) *ScenarioExecutor {
	return &ScenarioExecutor{
		env:    env,
		logger: env.Logger,
	}
}

// ExecuteCargoScenario runs a complete cargo shipping scenario
func (se *ScenarioExecutor) ExecuteCargoScenario(ctx context.Context, scenario CargoTestData, bookingService interface {
	BookNewCargo(ctx context.Context, origin, destination string, arrivalDeadline string) (bookingDomain.Cargo, error)
	RequestRouteCandidates(ctx context.Context, trackingId bookingDomain.TrackingId) ([]bookingDomain.Itinerary, error)
	AssignRouteToCargo(ctx context.Context, trackingId bookingDomain.TrackingId, itinerary bookingDomain.Itinerary) error
	GetCargoDetails(ctx context.Context, trackingId bookingDomain.TrackingId) (bookingDomain.Cargo, error)
}, handlingService interface {
	SubmitHandlingReport(ctx context.Context, report handlingPrimary.HandlingReport) error
}) error {

	se.logger.Info("Executing cargo scenario",
		"origin", scenario.Origin,
		"destination", scenario.Destination,
		"deadline", scenario.ArrivalDeadline.Format(time.RFC3339))

	// Step 1: Book the cargo
	cargo, err := bookingService.BookNewCargo(ctx, scenario.Origin, scenario.Destination, scenario.ArrivalDeadline.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to book cargo: %w", err)
	}
	se.logger.Info("Cargo booked successfully", "tracking_id", cargo.GetTrackingId().String())

	// Step 2: Request route candidates
	candidates, err := bookingService.RequestRouteCandidates(ctx, cargo.GetTrackingId())
	if err != nil {
		return fmt.Errorf("failed to get route candidates: %w", err)
	}
	se.logger.Info("Route candidates found", "count", len(candidates))

	// Step 3: Assign route if candidates are available
	if len(candidates) > 0 {
		err = bookingService.AssignRouteToCargo(ctx, cargo.GetTrackingId(), candidates[0])
		if err != nil {
			return fmt.Errorf("failed to assign route: %w", err)
		}
		se.logger.Info("Route assigned successfully")

		// Verify cargo has itinerary
		updatedCargo, err := bookingService.GetCargoDetails(ctx, cargo.GetTrackingId())
		if err != nil {
			return fmt.Errorf("failed to get updated cargo: %w", err)
		}
		if updatedCargo.GetItinerary() == nil {
			return fmt.Errorf("cargo should have an itinerary after route assignment")
		}
	}

	// Step 4: Execute handling events
	for i, eventData := range scenario.HandlingEvents {
		se.logger.Info("Executing handling event",
			"index", i,
			"type", string(eventData.EventType),
			"location", eventData.Location,
			"delay", eventData.Delay)

		// Wait for the specified delay
		if eventData.Delay > 0 {
			time.Sleep(eventData.Delay)
		}

		// Submit handling report
		handlingReport := handlingPrimary.HandlingReport{
			TrackingId:     cargo.GetTrackingId().String(),
			EventType:      string(eventData.EventType),
			Location:       eventData.Location,
			VoyageNumber:   eventData.VoyageNumber,
			CompletionTime: eventData.CompletionTime.Format(time.RFC3339),
		}

		err = handlingService.SubmitHandlingReport(ctx, handlingReport)
		if err != nil {
			return fmt.Errorf("failed to submit handling report %d: %w", i, err)
		}
		se.logger.Info("Handling event submitted successfully", "index", i)

		// Small delay to allow event processing
		time.Sleep(100 * time.Millisecond)
	}

	// Step 5: Verify final cargo state
	finalCargo, err := bookingService.GetCargoDetails(ctx, cargo.GetTrackingId())
	if err != nil {
		return fmt.Errorf("failed to get final cargo state: %w", err)
	}

	delivery := finalCargo.GetDelivery()
	se.logger.Info("Final cargo state",
		"tracking_id", cargo.GetTrackingId().String(),
		"transport_status", string(delivery.TransportStatus),
		"routing_status", string(delivery.RoutingStatus),
		"last_known_location", delivery.LastKnownLocation,
		"is_on_track", delivery.IsOnTrack(),
		"is_delivered", delivery.IsDelivered())

	return nil
}

// ExecuteAllScenarios runs all generated cargo scenarios
func (se *ScenarioExecutor) ExecuteAllScenarios(ctx context.Context, bookingService interface {
	BookNewCargo(ctx context.Context, origin, destination string, arrivalDeadline string) (bookingDomain.Cargo, error)
	RequestRouteCandidates(ctx context.Context, trackingId bookingDomain.TrackingId) ([]bookingDomain.Itinerary, error)
	AssignRouteToCargo(ctx context.Context, trackingId bookingDomain.TrackingId, itinerary bookingDomain.Itinerary) error
	GetCargoDetails(ctx context.Context, trackingId bookingDomain.TrackingId) (bookingDomain.Cargo, error)
}, handlingService interface {
	SubmitHandlingReport(ctx context.Context, report handlingPrimary.HandlingReport) error
}) error {

	scenarios := se.env.GetTestScenarios()
	se.logger.Info("Executing all cargo scenarios", "count", len(scenarios))

	for i, scenario := range scenarios {
		se.logger.Info("Starting scenario execution", "index", i, "total", len(scenarios))

		err := se.ExecuteCargoScenario(ctx, scenario, bookingService, handlingService)
		if err != nil {
			se.logger.Error("Scenario execution failed", "index", i, "error", err)
			return fmt.Errorf("scenario %d failed: %w", i, err)
		}

		se.logger.Info("Scenario completed successfully", "index", i)

		// Brief pause between scenarios
		time.Sleep(200 * time.Millisecond)
	}

	se.logger.Info("All scenarios executed successfully")
	return nil
}

// GenerateHandlingEventsForCargo creates additional handling events for an existing cargo
func (se *ScenarioExecutor) GenerateHandlingEventsForCargo(cargo bookingDomain.Cargo, eventCount int) []HandlingEventData {
	generator := NewTestDataGenerator(0, se.logger)

	// Get cargo details for context
	routeSpec := cargo.GetRouteSpecification()

	events := []HandlingEventData{}
	currentTime := time.Now().Add(-2 * time.Hour) // Start events in the past

	// Generate realistic event progression
	eventTypes := []handlingDomain.HandlingEventType{
		handlingDomain.HandlingEventTypeReceive,
		handlingDomain.HandlingEventTypeLoad,
		handlingDomain.HandlingEventTypeUnload,
		handlingDomain.HandlingEventTypeClaim,
	}

	for i := 0; i < eventCount && i < len(eventTypes); i++ {
		location := routeSpec.Origin
		if i >= 2 { // UNLOAD and CLAIM happen at destination
			location = routeSpec.Destination
		}

		voyageNumber := ""
		if eventTypes[i] == handlingDomain.HandlingEventTypeLoad || eventTypes[i] == handlingDomain.HandlingEventTypeUnload {
			// Use a generated voyage number for load/unload events
			voyageNumber = fmt.Sprintf("V%03d", generator.random.Intn(999)+1)
		}

		events = append(events, HandlingEventData{
			EventType:      eventTypes[i],
			Location:       location,
			VoyageNumber:   voyageNumber,
			CompletionTime: currentTime,
			Delay:          time.Duration(i*500) * time.Millisecond,
		})

		// Increment time for next event (keep in the past)
		currentTime = currentTime.Add(time.Duration(2+generator.random.Intn(4)) * time.Hour)
	}

	return events
}
