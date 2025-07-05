package in_memory_cargo_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/core/booking/domain"
	"go_hex/internal/core/booking/ports/secondary"
)

// InMemoryCargoRepository provides an in-memory implementation of the CargoRepository
type InMemoryCargoRepository struct {
	cargos map[string]domain.Cargo
	mutex  sync.RWMutex
}

// NewInMemoryCargoRepository creates a new in-memory cargo repository
func NewInMemoryCargoRepository() secondary.CargoRepository {
	return &InMemoryCargoRepository{
		cargos: make(map[string]domain.Cargo),
	}
}

// Store saves a cargo to the repository
func (r *InMemoryCargoRepository) Store(cargo domain.Cargo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	trackingId := cargo.GetTrackingId().String()
	r.cargos[trackingId] = cargo
	return nil
}

// FindByTrackingId retrieves a cargo by its tracking ID
func (r *InMemoryCargoRepository) FindByTrackingId(trackingId domain.TrackingId) (domain.Cargo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	cargo, exists := r.cargos[trackingId.String()]
	if !exists {
		return domain.Cargo{}, fmt.Errorf("cargo with tracking ID %s not found", trackingId.String())
	}
	return cargo, nil
}

// FindAll retrieves all cargos in the repository
func (r *InMemoryCargoRepository) FindAll() ([]domain.Cargo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	cargos := make([]domain.Cargo, 0, len(r.cargos))
	for _, cargo := range r.cargos {
		cargos = append(cargos, cargo)
	}
	return cargos, nil
}

// FindUnrouted retrieves all cargos that don't have an assigned itinerary
func (r *InMemoryCargoRepository) FindUnrouted() ([]domain.Cargo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var unroutedCargos []domain.Cargo
	for _, cargo := range r.cargos {
		if cargo.GetItinerary() == nil {
			unroutedCargos = append(unroutedCargos, cargo)
		}
	}
	return unroutedCargos, nil
}

// Update updates an existing cargo
func (r *InMemoryCargoRepository) Update(cargo domain.Cargo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	trackingId := cargo.GetTrackingId().String()
	r.cargos[trackingId] = cargo
	return nil
}
