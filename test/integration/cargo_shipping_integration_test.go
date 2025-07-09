package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"go_hex/internal/adapters/driven/event_bus"
	"go_hex/internal/adapters/driven/in_memory_cargo_repo"
	"go_hex/internal/adapters/driven/in_memory_handling_repo"
	"go_hex/internal/adapters/driven/in_memory_location_repo"
	"go_hex/internal/adapters/driven/in_memory_voyage_repo"
	"go_hex/internal/adapters/integration"
	"go_hex/internal/support/auth"

	"go_hex/internal/booking/bookingapplication"
	"go_hex/internal/booking/bookingdomain"
	"go_hex/internal/handling/handlingapplication"
	"go_hex/internal/handling/handlingdomain"
	routingApp "go_hex/internal/routing/application"
)

// createAuthenticatedContext creates a context with admin authentication for testing
func createAuthenticatedContext() context.Context {
	claims, _ := auth.NewClaims(
		"test-user-123",
		"testuser",
		"test@example.com",
		[]string{string(auth.RoleAdmin)},
		map[string]string{"test": "true"},
	)

	return context.WithValue(context.Background(), auth.ClaimsContextKey, claims)
}

func TestCargoShippingSystemIntegration(t *testing.T) {
	// Set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create event bus for inter-context communication
	eventBus := event_bus.NewInMemoryEventBus(logger)

	// Create repositories
	cargoRepo := in_memory_cargo_repo.NewInMemoryCargoRepository()
	voyageRepo := in_memory_voyage_repo.NewInMemoryVoyageRepository()
	locationRepo := in_memory_location_repo.NewInMemoryLocationRepository()
	handlingEventRepo := in_memory_handling_repo.NewInMemoryHandlingEventRepository()

	// Create Routing context application service
	routingService := routingApp.NewRoutingApplicationService(
		voyageRepo,
		locationRepo,
		logger,
	)

	// Create adapter for Booking->Routing integration (synchronous, customer-supplier)
	routingAdapter := integration.NewRoutingServiceAdapter(routingService)

	// Create Booking context application service
	bookingService := bookingapplication.NewBookingApplicationService(
		cargoRepo,
		routingAdapter, // Synchronous integration with routing
		eventBus,       // Event publisher
		logger,
	)

	// Create Handling context application services
	handlingReportService := handlingapplication.NewHandlingReportService(
		handlingEventRepo,
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

	ctx := createAuthenticatedContext()

	// Test 1: Book a new cargo
	t.Log("Test 1: Booking a new cargo")
	futureDeadline := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339) // 30 days from now
	cargo, err := bookingService.BookNewCargo(ctx, "SESTO", "NLRTM", futureDeadline)
	if err != nil {
		t.Fatalf("Failed to book cargo: %v", err)
	}
	t.Logf("Cargo booked successfully with tracking ID: %s", cargo.GetTrackingId().String())

	// Test 2: Request route candidates (Booking->Routing synchronous integration)
	t.Log("Test 2: Requesting route candidates")
	candidates, err := bookingService.RequestRouteCandidates(ctx, cargo.GetTrackingId())
	if err != nil {
		t.Fatalf("Failed to get route candidates: %v", err)
	}
	t.Logf("Found %d route candidates", len(candidates))

	if len(candidates) > 0 {
		// Test 3: Assign the first route candidate
		t.Log("Test 3: Assigning route to cargo")
		err = bookingService.AssignRouteToCargo(ctx, cargo.GetTrackingId(), candidates[0])
		if err != nil {
			t.Fatalf("Failed to assign route: %v", err)
		}
		t.Log("Route assigned successfully")

		// Verify cargo has itinerary
		updatedCargo, err := bookingService.GetCargoDetails(ctx, cargo.GetTrackingId())
		if err != nil {
			t.Fatalf("Failed to get updated cargo: %v", err)
		}
		if updatedCargo.GetItinerary() == nil {
			t.Fatal("Cargo should have an itinerary after route assignment")
		}
		t.Log("Cargo itinerary confirmed")
	}

	// Test 4: Submit handling report (Handling->Booking asynchronous integration)
	t.Log("Test 4: Submitting handling report")
	handlingReport := handlingdomain.HandlingReport{
		TrackingId:     cargo.GetTrackingId().String(),
		EventType:      string(handlingdomain.HandlingEventTypeLoad),
		Location:       "SESTO",
		VoyageNumber:   "V001",
		CompletionTime: time.Now().Format(time.RFC3339),
	}

	err = handlingReportService.SubmitHandlingReport(ctx, handlingReport)
	if err != nil {
		t.Fatalf("Failed to submit handling report: %v", err)
	}
	t.Log("Handling report submitted successfully")

	// Small delay to allow event processing
	time.Sleep(100 * time.Millisecond)

	// Test 5: Verify cargo delivery status was updated via event integration
	t.Log("Test 5: Verifying cargo delivery status update")
	finalCargo, err := bookingService.GetCargoDetails(ctx, cargo.GetTrackingId())
	if err != nil {
		t.Fatalf("Failed to get final cargo state: %v", err)
	}

	delivery := finalCargo.GetDelivery()
	if delivery.TransportStatus == bookingdomain.TransportStatusNotReceived {
		t.Error("Cargo delivery status should have been updated from handling event")
	} else {
		t.Logf("Cargo delivery status successfully updated: %s", delivery.TransportStatus)
	}

	t.Log("Integration test completed successfully!")
}
