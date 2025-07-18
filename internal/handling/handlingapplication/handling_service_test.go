package handlingapplication

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/handling/ports/handlingprimary"
	"go_hex/internal/support/auth"
	"go_hex/internal/support/basedomain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations
type MockHandlingEventRepository struct {
	mock.Mock
}

func (m *MockHandlingEventRepository) Store(event handlingdomain.HandlingEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockHandlingEventRepository) FindByTrackingId(trackingId string) ([]handlingdomain.HandlingEvent, error) {
	args := m.Called(trackingId)
	return args.Get(0).([]handlingdomain.HandlingEvent), args.Error(1)
}

func (m *MockHandlingEventRepository) FindAll() ([]handlingdomain.HandlingEvent, error) {
	args := m.Called()
	return args.Get(0).([]handlingdomain.HandlingEvent), args.Error(1)
}

func (m *MockHandlingEventRepository) FindById(id handlingdomain.HandlingEventId) (handlingdomain.HandlingEvent, error) {
	args := m.Called(id)
	return args.Get(0).(handlingdomain.HandlingEvent), args.Error(1)
}

type MockHandlingEventPublisher struct {
	mock.Mock
}

func (m *MockHandlingEventPublisher) Publish(event basedomain.DomainEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestHandlingReportService_SubmitHandlingReport(t *testing.T) {
	setup := func() (handlingprimary.HandlingReportService, *MockHandlingEventRepository, *MockHandlingEventPublisher) {
		repo := &MockHandlingEventRepository{}
		publisher := &MockHandlingEventPublisher{}

		jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

		logger := slog.New(jsonHandler)

		service := NewHandlingReportService(repo, publisher, logger)

		return service, repo, publisher
	}

	t.Run("should submit handling report successfully", func(t *testing.T) {
		service, repo, publisher := setup()

		// Setup mocks
		repo.On("Store", mock.AnythingOfType("handlingdomain.HandlingEvent")).Return(nil)
		publisher.On("Publish", mock.Anything).Return(nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Create test report
		report := handlingdomain.HandlingReport{
			TrackingId:     "TEST123",
			EventType:      "RECEIVE",
			Location:       "USNYC",
			VoyageNumber:   "",
			CompletionTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
		}

		// Execute
		err := service.SubmitHandlingReport(ctx, report)

		// Verify
		require.NoError(t, err)
		repo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _, _ := setup()

		// Create context without proper claims
		ctx := context.Background()

		report := handlingdomain.HandlingReport{
			TrackingId:     "TEST123",
			EventType:      "RECEIVE",
			Location:       "USNYC",
			VoyageNumber:   "",
			CompletionTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
		}

		// Execute
		err := service.SubmitHandlingReport(ctx, report)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should fail with invalid completion time format", func(t *testing.T) {
		service, _, _ := setup()

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		report := handlingdomain.HandlingReport{
			TrackingId:     "TEST123",
			EventType:      "RECEIVE",
			Location:       "USNYC",
			VoyageNumber:   "",
			CompletionTime: "invalid-time-format",
		}

		// Execute
		err := service.SubmitHandlingReport(ctx, report)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid completion time format")
	})

	t.Run("should fail when repository store fails", func(t *testing.T) {
		service, repo, _ := setup()

		// Setup mocks
		repo.On("Store", mock.AnythingOfType("handlingdomain.HandlingEvent")).Return(errors.New("storage error"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		report := handlingdomain.HandlingReport{
			TrackingId:     "TEST123",
			EventType:      "RECEIVE",
			Location:       "USNYC",
			VoyageNumber:   "",
			CompletionTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
		}

		// Execute
		err := service.SubmitHandlingReport(ctx, report)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store handling event")
		repo.AssertExpectations(t)
	})

	t.Run("should fail with invalid event type", func(t *testing.T) {
		service, _, _ := setup()

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		report := handlingdomain.HandlingReport{
			TrackingId:     "TEST123",
			EventType:      "INVALID_TYPE",
			Location:       "USNYC",
			VoyageNumber:   "",
			CompletionTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
		}

		// Execute
		err := service.SubmitHandlingReport(ctx, report)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create handling event")
	})

	t.Run("should fail when LOAD event missing voyage number", func(t *testing.T) {
		service, _, _ := setup()

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		report := handlingdomain.HandlingReport{
			TrackingId:     "TEST123",
			EventType:      "LOAD",
			Location:       "USNYC",
			VoyageNumber:   "", // Required for LOAD events
			CompletionTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
		}

		// Execute
		err := service.SubmitHandlingReport(ctx, report)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create handling event")
	})
}

func TestHandlingEventQueryService_GetHandlingHistory(t *testing.T) {
	setup := func() (handlingprimary.HandlingEventQueryService, *MockHandlingEventRepository) {
		repo := &MockHandlingEventRepository{}

		jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

		logger := slog.New(jsonHandler)

		service := NewHandlingEventQueryService(repo, logger)

		return service, repo
	}

	t.Run("should return handling history successfully", func(t *testing.T) {
		service, repo := setup()

		// Create test events
		events := createTestHandlingEvents(t)

		// Setup mocks
		repo.On("FindByTrackingId", "TEST123").Return(events, nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		history, err := service.GetHandlingHistory(ctx, "TEST123")

		// Verify
		require.NoError(t, err)
		assert.Equal(t, "TEST123", history.TrackingId)
		assert.Len(t, history.Events, len(events))
		repo.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _ := setup()

		// Create context without proper claims
		ctx := context.Background()

		// Execute
		_, err := service.GetHandlingHistory(ctx, "TEST123")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should fail when repository fails", func(t *testing.T) {
		service, repo := setup()

		// Setup mocks
		repo.On("FindByTrackingId", "TEST123").Return([]handlingdomain.HandlingEvent{}, errors.New("repository error"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		_, err := service.GetHandlingHistory(ctx, "TEST123")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find handling events")
		repo.AssertExpectations(t)
	})
}

func TestHandlingEventQueryService_ListAllHandlingEvents(t *testing.T) {
	setup := func() (handlingprimary.HandlingEventQueryService, *MockHandlingEventRepository) {
		repo := &MockHandlingEventRepository{}

		jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

		logger := slog.New(jsonHandler)

		service := NewHandlingEventQueryService(repo, logger)

		return service, repo
	}

	t.Run("should list all handling events successfully", func(t *testing.T) {
		service, repo := setup()

		// Create test events
		events := createTestHandlingEvents(t)

		// Setup mocks
		repo.On("FindAll").Return(events, nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		result, err := service.ListAllHandlingEvents(ctx)

		// Verify
		require.NoError(t, err)
		assert.Len(t, result, len(events))
		repo.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _ := setup()

		// Create context without proper claims
		ctx := context.Background()

		// Execute
		_, err := service.ListAllHandlingEvents(ctx)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should fail when repository fails", func(t *testing.T) {
		service, repo := setup()

		// Setup mocks
		repo.On("FindAll").Return([]handlingdomain.HandlingEvent{}, errors.New("repository error"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		_, err := service.ListAllHandlingEvents(ctx)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve all handling events")
		repo.AssertExpectations(t)
	})
}

func TestHandlingEventQueryService_GetHandlingEvent(t *testing.T) {
	setup := func() (handlingprimary.HandlingEventQueryService, *MockHandlingEventRepository) {
		repo := &MockHandlingEventRepository{}

		jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

		logger := slog.New(jsonHandler)

		service := NewHandlingEventQueryService(repo, logger)

		return service, repo
	}

	t.Run("should return handling event successfully", func(t *testing.T) {
		service, repo := setup()

		// Create test event
		event := createTestHandlingEvent(t)
		eventId := event.GetEventId()

		// Setup mocks
		repo.On("FindById", eventId).Return(event, nil)

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		result, err := service.GetHandlingEvent(ctx, eventId)

		// Verify
		require.NoError(t, err)
		assert.Equal(t, eventId, result.GetEventId())
		repo.AssertExpectations(t)
	})

	t.Run("should fail with unauthorized context", func(t *testing.T) {
		service, _ := setup()

		eventId := handlingdomain.NewHandlingEventId()
		ctx := context.Background()

		// Execute
		_, err := service.GetHandlingEvent(ctx, eventId)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should fail when event not found", func(t *testing.T) {
		service, repo := setup()

		eventId := handlingdomain.NewHandlingEventId()

		// Setup mocks
		repo.On("FindById", eventId).Return(handlingdomain.HandlingEvent{}, errors.New("not found"))

		// Create context with valid claims
		ctx := createContextWithClaims(t, []string{})

		// Execute
		_, err := service.GetHandlingEvent(ctx, eventId)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find handling event")
		repo.AssertExpectations(t)
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

func createTestHandlingEvent(t *testing.T) handlingdomain.HandlingEvent {
	event, err := handlingdomain.NewHandlingEvent(
		"TEST123",
		handlingdomain.HandlingEventTypeReceive,
		"USNYC",
		"",
		time.Now().Add(-time.Hour),
	)
	require.NoError(t, err)
	return event
}

func createTestHandlingEvents(t *testing.T) []handlingdomain.HandlingEvent {
	baseTime := time.Now().Add(-3 * time.Hour)

	receive, err := handlingdomain.NewHandlingEvent("TEST123", handlingdomain.HandlingEventTypeReceive, "USNYC", "", baseTime)
	require.NoError(t, err)

	load, err := handlingdomain.NewHandlingEvent("TEST123", handlingdomain.HandlingEventTypeLoad, "USNYC", "V001", baseTime.Add(time.Hour))
	require.NoError(t, err)

	unload, err := handlingdomain.NewHandlingEvent("TEST123", handlingdomain.HandlingEventTypeUnload, "DEHAM", "V001", baseTime.Add(2*time.Hour))
	require.NoError(t, err)

	return []handlingdomain.HandlingEvent{receive, load, unload}
}
