package integration

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"go_hex/internal/adapters/driven/event_bus"
	"go_hex/internal/adapters/integration"

	"go_hex/internal/booking/bookingapplication"
	"go_hex/internal/handling/handlingapplication"
	"go_hex/internal/handling/handlingdomain"

	"go_hex/test/testdata"
)

func TestCargoShippingSystemIntegrationWithGeneratedData(t *testing.T) {
	// Set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create test environment with generated data
	seed := time.Now().UnixNano()
	t.Logf("Using test data seed: %d", seed)

	testEnv, err := testdata.NewTestEnvironment(seed, logger)
	if err != nil {
		t.Fatalf("Failed to create test environment: %v", err)
	}

	ctx := createAuthenticatedContext()

	// Populate repositories with generated test data
	err = testEnv.PopulateWithTestData(ctx)
	if err != nil {
		t.Fatalf("Failed to populate test data: %v", err)
	}

	// Print test data summary for debugging
	testEnv.PrintTestDataSummary()

	// Create event bus for inter-context communication
	eventBus := event_bus.NewInMemoryEventBus(logger)

	// Create adapter for Booking->Routing integration (synchronous, customer-supplier)
	routingAdapter := integration.NewRoutingServiceAdapter(testEnv.RoutingService)

	// Create Booking context application service
	bookingService := bookingapplication.NewBookingApplicationService(
		testEnv.CargoRepo,
		routingAdapter, // Synchronous integration with routing
		eventBus,       // Event publisher
		logger,
	)

	// Create Handling context application services
	handlingReportService := handlingapplication.NewHandlingReportService(
		testEnv.HandlingEventRepo,
		eventBus, // Event publisher for handling events
		logger,
	)

	// Set up event-driven integration: Handling->Booking (asynchronous, ACL)
	handlingToBookingHandler := integration.NewHandlingToBookingEventHandler(bookingService, logger)

	// Subscribe to handling events
	eventBus.Subscribe(
		handlingdomain.HandlingEventRegisteredEvent{}.EventName(),
		handlingToBookingHandler.HandleCargoWasHandled,
	)

	// Create scenario executor
	executor := testdata.NewScenarioExecutor(testEnv)

	// Execute all generated scenarios
	t.Log("Executing all generated test scenarios")
	err = executor.ExecuteAllScenarios(ctx, bookingService, handlingReportService)
	if err != nil {
		t.Fatalf("Failed to execute test scenarios: %v", err)
	}

	t.Log("All generated scenarios executed successfully!")
}

// TestSpecificCargoScenario demonstrates testing a specific scenario
func TestSpecificCargoScenario(t *testing.T) {
	// Set up logger with debug level for detailed output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Use a fixed seed for reproducible tests
	seed := int64(12345)
	testEnv, err := testdata.NewTestEnvironment(seed, logger)
	if err != nil {
		t.Fatalf("Failed to create test environment: %v", err)
	}

	ctx := createAuthenticatedContext()

	// Populate repositories with generated test data
	err = testEnv.PopulateWithTestData(ctx)
	if err != nil {
		t.Fatalf("Failed to populate test data: %v", err)
	}

	// Create event bus for inter-context communication
	eventBus := event_bus.NewInMemoryEventBus(logger)

	// Create adapter for Booking->Routing integration
	routingAdapter := integration.NewRoutingServiceAdapter(testEnv.RoutingService)

	// Create Booking context application service
	bookingService := bookingapplication.NewBookingApplicationService(
		testEnv.CargoRepo,
		routingAdapter,
		eventBus,
		logger,
	)

	// Create Handling context application services
	handlingReportService := handlingapplication.NewHandlingReportService(
		testEnv.HandlingEventRepo,
		eventBus,
		logger,
	)

	// Set up event integration
	handlingToBookingHandler := integration.NewHandlingToBookingEventHandler(bookingService, logger)
	eventBus.Subscribe(
		handlingdomain.HandlingEventRegisteredEvent{}.EventName(),
		handlingToBookingHandler.HandleCargoWasHandled,
	)

	// Get the first generated scenario
	scenarios := testEnv.GetTestScenarios()
	if len(scenarios) == 0 {
		t.Fatal("No test scenarios were generated")
	}

	scenario := scenarios[0]
	t.Logf("Testing specific scenario: %s -> %s", scenario.Origin, scenario.Destination)

	// Create scenario executor and run single scenario
	executor := testdata.NewScenarioExecutor(testEnv)
	err = executor.ExecuteCargoScenario(ctx, scenario, bookingService, handlingReportService)
	if err != nil {
		t.Fatalf("Failed to execute specific scenario: %v", err)
	}

	t.Log("Specific scenario executed successfully!")
}

// TestStressWithMultipleDataSets tests the system with multiple different data sets
func TestStressWithMultipleDataSets(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	testIterations := 3
	for i := 0; i < testIterations; i++ {
		t.Run(fmt.Sprintf("DataSet_%d", i), func(t *testing.T) {
			// Use different seed for each iteration
			seed := time.Now().UnixNano() + int64(i*1000)

			testEnv, err := testdata.NewTestEnvironment(seed, logger)
			if err != nil {
				t.Fatalf("Failed to create test environment for iteration %d: %v", i, err)
			}

			ctx := createAuthenticatedContext()
			err = testEnv.PopulateWithTestData(ctx)
			if err != nil {
				t.Fatalf("Failed to populate test data for iteration %d: %v", i, err)
			}

			// Create minimal setup for this test
			eventBus := event_bus.NewInMemoryEventBus(logger)
			routingAdapter := integration.NewRoutingServiceAdapter(testEnv.RoutingService)

			bookingService := bookingapplication.NewBookingApplicationService(
				testEnv.CargoRepo,
				routingAdapter,
				eventBus,
				logger,
			)

			handlingReportService := handlingapplication.NewHandlingReportService(
				testEnv.HandlingEventRepo,
				eventBus,
				logger,
			)

			handlingToBookingHandler := integration.NewHandlingToBookingEventHandler(bookingService, logger)
			eventBus.Subscribe(
				handlingdomain.HandlingEventRegisteredEvent{}.EventName(),
				handlingToBookingHandler.HandleCargoWasHandled,
			)

			// Run scenarios for this iteration
			executor := testdata.NewScenarioExecutor(testEnv)
			err = executor.ExecuteAllScenarios(ctx, bookingService, handlingReportService)
			if err != nil {
				t.Fatalf("Failed to execute scenarios for iteration %d: %v", i, err)
			}

			t.Logf("Iteration %d completed successfully with %d scenarios",
				i, len(testEnv.GetTestScenarios()))
		})
	}
}
