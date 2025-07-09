package mock

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go_hex/internal/adapters/driven/in_memory_cargo_repo"
	"go_hex/internal/adapters/driven/in_memory_handling_repo"
	"go_hex/internal/adapters/driven/in_memory_location_repo"
	"go_hex/internal/adapters/driven/in_memory_voyage_repo"
	"go_hex/internal/adapters/driven/stdout_event_publisher"
	"go_hex/internal/adapters/integration"
	"go_hex/internal/booking/domain"
	bookingMock "go_hex/internal/booking/mock"
	handlingDomain "go_hex/internal/handling/domain"
	handlingMock "go_hex/internal/handling/mock"
	routingDomain "go_hex/internal/routing/domain"
	routingMock "go_hex/internal/routing/mock"
)

// MockTestEnvironment provides a complete test environment using mock applications
// that create test data through proper application layer methods
type MockTestEnvironment struct {
	// Mock Applications (embed real applications with test capabilities)
	BookingApp  *bookingMock.MockBookingApplication
	RoutingApp  *routingMock.MockRoutingApplication
	HandlingApp *handlingMock.MockHandlingApplication

	// Repositories (clean, no mock data)
	CargoRepo         *in_memory_cargo_repo.InMemoryCargoRepository
	VoyageRepo        *in_memory_voyage_repo.InMemoryVoyageRepository
	LocationRepo      *in_memory_location_repo.InMemoryLocationRepository
	HandlingEventRepo *in_memory_handling_repo.InMemoryHandlingEventRepository

	// Generated Test Data
	TestData TestDataSnapshot
	Logger   *slog.Logger
	Seed     int64
}

// NewMockTestEnvironment creates a complete test environment with mock applications
func NewMockTestEnvironment(seed int64, logger *slog.Logger) (*MockTestEnvironment, error) {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	logger.Info("Creating mock test environment", "seed", seed)

	// Create clean repositories (no mock data)
	cargoRepo := in_memory_cargo_repo.NewInMemoryCargoRepository().(*in_memory_cargo_repo.InMemoryCargoRepository)
	voyageRepo := in_memory_voyage_repo.NewInMemoryVoyageRepository().(*in_memory_voyage_repo.InMemoryVoyageRepository)
	locationRepo := in_memory_location_repo.NewInMemoryLocationRepository().(*in_memory_location_repo.InMemoryLocationRepository)
	handlingEventRepo := in_memory_handling_repo.NewInMemoryHandlingEventRepository().(*in_memory_handling_repo.InMemoryHandlingEventRepository)

	// Create event publisher
	eventPublisher := stdout_event_publisher.NewStdoutEventPublisher()

	// Create mock applications with embedded real applications
	routingApp := routingMock.NewMockRoutingApplication(voyageRepo, locationRepo, logger, seed)
	routingServiceAdapter := integration.NewRoutingServiceAdapter(routingApp.RoutingApplicationService)
	bookingApp := bookingMock.NewMockBookingApplication(cargoRepo, routingServiceAdapter, eventPublisher, logger, seed)
	handlingApp := handlingMock.NewMockHandlingApplication(handlingEventRepo, eventPublisher, logger, seed)

	return &MockTestEnvironment{
		BookingApp:        bookingApp,
		RoutingApp:        routingApp,
		HandlingApp:       handlingApp,
		CargoRepo:         cargoRepo,
		VoyageRepo:        voyageRepo,
		LocationRepo:      locationRepo,
		HandlingEventRepo: handlingEventRepo,
		Logger:            logger,
		Seed:              seed,
	}, nil
}

// PopulateTestData creates test data through application layer methods
func (env *MockTestEnvironment) PopulateTestData(ctx context.Context) error {
	env.Logger.Info("Populating test data through mock applications")

	// Step 1: Create locations through routing application
	locationSpecs := env.RoutingApp.GenerateStandardLocations(8 + (int(env.Seed) % 5)) // 8-12 locations
	locations, err := env.RoutingApp.PopulateTestLocations(ctx, locationSpecs)
	if err != nil {
		return fmt.Errorf("failed to populate locations: %w", err)
	}

	// Step 2: Create voyages through routing application
	voyageCount := 5 + (int(env.Seed) % 6) // 5-10 voyages
	voyages, err := env.RoutingApp.PopulateTestVoyages(ctx, locations, voyageCount)
	if err != nil {
		return fmt.Errorf("failed to populate voyages: %w", err)
	}

	// Step 3: Create cargo scenarios through booking application
	locationCodes := make([]string, len(locations))
	for i, loc := range locations {
		locationCodes[i] = loc.GetUnLocode().String()
	}

	cargoScenarios := env.BookingApp.GenerateCargoScenarios(locationCodes, 3+(int(env.Seed)%5)) // 3-7 cargo
	cargos, err := env.BookingApp.PopulateTestCargo(ctx, cargoScenarios)
	if err != nil {
		return fmt.Errorf("failed to populate cargo: %w", err)
	}

	// Step 4: Create handling events through handling application
	trackingIds := make([]string, len(cargos))
	for i, cargo := range cargos {
		trackingIds[i] = cargo.GetTrackingId().String()
	}

	handlingScenarios := env.HandlingApp.GenerateHandlingScenarios(trackingIds, locationCodes)
	handlingEvents, err := env.HandlingApp.PopulateTestHandlingEvents(ctx, handlingScenarios)
	if err != nil {
		return fmt.Errorf("failed to populate handling events: %w", err)
	}

	// Store the generated data for reference
	env.TestData = TestDataSnapshot{
		Locations:      locations,
		Voyages:        voyages,
		Cargos:         cargos,
		HandlingEvents: handlingEvents,
		Seed:           env.Seed,
		GeneratedAt:    time.Now(),
	}

	env.Logger.Info("Successfully populated test data",
		"locations", len(locations),
		"voyages", len(voyages),
		"cargos", len(cargos),
		"handling_events", len(handlingEvents))

	return nil
}

// GetTestDataSummary returns a summary of the generated test data
func (env *MockTestEnvironment) GetTestDataSummary() TestDataSnapshot {
	return env.TestData
}

// PrintTestDataSummary logs a detailed summary of the generated test data
func (env *MockTestEnvironment) PrintTestDataSummary() {
	env.Logger.Info("=== Test Data Summary ===")
	env.Logger.Info("Test environment seed", "seed", env.Seed)
	env.Logger.Info("Generated locations", "count", len(env.TestData.Locations))
	env.Logger.Info("Generated voyages", "count", len(env.TestData.Voyages))
	env.Logger.Info("Generated cargos", "count", len(env.TestData.Cargos))
	env.Logger.Info("Generated handling events", "count", len(env.TestData.HandlingEvents))

	// Print detailed location info
	for i, location := range env.TestData.Locations {
		env.Logger.Info("Location",
			"index", i,
			"code", location.GetUnLocode().String(),
			"name", location.GetName())
	}

	// Print detailed voyage info
	for i, voyage := range env.TestData.Voyages {
		movements := voyage.GetSchedule().Movements
		env.Logger.Info("Voyage",
			"index", i,
			"number", voyage.GetVoyageNumber().String(),
			"movements", len(movements))

		for j, movement := range movements {
			env.Logger.Info("Movement",
				"voyage_index", i,
				"movement_index", j,
				"from", movement.DepartureLocation.String(),
				"to", movement.ArrivalLocation.String(),
				"departure", movement.DepartureTime.Format("2006-01-02 15:04"),
				"arrival", movement.ArrivalTime.Format("2006-01-02 15:04"))
		}
	}

	// Print detailed cargo info
	for i, cargo := range env.TestData.Cargos {
		routeSpec := cargo.GetRouteSpecification()
		env.Logger.Info("Cargo",
			"index", i,
			"tracking_id", cargo.GetTrackingId().String(),
			"origin", routeSpec.Origin,
			"destination", routeSpec.Destination,
			"deadline", routeSpec.ArrivalDeadline.Format("2006-01-02 15:04"))
	}

	// Print handling events summary
	for i, event := range env.TestData.HandlingEvents {
		env.Logger.Info("Handling Event",
			"index", i,
			"tracking_id", event.GetTrackingId(),
			"event_type", string(event.GetEventType()),
			"location", event.GetLocation(),
			"voyage", event.GetVoyageNumber(),
			"completion_time", event.GetCompletionTime().Format("2006-01-02 15:04"))
	}
}

// TestDataSnapshot represents a snapshot of all generated test data
type TestDataSnapshot struct {
	Locations      []routingDomain.Location
	Voyages        []routingDomain.Voyage
	Cargos         []domain.Cargo
	HandlingEvents []handlingDomain.HandlingEvent
	Seed           int64
	GeneratedAt    time.Time
}
