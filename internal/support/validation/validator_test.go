package validation_test

import (
	"go_hex/internal/core/domain/authorization"
	"go_hex/internal/core/domain/friendship"
	"go_hex/internal/support/auth"
	"go_hex/internal/support/config"
	"go_hex/internal/support/validation"
	"testing"
)

func TestValidation_FriendData(t *testing.T) {
	t.Run("Valid FriendData", func(t *testing.T) {
		friendData := friendship.FriendData{
			Name:  "John Doe",
			Title: "Mr.",
		}

		if err := validation.Validate(friendData); err != nil {
			t.Errorf("Expected no error for valid friend data, got: %v", err)
		}
	})

	t.Run("Invalid FriendData - Empty Name", func(t *testing.T) {
		friendData := friendship.FriendData{
			Name:  "",
			Title: "Mr.",
		}

		if err := validation.Validate(friendData); err == nil {
			t.Error("Expected error for empty name, got nil")
		}
	})

	t.Run("Invalid FriendData - Short Name", func(t *testing.T) {
		friendData := friendship.FriendData{
			Name:  "A",
			Title: "Mr.",
		}

		if err := validation.Validate(friendData); err == nil {
			t.Error("Expected error for short name, got nil")
		}
	})

	t.Run("Invalid FriendData - Long Title", func(t *testing.T) {
		friendData := friendship.FriendData{
			Name:  "John Doe",
			Title: "This is an extremely long title that exceeds the maximum allowed length of 50 characters",
		}

		if err := validation.Validate(friendData); err == nil {
			t.Error("Expected error for long title, got nil")
		}
	})
}

func TestValidation_TokenClaims(t *testing.T) {
	t.Run("Valid TokenClaims", func(t *testing.T) {
		claims, err := auth.NewClaims(
			"user-123",
			"john.doe",
			"john@example.com",
			[]string{"user", "admin"},
			map[string]string{"dept": "engineering"},
		)
		if err != nil {
			t.Fatalf("Failed to create claims: %v", err)
		}

		if err := validation.Validate(claims); err != nil {
			t.Errorf("Expected no error for valid token claims, got: %v", err)
		}
	})

	t.Run("Invalid TokenClaims - Empty UserID", func(t *testing.T) {
		_, err := auth.NewClaims(
			"", // Empty UserID should fail
			"john.doe",
			"john@example.com",
			[]string{"user"},
			nil,
		)

		if err == nil {
			t.Error("Expected error for empty UserID, got nil")
		}
	})

	t.Run("Invalid TokenClaims - Invalid Email", func(t *testing.T) {
		_, err := auth.NewClaims(
			"user-123",
			"john.doe",
			"invalid-email", // Invalid email should fail
			[]string{"user"},
			nil,
		)

		if err == nil {
			t.Error("Expected error for invalid email, got nil")
		}
	})

	t.Run("Invalid TokenClaims - Invalid Role", func(t *testing.T) {
		_, err := auth.NewClaims(
			"user-123",
			"john.doe",
			"john@example.com",
			[]string{"invalid-role"}, // Invalid role should fail
			nil,
		)

		if err == nil {
			t.Error("Expected error for invalid role, got nil")
		}
	})
}

func TestValidation_Config(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		cfg := config.Config{
			Port:        8080,
			Environment: "development",
			LogLevel:    "info",
			JWT: config.JWTConfig{
				SecretKey: "this-is-a-very-long-secret-key-that-meets-minimum-length-requirements",
				Issuer:    "test-issuer",
				Audience:  "test-audience",
			},
		}

		if err := validation.Validate(cfg); err != nil {
			t.Errorf("Expected no error for valid config, got: %v", err)
		}
	})

	t.Run("Invalid Config - Invalid Port", func(t *testing.T) {
		cfg := config.Config{
			Port:        0,
			Environment: "development",
			LogLevel:    "info",
			JWT: config.JWTConfig{
				SecretKey: "this-is-a-very-long-secret-key-that-meets-minimum-length-requirements",
				Issuer:    "test-issuer",
				Audience:  "test-audience",
			},
		}

		if err := validation.Validate(cfg); err == nil {
			t.Error("Expected error for invalid port, got nil")
		}
	})

	t.Run("Invalid Config - Invalid Environment", func(t *testing.T) {
		cfg := config.Config{
			Port:        8080,
			Environment: "invalid-env",
			LogLevel:    "info",
			JWT: config.JWTConfig{
				SecretKey: "this-is-a-very-long-secret-key-that-meets-minimum-length-requirements",
				Issuer:    "test-issuer",
				Audience:  "test-audience",
			},
		}

		if err := validation.Validate(cfg); err == nil {
			t.Error("Expected error for invalid environment, got nil")
		}
	})

	t.Run("Invalid Config - Invalid LogLevel", func(t *testing.T) {
		cfg := config.Config{
			Port:        8080,
			Environment: "development",
			LogLevel:    "invalid-level",
			JWT: config.JWTConfig{
				SecretKey: "this-is-a-very-long-secret-key-that-meets-minimum-length-requirements",
				Issuer:    "test-issuer",
				Audience:  "test-audience",
			},
		}

		if err := validation.Validate(cfg); err == nil {
			t.Error("Expected error for invalid log level, got nil")
		}
	})
}

func TestValidation_AuthorizationContext(t *testing.T) {
	t.Run("Valid AuthorizationContext", func(t *testing.T) {
		authCtx := authorization.AuthorizationContext{
			UserID:   "user-123",
			Username: "john.doe",
			Roles:    []authorization.Role{authorization.RoleUser, authorization.RoleAdmin},
		}

		if err := validation.Validate(authCtx); err != nil {
			t.Errorf("Expected no error for valid authorization context, got: %v", err)
		}
	})

	t.Run("Invalid AuthorizationContext - Empty UserID", func(t *testing.T) {
		authCtx := authorization.AuthorizationContext{
			UserID:   "",
			Username: "john.doe",
			Roles:    []authorization.Role{authorization.RoleUser},
		}

		if err := validation.Validate(authCtx); err == nil {
			t.Error("Expected error for empty UserID, got nil")
		}
	})

	t.Run("Invalid AuthorizationContext - Short Username", func(t *testing.T) {
		authCtx := authorization.AuthorizationContext{
			UserID:   "user-123",
			Username: "a",
			Roles:    []authorization.Role{authorization.RoleUser},
		}

		if err := validation.Validate(authCtx); err == nil {
			t.Error("Expected error for short username, got nil")
		}
	})

	t.Run("Invalid AuthorizationContext - No Roles", func(t *testing.T) {
		authCtx := authorization.AuthorizationContext{
			UserID:   "user-123",
			Username: "john.doe",
			Roles:    []authorization.Role{},
		}

		if err := validation.Validate(authCtx); err == nil {
			t.Error("Expected error for empty roles, got nil")
		}
	})
}

func TestValidation_CustomValidators(t *testing.T) {
	t.Run("friend_name validator", func(t *testing.T) {
		// Valid names
		validNames := []string{"John", "Alice Smith", "José García"}
		for _, name := range validNames {
			if err := validation.ValidateVar(name, "friend_name"); err != nil {
				t.Errorf("Expected valid name '%s' to pass, got error: %v", name, err)
			}
		}

		// Invalid names
		invalidNames := []string{"", "A", string(rune(0x1F))} // empty, too short, control character
		for _, name := range invalidNames {
			if err := validation.ValidateVar(name, "friend_name"); err == nil {
				t.Errorf("Expected invalid name '%s' to fail validation", name)
			}
		}
	})

	t.Run("role validator", func(t *testing.T) {
		// Valid roles
		validRoles := []string{"admin", "user", "readonly"}
		for _, role := range validRoles {
			if err := validation.ValidateVar(role, "role"); err != nil {
				t.Errorf("Expected valid role '%s' to pass, got error: %v", role, err)
			}
		}

		// Invalid roles
		invalidRoles := []string{"", "invalid", "superadmin"}
		for _, role := range invalidRoles {
			if err := validation.ValidateVar(role, "role"); err == nil {
				t.Errorf("Expected invalid role '%s' to fail validation", role)
			}
		}
	})

	t.Run("environment validator", func(t *testing.T) {
		// Valid environments
		validEnvs := []string{"development", "staging", "production"}
		for _, env := range validEnvs {
			if err := validation.ValidateVar(env, "environment"); err != nil {
				t.Errorf("Expected valid environment '%s' to pass, got error: %v", env, err)
			}
		}

		// Invalid environments
		invalidEnvs := []string{"", "test", "local"}
		for _, env := range invalidEnvs {
			if err := validation.ValidateVar(env, "environment"); err == nil {
				t.Errorf("Expected invalid environment '%s' to fail validation", env)
			}
		}
	})

	t.Run("log_level validator", func(t *testing.T) {
		// Valid log levels
		validLevels := []string{"debug", "info", "warn", "error"}
		for _, level := range validLevels {
			if err := validation.ValidateVar(level, "log_level"); err != nil {
				t.Errorf("Expected valid log level '%s' to pass, got error: %v", level, err)
			}
		}

		// Invalid log levels
		invalidLevels := []string{"", "trace", "fatal"}
		for _, level := range invalidLevels {
			if err := validation.ValidateVar(level, "log_level"); err == nil {
				t.Errorf("Expected invalid log level '%s' to fail validation", level)
			}
		}
	})

	t.Run("phone_number validator", func(t *testing.T) {
		// Valid phone numbers
		validPhones := []string{"+1 555-123-4567", "555-123-4567", "(555) 123-4567"}
		for _, phone := range validPhones {
			if err := validation.ValidateVar(phone, "phone_number"); err != nil {
				t.Errorf("Expected valid phone '%s' to pass, got error: %v", phone, err)
			}
		}

		// Invalid phone numbers
		invalidPhones := []string{"123", "abc-def-ghij", "12345678901234567890"}
		for _, phone := range invalidPhones {
			if err := validation.ValidateVar(phone, "phone_number"); err == nil {
				t.Errorf("Expected invalid phone '%s' to fail validation", phone)
			}
		}
	})

	t.Run("postal_code validator", func(t *testing.T) {
		// Valid postal codes
		validCodes := []string{"12345", "K1A 0A6", "SW1A 1AA", "12345-6789"}
		for _, code := range validCodes {
			if err := validation.ValidateVar(code, "postal_code"); err != nil {
				t.Errorf("Expected valid postal code '%s' to pass, got error: %v", code, err)
			}
		}

		// Invalid postal codes
		invalidCodes := []string{"", "123", "12345678901", "ABC@123"}
		for _, code := range invalidCodes {
			if err := validation.ValidateVar(code, "postal_code"); err == nil {
				t.Errorf("Expected invalid postal code '%s' to fail validation", code)
			}
		}
	})

	t.Run("currency validator", func(t *testing.T) {
		// Valid currencies
		validCurrencies := []string{"USD", "EUR", "GBP", "JPY", "CAD"}
		for _, currency := range validCurrencies {
			if err := validation.ValidateVar(currency, "currency"); err != nil {
				t.Errorf("Expected valid currency '%s' to pass, got error: %v", currency, err)
			}
		}

		// Invalid currencies
		invalidCurrencies := []string{"", "XX", "DOLLAR", "US"}
		for _, currency := range invalidCurrencies {
			if err := validation.ValidateVar(currency, "currency"); err == nil {
				t.Errorf("Expected invalid currency '%s' to fail validation", currency)
			}
		}
	})
}
