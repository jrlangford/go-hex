package server

import (
	"context"
	"fmt"
	"go_hex/internal/adapters/driving/httpadapter"
	"go_hex/internal/support/config"
	"go_hex/internal/support/logging"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HTTPServer struct {
	server  *http.Server
	config  *config.Config
	handler *httpadapter.Handler
}

func New(cfg *config.Config, handler *httpadapter.Handler, _ interface{}) *HTTPServer {
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

func (s *HTTPServer) Start() error {
	logger := logging.Get()

	go func() {
		logger.Info("Starting HTTP server", "port", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited")
	return nil
}
