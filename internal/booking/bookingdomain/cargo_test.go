package bookingdomain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCargo(t *testing.T) {
	t.Run("should create cargo with valid parameters", func(t *testing.T) {
		origin := "USNYC"
		destination := "SEGOT"
		arrivalDeadline := time.Now().Add(30 * 24 * time.Hour) // 30 days from now

		cargo, err := NewCargo(origin, destination, arrivalDeadline)

		require.NoError(t, err)
		assert.Equal(t, origin, cargo.GetRouteSpecification().Origin)
		assert.Equal(t, destination, cargo.GetRouteSpecification().Destination)
		assert.Equal(t, arrivalDeadline, cargo.GetRouteSpecification().ArrivalDeadline)
		assert.False(t, cargo.IsRouted())
		assert.Equal(t, TransportStatusNotReceived, cargo.GetDelivery().TransportStatus)
		assert.Equal(t, RoutingStatusNotRouted, cargo.GetDelivery().RoutingStatus)
	})

	t.Run("should fail with invalid route specification", func(t *testing.T) {
		origin := "USNYC"
		destination := "USNYC" // Same as origin
		arrivalDeadline := time.Now().Add(30 * 24 * time.Hour)

		_, err := NewCargo(origin, destination, arrivalDeadline)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "origin and destination cannot be the same")
	})

	t.Run("should fail with past arrival deadline", func(t *testing.T) {
		origin := "USNYC"
		destination := "SEGOT"
		arrivalDeadline := time.Now().Add(-24 * time.Hour) // Yesterday

		_, err := NewCargo(origin, destination, arrivalDeadline)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "arrival deadline must be in the future")
	})
}

func TestCargo_AssignToRoute(t *testing.T) {
	t.Run("should assign valid itinerary", func(t *testing.T) {
		cargo := createTestCargo(t)
		itinerary := createTestItinerary(t, cargo.GetRouteSpecification())

		err := cargo.AssignToRoute(itinerary)

		require.NoError(t, err)
		assert.True(t, cargo.IsRouted())
		assert.Equal(t, RoutingStatusRouted, cargo.GetDelivery().RoutingStatus)
		assert.NotNil(t, cargo.GetItinerary())
	})

	t.Run("should fail if itinerary doesn't satisfy specification", func(t *testing.T) {
		cargo := createTestCargo(t)

		// Create itinerary with wrong destination
		wrongLeg, err := NewLeg("V001", "USNYC", "WRONG", time.Now().Add(time.Hour), time.Now().Add(2*time.Hour))
		require.NoError(t, err)
		wrongItinerary, err := NewItinerary([]Leg{wrongLeg})
		require.NoError(t, err)

		err = cargo.AssignToRoute(wrongItinerary)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "itinerary does not satisfy route specification")
	})

	t.Run("should fail if cargo is already delivered", func(t *testing.T) {
		cargo := createTestCargo(t)

		// Mark cargo as delivered
		deliveredStatus, err := NewDelivery(TransportStatusClaimed, RoutingStatusRouted, "SEGOT", "", true)
		require.NoError(t, err)
		cargo.Data.Delivery = deliveredStatus

		itinerary := createTestItinerary(t, cargo.GetRouteSpecification())
		err = cargo.AssignToRoute(itinerary)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot reassign route to already delivered cargo")
	})

	t.Run("should fail if itinerary exceeds deadline", func(t *testing.T) {
		cargo := createTestCargo(t)

		// Create itinerary that arrives after deadline but satisfies origin/destination
		// We need to create an itinerary that passes origin/destination check but fails deadline check
		futureTime := cargo.GetRouteSpecification().ArrivalDeadline.Add(24 * time.Hour)
		leg, err := NewLeg("V001", "USNYC", "SEGOT", time.Now().Add(time.Hour), futureTime)
		require.NoError(t, err)
		lateItinerary, err := NewItinerary([]Leg{leg})
		require.NoError(t, err)

		err = cargo.AssignToRoute(lateItinerary)

		assert.Error(t, err)
		// The SatisfiesSpecification method checks deadline as part of its logic,
		// so it returns "itinerary does not satisfy route specification"
		assert.Contains(t, err.Error(), "itinerary does not satisfy route specification")
	})
}

func TestCargo_CanBeRerouted(t *testing.T) {
	t.Run("should allow rerouting of undelivered cargo", func(t *testing.T) {
		cargo := createTestCargo(t)

		assert.True(t, cargo.CanBeRerouted())
	})

	t.Run("should not allow rerouting of delivered cargo", func(t *testing.T) {
		cargo := createTestCargo(t)

		// Mark cargo as delivered
		deliveredStatus, err := NewDelivery(TransportStatusClaimed, RoutingStatusRouted, "SEGOT", "", true)
		require.NoError(t, err)
		cargo.Data.Delivery = deliveredStatus

		assert.False(t, cargo.CanBeRerouted())
	})

	t.Run("should not allow rerouting of claimed cargo", func(t *testing.T) {
		cargo := createTestCargo(t)

		// Mark cargo as claimed
		claimedStatus, err := NewDelivery(TransportStatusClaimed, RoutingStatusRouted, "SEGOT", "", false)
		require.NoError(t, err)
		cargo.Data.Delivery = claimedStatus

		assert.False(t, cargo.CanBeRerouted())
	})
}

func TestCargo_IsReadyForPickup(t *testing.T) {
	t.Run("should be ready for pickup when routed and not received", func(t *testing.T) {
		cargo := createTestCargo(t)
		itinerary := createTestItinerary(t, cargo.GetRouteSpecification())
		err := cargo.AssignToRoute(itinerary)
		require.NoError(t, err)

		assert.True(t, cargo.IsReadyForPickup())
	})

	t.Run("should not be ready if not routed", func(t *testing.T) {
		cargo := createTestCargo(t)

		assert.False(t, cargo.IsReadyForPickup())
	})

	t.Run("should not be ready if already received", func(t *testing.T) {
		cargo := createTestCargo(t)
		itinerary := createTestItinerary(t, cargo.GetRouteSpecification())
		err := cargo.AssignToRoute(itinerary)
		require.NoError(t, err)

		// Mark as received
		receivedStatus, err := NewDelivery(TransportStatusInPort, RoutingStatusRouted, "USNYC", "", false)
		require.NoError(t, err)
		cargo.Data.Delivery = receivedStatus

		assert.False(t, cargo.IsReadyForPickup())
	})
}

func TestCargo_IsOverdue(t *testing.T) {
	t.Run("should be overdue if deadline passed and not delivered", func(t *testing.T) {
		origin := "USNYC"
		destination := "SEGOT"
		arrivalDeadline := time.Now().Add(-24 * time.Hour) // Yesterday

		// Create cargo with past deadline (bypassing validation for test)
		routeSpec := RouteSpecification{
			Origin:          origin,
			Destination:     destination,
			ArrivalDeadline: arrivalDeadline,
		}

		cargo, err := NewCargoFromExisting(
			NewTrackingId(),
			routeSpec,
			nil,
			NewInitialDelivery(),
		)
		require.NoError(t, err)

		assert.True(t, cargo.IsOverdue())
	})

	t.Run("should not be overdue if delivered", func(t *testing.T) {
		origin := "USNYC"
		destination := "SEGOT"
		arrivalDeadline := time.Now().Add(-24 * time.Hour) // Yesterday

		routeSpec := RouteSpecification{
			Origin:          origin,
			Destination:     destination,
			ArrivalDeadline: arrivalDeadline,
		}

		deliveredStatus, err := NewDelivery(TransportStatusClaimed, RoutingStatusRouted, "SEGOT", "", true)
		require.NoError(t, err)

		cargo, err := NewCargoFromExisting(
			NewTrackingId(),
			routeSpec,
			nil,
			deliveredStatus,
		)
		require.NoError(t, err)

		assert.False(t, cargo.IsOverdue())
	})
}

func TestCargo_GetEstimatedTimeOfArrival(t *testing.T) {
	t.Run("should return ETA when routed", func(t *testing.T) {
		cargo := createTestCargo(t)
		itinerary := createTestItinerary(t, cargo.GetRouteSpecification())
		err := cargo.AssignToRoute(itinerary)
		require.NoError(t, err)

		eta := cargo.GetEstimatedTimeOfArrival()

		assert.NotNil(t, eta)
		assert.Equal(t, itinerary.FinalArrivalTime(), *eta)
	})

	t.Run("should return nil when not routed", func(t *testing.T) {
		cargo := createTestCargo(t)

		eta := cargo.GetEstimatedTimeOfArrival()

		assert.Nil(t, eta)
	})
}

// Helper functions for tests

func createTestCargo(t *testing.T) *Cargo {
	origin := "USNYC"
	destination := "SEGOT"
	arrivalDeadline := time.Now().Add(30 * 24 * time.Hour)

	cargo, err := NewCargo(origin, destination, arrivalDeadline)
	require.NoError(t, err)
	return &cargo
}

func createTestItinerary(t *testing.T, routeSpec RouteSpecification) Itinerary {
	departureTime := time.Now().Add(24 * time.Hour)
	arrivalTime := departureTime.Add(24 * time.Hour)

	leg, err := NewLeg("V001", routeSpec.Origin, routeSpec.Destination, departureTime, arrivalTime)
	require.NoError(t, err)

	itinerary, err := NewItinerary([]Leg{leg})
	require.NoError(t, err)

	return itinerary
}
