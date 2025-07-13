package pim

import "testing"

func TestFileLoggingConfiguration(t *testing.T) {
	// Test default value
	if GetFileLogging() != false {
		t.Errorf("Expected default file logging to be false, got %v", GetFileLogging())
	}

	// Test enabling
	SetFileLogging(true)
	if !GetFileLogging() {
		t.Error("Expected GetFileLogging() to return true after SetFileLogging(true)")
	}

	// Test disabling
	SetFileLogging(false)
	if GetFileLogging() {
		t.Error("Expected GetFileLogging() to return false after SetFileLogging(false)")
	}
}
