# Mock Data Generation Strategy - New Implementation

## Overview

This document describes the refactored mock data generation strategy that follows Domain-Driven Design principles and ensures proper separation of concerns.

## 🎯 Requirements Fulfilled

✅ **In-memory repositories are pure** - No mock data hardcoded in repositories  
✅ **Dynamic date generation** - All dates generated relative to current date  
✅ **Application layer methods** - All entities created through proper business logic  
✅ **Mock applications** - Each bounded context has test-enabled application services  

## 🏗️ New Architecture

### Pure Repositories
- `InMemoryCargoRepository` - Clean, no seed data
- `InMemoryVoyageRepository` - Clean, no seed data  
- `InMemoryLocationRepository` - Clean, no seed data
- `InMemoryHandlingEventRepository` - Clean, no seed data

### Mock Applications
Each bounded context now has a mock application that embeds the real application service:

#### Booking Context: `MockBookingApplication`
- Embeds real `BookingApplicationService`
- Provides `PopulateTestCargo()` method
- Generates cargo scenarios with future deadlines (7-60 days)
- Uses proper authentication context for test operations

#### Routing Context: `MockRoutingApplication`  
- Embeds real `RoutingApplicationService`
- Provides `PopulateTestLocations()` and `PopulateTestVoyages()` methods
- Generates voyages with past completion times (1-7 days ago)
- Creates realistic carrier movement sequences

#### Handling Context: `MockHandlingApplication`
- Embeds real `HandlingReportService`
- Provides `PopulateTestHandlingEvents()` method
- Generates events with past completion times (24-96 hours ago)
- Creates realistic RECEIVE → LOAD → UNLOAD → CLAIM sequences

### Unified Test Environment: `MockTestEnvironment`
- Orchestrates all mock applications
- Ensures proper data creation order:
  1. Locations (through routing app)
  2. Voyages (through routing app)
  3. Cargo (through booking app)
  4. Handling events (through handling app)

## 🕒 Dynamic Date Generation

### Past Events (Completed)
- **Voyage movements**: 1-7 days ago
- **Handling events**: 24-96 hours ago (within 30-day validation window)

### Future Events (Planned)  
- **Cargo arrival deadlines**: 7-60 days in the future
- **Route specifications**: Always future-dated

### Current Validation
- All dates respect domain validation rules
- No hardcoded timestamps
- Seed-based reproducibility maintained

## 🔄 Business Logic Compliance

### Cargo Creation
```go
// Uses real application service
cargo, err := mockBookingApp.BookNewCargo(
    testCtx,
    scenario.Origin,
    scenario.Destination,
    arrivalDeadline.Format(time.RFC3339),
)
```

### Handling Events
```go
// Uses real application service
report := primary.HandlingReport{
    TrackingId:     scenario.TrackingId,
    EventType:      string(eventSpec.EventType),
    Location:       eventSpec.Location,
    VoyageNumber:   eventSpec.VoyageNumber,
    CompletionTime: completionTime.Format(time.RFC3339),
}
err := mockHandlingApp.SubmitHandlingReport(testCtx, report)
```

### Domain Events
- All domain events are properly published
- Event side effects trigger cross-context communication
- Integration patterns (Customer-Supplier, etc.) maintained

## 🧪 Usage in Tests

### Basic Setup
```go
mockEnv, err := mock.NewMockTestEnvironment(seed, logger)
if err != nil {
    t.Fatalf("Failed to create mock test environment: %v", err)
}

ctx := context.Background()
err = mockEnv.PopulateTestData(ctx)
if err != nil {
    t.Fatalf("Failed to populate test data: %v", err)
}
```

### Reproducible Testing
```go
seed := int64(12345) // Fixed seed for reproducible tests
mockEnv, _ := mock.NewMockTestEnvironment(seed, logger)
```

### Random Testing
```go
seed := time.Now().UnixNano() // Random seed for variety
mockEnv, _ := mock.NewMockTestEnvironment(seed, logger)
```

## 📊 Generated Data Characteristics

### Typical Test Run
- **8-12 locations** (European maritime ports)
- **5-10 voyages** with 2-4 movements each
- **3-7 cargo scenarios** with handling histories
- **15-25 handling events** (realistic sequences)

### Data Quality
- **Valid UN/LOCODE** location codes (SESTO, DEHAM, etc.)
- **Realistic timing** between movements and events
- **Proper sequencing** of handling events
- **Cross-referential integrity** between entities

## ✅ Benefits Achieved

### 1. **Domain Compliance**
- All data created through proper domain constructors
- Business rules enforced during generation
- Validation applied at every step

### 2. **Realistic Scenarios**  
- Dynamic date generation relative to current time
- Proper event sequencing and timing
- Maritime industry conventions followed

### 3. **Maintainability**
- Clean separation between repositories and test data
- Mock applications reuse real business logic
- Easy to extend with new scenarios

### 4. **Testability**
- Seed-based reproducibility
- Isolation between test runs
- Easy to generate various data volumes

## 🔄 Migration from Old System

### Before (Problems)
```go
// Hard-coded in repository constructor
func NewInMemoryLocationRepository() secondary.LocationRepository {
    repo := &InMemoryLocationRepository{
        locations: make(map[string]domain.Location),
    }
    repo.seedSampleData() // ❌ Hard-coded, fixed dates
    return repo
}
```

### After (Solution)
```go
// Clean repository
func NewInMemoryLocationRepository() secondary.LocationRepository {
    return &InMemoryLocationRepository{
        locations: make(map[string]domain.Location),
    }
}

// Data created through application
locationSpecs := routingApp.GenerateStandardLocations(8)
locations, err := routingApp.PopulateTestLocations(ctx, locationSpecs)
```

## 🚀 Future Extensions

### Additional Scenarios
- Error conditions and edge cases
- Performance testing with larger datasets
- Property-based testing integration

### Enhanced Realism
- Load test data from external sources (JSON/YAML)
- Integrate with real maritime scheduling data
- Generate seasonal shipping patterns

### Cross-Context Testing
- Complex multi-context workflows
- Event-driven scenario testing
- End-to-end integration validation

## 📝 Implementation Notes

### Key Files
- `test/mock/mock_test_environment.go` - Main orchestrator
- `internal/*/mock/mock_*_application.go` - Mock applications
- `test/integration/cargo_shipping_mock_applications_test.go` - Example usage

### Dependencies
- Uses existing routing adapter for cross-context communication
- Maintains compatibility with existing test infrastructure
- Preserves event bus and domain event patterns

This new strategy provides a robust, maintainable, and realistic foundation for integration testing while respecting DDD principles and hexagonal architecture constraints.
