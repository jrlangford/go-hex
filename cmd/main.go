package main

import (
	"go_hex/internal/adapters/driven/event_bus"
	"go_hex/internal/adapters/driven/in_memory_cargo_repo"
	"go_hex/internal/adapters/driven/in_memory_handling_repo"
	"go_hex/internal/adapters/driven/in_memory_location_repo"
	"go_hex/internal/adapters/driven/in_memory_voyage_repo"
	httpadapter "go_hex/internal/adapters/driving/http"
	"go_hex/internal/adapters/driving/http/middleware"
	"go_hex/internal/adapters/integration"

	bookingApp "go_hex/internal/booking/application"
	bookingPorts "go_hex/internal/booking/ports/primary"
	handlingApp "go_hex/internal/handling/application"
	handlingDomain "go_hex/internal/handling/domain"
	handlingPorts "go_hex/internal/handling/ports/primary"
	routingApp "go_hex/internal/routing/application"
	routingPorts "go_hex/internal/routing/ports/primary"

	mockBookingApp "go_hex/internal/booking/mock"
	mockHandlingApp "go_hex/internal/handling/mock"
	mockRoutingApp "go_hex/internal/routing/mock"

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

	var bookingService bookingPorts.BookingService
	var handlingReportService handlingPorts.HandlingReportService
	var routingService routingPorts.RouteFinder

	if cfg.IsMockMode() {
		logger.Info("Running in mock mode with pre-populated mock data", "mode", cfg.Mode, "isMockMode", cfg.IsMockMode())

		// Create Routing context application service
		routingService = mockRoutingApp.NewMockRoutingApplication(
			voyageRepo,
			locationRepo,
			logger,
			1017, // Use seed from config for reproducibility
		)

		// Create adapter for Booking->Routing integration (synchronous, customer-supplier)
		routingAdapter := integration.NewRoutingServiceAdapter(routingService)

		// Create Mock Booking context application service
		bookingService = mockBookingApp.NewMockBookingApplication(
			cargoRepo,
			routingAdapter, // Synchronous integration with routing
			eventBus,       // Event publisher
			logger,
			1017, // Use seed from config for reproducibility
		)

		handlingReportService = mockHandlingApp.NewMockHandlingApplication(
			handlingEventRepo,
			eventBus, // Event publisher for handling events
			logger,
			1017, // Use seed from config for reproducibility
		)

	} else {
		logger.Info("Running in live mode", "mode", cfg.Mode, "isMockMode", cfg.IsMockMode(), "isLiveMode", cfg.IsLiveMode())

		// Create Routing context application service
		routingService = routingApp.NewRoutingApplicationService(
			voyageRepo,
			locationRepo,
			logger,
		)

		// Create adapter for Booking->Routing integration (synchronous, customer-supplier)
		routingAdapter := integration.NewRoutingServiceAdapter(routingService)

		// Create Booking context application service
		bookingService = bookingApp.NewBookingApplicationService(
			cargoRepo,
			routingAdapter, // Synchronous integration with routing
			eventBus,       // Event publisher
			logger,
		)

		// Create Handling context application services
		handlingReportService = handlingApp.NewHandlingReportService(
			handlingEventRepo,
			eventBus, // Event publisher for handling events
			logger,
		)

	}

	handlingQueryService := handlingApp.NewHandlingEventQueryService(handlingEventRepo, logger)

	// Set up event-driven integration: Handling->Booking (asynchronous, ACL)
	handlingToBookingHandler := integration.NewHandlingToBookingEventHandler(bookingService, logger)

	// Subscribe to handling events
	eventBus.Subscribe(
		handlingDomain.HandlingEventRegisteredEvent{}.EventName(),
		handlingToBookingHandler.HandleCargoWasHandled,
	)

	// Wire up authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.SecretKey, cfg.JWT.Issuer, cfg.JWT.Audience)

	// Create HTTP handler with all application services
	httpHandler := httpadapter.NewHandler(
		authMiddleware,
		bookingService,
		routingService,
		handlingReportService,
		handlingQueryService,
	)

	logger.Info("Application dependencies wired successfully",
		"cargo_subscribers", eventBus.GetSubscriberCount(handlingDomain.HandlingEventRegisteredEvent{}.EventName()),
		"mode", cfg.Mode,
	)

	return httpHandler
}
