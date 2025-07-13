package pim

import "testing"

func TestCallerSkipFramesFunctions(t *testing.T) {
	// Test default value
	originalSkip := GetCallerSkipFrames()
	if originalSkip != 3 {
		t.Errorf("Expected default callerSkipFrames to be 3, got %d", originalSkip)
	}

	// Test setting valid values
	testValues := []int{0, 5, 10, 15}
	for _, val := range testValues {
		SetCallerSkipFrames(val)
		result := GetCallerSkipFrames()
		if result != val {
			t.Errorf("Expected GetCallerSkipFrames() to return %d after SetCallerSkipFrames(%d), got %d", val, val, result)
		}
	}

	// Test invalid values (should not change)
	SetCallerSkipFrames(5) // Set to known good value
	currentValue := GetCallerSkipFrames()

	invalidValues := []int{-1, 16, 100}
	for _, val := range invalidValues {
		SetCallerSkipFrames(val)
		result := GetCallerSkipFrames()
		if result != currentValue {
			t.Errorf("Expected callerSkipFrames to remain %d after invalid SetCallerSkipFrames(%d), got %d", currentValue, val, result)
		}
	}

	// Restore original value
	SetCallerSkipFrames(originalSkip)
}

func TestShowGoroutineIDDefault(t *testing.T) {
	// Test that default value is false
	if showGoroutineID != false {
		t.Errorf("Expected default showGoroutineID to be false, got %v", showGoroutineID)
	}
}
