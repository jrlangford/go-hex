# Test Data Generation for Integration Tests

This package provides comprehensive test data generation utilities for the cargo shipping system integration tests. All test data is generated dynamically to ensure consistency and avoid brittle tests, following the design guidelines.

## Overview

The test data generation system provides:

- **Dynamic test data generation**: All data is generated programmatically using application services
- **Reproducible tests**: Use seeds to recreate the same test data sets
- **Realistic scenarios**: Generated data follows real-world shipping patterns
- **Cross-context integration**: Test data spans all three bounded contexts (Booking, Routing, Handling)
- **Configurable complexity**: Generate different amounts of test entities

## Components

### Core Files

- `generator.go` - Main test data generator with configurable entities
- `environment.go` - Test environment setup with all dependencies wired
- `executor.go` - Scenario execution utilities for running complete workflows
- `generate_data.go` - Standalone script for generating test data
- `run_tests.go` - Script for running integration tests with multiple data sets

### Generated Test Entities

1. **Locations** - Maritime locations with realistic UN/LOCODEs
2. **Voyages** - Shipping voyages with realistic schedules and carrier movements
3. **Cargo Scenarios** - Complete cargo shipment scenarios with handling events
4. **Handling Events** - Realistic sequence of cargo handling events

## Usage

### Basic Usage in Tests

```go
import "go_hex/test/testdata"

// Create test environment with generated data
func TestWithGeneratedData(t *testing.T) {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
    
    // Use fixed seed for reproducible tests
    seed := int64(12345)
    testEnv, err := testdata.NewTestEnvironment(seed, logger)
    if err != nil {
        t.Fatalf("Failed to create test environment: %v", err)
    }
    
    // Populate repositories
    ctx := context.Background()
    err = testEnv.PopulateWithTestData(ctx)
    if err != nil {
        t.Fatalf("Failed to populate test data: %v", err)
    }
    
    // Use testEnv.TestData for your test scenarios
    scenarios := testEnv.GetTestScenarios()
    // ... run your tests
}
```

### Running Complete Scenarios

```go
// Execute all generated scenarios
executor := testdata.NewScenarioExecutor(testEnv)
err = executor.ExecuteAllScenarios(ctx, bookingService, handlingService)
if err != nil {
    t.Fatalf("Failed to execute scenarios: %v", err)
}
```

### Standalone Data Generation

Generate test data and explore its structure:

```bash
# Generate test data with random seed
go run test/testdata/generate_data.go

# Generate with specific seed for reproducibility
go run test/testdata/generate_data.go -seed=12345

# Generate with verbose output
go run test/testdata/generate_data.go -verbose

# Generate without summary
go run test/testdata/generate_data.go -summary=false
```

### Running Integration Tests with Multiple Data Sets

Test your system with different generated data sets:

```bash
# Run 5 iterations with different data sets
go run test/testdata/run_tests.go -iterations=5

# Run specific test function
go run test/testdata/run_tests.go -test=TestSpecificCargoScenario

# Run with verbose output and longer timeout
go run test/testdata/run_tests.go -verbose -timeout=10m
```

## Test Data Structure

### Generated Locations
- 8-12 realistic maritime locations
- Proper UN/LOCODE format (e.g., SESTO, DEHAM, NLRTM)
- Includes major European ports for realistic scenarios

### Generated Voyages
- 5-10 voyages with realistic schedules
- 2-4 carrier movements per voyage
- Realistic travel times between ports (4-16 hours)
- Port time for loading/unloading (2-6 hours)

### Generated Cargo Scenarios
- 3-7 complete cargo shipping scenarios
- Random origin/destination pairs
- Realistic arrival deadlines (7-60 days in future)
- Associated handling event sequences

### Generated Handling Events
- RECEIVE → LOAD → UNLOAD → CLAIM progression
- Realistic timing between events
- Appropriate voyage assignments for load/unload
- Configurable delays for event processing tests

## Design Principles

This test data generation follows the project's design guidelines:

### Generated, Not Hardcoded
- All test data is created through domain constructors
- No hardcoded IDs or magic strings
- Ensures data follows business rules and validation

### Consistency
- Generated data maintains referential integrity
- Location codes match between contexts
- Voyage schedules are realistic and connected

### Domain-Driven
- Uses proper domain entities and value objects
- Generated through application services
- Respects bounded context boundaries

### Configurable
- Seed-based generation for reproducibility
- Adjustable entity counts
- Different scenario complexities

## Testing Strategies

### Reproducible Tests
Use fixed seeds for tests that need to be deterministic:

```go
seed := int64(12345) // Fixed seed
testEnv, _ := testdata.NewTestEnvironment(seed, logger)
```

### Randomized Testing
Use random seeds to test with different data each run:

```go
seed := time.Now().UnixNano() // Random seed
testEnv, _ := testdata.NewTestEnvironment(seed, logger)
```

### Stress Testing
Run multiple iterations with different data sets:

```bash
go run test/testdata/run_tests.go -iterations=10
```

### Scenario-Specific Testing
Focus on specific cargo scenarios:

```go
scenarios := testEnv.GetTestScenarios()
scenario := scenarios[0] // Test first scenario
executor.ExecuteCargoScenario(ctx, scenario, bookingService, handlingService)
```

## Integration with Existing Tests

The generated test data works alongside existing integration tests:

- `cargo_shipping_integration_test.go` - Original hardcoded test
- `cargo_shipping_generated_data_test.go` - New generated data tests

Both approaches are valuable:
- Hardcoded tests for known scenarios and edge cases
- Generated tests for broader coverage and discovering issues

## Example Output

When running test data generation, you'll see output like:

```
=== Test Data Summary ===
Generated locations: 10
Generated voyages: 7
Generated cargo scenarios: 5

Generated locations:
Location index=0 code=SESTO name=Stockholm country=SE
Location index=1 code=DEHAM name=Hamburg country=DE
...

Generated voyages:
Voyage index=0 number=f47ac10b-... movements=3
Movement voyage_index=0 movement_index=0 from=SESTO to=FIHEL departure=2024-01-15 08:00 arrival=2024-01-15 14:00
...

Generated cargo scenarios:
Cargo scenario index=0 tracking_id=a1b2c3d4-... origin=SESTO destination=NLRTM deadline=2024-03-15 23:59 handling_events=4
...
```

## Benefits

1. **Comprehensive Coverage**: Tests system with varied, realistic data
2. **Early Problem Detection**: Finds issues that hardcoded tests might miss
3. **Realistic Scenarios**: Uses proper maritime codes and realistic timing
4. **Maintainable**: Changes to domain models automatically reflected in tests
5. **Reproducible**: Can recreate exact test conditions using seeds
6. **Scalable**: Easy to generate larger or more complex test scenarios

## Future Enhancements

Potential improvements to the test data generation:

- Load test data from external files (JSON/YAML)
- Generate data based on specific test requirements
- Add performance benchmarking with generated data
- Integration with property-based testing frameworks
- Generate test data for specific edge cases or error conditions
