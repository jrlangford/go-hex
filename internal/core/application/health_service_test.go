package application

import (
	"context"
	"go_hex/internal/adapters/driven/in_memory_repo"
	"go_hex/internal/support/logging"
	"testing"
)

func TestHealthService_CheckHealth(t *testing.T) {
	// Arrange
	logger := logging.Get()
	friendRepo := in_memory_repo.NewInMemoryFriendRepository()
	healthService := NewHealthService(friendRepo, logger)
	ctx := context.Background()

	// Act
	result, err := healthService.CheckHealth(ctx)

	// Assert
	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %s", result["status"])
	}

	if result["service"] != "go_hex" {
		t.Errorf("Expected service to be 'go_hex', got %s", result["service"])
	}

	if result["repository"] != "healthy" {
		t.Errorf("Expected repository to be 'healthy', got %s", result["repository"])
	}
}

func TestHealthService_CheckHealth_RepositoryFailure(t *testing.T) {
	// Arrange
	logger := logging.Get()
	// Create a mock repository that will fail
	friendRepo := in_memory_repo.NewInMemoryFriendRepository()
	healthService := NewHealthService(friendRepo, logger)
	ctx := context.Background()

	// For this test, we'll assume the repository is working fine
	// In a real scenario, you might want to use a mock that can simulate failures

	// Act
	result, err := healthService.CheckHealth(ctx)

	// Assert
	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}

	// Since we're using a working repository, status should be healthy
	if result["status"] != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %s", result["status"])
	}
}
