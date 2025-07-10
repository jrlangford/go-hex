// Package testdata provides test data population utilities for integration tests
package testdata

import (
	"context"
	"fmt"
	"log/slog"

	"go_hex/internal/adapters/driven/in_memory_cargo_repo"
	"go_hex/internal/adapters/driven/in_memory_handling_repo"
	"go_hex/internal/adapters/driven/in_memory_location_repo"
	"go_hex/internal/adapters/driven/in_memory_voyage_repo"
	"go_hex/internal/booking/bookingapplication"
	"go_hex/internal/handling/handlingapplication"
	"go_hex/internal/routing/routingapplication"
)

// TestEnvironment holds all the services and repositories needed for testing
type TestEnvironment struct {
	// Repositories
	CargoRepo         *in_memory_cargo_repo.InMemoryCargoRepository
	VoyageRepo        *in_memory_voyage_repo.InMemoryVoyageRepository
	LocationRepo      *in_memory_location_repo.InMemoryLocationRepository
	HandlingEventRepo *in_memory_handling_repo.InMemoryHandlingEventRepository

	// Application Services
	BookingService        *bookingapplication.BookingApplicationService
	RoutingService        *routingapplication.RoutingApplicationService
	HandlingReportService *handlingapplication.HandlingReportService

	// Test Data
	TestData TestDataSet
	Logger   *slog.Logger
}

// NewTestEnvironment creates a complete test environment with all dependencies wired
func NewTestEnvironment(seed int64, logger *slog.Logger) (*TestEnvironment, error) {
	// Create test data generator
	generator := NewTestDataGenerator(seed, logger)
	testData := generator.GenerateCompleteTestDataSet()

	// Create repositories (they auto-seed with default data)
	cargoRepo := in_memory_cargo_repo.NewInMemoryCargoRepository().(*in_memory_cargo_repo.InMemoryCargoRepository)
	voyageRepo := in_memory_voyage_repo.NewInMemoryVoyageRepository().(*in_memory_voyage_repo.InMemoryVoyageRepository)
	locationRepo := in_memory_location_repo.NewInMemoryLocationRepository().(*in_memory_location_repo.InMemoryLocationRepository)
	handlingEventRepo := in_memory_handling_repo.NewInMemoryHandlingEventRepository().(*in_memory_handling_repo.InMemoryHandlingEventRepository)

	// Create application services (this would normally be done in main.go)
	routingService := routingapplication.NewRoutingApplicationService(
		voyageRepo,
		locationRepo,
		logger,
	)

	return &TestEnvironment{
		CargoRepo:         cargoRepo,
		VoyageRepo:        voyageRepo,
		LocationRepo:      locationRepo,
		HandlingEventRepo: handlingEventRepo,
		RoutingService:    routingService,
		TestData:          testData,
		Logger:            logger,
	}, nil
}

// PopulateWithTestData adds the generated test data to repositories
func (env *TestEnvironment) PopulateWithTestData(ctx context.Context) error {
	env.Logger.Info("Populating repositories with generated test data")

	// Create repository interface wrapper
	repos := TestRepositories{
		LocationRepo:      env.LocationRepo,
		VoyageRepo:        env.VoyageRepo,
		CargoRepo:         env.CargoRepo,
		HandlingEventRepo: env.HandlingEventRepo,
	}

	if err := env.TestData.PopulateRepositories(ctx, repos); err != nil {
		return fmt.Errorf("failed to populate repositories: %w", err)
	}

	env.Logger.Info("Test data population completed successfully",
		"locations", len(env.TestData.Locations),
		"voyages", len(env.TestData.Voyages),
		"cargo_scenarios", len(env.TestData.CargoScenarios))

	return nil
}

// GetTestScenarios returns the generated cargo test scenarios
func (env *TestEnvironment) GetTestScenarios() []CargoTestData {
	return env.TestData.CargoScenarios
}

// PrintTestDataSummary logs a summary of the generated test data
func (env *TestEnvironment) PrintTestDataSummary() {
	env.Logger.Info("=== Test Data Summary ===")
	env.Logger.Info("Test data generation details",
		"seed", env.TestData.Seed,
		"generated_at", env.TestData.GeneratedAt)

	env.Logger.Info("Generated test entities",
		"locations", len(env.TestData.Locations),
		"voyages", len(env.TestData.Voyages),
		"cargo_scenarios", len(env.TestData.CargoScenarios))

	// Log location details
	env.Logger.Info("Generated locations:")
	for i, location := range env.TestData.Locations {
		env.Logger.Info("Location",
			"index", i,
			"code", location.GetUnLocode().String(),
			"name", location.GetName(),
			"country", location.GetCountry())
	}

	// Log voyage details
	env.Logger.Info("Generated voyages:")
	for i, voyage := range env.TestData.Voyages {
		schedule := voyage.GetSchedule()
		env.Logger.Info("Voyage",
			"index", i,
			"number", voyage.GetVoyageNumber().String(),
			"movements", len(schedule.Movements))

		for j, movement := range schedule.Movements {
			env.Logger.Info("Movement",
				"voyage_index", i,
				"movement_index", j,
				"from", movement.DepartureLocation.String(),
				"to", movement.ArrivalLocation.String(),
				"departure", movement.DepartureTime.Format("2006-01-02 15:04"),
				"arrival", movement.ArrivalTime.Format("2006-01-02 15:04"))
		}
	}

	// Log cargo scenario details
	env.Logger.Info("Generated cargo scenarios:")
	for i, scenario := range env.TestData.CargoScenarios {
		env.Logger.Info("Cargo scenario",
			"index", i,
			"tracking_id", scenario.Cargo.GetTrackingId().String(),
			"origin", scenario.Origin,
			"destination", scenario.Destination,
			"deadline", scenario.ArrivalDeadline.Format("2006-01-02 15:04"),
			"handling_events", len(scenario.HandlingEvents))

		for j, event := range scenario.HandlingEvents {
			env.Logger.Info("Handling event",
				"cargo_index", i,
				"event_index", j,
				"type", string(event.EventType),
				"location", event.Location,
				"voyage", event.VoyageNumber,
				"completion_time", event.CompletionTime.Format("2006-01-02 15:04:05"),
				"delay", event.Delay)
		}
	}
}
