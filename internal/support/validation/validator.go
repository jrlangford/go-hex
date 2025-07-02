package validation

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

var (
	validatorInstance *Validator
	validatorOnce     sync.Once
)

func GetValidator() *Validator {
	validatorOnce.Do(func() {
		validatorInstance = New()
	})
	return validatorInstance
}

func New() *Validator {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return fld.Name
		}
		return name
	})

	validate.RegisterValidation("friend_name", validateFriendName)
	validate.RegisterValidation("role", validateRole)
	validate.RegisterValidation("permission", validatePermission)
	validate.RegisterValidation("environment", validateEnvironment)
	validate.RegisterValidation("log_level", validateLogLevel)
	validate.RegisterValidation("phone_number", validatePhoneNumber)
	validate.RegisterValidation("postal_code", validatePostalCode)
	validate.RegisterValidation("currency", validateCurrency)

	return &Validator{
		validate: validate,
	}
}

func (v *Validator) Validate(s interface{}) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors []string
	for _, err := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, formatValidationError(err))
	}

	return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
}

func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", err.Field(), err.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	case "friend_name":
		return fmt.Sprintf("%s must be a valid friend name (2-100 characters, no special chars)", err.Field())
	case "role":
		return fmt.Sprintf("%s must be a valid role (admin, user, readonly)", err.Field())
	case "permission":
		return fmt.Sprintf("%s must be a valid permission", err.Field())
	case "environment":
		return fmt.Sprintf("%s must be a valid environment (development, staging, production)", err.Field())
	case "log_level":
		return fmt.Sprintf("%s must be a valid log level (debug, info, warn, error)", err.Field())
	case "phone_number":
		return fmt.Sprintf("%s must be a valid phone number (10-15 digits)", err.Field())
	case "postal_code":
		return fmt.Sprintf("%s must be a valid postal code (5-10 alphanumeric characters)", err.Field())
	case "currency":
		return fmt.Sprintf("%s must be a valid ISO 4217 currency code", err.Field())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}

// Custom validation functions

func validateFriendName(fl validator.FieldLevel) bool {
	name := strings.TrimSpace(fl.Field().String())
	if len(name) < 2 || len(name) > 100 {
		return false
	}

	// Check for disallowed characters (basic validation)
	for _, char := range name {
		if char < 32 || char == 127 { // Control characters
			return false
		}
	}

	return true
}

func validateRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := []string{"admin", "user", "readonly"}

	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

func validatePermission(fl validator.FieldLevel) bool {
	permission := fl.Field().String()
	validPermissions := []string{"add_friend", "view_friend", "update_friend", "delete_friend", "greet"}

	for _, validPermission := range validPermissions {
		if permission == validPermission {
			return true
		}
	}
	return false
}

func validateEnvironment(fl validator.FieldLevel) bool {
	env := fl.Field().String()
	validEnvs := []string{"development", "staging", "production"}

	for _, validEnv := range validEnvs {
		if env == validEnv {
			return true
		}
	}
	return false
}

func validateLogLevel(fl validator.FieldLevel) bool {
	level := fl.Field().String()
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
	phone := strings.TrimSpace(fl.Field().String())
	if phone == "" {
		return true // Optional field
	}

	// Basic phone number validation - digits, spaces, hyphens, parentheses, plus sign
	validChars := "0123456789 -()+"
	if len(phone) < 10 || len(phone) > 20 {
		return false
	}

	for _, char := range phone {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}

	// Count digits - should have at least 10 but no more than 15
	digitCount := 0
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}

	return digitCount >= 10 && digitCount <= 15
}

func validatePostalCode(fl validator.FieldLevel) bool {
	postalCode := strings.TrimSpace(fl.Field().String())
	if len(postalCode) < 5 || len(postalCode) > 10 {
		return false
	}

	// Allow alphanumeric characters, spaces, and hyphens
	for _, char := range postalCode {
		if !((char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			char == ' ' || char == '-') {
			return false
		}
	}

	return true
}

func validateCurrency(fl validator.FieldLevel) bool {
	currency := strings.ToUpper(strings.TrimSpace(fl.Field().String()))

	// Common currency codes - this could be expanded
	validCurrencies := []string{
		"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "CNY", "SEK", "NZD",
		"MXN", "SGD", "HKD", "NOK", "TRY", "ZAR", "BRL", "INR", "KRW", "PLN",
	}

	for _, validCurrency := range validCurrencies {
		if currency == validCurrency {
			return true
		}
	}

	return false
}

// Global convenience functions

func Validate(obj interface{}) error {
	return GetValidator().Validate(obj)
}

func ValidateVar(field interface{}, tag string) error {
	return GetValidator().ValidateVar(field, tag)
}
