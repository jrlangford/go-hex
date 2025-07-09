package mock

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"go_hex/internal/booking/application"
	"go_hex/internal/booking/domain"
	"go_hex/internal/booking/ports/primary"
	"go_hex/internal/booking/ports/secondary"
	"go_hex/internal/support/auth"
)

// MockBookingApplication embeds the real application service but provides test data population capabilities
type MockBookingApplication struct {
	*application.BookingApplicationService
	logger *slog.Logger
	random *rand.Rand
}

// NewMockBookingApplication creates a mock booking application with embedded real application service
func NewMockBookingApplication(
	cargoRepo secondary.CargoRepository,
	routingService secondary.RoutingService,
	eventPublisher secondary.EventPublisher,
	logger *slog.Logger,
	seed int64,
) *MockBookingApplication {
	realApp := application.NewBookingApplicationService(cargoRepo, routingService, eventPublisher, logger)

	return &MockBookingApplication{
		BookingApplicationService: realApp,
		logger:                    logger,
		random:                    rand.New(rand.NewSource(seed)),
	}
}

// PopulateTestCargo creates test cargo data using business logic through the application layer
func (m *MockBookingApplication) PopulateTestCargo(ctx context.Context, scenarios []TestCargoScenario) ([]domain.Cargo, error) {
	m.logger.Info("Populating test cargo through booking application", "scenarios", len(scenarios))

	// Create authenticated context for internal operations
	testCtx := m.createTestContext(ctx)

	var cargos []domain.Cargo

	for _, scenario := range scenarios {
		// Generate arrival deadline in the future (7-60 days from now)
		daysInFuture := 7 + m.random.Intn(54) // 7-60 days
		arrivalDeadline := time.Now().AddDate(0, 0, daysInFuture)

		// Use the real application service to book cargo
		cargo, err := m.BookingApplicationService.BookNewCargo(
			testCtx,
			scenario.Origin,
			scenario.Destination,
			arrivalDeadline.Format(time.RFC3339),
		)
		if err != nil {
			m.logger.Error("Failed to create test cargo", "error", err, "origin", scenario.Origin, "destination", scenario.Destination)
			return nil, fmt.Errorf("failed to create test cargo from %s to %s: %w", scenario.Origin, scenario.Destination, err)
		}

		// If an itinerary is provided, assign it
		if scenario.Itinerary != nil {
			err = m.BookingApplicationService.AssignRouteToCargo(testCtx, cargo.GetTrackingId(), *scenario.Itinerary)
			if err != nil {
				m.logger.Warn("Failed to assign route to test cargo", "error", err, "trackingId", cargo.GetTrackingId())
				// Continue with other cargo even if route assignment fails
			}
		}

		cargos = append(cargos, cargo)
		m.logger.Debug("Created test cargo", "trackingId", cargo.GetTrackingId(), "origin", scenario.Origin, "destination", scenario.Destination)
	}

	m.logger.Info("Successfully populated test cargo", "count", len(cargos))
	return cargos, nil
}

// GenerateCargoScenarios creates realistic cargo scenarios with current dates
func (m *MockBookingApplication) GenerateCargoScenarios(locations []string, count int) []TestCargoScenario {
	m.logger.Info("Generating cargo scenarios", "count", count, "available_locations", len(locations))

	if len(locations) < 2 {
		m.logger.Warn("Not enough locations to generate cargo scenarios", "locations", len(locations))
		return nil
	}

	scenarios := make([]TestCargoScenario, 0, count)

	for i := 0; i < count; i++ {
		// Pick random origin and destination (ensuring they're different)
		originIdx := m.random.Intn(len(locations))
		destinationIdx := m.random.Intn(len(locations))
		for destinationIdx == originIdx && len(locations) > 1 {
			destinationIdx = m.random.Intn(len(locations))
		}

		scenario := TestCargoScenario{
			Origin:      locations[originIdx],
			Destination: locations[destinationIdx],
			// Itinerary will be assigned later by routing service if needed
			Itinerary: nil,
		}

		scenarios = append(scenarios, scenario)
	}

	m.logger.Info("Generated cargo scenarios", "count", len(scenarios))
	return scenarios
}

// createTestContext creates an authenticated context for test operations
func (m *MockBookingApplication) createTestContext(ctx context.Context) context.Context {
	// Create test claims with admin permissions
	claims, _ := auth.NewClaims(
		"test-user",
		"test-system",
		"test@example.com",
		[]string{string(auth.RoleAdmin)},
		map[string]string{"test": "true"},
	)

	return context.WithValue(ctx, auth.ClaimsContextKey, claims)
}

// TestCargoScenario represents a test scenario for cargo booking
type TestCargoScenario struct {
	Origin      string
	Destination string
	Itinerary   *domain.Itinerary
}

// Ensure MockBookingApplication implements primary ports
var _ primary.BookingService = (*MockBookingApplication)(nil)
var _ primary.CargoTracker = (*MockBookingApplication)(nil)
