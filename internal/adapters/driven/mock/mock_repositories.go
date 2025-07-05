package mock

import (
	"go_hex/internal/adapters/driven/in_memory_cargo_repo"
	"go_hex/internal/adapters/driven/in_memory_handling_repo"
	"go_hex/internal/adapters/driven/in_memory_location_repo"
	"go_hex/internal/adapters/driven/in_memory_voyage_repo"
	bookingSecondary "go_hex/internal/core/booking/ports/secondary"
	handlingSecondary "go_hex/internal/core/handling/ports/secondary"
	routingSecondary "go_hex/internal/core/routing/ports/secondary"
	"go_hex/test/testdata"
	"log/slog"
)

// MockDataManager provides pre-populated repositories with generated test data
type MockDataManager struct {
	CargoRepo    bookingSecondary.CargoRepository
	VoyageRepo   routingSecondary.VoyageRepository
	LocationRepo routingSecondary.LocationRepository
	HandlingRepo handlingSecondary.HandlingEventRepository
	testDataSet  testdata.TestDataSet
}

// NewMockDataManager creates repositories pre-populated with generated test data
func NewMockDataManager(logger *slog.Logger, seed ...int64) *MockDataManager {
	// Create repositories
	cargoRepo := in_memory_cargo_repo.NewInMemoryCargoRepository()
	voyageRepo := in_memory_voyage_repo.NewInMemoryVoyageRepository()
	locationRepo := in_memory_location_repo.NewInMemoryLocationRepository()
	handlingRepo := in_memory_handling_repo.NewInMemoryHandlingEventRepository()

	// Generate test data
	var seedValue int64
	if len(seed) > 0 {
		seedValue = seed[0]
	} else {
		seedValue = 0 // Generator will use current time if 0
	}

	generator := testdata.NewTestDataGenerator(seedValue, logger)
	testDataSet := generator.GenerateCompleteTestDataSet()

	manager := &MockDataManager{
		CargoRepo:    cargoRepo,
		VoyageRepo:   voyageRepo,
		LocationRepo: locationRepo,
		HandlingRepo: handlingRepo,
		testDataSet:  testDataSet,
	}

	// Populate repositories with generated data
	manager.populateRepositories(logger)

	return manager
}

// GetTestDataSet returns the underlying test data set used to populate repositories
func (m *MockDataManager) GetTestDataSet() testdata.TestDataSet {
	return m.testDataSet
}

// populateRepositories fills all repositories with the generated test data
func (m *MockDataManager) populateRepositories(logger *slog.Logger) {
	logger.Info("Populating repositories with generated test data")

	// Populate locations
	for _, location := range m.testDataSet.Locations {
		if err := m.LocationRepo.Store(location); err != nil {
			logger.Error("Failed to store location", "error", err, "code", location.GetUnLocode().String())
		}
	}
	logger.Info("Populated locations", "count", len(m.testDataSet.Locations))

	// Populate voyages
	for _, voyage := range m.testDataSet.Voyages {
		if err := m.VoyageRepo.Store(voyage); err != nil {
			logger.Error("Failed to store voyage", "error", err, "number", voyage.GetVoyageNumber().String())
		}
	}
	logger.Info("Populated voyages", "count", len(m.testDataSet.Voyages))

	// For cargo scenarios, we'll store the cargo entities directly
	// In mock mode, we want pre-existing cargo for demonstration
	cargoCount := 0
	for _, scenario := range m.testDataSet.CargoScenarios {
		// Store the cargo entity (it should already be a proper domain.Cargo)
		if err := m.CargoRepo.Store(scenario.Cargo); err != nil {
			logger.Error("Failed to store cargo", "error", err, "origin", scenario.Origin, "destination", scenario.Destination)
			continue
		}
		cargoCount++

		// Note: HandlingEventData is test data structure, not domain entities
		// In a real implementation, you'd convert these to proper domain.HandlingEvent entities
		// For now, we skip storing handling events as they need proper conversion
	}
	logger.Info("Populated cargo", "cargo_count", cargoCount, "scenarios", len(m.testDataSet.CargoScenarios))
}
