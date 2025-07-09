// Package testdata provides utilities for generating test data for integration tests.
// All test data is generated dynamically to ensure consistency and avoid brittle tests.
package testdata

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"go_hex/internal/booking/bookingdomain"
	handlingDomain "go_hex/internal/handling/domain"
	routingDomain "go_hex/internal/routing/domain"
)

// TestDataGenerator provides methods to generate test data for all bounded contexts
type TestDataGenerator struct {
	random *rand.Rand
	logger *slog.Logger
}

// NewTestDataGenerator creates a new test data generator with optional seed for reproducibility
func NewTestDataGenerator(seed int64, logger *slog.Logger) *TestDataGenerator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &TestDataGenerator{
		random: rand.New(rand.NewSource(seed)),
		logger: logger,
	}
}

// Standard maritime locations (UN/LOCODEs) for realistic test data
var standardLocations = []struct {
	code    string
	name    string
	country string
}{
	{"SESTO", "Stockholm", "SE"},
	{"FIHEL", "Helsinki", "FI"},
	{"DEHAM", "Hamburg", "DE"},
	{"DKCPH", "Copenhagen", "DK"},
	{"NLRTM", "Rotterdam", "NL"},
	{"GBLON", "London", "GB"},
	{"FRPAR", "Paris", "FR"},
	{"ESBAR", "Barcelona", "ES"},
	{"ITGEN", "Genoa", "IT"},
	{"GRGOT", "Gothenburg", "SE"},
	{"NOTOS", "Tønsberg", "NO"},
	{"BEBRU", "Brussels", "BE"},
	{"CHZUR", "Zurich", "CH"},
	{"ATVIE", "Vienna", "AT"},
	{"PLGDA", "Gdansk", "PL"},
}

// GenerateLocations creates a collection of test locations
func (g *TestDataGenerator) GenerateLocations(count int) []routingDomain.Location {
	if count > len(standardLocations) {
		count = len(standardLocations)
	}

	locations := make([]routingDomain.Location, 0, count)

	// Shuffle the standard locations for variety
	shuffled := make([]int, len(standardLocations))
	for i := range shuffled {
		shuffled[i] = i
	}
	g.random.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i := 0; i < count; i++ {
		idx := shuffled[i]
		loc := standardLocations[idx]

		location, err := routingDomain.NewLocation(loc.code, loc.name, loc.country)
		if err != nil {
			g.logger.Error("Failed to create location", "error", err, "code", loc.code)
			continue
		}

		locations = append(locations, location)
	}

	return locations
}

// GenerateVoyages creates a collection of test voyages with realistic schedules
func (g *TestDataGenerator) GenerateVoyages(locations []routingDomain.Location, count int) []routingDomain.Voyage {
	if len(locations) < 2 {
		g.logger.Error("Need at least 2 locations to generate voyages")
		return []routingDomain.Voyage{}
	}

	voyages := make([]routingDomain.Voyage, 0, count)
	baseTime := time.Now().Add(24 * time.Hour) // Start voyages tomorrow

	for i := 0; i < count; i++ {
		voyage := g.generateSingleVoyage(locations, baseTime.Add(time.Duration(i*12)*time.Hour))
		if voyage != nil {
			voyages = append(voyages, *voyage)
		}
	}

	return voyages
}

// generateSingleVoyage creates a single voyage with multiple carrier movements
func (g *TestDataGenerator) generateSingleVoyage(locations []routingDomain.Location, startTime time.Time) *routingDomain.Voyage {
	// Generate 2-4 movements per voyage
	movementCount := 2 + g.random.Intn(3)
	movements := make([]routingDomain.CarrierMovement, 0, movementCount)

	// Create a path through different locations
	usedLocations := make(map[string]bool)
	currentTime := startTime

	var previousLocation *routingDomain.Location

	for i := 0; i < movementCount; i++ {
		var fromLocation, toLocation routingDomain.Location

		if previousLocation == nil {
			// First movement - pick random start location
			fromLocation = locations[g.random.Intn(len(locations))]
		} else {
			fromLocation = *previousLocation
		}

		// Pick destination that hasn't been used and isn't the same as origin
		attempts := 0
		for {
			toLocation = locations[g.random.Intn(len(locations))]
			if !usedLocations[toLocation.ID().String()] &&
				toLocation.ID().String() != fromLocation.ID().String() {
				break
			}
			attempts++
			if attempts > 10 {
				// Fallback to any different location
				for _, loc := range locations {
					if loc.ID().String() != fromLocation.ID().String() {
						toLocation = loc
						break
					}
				}
				break
			}
		}

		// Mark location as used
		usedLocations[toLocation.ID().String()] = true

		// Generate realistic travel times (4-16 hours between ports)
		travelHours := 4 + g.random.Intn(13)
		departureTime := currentTime
		arrivalTime := departureTime.Add(time.Duration(travelHours) * time.Hour)

		movement, err := routingDomain.NewCarrierMovement(
			fromLocation.ID(),
			toLocation.ID(),
			departureTime,
			arrivalTime,
		)

		if err != nil {
			g.logger.Error("Failed to create carrier movement", "error", err)
			return nil
		}

		movements = append(movements, movement)

		// Next movement starts after some port time (2-6 hours)
		portTimeHours := 2 + g.random.Intn(5)
		currentTime = arrivalTime.Add(time.Duration(portTimeHours) * time.Hour)
		previousLocation = &toLocation
	}

	voyage, err := routingDomain.NewVoyage(movements)
	if err != nil {
		g.logger.Error("Failed to create voyage", "error", err)
		return nil
	}

	return &voyage
}

// CargoTestData represents a complete cargo scenario for testing
type CargoTestData struct {
	Cargo           bookingdomain.Cargo
	Origin          string
	Destination     string
	ArrivalDeadline time.Time
	HandlingEvents  []HandlingEventData
}

// HandlingEventData represents a handling event for testing
type HandlingEventData struct {
	EventType      handlingDomain.HandlingEventType
	Location       string
	VoyageNumber   string
	CompletionTime time.Time
	Delay          time.Duration // Time to wait before applying this event
}

// GenerateCargoScenarios creates realistic cargo shipping scenarios for testing
func (g *TestDataGenerator) GenerateCargoScenarios(locations []routingDomain.Location, voyages []routingDomain.Voyage, count int) []CargoTestData {
	if len(locations) < 2 {
		g.logger.Error("Need at least 2 locations to generate cargo scenarios")
		return []CargoTestData{}
	}

	scenarios := make([]CargoTestData, 0, count)

	for i := 0; i < count; i++ {
		scenario := g.generateCargoScenario(locations, voyages, i)
		if scenario != nil {
			scenarios = append(scenarios, *scenario)
		}
	}

	return scenarios
}

// generateCargoScenario creates a single cargo scenario with associated handling events
func (g *TestDataGenerator) generateCargoScenario(locations []routingDomain.Location, voyages []routingDomain.Voyage, scenarioIndex int) *CargoTestData {
	// Pick random origin and destination
	originIdx := g.random.Intn(len(locations))
	var destIdx int
	for {
		destIdx = g.random.Intn(len(locations))
		if destIdx != originIdx {
			break
		}
	}

	origin := locations[originIdx].ID().String()
	destination := locations[destIdx].ID().String()

	// Generate arrival deadline 7-60 days in the future
	daysInFuture := 7 + g.random.Intn(54)
	arrivalDeadline := time.Now().Add(time.Duration(daysInFuture) * 24 * time.Hour)

	// Create cargo
	cargo, err := bookingdomain.NewCargo(origin, destination, arrivalDeadline)
	if err != nil {
		g.logger.Error("Failed to create cargo", "error", err)
		return nil
	}

	// Generate handling events for this cargo
	handlingEvents := g.generateHandlingEvents(origin, destination, voyages)

	return &CargoTestData{
		Cargo:           cargo,
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: arrivalDeadline,
		HandlingEvents:  handlingEvents,
	}
}

// generateHandlingEvents creates a sequence of realistic handling events
func (g *TestDataGenerator) generateHandlingEvents(origin, destination string, voyages []routingDomain.Voyage) []HandlingEventData {
	events := []HandlingEventData{}
	currentTime := time.Now().Add(-24 * time.Hour) // Start events 24 hours ago (well in the past)

	// Always start with RECEIVE at origin
	events = append(events, HandlingEventData{
		EventType:      handlingDomain.HandlingEventTypeReceive,
		Location:       origin,
		VoyageNumber:   "",
		CompletionTime: currentTime,
		Delay:          0,
	})

	// Add LOAD event
	if len(voyages) > 0 {
		selectedVoyage := voyages[g.random.Intn(len(voyages))]

		currentTime = currentTime.Add(time.Duration(2+g.random.Intn(4)) * time.Hour)
		events = append(events, HandlingEventData{
			EventType:      handlingDomain.HandlingEventTypeLoad,
			Location:       origin,
			VoyageNumber:   selectedVoyage.GetVoyageNumber().String(),
			CompletionTime: currentTime,
			Delay:          500 * time.Millisecond, // Small delay between events
		})

		// Add UNLOAD event if voyage has movements
		schedule := selectedVoyage.GetSchedule()
		if len(schedule.Movements) > 0 {
			// Pick a random movement for unloading
			movement := schedule.Movements[g.random.Intn(len(schedule.Movements))]
			unloadLocation := movement.ArrivalLocation.String()

			currentTime = currentTime.Add(time.Duration(4+g.random.Intn(8)) * time.Hour) // Use relative time instead of movement time
			events = append(events, HandlingEventData{
				EventType:      handlingDomain.HandlingEventTypeUnload,
				Location:       unloadLocation,
				VoyageNumber:   selectedVoyage.GetVoyageNumber().String(),
				CompletionTime: currentTime,
				Delay:          1 * time.Second,
			})

			// Add CLAIM event if unloaded at destination
			if unloadLocation == destination {
				currentTime = currentTime.Add(time.Duration(1+g.random.Intn(6)) * time.Hour)
				events = append(events, HandlingEventData{
					EventType:      handlingDomain.HandlingEventTypeClaim,
					Location:       destination,
					VoyageNumber:   "",
					CompletionTime: currentTime,
					Delay:          1500 * time.Millisecond,
				})
			}
		}
	}

	return events
}

// TestDataSet represents a complete set of test data for integration tests
type TestDataSet struct {
	Locations      []routingDomain.Location
	Voyages        []routingDomain.Voyage
	CargoScenarios []CargoTestData
	Seed           int64
	GeneratedAt    time.Time
}

// GenerateCompleteTestDataSet creates a full set of test data for integration tests
func (g *TestDataGenerator) GenerateCompleteTestDataSet() TestDataSet {
	g.logger.Info("Generating complete test data set")

	// Generate 8-12 locations
	locationCount := 8 + g.random.Intn(5)
	locations := g.GenerateLocations(locationCount)
	g.logger.Info("Generated locations", "count", len(locations))

	// Generate 5-10 voyages
	voyageCount := 5 + g.random.Intn(6)
	voyages := g.GenerateVoyages(locations, voyageCount)
	g.logger.Info("Generated voyages", "count", len(voyages))

	// Generate 3-7 cargo scenarios
	cargoCount := 3 + g.random.Intn(5)
	cargoScenarios := g.GenerateCargoScenarios(locations, voyages, cargoCount)
	g.logger.Info("Generated cargo scenarios", "count", len(cargoScenarios))

	return TestDataSet{
		Locations:      locations,
		Voyages:        voyages,
		CargoScenarios: cargoScenarios,
		Seed:           g.random.Int63(),
		GeneratedAt:    time.Now(),
	}
}

// PopulateRepositories fills the provided repositories with generated test data
func (dataset *TestDataSet) PopulateRepositories(ctx context.Context, repos TestRepositories) error {
	// Populate locations
	for _, location := range dataset.Locations {
		if err := repos.LocationRepo.Store(location); err != nil {
			return fmt.Errorf("failed to store location %s: %w", location.ID().String(), err)
		}
	}

	// Populate voyages
	for _, voyage := range dataset.Voyages {
		if err := repos.VoyageRepo.Store(voyage); err != nil {
			return fmt.Errorf("failed to store voyage %s: %w", voyage.GetVoyageNumber().String(), err)
		}
	}

	// Populate cargo (repositories will be populated as cargo is booked during tests)
	// This is intentionally left empty as cargo should be created through application services

	return nil
}

// TestRepositories holds references to test repositories for data population
type TestRepositories struct {
	LocationRepo interface {
		Store(routingDomain.Location) error
	}
	VoyageRepo interface {
		Store(routingDomain.Voyage) error
	}
	CargoRepo interface {
		Store(bookingdomain.Cargo) error
	}
	HandlingEventRepo interface {
		Store(handlingDomain.HandlingEvent) error
	}
}
