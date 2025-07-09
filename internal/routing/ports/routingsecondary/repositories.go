package routingsecondary

import (
	"go_hex/internal/routing/routingdomain"
)

// VoyageRepository defines the secondary port for voyage persistence
type VoyageRepository interface {
	// Store persists a voyage
	Store(voyage routingdomain.Voyage) error

	// FindByVoyageNumber retrieves a voyage by its number
	FindByVoyageNumber(voyageNumber routingdomain.VoyageNumber) (routingdomain.Voyage, error)

	// FindAll retrieves all voyages
	FindAll() ([]routingdomain.Voyage, error)

	// FindVoyagesConnecting finds voyages that connect two locations
	FindVoyagesConnecting(origin, destination routingdomain.UnLocode) ([]routingdomain.Voyage, error)
}

// LocationRepository defines the secondary port for location persistence
type LocationRepository interface {
	// Store persists a location
	Store(location routingdomain.Location) error

	// FindByUnLocode retrieves a location by its UN/LOCODE
	FindByUnLocode(unLocode routingdomain.UnLocode) (routingdomain.Location, error)

	// FindAll retrieves all locations
	FindAll() ([]routingdomain.Location, error)
}
