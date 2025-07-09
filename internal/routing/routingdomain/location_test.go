package routingdomain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUnLocode(t *testing.T) {
	t.Run("should create UnLocode with valid 5-character code", func(t *testing.T) {
		code := "USNYC"

		unLocode, err := NewUnLocode(code)

		require.NoError(t, err)
		assert.Equal(t, code, unLocode.Code)
		assert.Equal(t, code, unLocode.String())
	})

	t.Run("should fail with code shorter than 5 characters", func(t *testing.T) {
		code := "US"

		_, err := NewUnLocode(code)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UN/LOCODE format")
	})

	t.Run("should fail with code longer than 5 characters", func(t *testing.T) {
		code := "USNEWYORK"

		_, err := NewUnLocode(code)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UN/LOCODE format")
	})

	t.Run("should fail with empty code", func(t *testing.T) {
		code := ""

		_, err := NewUnLocode(code)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UN/LOCODE format")
	})
}

func TestNewLocation(t *testing.T) {
	t.Run("should create location with valid parameters", func(t *testing.T) {
		code := "USNYC"
		name := "New York"
		country := "US"

		location, err := NewLocation(code, name, country)

		require.NoError(t, err)
		assert.Equal(t, code, location.GetUnLocode().String())
		assert.Equal(t, name, location.GetName())
		assert.Equal(t, country, location.GetCountry())
	})

	t.Run("should fail with invalid UN/LOCODE", func(t *testing.T) {
		code := "US" // Too short
		name := "New York"
		country := "US"

		_, err := NewLocation(code, name, country)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UN/LOCODE format")
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		code := "USNYC"
		name := ""
		country := "US"

		_, err := NewLocation(code, name, country)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location data validation failed")
	})

	t.Run("should fail with name too long", func(t *testing.T) {
		code := "USNYC"
		name := "This is a very long location name that exceeds the maximum allowed length of 100 characters for a location name"
		country := "US"

		_, err := NewLocation(code, name, country)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location data validation failed")
	})

	t.Run("should fail with invalid country code length", func(t *testing.T) {
		code := "USNYC"
		name := "New York"
		country := "USA" // Should be 2 characters

		_, err := NewLocation(code, name, country)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location data validation failed")
	})

	t.Run("should fail with empty country code", func(t *testing.T) {
		code := "USNYC"
		name := "New York"
		country := ""

		_, err := NewLocation(code, name, country)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location data validation failed")
	})
}

func TestLocation_GetMethods(t *testing.T) {
	code := "DEHAM"
	name := "Hamburg"
	country := "DE"

	location, err := NewLocation(code, name, country)
	require.NoError(t, err)

	t.Run("should return correct UnLocode", func(t *testing.T) {
		assert.Equal(t, code, location.GetUnLocode().String())
	})

	t.Run("should return correct name", func(t *testing.T) {
		assert.Equal(t, name, location.GetName())
	})

	t.Run("should return correct country", func(t *testing.T) {
		assert.Equal(t, country, location.GetCountry())
	})
}
