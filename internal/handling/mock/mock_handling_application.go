package mock

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"go_hex/internal/handling/application"
	"go_hex/internal/handling/domain"
	"go_hex/internal/handling/ports/primary"
	"go_hex/internal/handling/ports/secondary"
	"go_hex/internal/support/auth"
)

// MockHandlingApplication embeds the real application service but provides test data population capabilities
type MockHandlingApplication struct {
	*application.HandlingReportService
	logger            *slog.Logger
	random            *rand.Rand
	handlingEventRepo secondary.HandlingEventRepository
}

// NewMockHandlingApplication creates a mock handling application with embedded real application service
func NewMockHandlingApplication(
	handlingEventRepo secondary.HandlingEventRepository,
	eventPublisher secondary.EventPublisher,
	logger *slog.Logger,
	seed int64,
) *MockHandlingApplication {
	realApp := application.NewHandlingReportService(handlingEventRepo, eventPublisher, logger)

	return &MockHandlingApplication{
		HandlingReportService: realApp.(*application.HandlingReportService),
		logger:                logger,
		random:                rand.New(rand.NewSource(seed)),
		handlingEventRepo:     handlingEventRepo,
	}
}

// PopulateTestHandlingEvents creates test handling events using business logic through the application layer
func (m *MockHandlingApplication) PopulateTestHandlingEvents(ctx context.Context, scenarios []TestHandlingScenario) ([]domain.HandlingEvent, error) {
	m.logger.Info("Populating test handling events through handling application", "scenarios", len(scenarios))

	// Create authenticated context for internal operations
	testCtx := m.createTestContext(ctx)

	var events []domain.HandlingEvent

	for _, scenario := range scenarios {
		for _, eventSpec := range scenario.EventSequence {
			// Generate completion time relative to current date
			// Events should be in the past (24+ hours ago to ensure validity)
			hoursAgo := 24 + m.random.Intn(72) // 24-96 hours ago
			completionTime := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

			// Create handling report for submission
			report := domain.HandlingReport{
				TrackingId:     scenario.TrackingId,
				EventType:      string(eventSpec.EventType),
				Location:       eventSpec.Location,
				VoyageNumber:   eventSpec.VoyageNumber,
				CompletionTime: completionTime.Format(time.RFC3339),
			}

			// Use the real application service to submit the handling report
			err := m.HandlingReportService.SubmitHandlingReport(testCtx, report)
			if err != nil {
				m.logger.Error("Failed to submit test handling report", "error", err,
					"trackingId", scenario.TrackingId, "eventType", eventSpec.EventType)
				return nil, fmt.Errorf("failed to submit handling report for %s: %w", scenario.TrackingId, err)
			}

			// Since we can't directly get the created event from the service,
			// we'll create one for our return list to maintain consistency
			event, err := domain.NewHandlingEvent(
				scenario.TrackingId,
				eventSpec.EventType,
				eventSpec.Location,
				eventSpec.VoyageNumber,
				completionTime,
			)
			if err != nil {
				m.logger.Warn("Failed to create domain event for tracking", "error", err)
				continue
			}

			events = append(events, event)
			m.logger.Debug("Submitted test handling event",
				"trackingId", scenario.TrackingId,
				"eventType", eventSpec.EventType,
				"location", eventSpec.Location)
		}
	}

	m.logger.Info("Successfully populated test handling events", "count", len(events))
	return events, nil
}

// GenerateHandlingScenarios creates realistic handling event sequences for cargo
func (m *MockHandlingApplication) GenerateHandlingScenarios(trackingIds []string, locations []string) []TestHandlingScenario {
	m.logger.Info("Generating handling scenarios", "trackingIds", len(trackingIds), "locations", len(locations))

	if len(locations) < 2 {
		m.logger.Warn("Not enough locations to generate handling scenarios", "locations", len(locations))
		return nil
	}

	scenarios := make([]TestHandlingScenario, 0, len(trackingIds))

	for _, trackingId := range trackingIds {
		// Generate a realistic sequence of handling events
		eventSequence := m.generateEventSequence(locations)

		scenario := TestHandlingScenario{
			TrackingId:    trackingId,
			EventSequence: eventSequence,
		}

		scenarios = append(scenarios, scenario)
	}

	m.logger.Info("Generated handling scenarios", "count", len(scenarios))
	return scenarios
}

// generateEventSequence creates a realistic sequence of handling events
func (m *MockHandlingApplication) generateEventSequence(locations []string) []TestHandlingEventSpec {
	var events []TestHandlingEventSpec

	// Always start with RECEIVE
	originLocation := locations[m.random.Intn(len(locations))]
	events = append(events, TestHandlingEventSpec{
		EventType:    domain.HandlingEventTypeReceive,
		Location:     originLocation,
		VoyageNumber: "", // No voyage for RECEIVE
	})

	// Add 1-3 LOAD/UNLOAD pairs for journey
	loadUnloadPairs := 1 + m.random.Intn(3) // 1-3 pairs

	for i := 0; i < loadUnloadPairs; i++ {
		// LOAD event
		loadLocation := locations[m.random.Intn(len(locations))]
		voyageNumber := fmt.Sprintf("V%d", 1000+m.random.Intn(9000))

		events = append(events, TestHandlingEventSpec{
			EventType:    domain.HandlingEventTypeLoad,
			Location:     loadLocation,
			VoyageNumber: voyageNumber,
		})

		// UNLOAD event (different location)
		unloadLocation := locations[m.random.Intn(len(locations))]
		for unloadLocation == loadLocation && len(locations) > 1 {
			unloadLocation = locations[m.random.Intn(len(locations))]
		}

		events = append(events, TestHandlingEventSpec{
			EventType:    domain.HandlingEventTypeUnload,
			Location:     unloadLocation,
			VoyageNumber: voyageNumber,
		})
	}

	// End with CLAIM if this is a complete journey
	if m.random.Float32() < 0.7 { // 70% chance of completed delivery
		finalLocation := events[len(events)-1].Location // Use last unload location
		events = append(events, TestHandlingEventSpec{
			EventType:    domain.HandlingEventTypeClaim,
			Location:     finalLocation,
			VoyageNumber: "", // No voyage for CLAIM
		})
	}

	return events
}

// createTestContext creates an authenticated context for test operations
func (m *MockHandlingApplication) createTestContext(ctx context.Context) context.Context {
	// Create test claims with admin permissions
	claims, _ := auth.NewClaims(
		"test-user",
		"test-system",
		"test@example.com",
		[]string{string(auth.RoleAdmin)},
		map[string]string{"test": "true"},
	)

	return context.WithValue(ctx, auth.ClaimsContextKey, claims)
}

// TestHandlingScenario represents a test scenario for handling events
type TestHandlingScenario struct {
	TrackingId    string
	EventSequence []TestHandlingEventSpec
}

// TestHandlingEventSpec defines the specification for creating a test handling event
type TestHandlingEventSpec struct {
	EventType    domain.HandlingEventType
	Location     string
	VoyageNumber string // Optional, empty for RECEIVE/CLAIM
}

// Ensure MockHandlingApplication implements primary ports
var _ primary.HandlingReportService = (*MockHandlingApplication)(nil)
