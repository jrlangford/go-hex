package in_memory_voyage_repo

import (
	"fmt"
	"sync"
	"time"

	"go_hex/internal/core/routing/domain"
	"go_hex/internal/core/routing/ports/secondary"
)

// InMemoryVoyageRepository provides an in-memory implementation of the VoyageRepository
type InMemoryVoyageRepository struct {
	voyages map[string]domain.Voyage
	mutex   sync.RWMutex
}

// NewInMemoryVoyageRepository creates a new in-memory voyage repository
func NewInMemoryVoyageRepository() secondary.VoyageRepository {
	repo := &InMemoryVoyageRepository{
		voyages: make(map[string]domain.Voyage),
	}

	// Seed with some sample data for testing
	repo.seedSampleData()

	return repo
}

// Store saves a voyage to the repository
func (r *InMemoryVoyageRepository) Store(voyage domain.Voyage) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	voyageNumber := voyage.GetVoyageNumber().String()
	r.voyages[voyageNumber] = voyage
	return nil
}

// FindByVoyageNumber retrieves a voyage by its voyage number
func (r *InMemoryVoyageRepository) FindByVoyageNumber(voyageNumber domain.VoyageNumber) (domain.Voyage, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	voyage, exists := r.voyages[voyageNumber.String()]
	if !exists {
		return domain.Voyage{}, fmt.Errorf("voyage with number %s not found", voyageNumber.String())
	}
	return voyage, nil
}

// FindAll retrieves all voyages in the repository
func (r *InMemoryVoyageRepository) FindAll() ([]domain.Voyage, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	voyages := make([]domain.Voyage, 0, len(r.voyages))
	for _, voyage := range r.voyages {
		voyages = append(voyages, voyage)
	}
	return voyages, nil
}

// FindVoyagesConnecting finds voyages that connect two locations
func (r *InMemoryVoyageRepository) FindVoyagesConnecting(origin, destination domain.UnLocode) ([]domain.Voyage, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var connectingVoyages []domain.Voyage
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

// seedSampleData adds some sample voyages for testing and demonstration
func (r *InMemoryVoyageRepository) seedSampleData() {
	// Create sample locations
	stockholm, _ := domain.NewUnLocode("SESTO")
	helsinki, _ := domain.NewUnLocode("FIHEL")
	hamburg, _ := domain.NewUnLocode("DEHAM")
	copenhagen, _ := domain.NewUnLocode("DKCPH")
	rotterdam, _ := domain.NewUnLocode("NLRTM")

	// Parse times for carrier movements
	time1, _ := time.Parse(time.RFC3339, "2024-01-15T08:00:00Z")
	time2, _ := time.Parse(time.RFC3339, "2024-01-15T14:00:00Z")
	time3, _ := time.Parse(time.RFC3339, "2024-01-15T16:00:00Z")
	time4, _ := time.Parse(time.RFC3339, "2024-01-16T10:00:00Z")
	time5, _ := time.Parse(time.RFC3339, "2024-01-16T12:00:00Z")
	time6, _ := time.Parse(time.RFC3339, "2024-01-16T18:00:00Z")
	time7, _ := time.Parse(time.RFC3339, "2024-01-16T20:00:00Z")
	time8, _ := time.Parse(time.RFC3339, "2024-01-17T08:00:00Z")
	time9, _ := time.Parse(time.RFC3339, "2024-01-17T10:00:00Z")
	time10, _ := time.Parse(time.RFC3339, "2024-01-17T20:00:00Z")
	time11, _ := time.Parse(time.RFC3339, "2024-01-17T22:00:00Z")
	time12, _ := time.Parse(time.RFC3339, "2024-01-18T06:00:00Z")
	time13, _ := time.Parse(time.RFC3339, "2024-01-18T08:00:00Z")
	time14, _ := time.Parse(time.RFC3339, "2024-01-18T16:00:00Z")

	// Create carrier movements
	mov1, _ := domain.NewCarrierMovement(stockholm, helsinki, time1, time2)
	mov2, _ := domain.NewCarrierMovement(helsinki, hamburg, time3, time4)
	mov3, _ := domain.NewCarrierMovement(hamburg, copenhagen, time5, time6)
	mov4, _ := domain.NewCarrierMovement(copenhagen, rotterdam, time7, time8)
	mov5, _ := domain.NewCarrierMovement(stockholm, copenhagen, time9, time10)
	mov6, _ := domain.NewCarrierMovement(copenhagen, hamburg, time11, time12)
	mov7, _ := domain.NewCarrierMovement(hamburg, rotterdam, time13, time14)

	// Create sample voyages
	voyage1, _ := domain.NewVoyage([]domain.CarrierMovement{mov1, mov2})
	voyage2, _ := domain.NewVoyage([]domain.CarrierMovement{mov3, mov4})
	voyage3, _ := domain.NewVoyage([]domain.CarrierMovement{mov5, mov6, mov7})

	r.voyages[voyage1.GetVoyageNumber().String()] = voyage1
	r.voyages[voyage2.GetVoyageNumber().String()] = voyage2
	r.voyages[voyage3.GetVoyageNumber().String()] = voyage3
}
