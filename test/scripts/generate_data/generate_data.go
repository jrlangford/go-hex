// Script to generate and populate test data for integration testing
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go_hex/test/testdata"
)

func main() {
	var (
		seed    = flag.Int64("seed", 0, "Seed for random data generation (0 for random)")
		verbose = flag.Bool("verbose", false, "Enable verbose logging")
		summary = flag.Bool("summary", true, "Print test data summary")
	)
	flag.Parse()

	// Set up logger
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	// Use current time as seed if not provided
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	logger.Info("Starting test data generation", "seed", *seed)

	// Create test environment
	testEnv, err := testdata.NewTestEnvironment(*seed, logger)
	if err != nil {
		logger.Error("Failed to create test environment", "error", err)
		os.Exit(1)
	}

	// Populate repositories
	ctx := context.Background()
	err = testEnv.PopulateWithTestData(ctx)
	if err != nil {
		logger.Error("Failed to populate test data", "error", err)
		os.Exit(1)
	}

	// Print summary if requested
	if *summary {
		testEnv.PrintTestDataSummary()
	}

	// Save seed for reproducibility
	fmt.Printf("\n=== Test Data Generation Complete ===\n")
	fmt.Printf("Seed used: %d\n", *seed)
	fmt.Printf("To reproduce this data set, use: go run %s -seed=%d\n", os.Args[0], *seed)
	fmt.Printf("Locations generated: %d\n", len(testEnv.TestData.Locations))
	fmt.Printf("Voyages generated: %d\n", len(testEnv.TestData.Voyages))
	fmt.Printf("Cargo scenarios generated: %d\n", len(testEnv.TestData.CargoScenarios))

	logger.Info("Test data generation completed successfully")
}
