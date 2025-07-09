package in_memory_location_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/routing/domain"
	"go_hex/internal/routing/ports/secondary"
)

// InMemoryLocationRepository provides an in-memory implementation of the LocationRepository
type InMemoryLocationRepository struct {
	locations map[string]domain.Location
	mutex     sync.RWMutex
}

// NewInMemoryLocationRepository creates a new in-memory location repository
func NewInMemoryLocationRepository() secondary.LocationRepository {
	return &InMemoryLocationRepository{
		locations: make(map[string]domain.Location),
	}
}

// Store saves a location to the repository
func (r *InMemoryLocationRepository) Store(location domain.Location) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	unLocode := location.GetUnLocode().String()
	r.locations[unLocode] = location
	return nil
}

// FindByUnLocode retrieves a location by its UN/LOCODE
func (r *InMemoryLocationRepository) FindByUnLocode(unLocode domain.UnLocode) (domain.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	location, exists := r.locations[unLocode.String()]
	if !exists {
		return domain.Location{}, fmt.Errorf("location with UN/LOCODE %s not found", unLocode.String())
	}
	return location, nil
}

// FindAll retrieves all locations in the repository
func (r *InMemoryLocationRepository) FindAll() ([]domain.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	locations := make([]domain.Location, 0, len(r.locations))
	for _, location := range r.locations {
		locations = append(locations, location)
	}
	return locations, nil
}
