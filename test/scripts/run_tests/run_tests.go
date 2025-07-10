// Script to run integration tests with various test data configurations
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

func main() {
	var (
		iterations = flag.Int("iterations", 3, "Number of test iterations with different data sets")
		verbose    = flag.Bool("verbose", false, "Enable verbose test output")
		testName   = flag.String("test", "", "Specific test function to run (optional)")
		timeout    = flag.Duration("timeout", 5*time.Minute, "Test timeout per iteration")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	fmt.Printf("Running integration tests with generated data\n")
	fmt.Printf("Iterations: %d\n", *iterations)
	fmt.Printf("Timeout per iteration: %v\n", *timeout)
	if *testName != "" {
		fmt.Printf("Specific test: %s\n", *testName)
	}
	fmt.Printf("==========================================\n\n")

	totalStart := time.Now()
	successCount := 0
	failureCount := 0

	for i := 0; i < *iterations; i++ {
		seed := time.Now().UnixNano() + int64(i*1000)
		iterationStart := time.Now()

		fmt.Printf("=== Iteration %d/%d (Seed: %d) ===\n", i+1, *iterations, seed)

		// Build test command
		args := []string{"test", "./test/integration/", "-v"}

		if *testName != "" {
			args = append(args, "-run", *testName)
		} else {
			args = append(args, "-run", "TestCargoShippingSystemIntegrationWithGeneratedData")
		}

		if *timeout > 0 {
			args = append(args, "-timeout", timeout.String())
		}

		if *verbose {
			args = append(args, "-test.v")
		}

		// Set environment variable for seed (tests can read this if needed)
		cmd := exec.Command("go", args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("TEST_DATA_SEED=%d", seed))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the test
		err := cmd.Run()
		iterationDuration := time.Since(iterationStart)

		if err != nil {
			logger.Error("Test iteration failed",
				"iteration", i+1,
				"seed", seed,
				"duration", iterationDuration,
				"error", err)
			failureCount++
			fmt.Printf("❌ Iteration %d FAILED after %v\n\n", i+1, iterationDuration)
		} else {
			logger.Info("Test iteration passed",
				"iteration", i+1,
				"seed", seed,
				"duration", iterationDuration)
			successCount++
			fmt.Printf("✅ Iteration %d PASSED in %v\n\n", i+1, iterationDuration)
		}

		// Brief pause between iterations
		if i < *iterations-1 {
			time.Sleep(1 * time.Second)
		}
	}

	totalDuration := time.Since(totalStart)

	fmt.Printf("==========================================\n")
	fmt.Printf("Test Results Summary:\n")
	fmt.Printf("Total iterations: %d\n", *iterations)
	fmt.Printf("Passed: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failureCount)
	fmt.Printf("Success rate: %.1f%%\n", float64(successCount)/float64(*iterations)*100)
	fmt.Printf("Total duration: %v\n", totalDuration)
	fmt.Printf("Average per iteration: %v\n", totalDuration/time.Duration(*iterations))

	if failureCount > 0 {
		logger.Error("Some test iterations failed", "failures", failureCount)
		os.Exit(1)
	} else {
		logger.Info("All test iterations passed successfully")
		fmt.Printf("🎉 All tests passed!\n")
	}
}
