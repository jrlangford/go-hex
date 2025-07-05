package main

import (
	"go_hex/internal/adapters/driven/event_bus"
	"go_hex/internal/adapters/driven/in_memory_cargo_repo"
	"go_hex/internal/adapters/driven/in_memory_handling_repo"
	"go_hex/internal/adapters/driven/in_memory_location_repo"
	"go_hex/internal/adapters/driven/in_memory_voyage_repo"
	"go_hex/internal/adapters/driven/mock"
	httpadapter "go_hex/internal/adapters/driving/http"
	"go_hex/internal/adapters/driving/http/middleware"
	"go_hex/internal/adapters/integration"

	bookingApp "go_hex/internal/core/booking/application"
	bookingSecondary "go_hex/internal/core/booking/ports/secondary"
	handlingApp "go_hex/internal/core/handling/application"
	handlingDomain "go_hex/internal/core/handling/domain"
	handlingSecondary "go_hex/internal/core/handling/ports/secondary"
	routingApp "go_hex/internal/core/routing/application"
	routingSecondary "go_hex/internal/core/routing/ports/secondary"

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

	// Create repositories based on application mode
	var cargoRepo bookingSecondary.CargoRepository
	var voyageRepo routingSecondary.VoyageRepository
	var locationRepo routingSecondary.LocationRepository
	var handlingEventRepo handlingSecondary.HandlingEventRepository

	if cfg.IsMockMode() {
		logger.Info("Running in mock mode with pre-populated mock data", "mode", cfg.Mode, "isMockMode", cfg.IsMockMode())

		// Use mock repositories with generated test data
		mockManager := mock.NewMockDataManager(logger)
		cargoRepo = mockManager.CargoRepo
		voyageRepo = mockManager.VoyageRepo
		locationRepo = mockManager.LocationRepo
		handlingEventRepo = mockManager.HandlingRepo
	} else {
		logger.Info("Running in live mode with empty repositories", "mode", cfg.Mode, "isMockMode", cfg.IsMockMode(), "isLiveMode", cfg.IsLiveMode())

		// Use empty in-memory repositories for live mode
		cargoRepo = in_memory_cargo_repo.NewInMemoryCargoRepository()
		voyageRepo = in_memory_voyage_repo.NewInMemoryVoyageRepository()
		locationRepo = in_memory_location_repo.NewInMemoryLocationRepository()
		handlingEventRepo = in_memory_handling_repo.NewInMemoryHandlingEventRepository()
	}

	// Create Routing context application service
	routingService := routingApp.NewRoutingApplicationService(
		voyageRepo,
		locationRepo,
		logger,
	)

	// Create adapter for Booking->Routing integration (synchronous, customer-supplier)
	routingAdapter := integration.NewRoutingServiceAdapter(routingService)

	// Create Booking context application service
	bookingService := bookingApp.NewBookingApplicationService(
		cargoRepo,
		routingAdapter, // Synchronous integration with routing
		eventBus,       // Event publisher
		logger,
	)

	// Create Handling context application services
	handlingReportService := handlingApp.NewHandlingReportService(
		handlingEventRepo,
		eventBus, // Event publisher for handling events
	)
	handlingQueryService := handlingApp.NewHandlingEventQueryService(handlingEventRepo)

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
