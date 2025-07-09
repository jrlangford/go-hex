package mock

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"go_hex/internal/routing/application"
	"go_hex/internal/routing/domain"
	"go_hex/internal/routing/ports/primary"
	"go_hex/internal/routing/ports/secondary"
)

// MockRoutingApplication embeds the real application service but provides test data population capabilities
type MockRoutingApplication struct {
	*application.RoutingApplicationService
	logger       *slog.Logger
	random       *rand.Rand
	voyageRepo   secondary.VoyageRepository
	locationRepo secondary.LocationRepository
}

// NewMockRoutingApplication creates a mock routing application with embedded real application service
func NewMockRoutingApplication(
	voyageRepo secondary.VoyageRepository,
	locationRepo secondary.LocationRepository,
	logger *slog.Logger,
	seed int64,
) *MockRoutingApplication {
	realApp := application.NewRoutingApplicationService(voyageRepo, locationRepo, logger)

	return &MockRoutingApplication{
		RoutingApplicationService: realApp,
		logger:                    logger,
		random:                    rand.New(rand.NewSource(seed)),
		voyageRepo:                voyageRepo,
		locationRepo:              locationRepo,
	}
}

func (m *MockRoutingApplication) GenerateTestData() {
	m.logger.Info("Generating test data for routing application")

	// Generate standard locations
	locationSpecs := m.GenerateStandardLocations(10)
	m.logger.Info("Generated standard locations", "count", len(locationSpecs))

	// Populate test locations
	locations, err := m.PopulateTestLocations(context.Background(), locationSpecs)
	if err != nil {
		m.logger.Error("Failed to populate test locations", "error", err)
		return
	}

	m.logger.Info("Successfully populated test locations", "count", len(locations))

	// Populate test voyages
	voyages, err := m.PopulateTestVoyages(context.Background(), locations, 20)
	if err != nil {
		m.logger.Error("Failed to populate test voyages", "error", err)
		return
	}

	m.logger.Info("Successfully populated test voyages", "count", len(voyages))
}

// PopulateTestLocations creates test location data through the domain layer
func (m *MockRoutingApplication) PopulateTestLocations(ctx context.Context, locationSpecs []TestLocationSpec) ([]domain.Location, error) {
	m.logger.Info("Populating test locations", "count", len(locationSpecs))

	var locations []domain.Location

	for _, spec := range locationSpecs {
		// Create location through domain constructor to ensure business rules
		location, err := domain.NewLocation(spec.Code, spec.Name, spec.Country)
		if err != nil {
			m.logger.Error("Failed to create test location", "error", err, "code", spec.Code)
			return nil, fmt.Errorf("failed to create location %s: %w", spec.Code, err)
		}

		// Store through repository
		if err := m.locationRepo.Store(location); err != nil {
			m.logger.Error("Failed to store test location", "error", err, "code", spec.Code)
			return nil, fmt.Errorf("failed to store location %s: %w", spec.Code, err)
		}

		locations = append(locations, location)
		m.logger.Debug("Created test location", "code", spec.Code, "name", spec.Name)
	}

	m.logger.Info("Successfully populated test locations", "count", len(locations))
	return locations, nil
}

// PopulateTestVoyages creates test voyage data with current-relative dates
func (m *MockRoutingApplication) PopulateTestVoyages(ctx context.Context, locations []domain.Location, count int) ([]domain.Voyage, error) {
	m.logger.Info("Populating test voyages", "count", count, "available_locations", len(locations))

	if len(locations) < 2 {
		return nil, fmt.Errorf("need at least 2 locations to create voyages, got %d", len(locations))
	}

	var voyages []domain.Voyage

	for i := 0; i < count; i++ {
		// Generate voyage with 2-4 carrier movements
		movementCount := 2 + m.random.Intn(3) // 2-4 movements
		movements := m.generateCarrierMovements(locations, movementCount)

		// Create voyage through domain constructor
		voyage, err := domain.NewVoyage(movements)
		if err != nil {
			m.logger.Error("Failed to create test voyage", "error", err)
			continue // Skip this voyage and try next
		}

		// Store through repository
		if err := m.voyageRepo.Store(voyage); err != nil {
			m.logger.Error("Failed to store test voyage", "error", err, "voyageNumber", voyage.GetVoyageNumber())
			continue
		}

		voyages = append(voyages, voyage)
		m.logger.Debug("Created test voyage", "voyageNumber", voyage.GetVoyageNumber(), "movements", len(movements))
	}

	m.logger.Info("Successfully populated test voyages", "count", len(voyages))
	return voyages, nil
}

// generateCarrierMovements creates a realistic sequence of carrier movements
func (m *MockRoutingApplication) generateCarrierMovements(locations []domain.Location, count int) []domain.CarrierMovement {
	if len(locations) < 2 || count < 1 {
		return nil
	}

	movements := make([]domain.CarrierMovement, 0, count)

	// Pick a random starting location
	currentLocationIdx := m.random.Intn(len(locations))

	// Calculate safe starting time to ensure ALL movements end in the past
	// Start time: random time in the past (2-7 days ago) to ensure completed journeys
	// We need to account for the total duration of all movements
	maxMovementDuration := time.Duration(count*(6+16)) * time.Hour // Max port time + max travel time per movement
	safeStartDaysAgo := 2 + m.random.Intn(6)                       // 2-7 days ago
	baseTime := time.Now().AddDate(0, 0, -safeStartDaysAgo).Add(-maxMovementDuration)
	currentTime := baseTime

	for i := 0; i < count; i++ {
		// Pick next location (different from current)
		nextLocationIdx := m.random.Intn(len(locations))
		for nextLocationIdx == currentLocationIdx && len(locations) > 1 {
			nextLocationIdx = m.random.Intn(len(locations))
		}

		fromLocation := locations[currentLocationIdx]
		toLocation := locations[nextLocationIdx]

		// Departure time (with small buffer from arrival at current location)
		departureTime := currentTime.Add(time.Duration(2+m.random.Intn(4)) * time.Hour) // 2-6 hours port time

		// Travel time: 4-16 hours depending on distance
		travelHours := 4 + m.random.Intn(13) // 4-16 hours
		arrivalTime := departureTime.Add(time.Duration(travelHours) * time.Hour)

		// Create movement through domain constructor
		movement, err := domain.NewCarrierMovement(
			fromLocation.GetUnLocode(),
			toLocation.GetUnLocode(),
			departureTime,
			arrivalTime,
		)
		if err != nil {
			m.logger.Error("Failed to create carrier movement", "error", err)
			continue
		}

		movements = append(movements, movement)

		// Update for next iteration
		currentLocationIdx = nextLocationIdx
		currentTime = arrivalTime
	}

	return movements
}

// GenerateStandardLocations creates a standard set of European maritime locations
func (m *MockRoutingApplication) GenerateStandardLocations(count int) []TestLocationSpec {
	standardSpecs := []TestLocationSpec{
		{"SESTO", "Stockholm", "SE"},
		{"FIHEL", "Helsinki", "FI"},
		{"DEHAM", "Hamburg", "DE"},
		{"DKCPH", "Copenhagen", "DK"},
		{"NLRTM", "Rotterdam", "NL"},
		{"GBLON", "London", "GB"},
		{"FRPAR", "Paris", "FR"},
		{"ESBAR", "Barcelona", "ES"},
		{"ITGEN", "Genoa", "IT"},
		{"SEGOT", "Gothenburg", "SE"},
		{"NOTOS", "Tønsberg", "NO"},
		{"BEBRU", "Brussels", "BE"},
	}

	if count > len(standardSpecs) {
		count = len(standardSpecs)
	}

	// Shuffle for variety
	m.random.Shuffle(len(standardSpecs), func(i, j int) {
		standardSpecs[i], standardSpecs[j] = standardSpecs[j], standardSpecs[i]
	})

	return standardSpecs[:count]
}

// TestLocationSpec defines the specification for creating a test location
type TestLocationSpec struct {
	Code    string
	Name    string
	Country string
}

// Ensure MockRoutingApplication implements primary ports
var _ primary.RouteFinder = (*MockRoutingApplication)(nil)
