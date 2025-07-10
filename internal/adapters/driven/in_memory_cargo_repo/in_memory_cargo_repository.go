package in_memory_cargo_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/booking/bookingdomain"
	"go_hex/internal/booking/ports/bookingsecondary"
)

// InMemoryCargoRepository provides an in-memory implementation of the CargoRepository
type InMemoryCargoRepository struct {
	cargos map[string]bookingdomain.Cargo
	mutex  sync.RWMutex
}

// NewInMemoryCargoRepository creates a new in-memory cargo repository
func NewInMemoryCargoRepository() bookingsecondary.CargoRepository {
	return &InMemoryCargoRepository{
		cargos: make(map[string]bookingdomain.Cargo),
	}
}

// Store saves a cargo to the repository
func (r *InMemoryCargoRepository) Store(cargo bookingdomain.Cargo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	trackingId := cargo.GetTrackingId().String()
	r.cargos[trackingId] = cargo
	return nil
}

// FindByTrackingId retrieves a cargo by its tracking ID
func (r *InMemoryCargoRepository) FindByTrackingId(trackingId bookingdomain.TrackingId) (bookingdomain.Cargo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	cargo, exists := r.cargos[trackingId.String()]
	if !exists {
		return bookingdomain.Cargo{}, fmt.Errorf("cargo with tracking ID %s not found", trackingId.String())
	}
	return cargo, nil
}

// FindAll retrieves all cargos in the repository
func (r *InMemoryCargoRepository) FindAll() ([]bookingdomain.Cargo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	cargos := make([]bookingdomain.Cargo, 0, len(r.cargos))
	for _, cargo := range r.cargos {
		cargos = append(cargos, cargo)
	}
	return cargos, nil
}

// FindUnrouted retrieves all cargos that don't have an assigned itinerary
func (r *InMemoryCargoRepository) FindUnrouted() ([]bookingdomain.Cargo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var unroutedCargos []bookingdomain.Cargo
	for _, cargo := range r.cargos {
		if cargo.GetItinerary() == nil {
			unroutedCargos = append(unroutedCargos, cargo)
		}
	}
	return unroutedCargos, nil
}

// Update updates an existing cargo
func (r *InMemoryCargoRepository) Update(cargo bookingdomain.Cargo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	trackingId := cargo.GetTrackingId().String()
	r.cargos[trackingId] = cargo
	return nil
}
