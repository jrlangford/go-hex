package main

import (
	"context"
	"go_hex/internal/adapters/driven/event_bus"
	"go_hex/internal/adapters/driven/in_memory_cargo_repo"
	"go_hex/internal/adapters/driven/in_memory_handling_repo"
	"go_hex/internal/adapters/driven/in_memory_location_repo"
	"go_hex/internal/adapters/driven/in_memory_voyage_repo"
	httpadapter "go_hex/internal/adapters/driving/httpadapter"
	"go_hex/internal/adapters/driving/httpadapter/httpmiddleware"
	"go_hex/internal/adapters/integration"

	"go_hex/internal/booking/bookingapplication"
	"go_hex/internal/booking/ports/bookingprimary"
	"go_hex/internal/handling/handlingapplication"
	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/handling/ports/handlingprimary"
	"go_hex/internal/routing/ports/routingprimary"
	"go_hex/internal/routing/routingapplication"

	"go_hex/internal/booking/bookingmock"
	"go_hex/internal/handling/handlingmock"
	"go_hex/internal/routing/routingmock"

	"go_hex/internal/support/auth"
	"go_hex/internal/support/config"
	"go_hex/internal/support/logging"
	"go_hex/internal/support/server"
	"log"
	"log/slog"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logging.Initialize(cfg)
	logger := logging.Get()

	httpHandler := wireAppDependencies(cfg, logger)

	httpServer := server.New(cfg, httpHandler, nil)
	if err := httpServer.Start(); err != nil {
		logger.Error("Server startup failed", "error", err)
	}

	log.Println("Server exited")
}

func wireAppDependencies(cfg *config.Config, logger *slog.Logger) *httpadapter.Handler {
	// Create event bus for inter-context communication
	eventBus := event_bus.NewInMemoryEventBus(logger)

	// Create repositories
	cargoRepo := in_memory_cargo_repo.NewInMemoryCargoRepository()
	voyageRepo := in_memory_voyage_repo.NewInMemoryVoyageRepository()
	locationRepo := in_memory_location_repo.NewInMemoryLocationRepository()
	handlingEventRepo := in_memory_handling_repo.NewInMemoryHandlingEventRepository()

	var bookingService bookingprimary.BookingService
	var handlingReportService handlingprimary.HandlingReportService
	var routingService routingprimary.RouteFinder

	if cfg.IsMockMode() {
		logger.Info("Running in mock mode with pre-populated mock data", "mode", cfg.Mode, "isMockMode", cfg.IsMockMode())

		// Create Routing context application service
		mockRoutingService := routingmock.NewMockRoutingApplication(
			voyageRepo,
			locationRepo,
			logger,
			1017, // Use seed or reproducibility
		)
		routingService = mockRoutingService

		// Create adapter for Booking->Routing integration (synchronous, customer-supplier)
		routingAdapter := integration.NewRoutingServiceAdapter(routingService)

		// Create Mock Booking context application service
		mockBookingService := bookingmock.NewMockBookingApplication(
			cargoRepo,
			routingAdapter, // Synchronous integration with routing
			eventBus,       // Event publisher
			logger,
			1017, // Use seed for reproducibility
		)
		bookingService = mockBookingService

		handlingReportService = handlingmock.NewMockHandlingApplication(
			handlingEventRepo,
			eventBus, // Event publisher for handling events
			logger,
			1017, // Use seed for reproducibility
		)

		claims, err := auth.NewClaims(
			"test-user",
			"test-system",
			"test@example.com",
			[]string{string(auth.RoleAdmin)},
			map[string]string{"test": "true"},
		)
		if err != nil {
			logger.Error("Failed to create test claims", "error", err)
			log.Panic("Failed to create test claims:", err)
		}

		ctx := context.WithValue(context.Background(), auth.ClaimsContextKey, claims)

		mockRoutingService.GenerateTestData() // Populate mock data

		locations, err := mockRoutingService.ListAllLocations(ctx)
		if err != nil {
			logger.Error("Failed to list locations in mock mode", "error", err)
			log.Panic("Failed to list locations in mock mode:", err)
		}

		locationStrings := make([]string, len(locations))
		for i, loc := range locations {
			locationStrings[i] = loc.GetUnLocode().String()
		}

		scenarios := mockBookingService.GenerateCargoScenarios(locationStrings, 10) // Generate test cargo scenarios

		_, err = mockBookingService.PopulateTestCargo(ctx, scenarios)
		if err != nil {
			logger.Error("Failed to populate test cargo in mock mode", "error", err)
			log.Panic("Failed to populate test cargo in mock mode:", err)
		}

	} else {
		logger.Info("Running in live mode", "mode", cfg.Mode, "isMockMode", cfg.IsMockMode(), "isLiveMode", cfg.IsLiveMode())

		// Create Routing context application service
		routingService = routingapplication.NewRoutingApplicationService(
			voyageRepo,
			locationRepo,
			logger,
		)

		// Create adapter for Booking->Routing integration (synchronous, customer-supplier)
		routingAdapter := integration.NewRoutingServiceAdapter(routingService)

		// Create Booking context application service
		bookingService = bookingapplication.NewBookingApplicationService(
			cargoRepo,
			routingAdapter, // Synchronous integration with routing
			eventBus,       // Event publisher
			logger,
		)

		// Create Handling context application services
		handlingReportService = handlingapplication.NewHandlingReportService(
			handlingEventRepo,
			eventBus, // Event publisher for handling events
			logger,
		)

	}

	handlingQueryService := handlingapplication.NewHandlingEventQueryService(handlingEventRepo, logger)

	// Set up event-driven integration: Handling->Booking (asynchronous, ACL)
	handlingToBookingHandler := integration.NewHandlingToBookingEventHandler(bookingService, logger)

	// Subscribe to handling events
	eventBus.Subscribe(
		handlingdomain.HandlingEventRegisteredEvent{}.EventName(),
		handlingToBookingHandler.HandleCargoWasHandled,
	)

	// Wire up authentication middleware
	authMiddleware := httpmiddleware.NewAuthMiddleware(cfg.JWT.SecretKey, cfg.JWT.Issuer, cfg.JWT.Audience)

	// Create HTTP handler with all application services
	httpHandler := httpadapter.NewHandler(
		authMiddleware,
		bookingService,
		routingService,
		handlingReportService,
		handlingQueryService,
	)

	logger.Info("Application dependencies wired successfully",
		"cargo_subscribers", eventBus.GetSubscriberCount(handlingdomain.HandlingEventRegisteredEvent{}.EventName()),
		"mode", cfg.Mode,
	)

	return httpHandler
}
