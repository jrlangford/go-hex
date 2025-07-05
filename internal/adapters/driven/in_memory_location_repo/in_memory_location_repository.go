package in_memory_location_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/core/routing/domain"
	"go_hex/internal/core/routing/ports/secondary"
)

// InMemoryLocationRepository provides an in-memory implementation of the LocationRepository
type InMemoryLocationRepository struct {
	locations map[string]domain.Location
	mutex     sync.RWMutex
}

// NewInMemoryLocationRepository creates a new in-memory location repository
func NewInMemoryLocationRepository() secondary.LocationRepository {
	repo := &InMemoryLocationRepository{
		locations: make(map[string]domain.Location),
	}

	// Seed with some sample data for testing
	repo.seedSampleData()

	return repo
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

// seedSampleData adds some sample locations for testing and demonstration
func (r *InMemoryLocationRepository) seedSampleData() {
	// Create sample locations
	loc1, _ := domain.NewLocation("SESTO", "Stockholm", "SE")
	loc2, _ := domain.NewLocation("FIHEL", "Helsinki", "FI")
	loc3, _ := domain.NewLocation("DEHAM", "Hamburg", "DE")
	loc4, _ := domain.NewLocation("DKCPH", "Copenhagen", "DK")
	loc5, _ := domain.NewLocation("NLRTM", "Rotterdam", "NL")

	r.locations[loc1.GetUnLocode().String()] = loc1
	r.locations[loc2.GetUnLocode().String()] = loc2
	r.locations[loc3.GetUnLocode().String()] = loc3
	r.locations[loc4.GetUnLocode().String()] = loc4
	r.locations[loc5.GetUnLocode().String()] = loc5
}
