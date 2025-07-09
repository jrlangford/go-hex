package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go_hex/internal/booking/bookingdomain"
	"go_hex/internal/handling/handlingdomain"
	"go_hex/internal/support/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing handlers
type MockBookingService struct {
	mock.Mock
}

func (m *MockBookingService) BookNewCargo(ctx context.Context, origin, destination, arrivalDeadline string) (bookingdomain.Cargo, error) {
	args := m.Called(ctx, origin, destination, arrivalDeadline)
	return args.Get(0).(bookingdomain.Cargo), args.Error(1)
}

func (m *MockBookingService) GetCargoDetails(ctx context.Context, trackingId bookingdomain.TrackingId) (bookingdomain.Cargo, error) {
	args := m.Called(ctx, trackingId)
	return args.Get(0).(bookingdomain.Cargo), args.Error(1)
}

func (m *MockBookingService) ListAllCargo(ctx context.Context) ([]bookingdomain.Cargo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]bookingdomain.Cargo), args.Error(1)
}

func (m *MockBookingService) AssignRouteToCargo(ctx context.Context, trackingId bookingdomain.TrackingId, itinerary bookingdomain.Itinerary) error {
	args := m.Called(ctx, trackingId, itinerary)
	return args.Error(0)
}

func (m *MockBookingService) RequestRouteCandidates(ctx context.Context, trackingId bookingdomain.TrackingId) ([]bookingdomain.Itinerary, error) {
	args := m.Called(ctx, trackingId)
	return args.Get(0).([]bookingdomain.Itinerary), args.Error(1)
}

func (m *MockBookingService) ListUnroutedCargo(ctx context.Context) ([]bookingdomain.Cargo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]bookingdomain.Cargo), args.Error(1)
}

func (m *MockBookingService) TrackCargo(ctx context.Context, trackingId bookingdomain.TrackingId) (bookingdomain.Cargo, error) {
	args := m.Called(ctx, trackingId)
	return args.Get(0).(bookingdomain.Cargo), args.Error(1)
}

func (m *MockBookingService) UpdateCargoDelivery(ctx context.Context, trackingId bookingdomain.TrackingId, handlingHistory []bookingdomain.HandlingEventSummary) error {
	args := m.Called(ctx, trackingId, handlingHistory)
	return args.Error(0)
}

type MockRoutingService struct {
	mock.Mock
}

func (m *MockRoutingService) ListAllVoyages(ctx context.Context) ([]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRoutingService) ListAllLocations(ctx context.Context) ([]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).([]interface{}), args.Error(1)
}

type MockHandlingReportService struct {
	mock.Mock
}

func (m *MockHandlingReportService) SubmitHandlingReport(ctx context.Context, report handlingdomain.HandlingReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

type MockHandlingQueryService struct {
	mock.Mock
}

func (m *MockHandlingQueryService) ListAllHandlingEvents(ctx context.Context) ([]handlingdomain.HandlingEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]handlingdomain.HandlingEvent), args.Error(1)
}

func (m *MockHandlingQueryService) GetHandlingHistory(ctx context.Context, trackingId string) (handlingdomain.HandlingHistory, error) {
	args := m.Called(ctx, trackingId)
	return args.Get(0).(handlingdomain.HandlingHistory), args.Error(1)
}

func (m *MockHandlingQueryService) GetHandlingEvent(ctx context.Context, eventId handlingdomain.HandlingEventId) (handlingdomain.HandlingEvent, error) {
	args := m.Called(ctx, eventId)
	return args.Get(0).(handlingdomain.HandlingEvent), args.Error(1)
}

func TestBookCargoHandler(t *testing.T) {
	t.Run("should call booking service to book new cargo", func(t *testing.T) {
		// Setup
		mockBookingService := &MockBookingService{}
		handler := createTestHandler(t, mockBookingService, nil, nil, nil)

		// Create test cargo
		testCargo := createTestCargo(t)
		mockBookingService.On("BookNewCargo", mock.Anything, "USNYC", "DEHAM", mock.AnythingOfType("string")).Return(testCargo, nil)

		// Create request
		futureDate := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339) // 30 days from now
		reqBody := BookCargoRequest{
			Origin:          "USNYC",
			Destination:     "DEHAM",
			ArrivalDeadline: futureDate,
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/cargos", bytes.NewBuffer(jsonBody))
		req = addAuthContext(req)
		w := httptest.NewRecorder()

		// Execute
		handler.BookCargoHandler(w, req)

		// Verify
		assert.Equal(t, http.StatusCreated, w.Code)
		mockBookingService.AssertExpectations(t)

		// Verify the service was called with correct parameters
		mockBookingService.AssertCalled(t, "BookNewCargo", mock.Anything, "USNYC", "DEHAM", mock.AnythingOfType("string"))
	})

	t.Run("should return validation error for invalid request", func(t *testing.T) {
		// Setup
		mockBookingService := &MockBookingService{}
		handler := createTestHandler(t, mockBookingService, nil, nil, nil)

		// Create invalid request (missing required fields)
		reqBody := BookCargoRequest{
			Origin: "USNYC",
			// Missing destination and arrival deadline
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/cargos", bytes.NewBuffer(jsonBody))
		req = addAuthContext(req)
		w := httptest.NewRecorder()

		// Execute
		handler.BookCargoHandler(w, req)

		// Verify
		assert.Equal(t, http.StatusBadRequest, w.Code)
		// Service should not be called for invalid requests
		mockBookingService.AssertNotCalled(t, "BookNewCargo", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestTrackCargoHandler(t *testing.T) {
	t.Run("should call booking service to get cargo details", func(t *testing.T) {
		// Setup
		mockBookingService := &MockBookingService{}
		handler := createTestHandler(t, mockBookingService, nil, nil, nil)

		// Create test cargo
		testCargo := createTestCargo(t)
		trackingId := testCargo.GetTrackingId()
		mockBookingService.On("GetCargoDetails", mock.Anything, trackingId).Return(testCargo, nil)

		// Create request
		req := httptest.NewRequest("GET", "/api/v1/cargos/"+trackingId.String(), nil)
		req = addAuthContext(req)
		w := httptest.NewRecorder()

		// Execute
		handler.TrackCargoHandler(w, req)

		// Verify
		assert.Equal(t, http.StatusOK, w.Code)
		mockBookingService.AssertExpectations(t)
		mockBookingService.AssertCalled(t, "GetCargoDetails", mock.Anything, trackingId)
	})
}

func TestRequestRouteCandidatesHandler(t *testing.T) {
	t.Run("should call booking service to request route candidates", func(t *testing.T) {
		// Setup
		mockBookingService := &MockBookingService{}
		handler := createTestHandler(t, mockBookingService, nil, nil, nil)

		// Create test data
		testCargo := createTestCargo(t)
		trackingId := testCargo.GetTrackingId()
		testItineraries := []bookingdomain.Itinerary{createTestItinerary(t)}
		mockBookingService.On("RequestRouteCandidates", mock.Anything, trackingId).Return(testItineraries, nil)

		// Create request
		reqBody := struct {
			TrackingId string `json:"trackingId"`
		}{
			TrackingId: trackingId.String(),
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/route-candidates", bytes.NewBuffer(jsonBody))
		req = addAuthContext(req)
		w := httptest.NewRecorder()

		// Execute
		handler.RequestRouteCandidatesHandler(w, req)

		// Verify
		assert.Equal(t, http.StatusOK, w.Code)
		mockBookingService.AssertExpectations(t)
		mockBookingService.AssertCalled(t, "RequestRouteCandidates", mock.Anything, trackingId)
	})
}

func TestSubmitHandlingReportHandler(t *testing.T) {
	t.Run("should call handling service to submit handling report", func(t *testing.T) {
		// Setup
		mockHandlingService := &MockHandlingReportService{}
		handler := createTestHandler(t, nil, nil, mockHandlingService, nil)

		// Setup mock
		mockHandlingService.On("SubmitHandlingReport", mock.Anything, mock.Anything).Return(nil)

		// Create request
		reqBody := HandlingEventRequest{
			TrackingId:     "550e8400-e29b-41d4-a716-446655440000", // Valid UUID
			EventType:      "RECEIVE",
			Location:       "USNYC",
			VoyageNumber:   "",
			CompletionTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/handling-events", bytes.NewBuffer(jsonBody))
		req = addAuthContext(req)
		w := httptest.NewRecorder()

		// Execute
		handler.SubmitHandlingReportHandler(w, req)

		// Verify
		assert.Equal(t, http.StatusCreated, w.Code)
		mockHandlingService.AssertExpectations(t)
		mockHandlingService.AssertCalled(t, "SubmitHandlingReport", mock.Anything, mock.Anything)
	})
}

func TestAssignRouteHandler(t *testing.T) {
	t.Run("should call booking service to assign route", func(t *testing.T) {
		// Setup
		mockBookingService := &MockBookingService{}
		handler := createTestHandler(t, mockBookingService, nil, nil, nil)

		// Create test data
		testCargo := createTestCargo(t)
		trackingId := testCargo.GetTrackingId()
		mockBookingService.On("AssignRouteToCargo", mock.Anything, trackingId, mock.AnythingOfType("bookingdomain.Itinerary")).Return(nil)

		// Create request
		reqBody := AssignRouteRequest{
			Legs: []LegDTO{
				{
					VoyageNumber:   "V001",
					LoadLocation:   "USNYC",
					UnloadLocation: "DEHAM",
					LoadTime:       time.Now().Add(time.Hour).Format(time.RFC3339),
					UnloadTime:     time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				},
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/cargos/"+trackingId.String()+"/route", bytes.NewBuffer(jsonBody))
		req = addAuthContext(req)
		w := httptest.NewRecorder()

		// Execute
		handler.AssignRouteHandler(w, req)

		// Verify
		assert.Equal(t, http.StatusOK, w.Code)
		mockBookingService.AssertExpectations(t)
		mockBookingService.AssertCalled(t, "AssignRouteToCargo", mock.Anything, trackingId, mock.AnythingOfType("bookingdomain.Itinerary"))
	})
}

func TestListCargoHandler(t *testing.T) {
	t.Run("should call booking service to list all cargo", func(t *testing.T) {
		// Setup
		mockBookingService := &MockBookingService{}
		handler := createTestHandler(t, mockBookingService, nil, nil, nil)

		// Setup mock
		testCargos := []bookingdomain.Cargo{createTestCargo(t)}
		mockBookingService.On("ListUnroutedCargo", mock.Anything).Return(testCargos, nil)

		// Create request
		req := httptest.NewRequest("GET", "/api/v1/cargos", nil)
		req = addAuthContext(req)
		w := httptest.NewRecorder()

		// Execute
		handler.ListCargoHandler(w, req)

		// Verify
		assert.Equal(t, http.StatusOK, w.Code)
		mockBookingService.AssertExpectations(t)
		mockBookingService.AssertCalled(t, "ListUnroutedCargo", mock.Anything)
	})
}

// Helper functions

func createTestHandler(t *testing.T, bookingService *MockBookingService, routingService *MockRoutingService, handlingReportService *MockHandlingReportService, handlingQueryService *MockHandlingQueryService) *Handler {
	return &Handler{
		authMiddleware:        nil, // Not needed for unit tests
		bookingService:        bookingService,
		routingService:        nil, // Routing service not used in most tests
		handlingReportService: handlingReportService,
		handlingQueryService:  handlingQueryService,
	}
}

func addAuthContext(req *http.Request) *http.Request {
	claims, err := auth.NewClaims(
		"test-user",
		"testuser",
		"test@example.com",
		[]string{"admin"},
		map[string]string{"test": "true"},
	)
	if err != nil {
		panic(err)
	}

	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, claims)
	return req.WithContext(ctx)
}

func createTestCargo(t *testing.T) bookingdomain.Cargo {
	cargo, err := bookingdomain.NewCargo("USNYC", "DEHAM", time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	return cargo
}

func createTestItinerary(t *testing.T) bookingdomain.Itinerary {
	leg, err := bookingdomain.NewLeg("V001", "USNYC", "DEHAM", time.Now().Add(time.Hour), time.Now().Add(2*time.Hour))
	require.NoError(t, err)

	itinerary, err := bookingdomain.NewItinerary([]bookingdomain.Leg{leg})
	require.NoError(t, err)

	return itinerary
}
