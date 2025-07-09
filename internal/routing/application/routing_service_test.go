package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go_hex/internal/core/routing/domain"
	"go_hex/internal/support/auth"
	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations
type MockVoyageRepository struct {
	mock.Mock
}

func (m *MockVoyageRepository) Store(voyage domain.Voyage) error {
	args := m.Called(voyage)
	return args.Error(0)
}

func (m *MockVoyageRepository) FindAll() ([]domain.Voyage, error) {
	args := m.Called()
	return args.Get(0).([]domain.Voyage), args.Error(1)
}

func (m *MockVoyageRepository) FindByNumber(number domain.VoyageNumber) (domain.Voyage, error) {
	args := m.Called(number)
	return args.Get(0).(domain.Voyage), args.Error(1)
}

func (m *MockVoyageRepository) FindByVoyageNumber(voyageNumber domain.VoyageNumber) (domain.Voyage, error) {
	args := m.Called(voyageNumber)
	return args.Get(0).(domain.Voyage), args.Error(1)
}

func (m *MockVoyageRepository) FindVoyagesConnecting(origin, destination domain.UnLocode) ([]domain.Voyage, error) {
	args := m.Called(origin, destination)
	return args.Get(0).([]domain.Voyage), args.Error(1)
}

type MockLocationRepository struct {
	mock.Mock
}

func (m *MockLocationRepository) Store(location domain.Location) error {
	args := m.Called(location)
	return args.Error(0)
}

func (m *MockLocationRepository) FindAll() ([]domain.Location, error) {
	args := m.Called()
	return args.Get(0).([]domain.Location), args.Error(1)
}

func (m *MockLocationRepository) FindByUnLocode(code domain.UnLocode) (domain.Location, error) {
	args := m.Called(code)
	return args.Get(0).(domain.Location), args.Error(1)
}

func TestRoutingApplicationService_FindOptimalItineraries(t *testing.T) {
	setup := func() (*RoutingApplicationService, *MockVoyageRepository, *MockLocationRepository) {
		voyageRepo := &MockVoyageRepository{}
		locationRepo := &MockLocationRepository{}
		logger := slog.Default()

		service := NewRoutingApplicationService(voyageRepo, locationRepo, logger)

		return service, voyageRepo, locationRepo
	}

	t.Run("should find optimal itineraries successfully", func(t *testing.T) {
		service, voyageRepo, _ := setup()

		// Create test voyages
		voyages := createTestVoyages(t)

		// Setup mocks
		voyageRepo.On("FindAll").Return(voyages, nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Create route specification
		routeSpec := domain.RouteSpecification{
			Origin:          "USNYC",
			Destination:     "DEHAM",
			ArrivalDeadline: time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		}

		// Execute
		itineraries, err := service.FindOptimalItineraries(ctx, routeSpec)

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, itineraries)
		voyageRepo.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _, _ := setup()

		// Create context without proper claims
		ctx := context.Background()

		routeSpec := domain.RouteSpecification{
			Origin:          "USNYC",
			Destination:     "DEHAM",
			ArrivalDeadline: time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		}

		// Execute
		_, err := service.FindOptimalItineraries(ctx, routeSpec)

		// Verify
		assert.Error(t, err)
	})

	t.Run("should fail with invalid arrival deadline format", func(t *testing.T) {
		service, _, _ := setup()

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		routeSpec := domain.RouteSpecification{
			Origin:          "USNYC",
			Destination:     "DEHAM",
			ArrivalDeadline: "invalid-date-format",
		}

		// Execute
		_, err := service.FindOptimalItineraries(ctx, routeSpec)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arrival deadline format")
	})

	t.Run("should fail with invalid origin UN/LOCODE", func(t *testing.T) {
		service, _, _ := setup()

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		routeSpec := domain.RouteSpecification{
			Origin:          "XX", // Invalid - too short
			Destination:     "DEHAM",
			ArrivalDeadline: time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		}

		// Execute
		_, err := service.FindOptimalItineraries(ctx, routeSpec)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UN/LOCODE format")
	})

	t.Run("should fail with invalid destination UN/LOCODE", func(t *testing.T) {
		service, _, _ := setup()

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		routeSpec := domain.RouteSpecification{
			Origin:          "USNYC",
			Destination:     "XX", // Invalid - too short
			ArrivalDeadline: time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		}

		// Execute
		_, err := service.FindOptimalItineraries(ctx, routeSpec)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UN/LOCODE format")
	})

	t.Run("should fail when voyage repository fails", func(t *testing.T) {
		service, voyageRepo, _ := setup()

		// Setup mocks
		voyageRepo.On("FindAll").Return([]domain.Voyage{}, errors.New("repository error"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		routeSpec := domain.RouteSpecification{
			Origin:          "USNYC",
			Destination:     "DEHAM",
			ArrivalDeadline: time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		}

		// Execute
		_, err := service.FindOptimalItineraries(ctx, routeSpec)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve voyages")
		voyageRepo.AssertExpectations(t)
	})

	t.Run("should return empty list when no suitable routes found", func(t *testing.T) {
		service, voyageRepo, _ := setup()

		// Setup mocks with empty voyage list
		voyageRepo.On("FindAll").Return([]domain.Voyage{}, nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		routeSpec := domain.RouteSpecification{
			Origin:          "USNYC",
			Destination:     "DEHAM",
			ArrivalDeadline: time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		}

		// Execute
		itineraries, err := service.FindOptimalItineraries(ctx, routeSpec)

		// Verify
		require.NoError(t, err)
		assert.Empty(t, itineraries)
		voyageRepo.AssertExpectations(t)
	})
}

// Helper functions

func createContextWithClaims(t *testing.T, permissions []string) context.Context {
	// Create claims with admin role to ensure all permissions are available
	claims, err := auth.NewClaims(
		"test-user",
		"testuser",
		"test@example.com",
		[]string{"admin"}, // admin role has all permissions
		map[string]string{"test": "true"},
	)
	require.NoError(t, err)

	return context.WithValue(context.Background(), auth.ClaimsContextKey, claims)
}

func createTestVoyages(t *testing.T) []domain.Voyage {
	// Create test UN/LOCODEs
	usnyc, err := domain.NewUnLocode("USNYC")
	require.NoError(t, err)
	deham, err := domain.NewUnLocode("DEHAM")
	require.NoError(t, err)
	segot, err := domain.NewUnLocode("SEGOT")
	require.NoError(t, err)

	baseTime := time.Now().Add(time.Hour)

	// Create first voyage: USNYC -> DEHAM
	movement1, err := domain.NewCarrierMovement(
		usnyc, deham,
		baseTime,
		baseTime.Add(24*time.Hour),
	)
	require.NoError(t, err)

	voyage1, err := domain.NewVoyage([]domain.CarrierMovement{movement1})
	require.NoError(t, err)

	// Create second voyage: DEHAM -> SEGOT
	movement2, err := domain.NewCarrierMovement(
		deham, segot,
		baseTime.Add(26*time.Hour), // Allow time for transshipment
		baseTime.Add(48*time.Hour),
	)
	require.NoError(t, err)

	voyage2, err := domain.NewVoyage([]domain.CarrierMovement{movement2})
	require.NoError(t, err)

	return []domain.Voyage{voyage1, voyage2}
}
