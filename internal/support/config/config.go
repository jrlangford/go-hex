package config

import (
	"fmt"
	"go_hex/internal/support/validation"
	"os"
	"strconv"
)

// Config holds application configuration.
type Config struct {
	Port        int       `json:"port" validate:"required,min=1,max=65535"`
	Environment string    `json:"environment" validate:"required,environment"`
	LogLevel    string    `json:"log_level" validate:"required,log_level"`
	JWT         JWTConfig `json:"jwt"`
}

// JWTConfig holds JWT-specific configuration.
type JWTConfig struct {
	SecretKey string `json:"secret_key" validate:"required,min=32"`
	Issuer    string `json:"issuer" validate:"required"`
	Audience  string `json:"audience" validate:"required"`
}

// New creates configuration from environment variables with validation.
func New() (*Config, error) {
	config := &Config{
		Port:        8080,
		Environment: "development",
		LogLevel:    "info",
		JWT: JWTConfig{
			SecretKey: "your-secret-key-7890123456789012", // Default for development
			Issuer:    "go-hex-service",
			Audience:  "go-hex-api",
		},
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err != nil {
			return nil, fmt.Errorf("invalid PORT value: %w", err)
		} else {
			config.Port = p
		}
	}

	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// JWT configuration from environment variables
	if jwtSecret := os.Getenv("JWT_SECRET_KEY"); jwtSecret != "" {
		config.JWT.SecretKey = jwtSecret
	}

	if jwtIssuer := os.Getenv("JWT_ISSUER"); jwtIssuer != "" {
		config.JWT.Issuer = jwtIssuer
	}

	if jwtAudience := os.Getenv("JWT_AUDIENCE"); jwtAudience != "" {
		config.JWT.Audience = jwtAudience
	}

	// Annotation-based validation handles all validation rules
	if err := validation.Validate(config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
