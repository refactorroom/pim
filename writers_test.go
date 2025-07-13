package pim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testLoggerName           = "test-logger"
	testJSONMessage          = "Test JSON message"
	expectedResultFormat     = "Expected result to contain '%s', got: %s"
	failedToCreateFileWriter = "Failed to create FileWriter: %v"
)

func TestNewConsoleWriter(t *testing.T) {
	config := LoggerConfig{
		Level:           InfoLevel,
		TimestampFormat: time.RFC3339,
		EnableJSON:      false,
		ThemeName:       "default",
	}

	writer := NewConsoleWriter(config)

	if writer == nil {
		t.Fatal("Expected NewConsoleWriter to return non-nil writer")
	}

	if writer.config.Level != InfoLevel {
		t.Errorf("Expected config level to be InfoLevel, got %v", writer.config.Level)
	}

	if writer.themeManager == nil {
		t.Error("Expected themeManager to be initialized")
	}
}

func TestConsoleWriterWrite(t *testing.T) {
	config := LoggerConfig{
		Level:           InfoLevel,
		TimestampFormat: time.RFC3339,
		EnableJSON:      false,
	}

	writer := NewConsoleWriter(config)

	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		Message:     "Test message",
		ServiceName: "test-logger",
		File:        "test.go",
		Line:        10,
		Function:    "TestFunc",
		Package:     "main",
		Context: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	err := writer.Write(entry)
	if err != nil {
		t.Errorf("Expected Write to succeed, got error: %v", err)
	}
}

func TestConsoleWriterWriteJSON(t *testing.T) {
	config := LoggerConfig{
		Level:           InfoLevel,
		TimestampFormat: time.RFC3339,
		EnableJSON:      true,
	}

	writer := NewConsoleWriter(config)

	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		Message:     "Test JSON message",
		ServiceName: "test-logger",
	}

	err := writer.Write(entry)
	if err != nil {
		t.Errorf("Expected JSON Write to succeed, got error: %v", err)
	}
}

func TestConsoleWriterFormatContext(t *testing.T) {
	writer := NewConsoleWriter(LoggerConfig{})

	context := map[string]interface{}{
		"string_key": "string_value",
		"int_key":    42,
		"bool_key":   true,
		"float_key":  3.14,
	}

	result := writer.formatContext(context)

	if result == "" {
		t.Error("Expected formatContext to return non-empty string")
	}

	// Should contain all keys and values
	expectedSubstrings := []string{"string_key", "string_value", "int_key", "42", "bool_key", "true", "float_key", "3.14"}
	for _, expected := range expectedSubstrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', got: %s", expected, result)
		}
	}
}

func TestConsoleWriterFormatStackTrace(t *testing.T) {
	writer := NewConsoleWriter(LoggerConfig{})

	frames := []StackFrame{
		{
			File:     "test1.go",
			Line:     10,
			Function: "func1",
			Package:  "pkg1",
		},
		{
			File:     "test2.go",
			Line:     20,
			Function: "func2",
			Package:  "pkg2",
		},
	}

	result := writer.formatStackTrace(frames)

	if result == "" {
		t.Error("Expected formatStackTrace to return non-empty string")
	}

	// Should contain file names, line numbers, and function names
	expectedSubstrings := []string{"test1.go", "10", "func1", "test2.go", "20", "func2"}
	for _, expected := range expectedSubstrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', got: %s", expected, result)
		}
	}
}

func TestConsoleWriterClose(t *testing.T) {
	writer := NewConsoleWriter(LoggerConfig{})

	err := writer.Close()
	if err != nil {
		t.Errorf("Expected Close to succeed, got error: %v", err)
	}
}

func TestConsoleWriterFlush(t *testing.T) {
	writer := NewConsoleWriter(LoggerConfig{})

	err := writer.Flush()
	if err != nil {
		t.Errorf("Expected Flush to succeed, got error: %v", err)
	}
}

func TestNewFileWriter(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	config := LoggerConfig{
		Level:           InfoLevel,
		TimestampFormat: time.RFC3339,
		EnableJSON:      false,
	}

	rotationConfig := RotationConfig{
		MaxSize:         1024 * 1024, // 1MB
		MaxAge:          24 * time.Hour,
		MaxFiles:        5,
		Compress:        true,
		RotateTime:      time.Hour,
		CleanupInterval: time.Hour,
		VerboseCleanup:  false,
	}

	writer, err := NewFileWriter(filename, config, rotationConfig)
	if err != nil {
		t.Fatalf("Expected NewFileWriter to succeed, got error: %v", err)
	}
	defer writer.Close()

	if writer == nil {
		t.Fatal("Expected NewFileWriter to return non-nil writer")
	}

	// Check that file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

func TestFileWriterWrite(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test-write.log")
	config := LoggerConfig{
		Level:           InfoLevel,
		TimestampFormat: time.RFC3339,
		EnableJSON:      false,
	}

	rotationConfig := RotationConfig{
		MaxSize:  1024 * 1024,
		MaxFiles: 5,
	}

	writer, err := NewFileWriter(filename, config, rotationConfig)
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer writer.Close()
	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		Message:     "Test file write message",
		ServiceName: "test-logger",
		Context: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	err = writer.Write(entry)
	if err != nil {
		t.Errorf("Expected Write to succeed, got error: %v", err)
	}

	// Flush to ensure write is complete
	writer.Flush()

	// Check that content was written
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Test file write message") {
		t.Errorf("Expected log file to contain message, got: %s", contentStr)
	}

	if !strings.Contains(contentStr, "test_key") {
		t.Errorf("Expected log file to contain context, got: %s", contentStr)
	}
}

func TestFileWriterFormatLogEntry(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test-format.log")

	config := LoggerConfig{
		Level:           InfoLevel,
		TimestampFormat: time.RFC3339,
		EnableJSON:      false,
	}

	writer, err := NewFileWriter(filename, config, RotationConfig{})
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer writer.Close()

	entry := CoreLogEntry{
		Timestamp:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Level:       InfoLevel,
		Message:     "Test format message",
		ServiceName: "test-logger",
		File:        "test.go",
		Line:        42,
		Function:    "TestFunc",
		Package:     "main",
		Context: map[string]interface{}{
			"key": "value",
		},
	}

	result := writer.formatLogEntry(entry)

	if result == "" {
		t.Error("Expected formatLogEntry to return non-empty string")
	}

	// Should contain timestamp, level, message, and context
	expectedSubstrings := []string{
		"2023-01-01T12:00:00Z",
		"INFO",
		"Test format message",
		"test-logger",
		"test.go:42",
		"main.TestFunc",
		"key=value",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', got: %s", expected, result)
		}
	}
}

func TestFileWriterShouldRotate(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test-rotate.log")

	// Create a config with small max size for testing
	config := LoggerConfig{Level: InfoLevel}
	rotationConfig := RotationConfig{
		MaxSize:  100, // Very small size to trigger rotation
		MaxFiles: 3,
	}

	writer, err := NewFileWriter(filename, config, rotationConfig)
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer writer.Close()

	// Initially should not need rotation
	if writer.shouldRotate() {
		t.Error("Expected new file to not need rotation")
	}
	// Write enough data to exceed max size
	longMessage := strings.Repeat("This is a long message that will exceed the max size limit. ", 10)
	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		Message:     longMessage,
		ServiceName: "test-logger",
	}

	// Write multiple entries to exceed size limit
	for i := 0; i < 5; i++ {
		writer.Write(entry)
	}
	writer.Flush()

	// Now should need rotation
	if !writer.shouldRotate() {
		t.Error("Expected file to need rotation after exceeding max size")
	}
}

func TestRotationConfig(t *testing.T) {
	config := RotationConfig{
		MaxSize:         1024 * 1024,
		MaxAge:          24 * time.Hour,
		MaxFiles:        10,
		Compress:        true,
		RotateTime:      time.Hour,
		CleanupInterval: 30 * time.Minute,
		VerboseCleanup:  true,
	}

	// Test that all fields are set correctly
	if config.MaxSize != 1024*1024 {
		t.Errorf("Expected MaxSize to be %d, got %d", 1024*1024, config.MaxSize)
	}

	if config.MaxAge != 24*time.Hour {
		t.Errorf("Expected MaxAge to be %v, got %v", 24*time.Hour, config.MaxAge)
	}

	if config.MaxFiles != 10 {
		t.Errorf("Expected MaxFiles to be 10, got %d", config.MaxFiles)
	}

	if !config.Compress {
		t.Error("Expected Compress to be true")
	}

	if config.RotateTime != time.Hour {
		t.Errorf("Expected RotateTime to be %v, got %v", time.Hour, config.RotateTime)
	}

	if config.CleanupInterval != 30*time.Minute {
		t.Errorf("Expected CleanupInterval to be %v, got %v", 30*time.Minute, config.CleanupInterval)
	}

	if !config.VerboseCleanup {
		t.Error("Expected VerboseCleanup to be true")
	}
}

func TestFileWriterJSON(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test-json.log")

	config := LoggerConfig{
		Level:           InfoLevel,
		TimestampFormat: time.RFC3339,
		EnableJSON:      true,
	}

	writer, err := NewFileWriter(filename, config, RotationConfig{})
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer writer.Close()

	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		Message:     "Test JSON message",
		ServiceName: "test-logger",
		Context: map[string]interface{}{
			"json_key": "json_value",
			"number":   42,
		},
	}

	err = writer.Write(entry)
	if err != nil {
		t.Errorf("Expected JSON Write to succeed, got error: %v", err)
	}

	writer.Flush()

	// Check that valid JSON was written
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read JSON log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Test JSON message") {
		t.Errorf("Expected JSON log to contain message, got: %s", contentStr)
	}

	if !strings.Contains(contentStr, "json_key") {
		t.Errorf("Expected JSON log to contain context key, got: %s", contentStr)
	}
}

func TestFileWriterClose(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test-close.log")

	writer, err := NewFileWriter(filename, LoggerConfig{}, RotationConfig{})
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}

	// Write some data
	entry := CoreLogEntry{
		Timestamp: time.Now(),
		Level:     InfoLevel,
		Message:   "Test close message",
	}

	writer.Write(entry)

	// Close should succeed
	err = writer.Close()
	if err != nil {
		t.Errorf("Expected Close to succeed, got error: %v", err)
	}

	// Subsequent writes should fail or be ignored
	err = writer.Write(entry)
	if err == nil {
		t.Error("Expected Write after Close to fail")
	}
}

func BenchmarkConsoleWriterWrite(b *testing.B) {
	writer := NewConsoleWriter(LoggerConfig{Level: InfoLevel})
	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		Message:     "Benchmark message",
		ServiceName: "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(entry)
	}
}

func BenchmarkFileWriterWrite(b *testing.B) {
	tempDir := b.TempDir()
	filename := filepath.Join(tempDir, "benchmark.log")

	writer, err := NewFileWriter(filename, LoggerConfig{Level: InfoLevel}, RotationConfig{})
	if err != nil {
		b.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer writer.Close()
	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		Message:     "Benchmark message",
		ServiceName: "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(entry)
	}
}
