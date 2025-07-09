package routingdomain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCarrierMovement(t *testing.T) {
	t.Run("should create carrier movement with valid parameters", func(t *testing.T) {
		depLoc, err := NewUnLocode("USNYC")
		require.NoError(t, err)
		arrLoc, err := NewUnLocode("DEHAM")
		require.NoError(t, err)

		depTime := time.Now().Add(time.Hour)
		arrTime := time.Now().Add(2 * time.Hour)

		movement, err := NewCarrierMovement(depLoc, arrLoc, depTime, arrTime)

		require.NoError(t, err)
		assert.Equal(t, depLoc, movement.DepartureLocation)
		assert.Equal(t, arrLoc, movement.ArrivalLocation)
		assert.Equal(t, depTime, movement.DepartureTime)
		assert.Equal(t, arrTime, movement.ArrivalTime)
	})

	t.Run("should fail when departure and arrival locations are the same", func(t *testing.T) {
		location, err := NewUnLocode("USNYC")
		require.NoError(t, err)

		depTime := time.Now().Add(time.Hour)
		arrTime := time.Now().Add(2 * time.Hour)

		_, err = NewCarrierMovement(location, location, depTime, arrTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "departure and arrival locations cannot be the same")
	})

	t.Run("should fail when arrival time is not after departure time", func(t *testing.T) {
		depLoc, err := NewUnLocode("USNYC")
		require.NoError(t, err)
		arrLoc, err := NewUnLocode("DEHAM")
		require.NoError(t, err)

		depTime := time.Now().Add(2 * time.Hour)
		arrTime := time.Now().Add(time.Hour) // Earlier than departure

		_, err = NewCarrierMovement(depLoc, arrLoc, depTime, arrTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "arrival time must be after departure time")
	})
}

func TestNewSchedule(t *testing.T) {
	t.Run("should create schedule with valid movements", func(t *testing.T) {
		movements := createTestMovements(t)

		schedule, err := NewSchedule(movements)

		require.NoError(t, err)
		assert.Len(t, schedule.Movements, 2)
	})

	t.Run("should fail with empty movements", func(t *testing.T) {
		_, err := NewSchedule([]CarrierMovement{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schedule must contain at least one movement")
	})

	t.Run("should fail when movements are not connected", func(t *testing.T) {
		usnyc, err := NewUnLocode("USNYC")
		require.NoError(t, err)
		deham, err := NewUnLocode("DEHAM")
		require.NoError(t, err)
		segot, err := NewUnLocode("SEGOT")
		require.NoError(t, err)
		wrong, err := NewUnLocode("WRONG")
		require.NoError(t, err)

		baseTime := time.Now()

		movement1, err := NewCarrierMovement(usnyc, deham, baseTime, baseTime.Add(time.Hour))
		require.NoError(t, err)

		// Second movement doesn't connect (starts from WRONG instead of DEHAM)
		movement2, err := NewCarrierMovement(wrong, segot, baseTime.Add(2*time.Hour), baseTime.Add(3*time.Hour))
		require.NoError(t, err)

		_, err = NewSchedule([]CarrierMovement{movement1, movement2})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "movements must be connected")
	})

	t.Run("should fail when there's insufficient time between movements", func(t *testing.T) {
		usnyc, err := NewUnLocode("USNYC")
		require.NoError(t, err)
		deham, err := NewUnLocode("DEHAM")
		require.NoError(t, err)
		segot, err := NewUnLocode("SEGOT")
		require.NoError(t, err)

		baseTime := time.Now()

		movement1, err := NewCarrierMovement(usnyc, deham, baseTime, baseTime.Add(time.Hour))
		require.NoError(t, err)

		// Second movement starts exactly when first ends (no time for transshipment)
		movement2, err := NewCarrierMovement(deham, segot, baseTime.Add(time.Hour), baseTime.Add(2*time.Hour))
		require.NoError(t, err)

		_, err = NewSchedule([]CarrierMovement{movement1, movement2})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient time between movements")
	})
}

func TestSchedule_Methods(t *testing.T) {
	movements := createTestMovements(t)
	schedule, err := NewSchedule(movements)
	require.NoError(t, err)

	t.Run("should return correct initial departure location", func(t *testing.T) {
		expected, _ := NewUnLocode("USNYC")
		assert.Equal(t, expected, schedule.InitialDepartureLocation())
	})

	t.Run("should return correct final arrival location", func(t *testing.T) {
		expected, _ := NewUnLocode("SEGOT")
		assert.Equal(t, expected, schedule.FinalArrivalLocation())
	})

	t.Run("should return correct initial departure time", func(t *testing.T) {
		assert.Equal(t, movements[0].DepartureTime, schedule.InitialDepartureTime())
	})

	t.Run("should return correct final arrival time", func(t *testing.T) {
		assert.Equal(t, movements[1].ArrivalTime, schedule.FinalArrivalTime())
	})
}

func TestNewVoyage(t *testing.T) {
	t.Run("should create voyage with valid movements", func(t *testing.T) {
		movements := createTestMovements(t)

		voyage, err := NewVoyage(movements)

		require.NoError(t, err)
		assert.Len(t, voyage.Data.Schedule.Movements, 2)
		assert.False(t, voyage.GetVoyageNumber().String() == "")
	})

	t.Run("should fail with invalid movements", func(t *testing.T) {
		_, err := NewVoyage([]CarrierMovement{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schedule must contain at least one movement")
	})
}

func TestVoyage_Methods(t *testing.T) {
	movements := createTestMovements(t)
	voyage, err := NewVoyage(movements)
	require.NoError(t, err)

	t.Run("should return correct departure location", func(t *testing.T) {
		expected, _ := NewUnLocode("USNYC")
		assert.Equal(t, expected, voyage.GetDepartureLocation())
	})

	t.Run("should return correct arrival location", func(t *testing.T) {
		expected, _ := NewUnLocode("SEGOT")
		assert.Equal(t, expected, voyage.GetArrivalLocation())
	})

	t.Run("should return correct departure time", func(t *testing.T) {
		assert.Equal(t, movements[0].DepartureTime, voyage.GetDepartureTime())
	})

	t.Run("should return correct arrival time", func(t *testing.T) {
		assert.Equal(t, movements[1].ArrivalTime, voyage.GetArrivalTime())
	})
}

func TestVoyage_CanCarryCargoFrom(t *testing.T) {
	movements := createTestMovements(t)
	voyage, err := NewVoyage(movements)
	require.NoError(t, err)

	t.Run("should return true for departure location in schedule", func(t *testing.T) {
		usnyc, _ := NewUnLocode("USNYC")
		assert.True(t, voyage.CanCarryCargoFrom(usnyc))
	})

	t.Run("should return true for intermediate departure location", func(t *testing.T) {
		deham, _ := NewUnLocode("DEHAM")
		assert.True(t, voyage.CanCarryCargoFrom(deham))
	})

	t.Run("should return false for location not in schedule", func(t *testing.T) {
		wrong, _ := NewUnLocode("WRONG")
		assert.False(t, voyage.CanCarryCargoFrom(wrong))
	})
}

func TestVoyage_CanDeliverCargoTo(t *testing.T) {
	movements := createTestMovements(t)
	voyage, err := NewVoyage(movements)
	require.NoError(t, err)

	t.Run("should return true for arrival location in schedule", func(t *testing.T) {
		segot, _ := NewUnLocode("SEGOT")
		assert.True(t, voyage.CanDeliverCargoTo(segot))
	})

	t.Run("should return true for intermediate arrival location", func(t *testing.T) {
		deham, _ := NewUnLocode("DEHAM")
		assert.True(t, voyage.CanDeliverCargoTo(deham))
	})

	t.Run("should return false for location not in schedule", func(t *testing.T) {
		wrong, _ := NewUnLocode("WRONG")
		assert.False(t, voyage.CanDeliverCargoTo(wrong))
	})
}

func TestVoyage_IsOperational(t *testing.T) {
	t.Run("should return true for future voyage", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour)
		movements := createTestMovementsWithTime(t, futureTime)
		voyage, err := NewVoyage(movements)
		require.NoError(t, err)

		assert.True(t, voyage.IsOperational())
	})

	t.Run("should return false for past voyage", func(t *testing.T) {
		pastTime := time.Now().Add(-24 * time.Hour)
		movements := createTestMovementsWithTime(t, pastTime)
		voyage, err := NewVoyage(movements)
		require.NoError(t, err)

		assert.False(t, voyage.IsOperational())
	})
}

// Helper functions

func createTestMovements(t *testing.T) []CarrierMovement {
	baseTime := time.Now().Add(time.Hour) // Start in the future
	return createTestMovementsWithTime(t, baseTime)
}

func createTestMovementsWithTime(t *testing.T, baseTime time.Time) []CarrierMovement {
	usnyc, err := NewUnLocode("USNYC")
	require.NoError(t, err)
	deham, err := NewUnLocode("DEHAM")
	require.NoError(t, err)
	segot, err := NewUnLocode("SEGOT")
	require.NoError(t, err)

	movement1, err := NewCarrierMovement(
		usnyc, deham,
		baseTime,
		baseTime.Add(time.Hour),
	)
	require.NoError(t, err)

	movement2, err := NewCarrierMovement(
		deham, segot,
		baseTime.Add(2*time.Hour), // Allow time for transshipment
		baseTime.Add(3*time.Hour),
	)
	require.NoError(t, err)

	return []CarrierMovement{movement1, movement2}
}
