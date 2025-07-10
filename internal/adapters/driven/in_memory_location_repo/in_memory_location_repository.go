package in_memory_location_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/routing/ports/routingsecondary"
	"go_hex/internal/routing/routingdomain"
)

// InMemoryLocationRepository provides an in-memory implementation of the LocationRepository
type InMemoryLocationRepository struct {
	locations map[string]routingdomain.Location
	mutex     sync.RWMutex
}

// NewInMemoryLocationRepository creates a new in-memory location repository
func NewInMemoryLocationRepository() routingsecondary.LocationRepository {
	return &InMemoryLocationRepository{
		locations: make(map[string]routingdomain.Location),
	}
}

// Store saves a location to the repository
func (r *InMemoryLocationRepository) Store(location routingdomain.Location) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	unLocode := location.GetUnLocode().String()
	r.locations[unLocode] = location
	return nil
}

// FindByUnLocode retrieves a location by its UN/LOCODE
func (r *InMemoryLocationRepository) FindByUnLocode(unLocode routingdomain.UnLocode) (routingdomain.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	location, exists := r.locations[unLocode.String()]
	if !exists {
		return routingdomain.Location{}, fmt.Errorf("location with UN/LOCODE %s not found", unLocode.String())
	}
	return location, nil
}

// FindAll retrieves all locations in the repository
func (r *InMemoryLocationRepository) FindAll() ([]routingdomain.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	locations := make([]routingdomain.Location, 0, len(r.locations))
	for _, location := range r.locations {
		locations = append(locations, location)
	}
	return locations, nil
}
