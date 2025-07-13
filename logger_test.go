package pim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testFileName    = "logger_test.go"
	testInfoMessage = "Test info message"
)

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	// Temporarily redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute function
	fn()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func TestGetCallInfo(t *testing.T) {
	// Test basic call info retrieval
	info := getCallInfo(0)

	if info.File == "" {
		t.Error("Expected file name to be set")
	}
	if info.Line == 0 {
		t.Error("Expected line number to be greater than 0")
	}
	if info.Function == "" {
		t.Error("Expected function name to be set")
	}
	if info.Package == "" {
		t.Error("Expected package name to be set")
	}

	// Test that function name contains the test function
	if !strings.Contains(info.Function, "TestGetCallInfo") {
		t.Errorf("Expected function name to contain 'TestGetCallInfo', got '%s'", info.Function)
	}
}

func TestGetFileInfo(t *testing.T) {
	// Test with file line enabled
	originalShowFileLine := showFileLine
	showFileLine = true
	defer func() { showFileLine = originalShowFileLine }()

	fileInfo := getFileInfo()
	if fileInfo == "" {
		t.Error("Expected file info to be non-empty when showFileLine is true")
	}
	// Should contain file name and line number
	if !strings.Contains(fileInfo, testFileName) && !strings.Contains(fileInfo, "L") {
		t.Errorf("Expected file info to contain file name and line number, got '%s'", fileInfo)
	}

	// Test with file line disabled
	showFileLine = false
	fileInfo = getFileInfo()
	if fileInfo != "" {
		t.Error("Expected file info to be empty when showFileLine is false")
	}
}

func TestGetGoroutineID(t *testing.T) {
	// Test with goroutine ID enabled
	originalShowGoroutineID := showGoroutineID
	showGoroutineID = true
	defer func() { showGoroutineID = originalShowGoroutineID }()

	goroutineID := getGoroutineID()
	if goroutineID == "" {
		t.Error("Expected goroutine ID to be non-empty when showGoroutineID is true")
	}

	// Should contain "goroutine" and a number
	if !strings.Contains(goroutineID, "goroutine") {
		t.Errorf("Expected goroutine ID to contain 'goroutine', got '%s'", goroutineID)
	}

	// Test with goroutine ID disabled
	showGoroutineID = false
	goroutineID = getGoroutineID()
	if goroutineID != "" {
		t.Error("Expected goroutine ID to be empty when showGoroutineID is false")
	}
}

func TestGetStackTrace(t *testing.T) {
	// Set a reasonable stack depth for testing
	originalStackDepth := stackDepth
	stackDepth = 3
	defer func() { stackDepth = originalStackDepth }()

	frames := getStackTrace(0)

	if len(frames) == 0 {
		t.Error("Expected stack trace to contain at least one frame")
	}

	// Check first frame
	frame := frames[0]
	if frame.File == "" {
		t.Error("Expected frame file to be set")
	}
	if frame.Line == 0 {
		t.Error("Expected frame line to be greater than 0")
	}
	if frame.Function == "" {
		t.Error("Expected frame function to be set")
	}

	// Should contain the test function
	found := false
	for _, frame := range frames {
		if strings.Contains(frame.Function, "TestGetStackTrace") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected stack trace to contain the test function")
	}
}

func TestFormatStackTrace(t *testing.T) {
	frames := []StackFrame{
		{
			File:     "test.go",
			Line:     10,
			Function: "TestFunc",
			Package:  "main",
		},
		{
			File:     "main.go",
			Line:     5,
			Function: "main",
			Package:  "main",
		},
	}

	originalShowFunctionName := showFunctionName
	originalShowPackageName := showPackageName
	showFunctionName = true
	showPackageName = true
	defer func() {
		showFunctionName = originalShowFunctionName
		showPackageName = originalShowPackageName
	}()

	formatted := formatStackTrace(frames)

	if formatted == "" {
		t.Error("Expected formatted stack trace to be non-empty")
	}

	// Should contain file names, line numbers, and function names
	if !strings.Contains(formatted, "test.go") {
		t.Error("Expected formatted stack trace to contain 'test.go'")
	}
	if !strings.Contains(formatted, "L10") {
		t.Error("Expected formatted stack trace to contain 'L10'")
	}
	if !strings.Contains(formatted, "main.TestFunc") {
		t.Error("Expected formatted stack trace to contain 'main.TestFunc'")
	}
	if !strings.Contains(formatted, "↳") {
		t.Error("Expected formatted stack trace to contain '↳' symbol")
	}

	// Test with empty frames
	emptyFormatted := formatStackTrace([]StackFrame{})
	if emptyFormatted != "" {
		t.Error("Expected empty stack trace to return empty string")
	}
}

func TestInfo(t *testing.T) {
	// Test without arguments
	output := captureOutput(func() {
		Info("Test info message")
	})

	if !strings.Contains(output, "Test info message") {
		t.Errorf("Expected output to contain 'Test info message', got: %s", output)
	}

	// Test with arguments
	output = captureOutput(func() {
		Info("Test with args", "arg1", 123)
	})

	if !strings.Contains(output, "Test with args: arg1 123") {
		t.Errorf("Expected output to contain formatted message with args, got: %s", output)
	}
}

func TestSuccess(t *testing.T) {
	output := captureOutput(func() {
		Success("Operation completed successfully")
	})

	if !strings.Contains(output, "Operation completed successfully") {
		t.Errorf("Expected output to contain success message, got: %s", output)
	}
}

func TestInit(t *testing.T) {
	output := captureOutput(func() {
		Init("Initializing application", "version", "1.0.0")
	})

	if !strings.Contains(output, "Initializing application: version 1.0.0") {
		t.Errorf("Expected output to contain init message with args, got: %s", output)
	}
}

func TestConfig(t *testing.T) {
	output := captureOutput(func() {
		Config("Configuration loaded", "file", "config.json")
	})

	if !strings.Contains(output, "Configuration loaded: file config.json") {
		t.Errorf("Expected output to contain config message with args, got: %s", output)
	}
}

func TestWarning(t *testing.T) {
	output := captureOutput(func() {
		Warning("This is a warning", "code", 404)
	})

	if !strings.Contains(output, "This is a warning: code 404") {
		t.Errorf("Expected output to contain warning message with args, got: %s", output)
	}
}

func TestError(t *testing.T) {
	output := captureOutput(func() {
		Error("Something went wrong", "error", "file not found")
	})

	if !strings.Contains(output, "Something went wrong: error file not found") {
		t.Errorf("Expected output to contain error message with args, got: %s", output)
	}

	// Error should also include stack trace
	if !strings.Contains(output, "↳") {
		t.Errorf("Expected error output to contain stack trace with '↳' symbol, got: %s", output)
	}
}

func TestDebug(t *testing.T) {
	// Set log level to debug to ensure debug messages are shown
	originalLogLevel := currentLogLevel
	currentLogLevel = DebugLevel
	defer func() { currentLogLevel = originalLogLevel }()

	output := captureOutput(func() {
		Debug("Debug information", "variable", "value")
	})

	if !strings.Contains(output, "Debug information: variable value") {
		t.Errorf("Expected output to contain debug message with args, got: %s", output)
	}
}

func TestTrace(t *testing.T) {
	// Set log level to trace to ensure trace messages are shown
	originalLogLevel := currentLogLevel
	currentLogLevel = TraceLevel
	defer func() { currentLogLevel = originalLogLevel }()

	output := captureOutput(func() {
		Trace("Trace information", "step", 1)
	})

	if !strings.Contains(output, "Trace information: step 1") {
		t.Errorf("Expected output to contain trace message with args, got: %s", output)
	}
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Panic function to cause a panic")
		} else {
			panicMsg := r.(string)
			if !strings.Contains(panicMsg, "Critical error: system failure") {
				t.Errorf("Expected panic message to contain formatted message, got: %s", panicMsg)
			}
		}
	}()

	Panic("Critical error", "system failure")
}

func TestMetric(t *testing.T) {
	output := captureOutput(func() {
		Metric("cpu_usage", 75.5, "server1", "production")
	})

	if !strings.Contains(output, "cpu_usage: 75.5") {
		t.Errorf("Expected output to contain metric name and value, got: %s", output)
	}

	if !strings.Contains(output, "[server1, production]") {
		t.Errorf("Expected output to contain tags, got: %s", output)
	}

	// Test metric without tags
	output = captureOutput(func() {
		Metric("memory_usage", "2GB")
	})

	if !strings.Contains(output, "memory_usage: 2GB") {
		t.Errorf("Expected output to contain metric without tags, got: %s", output)
	}
}

func TestLogWithTimestamp(t *testing.T) {
	// Test that log level filtering works
	originalLogLevel := currentLogLevel
	currentLogLevel = ErrorLevel
	defer func() { currentLogLevel = originalLogLevel }()

	// Debug message should not appear
	output := captureOutput(func() {
		LogWithTimestamp("🐛", "Debug message", DebugLevel)
	})

	if output != "" {
		t.Errorf("Expected no output for debug message when log level is Error, got: %s", output)
	}

	// Error message should appear
	output = captureOutput(func() {
		LogWithTimestamp("❌", "Error message", ErrorLevel)
	})

	if !strings.Contains(output, "Error message") {
		t.Errorf("Expected output to contain error message, got: %s", output)
	}

	// Should contain timestamp
	if !strings.Contains(output, "UTC") {
		t.Errorf("Expected output to contain timestamp with UTC, got: %s", output)
	}
}

func TestLogWithStackTrace(t *testing.T) {
	originalStackDepth := stackDepth
	stackDepth = 2
	defer func() { stackDepth = originalStackDepth }()

	output := captureOutput(func() {
		LogWithStackTrace("❌", "Error with stack trace", ErrorLevel)
	})

	if !strings.Contains(output, "Error with stack trace") {
		t.Errorf("Expected output to contain error message, got: %s", output)
	}

	// Should contain stack trace
	if !strings.Contains(output, "↳") {
		t.Errorf("Expected output to contain stack trace with '↳' symbol, got: %s", output)
	}
}

func TestGetCallInfoExternalUse(t *testing.T) {
	info := GetCallInfo()

	if info.File == "" {
		t.Error("Expected GetCallInfo to return non-empty file name")
	}
	if info.Line == 0 {
		t.Error("Expected GetCallInfo to return line number greater than 0")
	}
	if !strings.Contains(info.Function, "TestGetCallInfoExternalUse") {
		t.Errorf("Expected function name to contain test name, got: %s", info.Function)
	}
}

func TestGetStackTraceExternalUse(t *testing.T) {
	originalStackDepth := stackDepth
	stackDepth = 3
	defer func() { stackDepth = originalStackDepth }()

	frames := GetStackTrace()

	if len(frames) == 0 {
		t.Error("Expected GetStackTrace to return at least one frame")
	}

	// Should contain the test function
	found := false
	for _, frame := range frames {
		if strings.Contains(frame.Function, "TestGetStackTraceExternalUse") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected stack trace to contain the test function")
	}
}

func TestLoggingWithFileOutput(t *testing.T) {
	// Setup temporary directory for test logs
	tempDir := t.TempDir()

	// Initialize file logging
	err := InitializeFileLogging(tempDir, "test-service")
	if err != nil {
		t.Fatalf("Failed to initialize file logging: %v", err)
	}
	defer CloseLogFiles()

	// Test that files are created for all log levels
	expectedFiles := []string{
		"panic.jaeger.json",
		"error.jaeger.json",
		"warning.jaeger.json",
		"info.jaeger.json",
		"debug.jaeger.json",
		"trace.jaeger.json",
	}

	// Generate some log messages
	Info("Test info message")
	Error("Test error message")
	Warning("Test warning message")

	// Give a moment for file operations
	time.Sleep(100 * time.Millisecond)

	logDir := filepath.Join(tempDir, ".log/pim")

	// Check that log files exist
	for _, fileName := range expectedFiles {
		filePath := filepath.Join(logDir, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected log file %s to exist", fileName)
		}
	}
}

func TestLoggingConfigurationSettings(t *testing.T) {
	// Test various configuration settings
	originalShowFileLine := showFileLine
	originalShowGoroutineID := showGoroutineID
	originalShowFunctionName := showFunctionName
	originalShowPackageName := showPackageName
	originalShowFullPath := showFullPath

	defer func() {
		showFileLine = originalShowFileLine
		showGoroutineID = originalShowGoroutineID
		showFunctionName = originalShowFunctionName
		showPackageName = originalShowPackageName
		showFullPath = originalShowFullPath
	}()

	// Test with all options enabled
	showFileLine = true
	showGoroutineID = true
	showFunctionName = true
	showPackageName = true
	showFullPath = false

	output := captureOutput(func() {
		Info("Test message with all options")
	})
	if !strings.Contains(output, testFileName) {
		t.Error("Expected output to contain file name when showFileLine is true")
	}
	if !strings.Contains(output, "goroutine") {
		t.Error("Expected output to contain goroutine info when showGoroutineID is true")
	}

	// Test with all options disabled
	showFileLine = false
	showGoroutineID = false

	output = captureOutput(func() {
		Info("Test message with options disabled")
	})
	if strings.Contains(output, testFileName) {
		t.Error("Expected output to not contain file name when showFileLine is false")
	}
	if strings.Contains(output, "goroutine") {
		t.Error("Expected output to not contain goroutine info when showGoroutineID is false")
	}
}

func BenchmarkLoggerInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Info("Benchmark message", "iteration", i)
	}
}

func BenchmarkLoggerError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Error("Benchmark error", "iteration", i)
	}
}

func BenchmarkGetCallInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getCallInfo(2)
	}
}

func BenchmarkGetStackTrace(b *testing.B) {
	originalStackDepth := stackDepth
	stackDepth = 5
	defer func() { stackDepth = originalStackDepth }()

	for i := 0; i < b.N; i++ {
		getStackTrace(2)
	}
}
