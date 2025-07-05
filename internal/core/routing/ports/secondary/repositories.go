package secondary

import (
	"go_hex/internal/core/routing/domain"
)

// VoyageRepository defines the secondary port for voyage persistence
type VoyageRepository interface {
	// Store persists a voyage
	Store(voyage domain.Voyage) error

	// FindByVoyageNumber retrieves a voyage by its number
	FindByVoyageNumber(voyageNumber domain.VoyageNumber) (domain.Voyage, error)

	// FindAll retrieves all voyages
	FindAll() ([]domain.Voyage, error)

	// FindVoyagesConnecting finds voyages that connect two locations
	FindVoyagesConnecting(origin, destination domain.UnLocode) ([]domain.Voyage, error)
}

// LocationRepository defines the secondary port for location persistence
type LocationRepository interface {
	// Store persists a location
	Store(location domain.Location) error

	// FindByUnLocode retrieves a location by its UN/LOCODE
	FindByUnLocode(unLocode domain.UnLocode) (domain.Location, error)

	// FindAll retrieves all locations
	FindAll() ([]domain.Location, error)
}
