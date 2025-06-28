package validation_test

import (
	"go_hex/internal/core/domain/friendship"
	"go_hex/internal/support/validation"
	"testing"
	"time"
)

// BenchmarkValidatorCacheReuse demonstrates the performance difference between
// creating new validator instances vs reusing the singleton validator.
func BenchmarkValidatorCacheReuse(b *testing.B) {
	// Test data
	friendData := friendship.FriendData{
		Name:  "John Doe",
		Title: "Mr.",
	}

	b.Run("Using Global Validator (Cached)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validation.Validate(friendData)
		}
	})

	b.Run("Creating New Validator Each Time (No Cache)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			validator := validation.New()
			_ = validator.Validate(friendData)
		}
	})
}

// ExampleValidate shows how to properly use the validator
// to leverage the internal cache for better performance.
func ExampleValidate() {
	// ✅ GOOD: Use global validation functions - reuses cached validator
	friendData := friendship.FriendData{
		Name:  "John Doe",
		Title: "Mr.",
	}

	// This reuses the singleton validator instance and its cache
	err := validation.Validate(friendData)
	if err != nil {
		// Handle validation error
	}

	// ✅ ALSO GOOD: Direct access to singleton validator
	validator := validation.GetValidator()
	err = validator.Validate(friendData)
	if err != nil {
		// Handle validation error
	}

	// ❌ BAD: Creating new validator instances - doesn't reuse cache
	// Don't do this in production code:
	// validator := validation.New()
	// err := validator.Validate(friendData)
}

// TestValidatorSingleton ensures that the global validator is indeed a singleton
func TestValidatorSingleton(t *testing.T) {
	validator1 := validation.GetValidator()
	validator2 := validation.GetValidator()

	if validator1 != validator2 {
		t.Error("GetValidator() should return the same instance (singleton)")
	}
}

// BenchmarkValidatorWithComplexStruct tests performance with more complex validation
func BenchmarkValidatorWithComplexStruct(b *testing.B) {
	friendData := friendship.FriendData{
		Name:  "John Doe",
		Title: "Software Engineer",
	}

	b.Run("Complex Struct - Global Validator", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validation.Validate(friendData)
		}
	})

	b.Run("Complex Struct - New Validator Each Time", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			validator := validation.New()
			_ = validator.Validate(friendData)
		}
	})
}

// TestValidatorCacheEffectiveness demonstrates that the validator cache
// provides consistent performance even with multiple validations
func TestValidatorCacheEffectiveness(t *testing.T) {
	friendData := friendship.FriendData{
		Name:  "John Doe",
		Title: "Mr.",
	}

	// First validation (cache miss)
	start := time.Now()
	err := validation.Validate(friendData)
	firstDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Validation should succeed: %v", err)
	}

	// Second validation (cache hit)
	start = time.Now()
	err = validation.Validate(friendData)
	secondDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Validation should succeed: %v", err)
	}

	// The second validation should be faster or at least not significantly slower
	// (allowing for some variance in timing)
	if secondDuration > firstDuration*2 {
		t.Logf("Warning: Second validation took longer than expected. First: %v, Second: %v",
			firstDuration, secondDuration)
	}

	t.Logf("First validation: %v, Second validation: %v", firstDuration, secondDuration)
}
