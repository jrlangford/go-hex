package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"go_hex/test/mock"
)

// TestCargoShippingWithMockApplications demonstrates the new mock data generation strategy
// where data is created through application layer methods instead of direct repository population
func TestCargoShippingWithMockApplications(t *testing.T) {
	// Set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create test environment with mock applications
	seed := time.Now().UnixNano()
	t.Logf("Using test data seed: %d", seed)

	mockEnv, err := mock.NewMockTestEnvironment(seed, logger)
	if err != nil {
		t.Fatalf("Failed to create mock test environment: %v", err)
	}

	ctx := context.Background()

	// Populate test data through application layer methods
	err = mockEnv.PopulateTestData(ctx)
	if err != nil {
		t.Fatalf("Failed to populate test data: %v", err)
	}

	// Print test data summary
	mockEnv.PrintTestDataSummary()

	// Verify that data was created properly
	testData := mockEnv.GetTestDataSummary()

	// Verify locations were created
	if len(testData.Locations) < 5 {
		t.Errorf("Expected at least 5 locations, got %d", len(testData.Locations))
	}

	// Verify voyages were created with proper movements
	if len(testData.Voyages) < 3 {
		t.Errorf("Expected at least 3 voyages, got %d", len(testData.Voyages))
	}

	for i, voyage := range testData.Voyages {
		movements := voyage.GetSchedule().Movements
		if len(movements) < 2 {
			t.Errorf("Voyage %d should have at least 2 movements, got %d", i, len(movements))
		}

		// Verify movement times are in the past (since they're completed journeys)
		for j, movement := range movements {
			if movement.ArrivalTime.After(time.Now()) {
				t.Errorf("Voyage %d movement %d arrival time should be in the past, got %v",
					i, j, movement.ArrivalTime)
			}
		}
	}

	// Verify cargo was created with future deadlines
	if len(testData.Cargos) < 2 {
		t.Errorf("Expected at least 2 cargos, got %d", len(testData.Cargos))
	}

	for i, cargo := range testData.Cargos {
		routeSpec := cargo.GetRouteSpecification()

		// Verify arrival deadline is in the future
		if routeSpec.ArrivalDeadline.Before(time.Now()) {
			t.Errorf("Cargo %d arrival deadline should be in the future, got %v",
				i, routeSpec.ArrivalDeadline)
		}

		// Verify tracking ID is properly set
		if cargo.GetTrackingId().String() == "" {
			t.Errorf("Cargo %d should have a valid tracking ID", i)
		}
	}

	// Verify handling events were created with past completion times
	if len(testData.HandlingEvents) < 2 {
		t.Errorf("Expected at least 2 handling events, got %d", len(testData.HandlingEvents))
	}

	for i, event := range testData.HandlingEvents {
		// Verify completion time is in the past
		if event.GetCompletionTime().After(time.Now()) {
			t.Errorf("Handling event %d completion time should be in the past, got %v",
				i, event.GetCompletionTime())
		}

		// Verify completion time is not too far in the past (within 30 days)
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		if event.GetCompletionTime().Before(thirtyDaysAgo) {
			t.Errorf("Handling event %d completion time should be within 30 days, got %v",
				i, event.GetCompletionTime())
		}
	}

	// Verify repositories contain the data (created through business logic)
	allLocations, err := mockEnv.LocationRepo.FindAll()
	if err != nil {
		t.Fatalf("Failed to retrieve locations from repository: %v", err)
	}
	if len(allLocations) != len(testData.Locations) {
		t.Errorf("Repository should contain %d locations, got %d",
			len(testData.Locations), len(allLocations))
	}

	allVoyages, err := mockEnv.VoyageRepo.FindAll()
	if err != nil {
		t.Fatalf("Failed to retrieve voyages from repository: %v", err)
	}
	if len(allVoyages) != len(testData.Voyages) {
		t.Errorf("Repository should contain %d voyages, got %d",
			len(testData.Voyages), len(allVoyages))
	}

	allCargos, err := mockEnv.CargoRepo.FindAll()
	if err != nil {
		t.Fatalf("Failed to retrieve cargos from repository: %v", err)
	}
	if len(allCargos) != len(testData.Cargos) {
		t.Errorf("Repository should contain %d cargos, got %d",
			len(testData.Cargos), len(allCargos))
	}

	t.Logf("✅ Successfully tested mock data generation through application layer")
	t.Logf("   - %d locations created", len(testData.Locations))
	t.Logf("   - %d voyages created", len(testData.Voyages))
	t.Logf("   - %d cargos created", len(testData.Cargos))
	t.Logf("   - %d handling events created", len(testData.HandlingEvents))
}

// TestMockApplicationsReproducibility verifies that using the same seed produces identical data
func TestMockApplicationsReproducibility(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Fixed seed for reproducibility
	seed := int64(12345)

	// Create first environment
	env1, err := mock.NewMockTestEnvironment(seed, logger)
	if err != nil {
		t.Fatalf("Failed to create first test environment: %v", err)
	}

	ctx := context.Background()
	err = env1.PopulateTestData(ctx)
	if err != nil {
		t.Fatalf("Failed to populate first test data: %v", err)
	}

	// Create second environment with same seed
	env2, err := mock.NewMockTestEnvironment(seed, logger)
	if err != nil {
		t.Fatalf("Failed to create second test environment: %v", err)
	}

	err = env2.PopulateTestData(ctx)
	if err != nil {
		t.Fatalf("Failed to populate second test data: %v", err)
	}

	// Compare data sets
	data1 := env1.GetTestDataSummary()
	data2 := env2.GetTestDataSummary()

	// Should have same number of entities
	if len(data1.Locations) != len(data2.Locations) {
		t.Errorf("Location count should be identical: %d vs %d",
			len(data1.Locations), len(data2.Locations))
	}

	if len(data1.Voyages) != len(data2.Voyages) {
		t.Errorf("Voyage count should be identical: %d vs %d",
			len(data1.Voyages), len(data2.Voyages))
	}

	if len(data1.Cargos) != len(data2.Cargos) {
		t.Errorf("Cargo count should be identical: %d vs %d",
			len(data1.Cargos), len(data2.Cargos))
	}

	// Location codes should be identical (same order due to same seed)
	for i := range data1.Locations {
		if i < len(data2.Locations) {
			code1 := data1.Locations[i].GetUnLocode().String()
			code2 := data2.Locations[i].GetUnLocode().String()
			if code1 != code2 {
				t.Errorf("Location %d code should be identical: %s vs %s", i, code1, code2)
			}
		}
	}

	t.Logf("✅ Reproducibility test passed - same seed produces identical data structure")
}
