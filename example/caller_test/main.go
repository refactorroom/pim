package main

import (
	"fmt"

	pim "github.com/refactorroom/pim"
)

func testFunction() {
	pim.Info("This is a test from testFunction")
}

func intermediateFunction() {
	testFunction()
}

func main() {
	fmt.Println("=== Testing Default Caller Skip Frames ===")
	intermediateFunction()

	fmt.Println("\n=== Testing Increased Caller Skip Frames ===")
	pim.SetCallerSkipFrames(5) // Skip more frames
	intermediateFunction()

	fmt.Println("\n=== Testing Reduced Caller Skip Frames ===")
	pim.SetCallerSkipFrames(2) // Skip fewer frames
	intermediateFunction()

	fmt.Println("\n=== Testing Goroutine ID Disabled by Default ===")
	pim.SetCallerSkipFrames(3) // Reset to default
	pim.Info("Goroutine ID should not be shown")

	fmt.Println("\n=== Testing Goroutine ID Enabled ===")
	pim.SetShowGoroutineID(true)
	pim.Info("Goroutine ID should be shown")

	// Test configuration functions
	fmt.Println("\n=== Testing Configuration Functions ===")
	fmt.Printf("Current caller skip frames: %d\n", pim.GetCallerSkipFrames())

	// Test edge cases
	pim.SetCallerSkipFrames(-1) // Should not change
	fmt.Printf("After invalid -1: %d\n", pim.GetCallerSkipFrames())

	pim.SetCallerSkipFrames(20) // Should not change
	fmt.Printf("After invalid 20: %d\n", pim.GetCallerSkipFrames())

	pim.SetCallerSkipFrames(7) // Should change
	fmt.Printf("After valid 7: %d\n", pim.GetCallerSkipFrames())
}
