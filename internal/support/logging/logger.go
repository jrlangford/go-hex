package logging

import (
	"go_hex/internal/support/config"
	"log/slog"
	"os"
)

// Logger is a global logger instance.
var Logger *slog.Logger

// Initialize sets up the global logger based on configuration.
func Initialize(cfg *config.Config) {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.IsProduction() {
		// JSON logging for production
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text logging for development
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// Get returns the current logger instance.
func Get() *slog.Logger {
	if Logger == nil {
		// Fallback to default logger if not initialized
		return slog.Default()
	}
	return Logger
}
