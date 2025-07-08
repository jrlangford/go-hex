package domain

import (
	"go_hex/internal/support/basedomain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandlingEvent(t *testing.T) {
	t.Run("should create valid handling event", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeReceive
		location := "USNYC"
		voyageNumber := ""
		completionTime := time.Now().Add(-1 * time.Hour)

		event, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		require.NoError(t, err)
		assert.Equal(t, trackingId, event.GetTrackingId())
		assert.Equal(t, eventType, event.GetEventType())
		assert.Equal(t, location, event.GetLocation())
		assert.Equal(t, voyageNumber, event.GetVoyageNumber())
		assert.Equal(t, completionTime, event.GetCompletionTime())
	})

	t.Run("should fail when completion time is in the future", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeReceive
		location := "USNYC"
		voyageNumber := ""
		completionTime := time.Now().Add(1 * time.Hour) // Future

		_, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "completion time cannot be in the future")
	})

	t.Run("should fail when completion time is too far in the past", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeReceive
		location := "USNYC"
		voyageNumber := ""
		completionTime := time.Now().AddDate(0, 0, -31) // 31 days ago

		_, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "completion time cannot be more than 30 days in the past")
	})

	t.Run("should require voyage number for LOAD events", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeLoad
		location := "USNYC"
		voyageNumber := "" // Missing
		completionTime := time.Now().Add(-1 * time.Hour)

		_, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "voyage number is required for LOAD and UNLOAD events")
	})

	t.Run("should require voyage number for UNLOAD events", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeUnload
		location := "SEGOT"
		voyageNumber := "" // Missing
		completionTime := time.Now().Add(-1 * time.Hour)

		_, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "voyage number is required for LOAD and UNLOAD events")
	})

	t.Run("should not allow voyage number for RECEIVE events", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeReceive
		location := "USNYC"
		voyageNumber := "V001" // Should not be provided
		completionTime := time.Now().Add(-1 * time.Hour)

		_, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "voyage number should not be provided for RECEIVE and CLAIM events")
	})

	t.Run("should not allow voyage number for CLAIM events", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeClaim
		location := "SEGOT"
		voyageNumber := "V001" // Should not be provided
		completionTime := time.Now().Add(-1 * time.Hour)

		_, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "voyage number should not be provided for RECEIVE and CLAIM events")
	})

	t.Run("should fail with invalid location format", func(t *testing.T) {
		trackingId := "test-tracking-id"
		eventType := HandlingEventTypeReceive
		location := "US" // Too short
		voyageNumber := ""
		completionTime := time.Now().Add(-1 * time.Hour)

		_, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location must be a valid 5-character UN/LOCODE")
	})
}

func TestNewHandlingHistory(t *testing.T) {
	t.Run("should create valid handling history", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		events := []HandlingEvent{event}

		history, err := NewHandlingHistory(trackingId, events)

		require.NoError(t, err)
		assert.Equal(t, trackingId, history.TrackingId)
		assert.Len(t, history.Events, 1)
	})

	t.Run("should create empty handling history", func(t *testing.T) {
		trackingId := "test-tracking-id"
		events := []HandlingEvent{}

		history, err := NewHandlingHistory(trackingId, events)

		require.NoError(t, err)
		assert.Equal(t, trackingId, history.TrackingId)
		assert.Len(t, history.Events, 0)
	})
}

func TestHandlingHistory_GetMostRecentEvent(t *testing.T) {
	t.Run("should return most recent event", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event1 := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		event2 := createTestHandlingEventAfter(t, event1, HandlingEventTypeLoad, "USNYC", "V001")
		events := []HandlingEvent{event1, event2}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		mostRecent := history.GetMostRecentEvent()

		assert.NotNil(t, mostRecent)
		assert.Equal(t, event2.GetEventId(), mostRecent.GetEventId())
	})

	t.Run("should return nil for empty history", func(t *testing.T) {
		trackingId := "test-tracking-id"
		events := []HandlingEvent{}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		mostRecent := history.GetMostRecentEvent()

		assert.Nil(t, mostRecent)
	})
}

func TestHandlingHistory_HasEventType(t *testing.T) {
	t.Run("should find existing event type", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		events := []HandlingEvent{event}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		result := history.HasEventType(HandlingEventTypeReceive)

		assert.True(t, result)
	})

	t.Run("should not find missing event type", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		events := []HandlingEvent{event}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		result := history.HasEventType(HandlingEventTypeClaim)

		assert.False(t, result)
	})
}

func TestHandlingHistory_GetLastEventAtLocation(t *testing.T) {
	t.Run("should find last event at location", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event1 := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		event2 := createTestHandlingEventAfter(t, event1, HandlingEventTypeLoad, "USNYC", "V001")
		event3 := createTestHandlingEventAfter(t, event2, HandlingEventTypeUnload, "SEGOT", "V001")
		events := []HandlingEvent{event1, event2, event3}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		lastAtUSNYC := history.GetLastEventAtLocation("USNYC")

		assert.NotNil(t, lastAtUSNYC)
		assert.Equal(t, event2.GetEventId(), lastAtUSNYC.GetEventId())
	})

	t.Run("should return nil for location not found", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		events := []HandlingEvent{event}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		lastAtWRONG := history.GetLastEventAtLocation("WRONG")

		assert.Nil(t, lastAtWRONG)
	})
}

func TestHandlingHistory_IsValidSequence(t *testing.T) {
	t.Run("should validate correct sequence", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event1 := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		event2 := createTestHandlingEventAfter(t, event1, HandlingEventTypeLoad, "USNYC", "V001")
		event3 := createTestHandlingEventAfter(t, event2, HandlingEventTypeUnload, "SEGOT", "V001")
		event4 := createTestHandlingEventAfter(t, event3, HandlingEventTypeClaim, "SEGOT", "")
		events := []HandlingEvent{event1, event2, event3, event4}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		err = history.IsValidSequence()

		assert.NoError(t, err)
	})

	t.Run("should fail if first event is not RECEIVE", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event := createTestHandlingEvent(t, HandlingEventTypeLoad, "USNYC", "V001")
		events := []HandlingEvent{event}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		err = history.IsValidSequence()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first handling event must be RECEIVE")
	})

	t.Run("should fail if events are not in chronological order", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event1 := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")

		// Create event2 with earlier time than event1
		event2Data := HandlingEventData{
			EventId:          NewHandlingEventId(),
			TrackingId:       trackingId,
			EventType:        HandlingEventTypeLoad,
			Location:         "USNYC",
			VoyageNumber:     "V001",
			CompletionTime:   event1.GetCompletionTime().Add(-1 * time.Hour), // Earlier
			RegistrationTime: time.Now(),
		}
		event2 := HandlingEvent{
			BaseEntity: basedomain.NewBaseEntity(event2Data.EventId),
			Data:       event2Data,
		}

		events := []HandlingEvent{event1, event2}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		err = history.IsValidSequence()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "events must be in chronological order")
	})

	t.Run("should fail if CLAIM comes before UNLOAD", func(t *testing.T) {
		trackingId := "test-tracking-id"
		event1 := createTestHandlingEvent(t, HandlingEventTypeReceive, "USNYC", "")
		event2 := createTestHandlingEventAfter(t, event1, HandlingEventTypeClaim, "SEGOT", "")
		events := []HandlingEvent{event1, event2}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		err = history.IsValidSequence()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot claim cargo before unloading")
	})

	t.Run("should validate empty history", func(t *testing.T) {
		trackingId := "test-tracking-id"
		events := []HandlingEvent{}

		history, err := NewHandlingHistory(trackingId, events)
		require.NoError(t, err)

		err = history.IsValidSequence()

		assert.NoError(t, err)
	})
}

// Helper functions for tests

func createTestHandlingEvent(t *testing.T, eventType HandlingEventType, location, voyageNumber string) HandlingEvent {
	trackingId := "test-tracking-id"
	completionTime := time.Now().Add(-2 * time.Hour)

	event, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)
	require.NoError(t, err)
	return event
}

func createTestHandlingEventAfter(t *testing.T, previousEvent HandlingEvent, eventType HandlingEventType, location, voyageNumber string) HandlingEvent {
	trackingId := previousEvent.GetTrackingId()
	// Add time but ensure it's still in the past - use min of 30 minutes after previous event or 30 minutes ago
	timeSincePrevious := previousEvent.GetCompletionTime().Add(30 * time.Minute)
	thirtyMinsAgo := time.Now().Add(-30 * time.Minute)
	
	var completionTime time.Time
	if timeSincePrevious.Before(thirtyMinsAgo) {
		completionTime = timeSincePrevious
	} else {
		completionTime = thirtyMinsAgo
	}

	event, err := NewHandlingEvent(trackingId, eventType, location, voyageNumber, completionTime)
	require.NoError(t, err)
	return event
}
