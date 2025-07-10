package validation

import (
	"testing"
)

func TestValidate_BasicStructure(t *testing.T) {
	t.Run("Valid struct with required field", func(t *testing.T) {
		testStruct := struct {
			Name string `validate:"required"`
		}{
			Name: "Test Value",
		}

		if err := Validate(testStruct); err != nil {
			t.Errorf("Expected no error for valid struct, got: %v", err)
		}
	})

	t.Run("Invalid struct - missing required field", func(t *testing.T) {
		testStruct := struct {
			Name string `validate:"required"`
		}{
			Name: "",
		}

		if err := Validate(testStruct); err == nil {
			t.Error("Expected error for missing required field, got nil")
		}
	})
}

func TestValidate_Email(t *testing.T) {
	t.Run("Valid Email", func(t *testing.T) {
		testStruct := struct {
			Email string `validate:"email"`
		}{
			Email: "test@example.com",
		}

		if err := Validate(testStruct); err != nil {
			t.Errorf("Expected no error for valid email, got: %v", err)
		}
	})

	t.Run("Invalid Email", func(t *testing.T) {
		testStruct := struct {
			Email string `validate:"email"`
		}{
			Email: "invalid-email",
		}

		if err := Validate(testStruct); err == nil {
			t.Error("Expected error for invalid email, got nil")
		}
	})
}

func TestValidate_MinLength(t *testing.T) {
	t.Run("Valid min length", func(t *testing.T) {
		testStruct := struct {
			Name string `validate:"min=3"`
		}{
			Name: "John",
		}

		if err := Validate(testStruct); err != nil {
			t.Errorf("Expected no error for valid min length, got: %v", err)
		}
	})

	t.Run("Invalid min length", func(t *testing.T) {
		testStruct := struct {
			Name string `validate:"min=3"`
		}{
			Name: "Jo",
		}

		if err := Validate(testStruct); err == nil {
			t.Error("Expected error for invalid min length, got nil")
		}
	})
}
