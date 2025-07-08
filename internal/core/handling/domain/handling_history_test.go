package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlingHistory_IsCompleted(t *testing.T) {
	t.Run("should return true when history contains CLAIM event", func(t *testing.T) {
		events := createTestHandlingEventsWithClaim(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		assert.True(t, history.IsCompleted())
	})

	t.Run("should return false when history does not contain CLAIM event", func(t *testing.T) {
		events := createTestHandlingEventsWithoutClaim(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		assert.False(t, history.IsCompleted())
	})

	t.Run("should return false for empty history", func(t *testing.T) {
		history, err := NewHandlingHistory("TEST123", []HandlingEvent{})
		require.NoError(t, err)

		assert.False(t, history.IsCompleted())
	})
}

func TestHandlingHistory_IsReceived(t *testing.T) {
	t.Run("should return true when history contains RECEIVE event", func(t *testing.T) {
		events := createTestHandlingEventsWithReceive(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		assert.True(t, history.IsReceived())
	})

	t.Run("should return false when history does not contain RECEIVE event", func(t *testing.T) {
		// Create events starting with LOAD (invalid but for testing)
		event, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "USNYC", "V001", time.Now().Add(-time.Hour))
		require.NoError(t, err)

		history, err := NewHandlingHistory("TEST123", []HandlingEvent{event})
		require.NoError(t, err)

		assert.False(t, history.IsReceived())
	})
}

func TestHandlingHistory_GetCurrentLocation(t *testing.T) {
	t.Run("should return location of most recent event", func(t *testing.T) {
		events := createTestHandlingEventsWithMultipleLocations(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		// Last event should be at SEGOT
		assert.Equal(t, "SEGOT", history.GetCurrentLocation())
	})

	t.Run("should return empty string for empty history", func(t *testing.T) {
		history, err := NewHandlingHistory("TEST123", []HandlingEvent{})
		require.NoError(t, err)

		assert.Empty(t, history.GetCurrentLocation())
	})
}

func TestHandlingHistory_GetCurrentVoyage(t *testing.T) {
	t.Run("should return voyage of most recent event", func(t *testing.T) {
		events := createTestHandlingEventsWithVoyages(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		// Last event should be on V002
		assert.Equal(t, "V002", history.GetCurrentVoyage())
	})

	t.Run("should return empty string when most recent event has no voyage", func(t *testing.T) {
		// CLAIM event typically has no voyage
		event, err := NewHandlingEvent("TEST123", HandlingEventTypeClaim, "USNYC", "", time.Now())
		require.NoError(t, err)

		history, err := NewHandlingHistory("TEST123", []HandlingEvent{event})
		require.NoError(t, err)

		assert.Empty(t, history.GetCurrentVoyage())
	})
}

func TestHandlingHistory_GetEventsOfType(t *testing.T) {
	t.Run("should return all events of specified type", func(t *testing.T) {
		events := createTestHandlingEventsWithMultipleTypes(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		loadEvents := history.GetEventsOfType(HandlingEventTypeLoad)
		unloadEvents := history.GetEventsOfType(HandlingEventTypeUnload)

		assert.Len(t, loadEvents, 2)   // Two LOAD events
		assert.Len(t, unloadEvents, 1) // One UNLOAD event

		for _, event := range loadEvents {
			assert.Equal(t, HandlingEventTypeLoad, event.GetEventType())
		}

		for _, event := range unloadEvents {
			assert.Equal(t, HandlingEventTypeUnload, event.GetEventType())
		}
	})

	t.Run("should return empty slice for non-existent type", func(t *testing.T) {
		events := createTestHandlingEventsWithReceive(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		claimEvents := history.GetEventsOfType(HandlingEventTypeClaim)

		assert.Empty(t, claimEvents)
	})
}

func TestHandlingHistory_GetEventsAtLocation(t *testing.T) {
	t.Run("should return all events at specified location", func(t *testing.T) {
		events := createTestHandlingEventsWithMultipleLocations(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		nycEvents := history.GetEventsAtLocation("USNYC")
		hamEvents := history.GetEventsAtLocation("DEHAM")

		assert.Len(t, nycEvents, 2) // RECEIVE and LOAD at USNYC
		assert.Len(t, hamEvents, 2) // UNLOAD and LOAD at DEHAM

		for _, event := range nycEvents {
			assert.Equal(t, "USNYC", event.GetLocation())
		}

		for _, event := range hamEvents {
			assert.Equal(t, "DEHAM", event.GetLocation())
		}
	})

	t.Run("should return empty slice for non-existent location", func(t *testing.T) {
		events := createTestHandlingEventsWithReceive(t)
		history, err := NewHandlingHistory("TEST123", events)
		require.NoError(t, err)

		wrongEvents := history.GetEventsAtLocation("WRONG")

		assert.Empty(t, wrongEvents)
	})
}

// Helper functions for creating test data

func createTestHandlingEventsWithReceive(t *testing.T) []HandlingEvent {
	event, err := NewHandlingEvent("TEST123", HandlingEventTypeReceive, "USNYC", "", time.Now().Add(-time.Hour))
	require.NoError(t, err)
	return []HandlingEvent{event}
}

func createTestHandlingEventsWithClaim(t *testing.T) []HandlingEvent {
	baseTime := time.Now().Add(-3 * time.Hour)

	receive, err := NewHandlingEvent("TEST123", HandlingEventTypeReceive, "USNYC", "", baseTime)
	require.NoError(t, err)

	load, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "USNYC", "V001", baseTime.Add(time.Hour))
	require.NoError(t, err)

	unload, err := NewHandlingEvent("TEST123", HandlingEventTypeUnload, "DEHAM", "V001", baseTime.Add(2*time.Hour))
	require.NoError(t, err)

	claim, err := NewHandlingEvent("TEST123", HandlingEventTypeClaim, "DEHAM", "", baseTime.Add(3*time.Hour))
	require.NoError(t, err)

	return []HandlingEvent{receive, load, unload, claim}
}

func createTestHandlingEventsWithoutClaim(t *testing.T) []HandlingEvent {
	baseTime := time.Now().Add(-2 * time.Hour)

	receive, err := NewHandlingEvent("TEST123", HandlingEventTypeReceive, "USNYC", "", baseTime)
	require.NoError(t, err)

	load, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "USNYC", "V001", baseTime.Add(time.Hour))
	require.NoError(t, err)

	return []HandlingEvent{receive, load}
}

func createTestHandlingEventsWithMultipleLocations(t *testing.T) []HandlingEvent {
	baseTime := time.Now().Add(-4 * time.Hour)

	receive, err := NewHandlingEvent("TEST123", HandlingEventTypeReceive, "USNYC", "", baseTime)
	require.NoError(t, err)

	load1, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "USNYC", "V001", baseTime.Add(time.Hour))
	require.NoError(t, err)

	unload1, err := NewHandlingEvent("TEST123", HandlingEventTypeUnload, "DEHAM", "V001", baseTime.Add(2*time.Hour))
	require.NoError(t, err)

	load2, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "DEHAM", "V002", baseTime.Add(3*time.Hour))
	require.NoError(t, err)

	unload2, err := NewHandlingEvent("TEST123", HandlingEventTypeUnload, "SEGOT", "V002", baseTime.Add(4*time.Hour))
	require.NoError(t, err)

	return []HandlingEvent{receive, load1, unload1, load2, unload2}
}

func createTestHandlingEventsWithVoyages(t *testing.T) []HandlingEvent {
	baseTime := time.Now().Add(-3 * time.Hour)

	receive, err := NewHandlingEvent("TEST123", HandlingEventTypeReceive, "USNYC", "", baseTime)
	require.NoError(t, err)

	load1, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "USNYC", "V001", baseTime.Add(time.Hour))
	require.NoError(t, err)

	unload1, err := NewHandlingEvent("TEST123", HandlingEventTypeUnload, "DEHAM", "V001", baseTime.Add(2*time.Hour))
	require.NoError(t, err)

	load2, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "DEHAM", "V002", baseTime.Add(3*time.Hour))
	require.NoError(t, err)

	return []HandlingEvent{receive, load1, unload1, load2}
}

func createTestHandlingEventsWithMultipleTypes(t *testing.T) []HandlingEvent {
	baseTime := time.Now().Add(-4 * time.Hour)

	receive, err := NewHandlingEvent("TEST123", HandlingEventTypeReceive, "USNYC", "", baseTime)
	require.NoError(t, err)

	load1, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "USNYC", "V001", baseTime.Add(time.Hour))
	require.NoError(t, err)

	unload, err := NewHandlingEvent("TEST123", HandlingEventTypeUnload, "DEHAM", "V001", baseTime.Add(2*time.Hour))
	require.NoError(t, err)

	load2, err := NewHandlingEvent("TEST123", HandlingEventTypeLoad, "DEHAM", "V002", baseTime.Add(3*time.Hour))
	require.NoError(t, err)

	customs, err := NewHandlingEvent("TEST123", HandlingEventTypeCustoms, "DEHAM", "", baseTime.Add(4*time.Hour))
	require.NoError(t, err)

	return []HandlingEvent{receive, load1, unload, load2, customs}
}
