package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDelivery(t *testing.T) {
	t.Run("should create delivery with valid parameters", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)

		require.NoError(t, err)
		assert.Equal(t, TransportStatusInPort, delivery.TransportStatus)
		assert.Equal(t, RoutingStatusRouted, delivery.RoutingStatus)
		assert.Equal(t, "USNYC", delivery.LastKnownLocation)
		assert.Equal(t, "V001", delivery.CurrentVoyage)
		assert.False(t, delivery.IsUnloadedAtDest)
		assert.False(t, delivery.CalculatedAt.IsZero())
	})
}

func TestNewInitialDelivery(t *testing.T) {
	t.Run("should create initial delivery status", func(t *testing.T) {
		delivery := NewInitialDelivery()

		assert.Equal(t, TransportStatusNotReceived, delivery.TransportStatus)
		assert.Equal(t, RoutingStatusNotRouted, delivery.RoutingStatus)
		assert.Empty(t, delivery.LastKnownLocation)
		assert.Empty(t, delivery.CurrentVoyage)
		assert.False(t, delivery.IsUnloadedAtDest)
		assert.False(t, delivery.CalculatedAt.IsZero())
	})
}

func TestDelivery_IsDelivered(t *testing.T) {
	t.Run("should return true when cargo is claimed and unloaded at destination", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusClaimed,
			RoutingStatusRouted,
			"USNYC",
			"",
			true,
		)
		require.NoError(t, err)

		assert.True(t, delivery.IsDelivered())
	})

	t.Run("should return false when cargo is claimed but not unloaded at destination", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusClaimed,
			RoutingStatusRouted,
			"DEHAM",
			"",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.IsDelivered())
	})

	t.Run("should return false when cargo is unloaded at destination but not claimed", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"",
			true,
		)
		require.NoError(t, err)

		assert.False(t, delivery.IsDelivered())
	})
}

func TestDelivery_IsOnTrack(t *testing.T) {
	t.Run("should return true when routing status is ROUTED", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.True(t, delivery.IsOnTrack())
	})

	t.Run("should return false when routing status is MISDIRECTED", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusMisdirected,
			"WRONG",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.IsOnTrack())
	})

	t.Run("should return false when routing status is NOT_ROUTED", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusNotReceived,
			RoutingStatusNotRouted,
			"",
			"",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.IsOnTrack())
	})
}

func TestDelivery_IsMisdirected(t *testing.T) {
	t.Run("should return true when routing status is MISDIRECTED", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusMisdirected,
			"WRONG",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.True(t, delivery.IsMisdirected())
	})

	t.Run("should return false when routing status is ROUTED", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.IsMisdirected())
	})
}

func TestDelivery_IsInTransit(t *testing.T) {
	t.Run("should return true when transport status is ONBOARD_CARRIER", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusOnboardCarrier,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.True(t, delivery.IsInTransit())
	})

	t.Run("should return false when transport status is IN_PORT", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.IsInTransit())
	})
}

func TestDelivery_IsAtPort(t *testing.T) {
	t.Run("should return true when transport status is IN_PORT", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.True(t, delivery.IsAtPort())
	})

	t.Run("should return false when transport status is ONBOARD_CARRIER", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusOnboardCarrier,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.IsAtPort())
	})
}

func TestDelivery_HasBeenReceived(t *testing.T) {
	t.Run("should return false when transport status is NOT_RECEIVED", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusNotReceived,
			RoutingStatusNotRouted,
			"",
			"",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.HasBeenReceived())
	})

	t.Run("should return true when transport status is IN_PORT", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.True(t, delivery.HasBeenReceived())
	})

	t.Run("should return true when transport status is ONBOARD_CARRIER", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusOnboardCarrier,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			false,
		)
		require.NoError(t, err)

		assert.True(t, delivery.HasBeenReceived())
	})
}

func TestDelivery_CanBeClaimed(t *testing.T) {
	t.Run("should return true when cargo is unloaded at destination and at port", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"USNYC",
			"",
			true,
		)
		require.NoError(t, err)

		assert.True(t, delivery.CanBeClaimed())
	})

	t.Run("should return false when cargo is not unloaded at destination", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusInPort,
			RoutingStatusRouted,
			"DEHAM",
			"",
			false,
		)
		require.NoError(t, err)

		assert.False(t, delivery.CanBeClaimed())
	})

	t.Run("should return false when cargo is unloaded at destination but not at port", func(t *testing.T) {
		delivery, err := NewDelivery(
			TransportStatusOnboardCarrier,
			RoutingStatusRouted,
			"USNYC",
			"V001",
			true,
		)
		require.NoError(t, err)

		assert.False(t, delivery.CanBeClaimed())
	})
}
