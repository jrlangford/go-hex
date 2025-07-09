package in_memory_voyage_repo

import (
	"fmt"
	"sync"

	"go_hex/internal/routing/ports/routingsecondary"
	"go_hex/internal/routing/routingdomain"
)

// InMemoryVoyageRepository provides an in-memory implementation of the VoyageRepository
type InMemoryVoyageRepository struct {
	voyages map[string]routingdomain.Voyage
	mutex   sync.RWMutex
}

// NewInMemoryVoyageRepository creates a new in-memory voyage repository
func NewInMemoryVoyageRepository() routingsecondary.VoyageRepository {
	return &InMemoryVoyageRepository{
		voyages: make(map[string]routingdomain.Voyage),
	}
}

// Store saves a voyage to the repository
func (r *InMemoryVoyageRepository) Store(voyage routingdomain.Voyage) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	voyageNumber := voyage.GetVoyageNumber().String()
	r.voyages[voyageNumber] = voyage
	return nil
}

// FindByVoyageNumber retrieves a voyage by its voyage number
func (r *InMemoryVoyageRepository) FindByVoyageNumber(voyageNumber routingdomain.VoyageNumber) (routingdomain.Voyage, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	voyage, exists := r.voyages[voyageNumber.String()]
	if !exists {
		return routingdomain.Voyage{}, fmt.Errorf("voyage with number %s not found", voyageNumber.String())
	}
	return voyage, nil
}

// FindAll retrieves all voyages in the repository
func (r *InMemoryVoyageRepository) FindAll() ([]routingdomain.Voyage, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	voyages := make([]routingdomain.Voyage, 0, len(r.voyages))
	for _, voyage := range r.voyages {
		voyages = append(voyages, voyage)
	}
	return voyages, nil
}

// FindVoyagesConnecting finds voyages that connect two locations
func (r *InMemoryVoyageRepository) FindVoyagesConnecting(origin, destination routingdomain.UnLocode) ([]routingdomain.Voyage, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var connectingVoyages []routingdomain.Voyage
	for _, voyage := range r.voyages {
		schedule := voyage.GetSchedule()
		// Check if voyage starts at origin and ends at destination
		if schedule.InitialDepartureLocation() == origin && schedule.FinalArrivalLocation() == destination {
			connectingVoyages = append(connectingVoyages, voyage)
			continue
		}

		// Check if voyage has any movement from origin to destination
		for _, movement := range schedule.Movements {
			if movement.DepartureLocation == origin && movement.ArrivalLocation == destination {
				connectingVoyages = append(connectingVoyages, voyage)
				break
			}
		}
	}
	return connectingVoyages, nil
}
