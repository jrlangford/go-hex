package main

import (
	"go_hex/internal/adapters/driven/in_memory_repo"
	"go_hex/internal/adapters/driven/stdout_event_publisher"
	httpadapter "go_hex/internal/adapters/driving/http"
	"go_hex/internal/adapters/driving/http/middleware"
	"go_hex/internal/core/application"
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

	httpServer := server.New(cfg, httpHandler)
	if err := httpServer.Start(); err != nil {
		logger.Error("Server startup failed", "error", err)
	}

	log.Println("Server exited")
}

func wireAppDependencies(cfg *config.Config, logger *slog.Logger) *httpadapter.Handler {
	friendRepo := in_memory_repo.NewInMemoryFriendRepository()
	eventPublisher := stdout_event_publisher.NewStdoutEventPublisher()

	greeterService := application.NewHelloService(friendRepo, eventPublisher, logger)
	healthService := application.NewHealthService(friendRepo, logger)

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.SecretKey, cfg.JWT.Issuer, cfg.JWT.Audience)

	httpHandler := httpadapter.NewHandler(greeterService, healthService, authMiddleware)

	return httpHandler
}
