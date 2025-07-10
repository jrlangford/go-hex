package bookingdomain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouteSpecification(t *testing.T) {
	t.Run("should create valid route specification", func(t *testing.T) {
		origin := "USNYC"
		destination := "SEGOT"
		arrivalDeadline := time.Now().Add(30 * 24 * time.Hour)

		spec, err := NewRouteSpecification(origin, destination, arrivalDeadline)

		require.NoError(t, err)
		assert.Equal(t, origin, spec.Origin)
		assert.Equal(t, destination, spec.Destination)
		assert.Equal(t, arrivalDeadline, spec.ArrivalDeadline)
	})

	t.Run("should fail when origin equals destination", func(t *testing.T) {
		origin := "USNYC"
		destination := "USNYC"
		arrivalDeadline := time.Now().Add(30 * 24 * time.Hour)

		_, err := NewRouteSpecification(origin, destination, arrivalDeadline)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "origin and destination cannot be the same")
	})

	t.Run("should fail when arrival deadline is in the past", func(t *testing.T) {
		origin := "USNYC"
		destination := "SEGOT"
		arrivalDeadline := time.Now().Add(-24 * time.Hour)

		_, err := NewRouteSpecification(origin, destination, arrivalDeadline)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "arrival deadline must be in the future")
	})

	t.Run("should fail with invalid UN/LOCODE format", func(t *testing.T) {
		origin := "US" // Too short
		destination := "SEGOT"
		arrivalDeadline := time.Now().Add(30 * 24 * time.Hour)

		_, err := NewRouteSpecification(origin, destination, arrivalDeadline)

		assert.Error(t, err)
	})
}

func TestNewLeg(t *testing.T) {
	t.Run("should create valid leg", func(t *testing.T) {
		voyageNumber := "V001"
		loadLocation := "USNYC"
		unloadLocation := "SEGOT"
		loadTime := time.Now().Add(24 * time.Hour)
		unloadTime := loadTime.Add(48 * time.Hour)

		leg, err := NewLeg(voyageNumber, loadLocation, unloadLocation, loadTime, unloadTime)

		require.NoError(t, err)
		assert.Equal(t, voyageNumber, leg.VoyageNumber)
		assert.Equal(t, loadLocation, leg.LoadLocation)
		assert.Equal(t, unloadLocation, leg.UnloadLocation)
		assert.Equal(t, loadTime, leg.LoadTime)
		assert.Equal(t, unloadTime, leg.UnloadTime)
	})

	t.Run("should fail when load and unload locations are the same", func(t *testing.T) {
		voyageNumber := "V001"
		loadLocation := "USNYC"
		unloadLocation := "USNYC"
		loadTime := time.Now().Add(24 * time.Hour)
		unloadTime := loadTime.Add(48 * time.Hour)

		_, err := NewLeg(voyageNumber, loadLocation, unloadLocation, loadTime, unloadTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load and unload locations cannot be the same")
	})

	t.Run("should fail when unload time is not after load time", func(t *testing.T) {
		voyageNumber := "V001"
		loadLocation := "USNYC"
		unloadLocation := "SEGOT"
		loadTime := time.Now().Add(24 * time.Hour)
		unloadTime := loadTime.Add(-1 * time.Hour) // Before load time

		_, err := NewLeg(voyageNumber, loadLocation, unloadLocation, loadTime, unloadTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unload time must be after load time")
	})
}

func TestNewItinerary(t *testing.T) {
	t.Run("should create valid single-leg itinerary", func(t *testing.T) {
		leg := createTestLeg(t, "V001", "USNYC", "SEGOT")

		itinerary, err := NewItinerary([]Leg{leg})

		require.NoError(t, err)
		assert.Len(t, itinerary.Legs, 1)
		assert.Equal(t, leg, itinerary.Legs[0])
	})

	t.Run("should create valid multi-leg itinerary", func(t *testing.T) {
		leg1 := createTestLeg(t, "V001", "USNYC", "DEHAM")
		leg2 := createTestLegAfter(t, leg1, "V002", "DEHAM", "SEGOT")

		itinerary, err := NewItinerary([]Leg{leg1, leg2})

		require.NoError(t, err)
		assert.Len(t, itinerary.Legs, 2)
	})

	t.Run("should fail with empty legs", func(t *testing.T) {
		_, err := NewItinerary([]Leg{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "itinerary must contain at least one leg")
	})

	t.Run("should fail with disconnected legs", func(t *testing.T) {
		leg1 := createTestLeg(t, "V001", "USNYC", "DEHAM")
		leg2 := createTestLegAfter(t, leg1, "V002", "WRONG", "SEGOT") // Wrong connection

		_, err := NewItinerary([]Leg{leg1, leg2})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "legs must be connected")
	})

	t.Run("should fail with insufficient time between legs", func(t *testing.T) {
		leg1 := createTestLeg(t, "V001", "USNYC", "DEHAM")
		// Create leg2 that starts before leg1 ends
		leg2, err := NewLeg("V002", "DEHAM", "SEGOT", leg1.UnloadTime.Add(-1*time.Hour), leg1.UnloadTime.Add(1*time.Hour))
		require.NoError(t, err)

		_, err = NewItinerary([]Leg{leg1, leg2})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient time between legs")
	})
}

func TestItinerary_SatisfiesSpecification(t *testing.T) {
	t.Run("should satisfy matching specification", func(t *testing.T) {
		spec := createTestRouteSpec(t, "USNYC", "SEGOT")
		leg := createTestLeg(t, "V001", "USNYC", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.SatisfiesSpecification(spec)

		assert.True(t, result)
	})

	t.Run("should not satisfy with wrong origin", func(t *testing.T) {
		spec := createTestRouteSpec(t, "USNYC", "SEGOT")
		leg := createTestLeg(t, "V001", "WRONG", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.SatisfiesSpecification(spec)

		assert.False(t, result)
	})

	t.Run("should not satisfy with wrong destination", func(t *testing.T) {
		spec := createTestRouteSpec(t, "USNYC", "SEGOT")
		leg := createTestLeg(t, "V001", "USNYC", "WRONG")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.SatisfiesSpecification(spec)

		assert.False(t, result)
	})

	t.Run("should not satisfy when arriving after deadline", func(t *testing.T) {
		spec := createTestRouteSpec(t, "USNYC", "SEGOT")

		// Create leg that arrives after deadline
		voyageNumber := "V001"
		loadLocation := "USNYC"
		unloadLocation := "SEGOT"
		loadTime := spec.ArrivalDeadline.Add(-1 * time.Hour)
		unloadTime := spec.ArrivalDeadline.Add(1 * time.Hour) // After deadline

		leg, err := NewLeg(voyageNumber, loadLocation, unloadLocation, loadTime, unloadTime)
		require.NoError(t, err)

		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.SatisfiesSpecification(spec)

		assert.False(t, result)
	})
}

func TestItinerary_FinalArrivalTime(t *testing.T) {
	t.Run("should return final arrival time for single leg", func(t *testing.T) {
		leg := createTestLeg(t, "V001", "USNYC", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		finalTime := itinerary.FinalArrivalTime()

		assert.Equal(t, leg.UnloadTime, finalTime)
	})

	t.Run("should return final arrival time for multi-leg itinerary", func(t *testing.T) {
		leg1 := createTestLeg(t, "V001", "USNYC", "DEHAM")
		leg2 := createTestLegAfter(t, leg1, "V002", "DEHAM", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg1, leg2})
		require.NoError(t, err)

		finalTime := itinerary.FinalArrivalTime()

		assert.Equal(t, leg2.UnloadTime, finalTime)
	})

	t.Run("should return zero time for empty itinerary", func(t *testing.T) {
		itinerary := Itinerary{Legs: []Leg{}}

		finalTime := itinerary.FinalArrivalTime()

		assert.True(t, finalTime.IsZero())
	})
}

func TestItinerary_InitialDepartureTime(t *testing.T) {
	t.Run("should return initial departure time", func(t *testing.T) {
		leg1 := createTestLeg(t, "V001", "USNYC", "DEHAM")
		leg2 := createTestLegAfter(t, leg1, "V002", "DEHAM", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg1, leg2})
		require.NoError(t, err)

		initialTime := itinerary.InitialDepartureTime()

		assert.Equal(t, leg1.LoadTime, initialTime)
	})
}

func TestItinerary_IsOnTrack(t *testing.T) {
	t.Run("should be on track for matching voyage and location", func(t *testing.T) {
		leg := createTestLeg(t, "V001", "USNYC", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.IsOnTrack("USNYC", "V001")

		assert.True(t, result)
	})

	t.Run("should be on track for unload location", func(t *testing.T) {
		leg := createTestLeg(t, "V001", "USNYC", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.IsOnTrack("SEGOT", "V001")

		assert.True(t, result)
	})

	t.Run("should not be on track for wrong voyage", func(t *testing.T) {
		leg := createTestLeg(t, "V001", "USNYC", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.IsOnTrack("USNYC", "WRONG")

		assert.False(t, result)
	})

	t.Run("should not be on track for wrong location", func(t *testing.T) {
		leg := createTestLeg(t, "V001", "USNYC", "SEGOT")
		itinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		result := itinerary.IsOnTrack("WRONG", "V001")

		assert.False(t, result)
	})
}

// Helper functions for tests

func createTestRouteSpec(t *testing.T, origin, destination string) RouteSpecification {
	arrivalDeadline := time.Now().Add(30 * 24 * time.Hour)
	spec, err := NewRouteSpecification(origin, destination, arrivalDeadline)
	require.NoError(t, err)
	return spec
}

func createTestLeg(t *testing.T, voyageNumber, loadLocation, unloadLocation string) Leg {
	loadTime := time.Now().Add(24 * time.Hour)
	unloadTime := loadTime.Add(48 * time.Hour)

	leg, err := NewLeg(voyageNumber, loadLocation, unloadLocation, loadTime, unloadTime)
	require.NoError(t, err)
	return leg
}

func createTestLegAfter(t *testing.T, previousLeg Leg, voyageNumber, loadLocation, unloadLocation string) Leg {
	loadTime := previousLeg.UnloadTime.Add(2 * time.Hour) // 2 hours for transfer
	unloadTime := loadTime.Add(48 * time.Hour)

	leg, err := NewLeg(voyageNumber, loadLocation, unloadLocation, loadTime, unloadTime)
	require.NoError(t, err)
	return leg
}
