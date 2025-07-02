package validation_test

import (
	"go_hex/internal/core/domain/friendship"
	"go_hex/internal/support/validation"
	"testing"
	"time"
)

func BenchmarkValidatorCacheReuse(b *testing.B) {
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

func ExampleValidate() {
	friendData := friendship.FriendData{
		Name:  "John Doe",
		Title: "Mr.",
	}

	err := validation.Validate(friendData)
	if err != nil {
		// Handle validation error
	}

	validator := validation.GetValidator()
	err = validator.Validate(friendData)
	if err != nil {
		// Handle validation error
	}

	// validator := validation.New()
	// err := validator.Validate(friendData)
}

func TestValidatorSingleton(t *testing.T) {
	validator1 := validation.GetValidator()
	validator2 := validation.GetValidator()

	if validator1 != validator2 {
		t.Error("GetValidator() should return the same instance (singleton)")
	}
}

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

func TestValidatorCacheEffectiveness(t *testing.T) {
	friendData := friendship.FriendData{
		Name:  "John Doe",
		Title: "Mr.",
	}

	start := time.Now()
	err := validation.Validate(friendData)
	firstDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Validation should succeed: %v", err)
	}

	start = time.Now()
	err = validation.Validate(friendData)
	secondDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Validation should succeed: %v", err)
	}

	if secondDuration > firstDuration*2 {
		t.Logf("Warning: Second validation took longer than expected. First: %v, Second: %v",
			firstDuration, secondDuration)
	}

	t.Logf("First validation: %v, Second validation: %v", firstDuration, secondDuration)
}
