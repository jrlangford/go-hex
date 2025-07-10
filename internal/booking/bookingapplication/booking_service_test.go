package bookingapplication

import (
	"context"
	"errors"
	"testing"
	"time"

	"go_hex/internal/booking/bookingdomain"
	"go_hex/internal/support/auth"
	"go_hex/internal/support/basedomain"
	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations
type MockCargoRepository struct {
	mock.Mock
}

func (m *MockCargoRepository) Store(cargo bookingdomain.Cargo) error {
	args := m.Called(cargo)
	return args.Error(0)
}

func (m *MockCargoRepository) FindByTrackingId(id bookingdomain.TrackingId) (bookingdomain.Cargo, error) {
	args := m.Called(id)
	return args.Get(0).(bookingdomain.Cargo), args.Error(1)
}

func (m *MockCargoRepository) FindAll() ([]bookingdomain.Cargo, error) {
	args := m.Called()
	return args.Get(0).([]bookingdomain.Cargo), args.Error(1)
}

func (m *MockCargoRepository) FindUnrouted() ([]bookingdomain.Cargo, error) {
	args := m.Called()
	return args.Get(0).([]bookingdomain.Cargo), args.Error(1)
}

func (m *MockCargoRepository) Update(cargo bookingdomain.Cargo) error {
	args := m.Called(cargo)
	return args.Error(0)
}

type MockRoutingService struct {
	mock.Mock
}

func (m *MockRoutingService) FindOptimalItineraries(ctx context.Context, routeSpec bookingdomain.RouteSpecification) ([]bookingdomain.Itinerary, error) {
	args := m.Called(ctx, routeSpec)
	return args.Get(0).([]bookingdomain.Itinerary), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(event basedomain.DomainEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestBookingApplicationService_BookNewCargo(t *testing.T) {
	setup := func() (*BookingApplicationService, *MockCargoRepository, *MockRoutingService, *MockEventPublisher) {
		cargoRepo := &MockCargoRepository{}
		routingService := &MockRoutingService{}
		eventPublisher := &MockEventPublisher{}
		logger := slog.Default()

		service := NewBookingApplicationService(cargoRepo, routingService, eventPublisher, logger)

		return service, cargoRepo, routingService, eventPublisher
	}

	t.Run("should book new cargo successfully", func(t *testing.T) {
		service, cargoRepo, _, eventPublisher := setup()

		// Setup mocks
		cargoRepo.On("Store", mock.AnythingOfType("bookingdomain.Cargo")).Return(nil)
		eventPublisher.On("Publish", mock.Anything).Return(nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{}) // admin role has all permissions

		// Execute
		futureDate := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339) // 30 days from now
		cargo, err := service.BookNewCargo(ctx, "USNYC", "DEHAM", futureDate)

		// Verify
		require.NoError(t, err)
		assert.Equal(t, "USNYC", cargo.GetRouteSpecification().Origin)
		assert.Equal(t, "DEHAM", cargo.GetRouteSpecification().Destination)
		assert.False(t, cargo.IsRouted())
		cargoRepo.AssertExpectations(t)
		eventPublisher.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _, _, _ := setup()

		// Create context without proper claims
		ctx := context.Background()

		// Execute
		futureDate := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339) // 30 days from now
		_, err := service.BookNewCargo(ctx, "USNYC", "DEHAM", futureDate)

		// Verify
		assert.Error(t, err)
	})

	t.Run("should fail with invalid arrival deadline", func(t *testing.T) {
		service, _, _, _ := setup()

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute with invalid date format
		_, err := service.BookNewCargo(ctx, "USNYC", "DEHAM", "invalid-date")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arrival deadline format")
	})

	t.Run("should fail when repository store fails", func(t *testing.T) {
		service, cargoRepo, _, _ := setup()

		// Setup mocks
		cargoRepo.On("Store", mock.AnythingOfType("bookingdomain.Cargo")).Return(errors.New("storage error"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		futureDate := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339) // 30 days from now
		_, err := service.BookNewCargo(ctx, "USNYC", "DEHAM", futureDate)

		// Verify
		assert.Error(t, err)
		cargoRepo.AssertExpectations(t)
	})
}

func TestBookingApplicationService_GetCargoDetails(t *testing.T) {
	setup := func() (*BookingApplicationService, *MockCargoRepository, *MockRoutingService, *MockEventPublisher) {
		cargoRepo := &MockCargoRepository{}
		routingService := &MockRoutingService{}
		eventPublisher := &MockEventPublisher{}
		logger := slog.Default()

		service := NewBookingApplicationService(cargoRepo, routingService, eventPublisher, logger)

		return service, cargoRepo, routingService, eventPublisher
	}

	t.Run("should return cargo details successfully", func(t *testing.T) {
		service, cargoRepo, _, _ := setup()

		// Create test cargo
		cargo := createTestCargo(t)
		trackingId := cargo.GetTrackingId()

		// Setup mocks
		cargoRepo.On("FindByTrackingId", trackingId).Return(cargo, nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		result, err := service.GetCargoDetails(ctx, trackingId)

		// Verify
		require.NoError(t, err)
		assert.Equal(t, trackingId, result.GetTrackingId())
		cargoRepo.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _, _, _ := setup()

		trackingId := bookingdomain.NewTrackingId()
		ctx := context.Background()

		// Execute
		_, err := service.GetCargoDetails(ctx, trackingId)

		// Verify
		assert.Error(t, err)
	})

	t.Run("should fail when cargo not found", func(t *testing.T) {
		service, cargoRepo, _, _ := setup()

		trackingId := bookingdomain.NewTrackingId()

		// Setup mocks
		cargoRepo.On("FindByTrackingId", trackingId).Return(bookingdomain.Cargo{}, errors.New("not found"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		_, err := service.GetCargoDetails(ctx, trackingId)

		// Verify
		assert.Error(t, err)
		cargoRepo.AssertExpectations(t)
	})
}

func TestBookingApplicationService_AssignRouteToCargo(t *testing.T) {
	setup := func() (*BookingApplicationService, *MockCargoRepository, *MockRoutingService, *MockEventPublisher) {
		cargoRepo := &MockCargoRepository{}
		routingService := &MockRoutingService{}
		eventPublisher := &MockEventPublisher{}
		logger := slog.Default()

		service := NewBookingApplicationService(cargoRepo, routingService, eventPublisher, logger)

		return service, cargoRepo, routingService, eventPublisher
	}

	t.Run("should assign route successfully", func(t *testing.T) {
		service, cargoRepo, _, eventPublisher := setup()

		// Create test cargo and itinerary
		cargo := createTestCargo(t)
		trackingId := cargo.GetTrackingId()
		itinerary := createTestItinerary(t, cargo.GetRouteSpecification())

		// Setup mocks
		cargoRepo.On("FindByTrackingId", trackingId).Return(cargo, nil)
		cargoRepo.On("Update", mock.AnythingOfType("bookingdomain.Cargo")).Return(nil)
		eventPublisher.On("Publish", mock.Anything).Return(nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		err := service.AssignRouteToCargo(ctx, trackingId, itinerary)

		// Verify
		require.NoError(t, err)
		cargoRepo.AssertExpectations(t)
		eventPublisher.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _, _, _ := setup()

		trackingId := bookingdomain.NewTrackingId()
		itinerary := bookingdomain.Itinerary{}
		ctx := context.Background()

		// Execute
		err := service.AssignRouteToCargo(ctx, trackingId, itinerary)

		// Verify
		assert.Error(t, err)
	})

	t.Run("should fail when cargo not found", func(t *testing.T) {
		service, cargoRepo, _, _ := setup()

		trackingId := bookingdomain.NewTrackingId()
		itinerary := bookingdomain.Itinerary{}

		// Setup mocks
		cargoRepo.On("FindByTrackingId", trackingId).Return(bookingdomain.Cargo{}, errors.New("not found"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		err := service.AssignRouteToCargo(ctx, trackingId, itinerary)

		// Verify
		assert.Error(t, err)
		cargoRepo.AssertExpectations(t)
	})
}

func TestBookingApplicationService_ListAllCargo(t *testing.T) {
	setup := func() (*BookingApplicationService, *MockCargoRepository, *MockRoutingService, *MockEventPublisher) {
		cargoRepo := &MockCargoRepository{}
		routingService := &MockRoutingService{}
		eventPublisher := &MockEventPublisher{}
		logger := slog.Default()

		service := NewBookingApplicationService(cargoRepo, routingService, eventPublisher, logger)

		return service, cargoRepo, routingService, eventPublisher
	}

	t.Run("should list all cargo successfully", func(t *testing.T) {
		service, cargoRepo, _, _ := setup()

		// Create test cargo list
		cargoList := []bookingdomain.Cargo{createTestCargo(t), createTestCargo(t)}

		// Setup mocks
		cargoRepo.On("FindAll").Return(cargoList, nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		result, err := service.ListAllCargo(ctx)

		// Verify
		require.NoError(t, err)
		assert.Len(t, result, 2)
		cargoRepo.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _, _, _ := setup()

		ctx := context.Background()

		// Execute
		_, err := service.ListAllCargo(ctx)

		// Verify
		assert.Error(t, err)
	})

	t.Run("should fail when repository fails", func(t *testing.T) {
		service, cargoRepo, _, _ := setup()

		// Setup mocks
		cargoRepo.On("FindAll").Return([]bookingdomain.Cargo{}, errors.New("repository error"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		_, err := service.ListAllCargo(ctx)

		// Verify
		assert.Error(t, err)
		cargoRepo.AssertExpectations(t)
	})
}

func TestBookingApplicationService_UpdateCargoDelivery(t *testing.T) {
	setup := func() (*BookingApplicationService, *MockCargoRepository, *MockRoutingService, *MockEventPublisher) {
		cargoRepo := &MockCargoRepository{}
		routingService := &MockRoutingService{}
		eventPublisher := &MockEventPublisher{}
		logger := slog.Default()

		service := NewBookingApplicationService(cargoRepo, routingService, eventPublisher, logger)

		return service, cargoRepo, routingService, eventPublisher
	}

	t.Run("should update cargo delivery successfully", func(t *testing.T) {
		service, cargoRepo, _, eventPublisher := setup()

		// Create test cargo and handling history
		cargo := createTestCargo(t)
		trackingId := cargo.GetTrackingId()
		handlingHistory := []bookingdomain.HandlingEventSummary{
			{
				Type:         "RECEIVE",
				Location:     "USNYC",
				VoyageNumber: "",
				Timestamp:    time.Now(),
			},
		}

		// Setup mocks
		cargoRepo.On("FindByTrackingId", trackingId).Return(cargo, nil)
		cargoRepo.On("Update", mock.AnythingOfType("bookingdomain.Cargo")).Return(nil)
		eventPublisher.On("Publish", mock.Anything).Return(nil)

		// Execute
		ctx := context.Background()
		err := service.UpdateCargoDelivery(ctx, trackingId, handlingHistory)

		// Verify
		require.NoError(t, err)
		cargoRepo.AssertExpectations(t)
		eventPublisher.AssertExpectations(t)
	})

	t.Run("should fail when cargo not found", func(t *testing.T) {
		service, cargoRepo, _, _ := setup()

		trackingId := bookingdomain.NewTrackingId()
		handlingHistory := []bookingdomain.HandlingEventSummary{}

		// Setup mocks
		cargoRepo.On("FindByTrackingId", trackingId).Return(bookingdomain.Cargo{}, errors.New("not found"))

		// Execute
		ctx := context.Background()
		err := service.UpdateCargoDelivery(ctx, trackingId, handlingHistory)

		// Verify
		assert.Error(t, err)
		cargoRepo.AssertExpectations(t)
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

func createTestCargo(t *testing.T) bookingdomain.Cargo {
	cargo, err := bookingdomain.NewCargo("USNYC", "DEHAM", time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	return cargo
}

func createTestItinerary(t *testing.T, routeSpec bookingdomain.RouteSpecification) bookingdomain.Itinerary {
	leg, err := bookingdomain.NewLeg(
		"V001",
		routeSpec.Origin,
		routeSpec.Destination,
		time.Now().Add(time.Hour),
		time.Now().Add(2*time.Hour),
	)
	require.NoError(t, err)

	itinerary, err := bookingdomain.NewItinerary([]bookingdomain.Leg{leg})
	require.NoError(t, err)

	return itinerary
}
