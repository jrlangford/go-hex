package server

import (
	"context"
	"fmt"
	httpadapter "go_hex/internal/adapters/driving/http"
	"go_hex/internal/support/config"
	"go_hex/internal/support/logging"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HTTPServer manages the HTTP server lifecycle.
type HTTPServer struct {
	server  *http.Server
	config  *config.Config
	handler *httpadapter.Handler
}

// New creates a new HTTP server instance.
func New(cfg *config.Config, handler *httpadapter.Handler) *HTTPServer {
	mux := http.NewServeMux()
	httpadapter.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return &HTTPServer{
		server:  server,
		config:  cfg,
		handler: handler,
	}
}

// Start starts the HTTP server and handles graceful shutdown.
func (s *HTTPServer) Start() error {
	logger := logging.Get()

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited")
	return nil
}
