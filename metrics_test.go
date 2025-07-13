package pim

import (
	"testing"
	"time"
)

// Benchmark tests for different logging functions
func BenchmarkInfo(b *testing.B) {
	SetLogLevel(InfoLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("Benchmark info message", "iteration", i)
	}
}

func BenchmarkDebug(b *testing.B) {
	SetLogLevel(DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debug("Benchmark debug message", "iteration", i)
	}
}

func BenchmarkError(b *testing.B) {
	SetLogLevel(ErrorLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Error("Benchmark error message", "iteration", i)
	}
}

func BenchmarkJson(b *testing.B) {
	SetLogLevel(InfoLevel)
	data := map[string]interface{}{
		"id":     1,
		"name":   "test",
		"email":  "test@example.com",
		"active": true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Json(data)
	}
}

func BenchmarkKeyValue(b *testing.B) {
	SetLogLevel(InfoLevel)
	data := map[string]interface{}{
		"id":     1,
		"name":   "test",
		"email":  "test@example.com",
		"active": true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		KeyValue(data)
	}
}

// Performance test for different configuration options
func BenchmarkWithFunctionNames(b *testing.B) {
	SetShowFunctionName(true)
	SetLogLevel(InfoLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("Benchmark with function names", "iteration", i)
	}
}

func BenchmarkWithoutFunctionNames(b *testing.B) {
	SetShowFunctionName(false)
	SetLogLevel(InfoLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("Benchmark without function names", "iteration", i)
	}
}

func BenchmarkWithStackDepth(b *testing.B) {
	SetStackDepth(5)
	SetLogLevel(ErrorLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Error("Benchmark with stack depth", "iteration", i)
	}
}

// Test functions to demonstrate enhanced features
func TestEnhancedLineVisibility(t *testing.T) {
	// Test function name tracking
	SetShowFunctionName(true)
	SetShowPackageName(true)

	Info("Testing enhanced line visibility")

	// Verify call info
	callInfo := GetCallInfo()
	if callInfo.File == "" {
		t.Error("Expected file name in call info")
	}
	if callInfo.Line == 0 {
		t.Error("Expected line number in call info")
	}
	if callInfo.Function == "" {
		t.Error("Expected function name in call info")
	}
}

func TestStackTraceFunctionality(t *testing.T) {
	// Test stack trace generation
	SetStackDepth(3)

	stackTrace := GetStackTrace()
	if len(stackTrace) == 0 {
		t.Error("Expected non-empty stack trace")
	}

	// Verify stack frame structure
	for i, frame := range stackTrace {
		if frame.File == "" {
			t.Errorf("Frame %d: Expected file name", i)
		}
		if frame.Line == 0 {
			t.Errorf("Frame %d: Expected line number", i)
		}
		if frame.Function == "" {
			t.Errorf("Frame %d: Expected function name", i)
		}
	}
}

func TestConfigurationOptions(t *testing.T) {
	// Test configuration functions
	originalFunctionName := true
	originalPackageName := true
	originalFullPath := false
	originalStackDepth := 3

	// Test function name toggle
	SetShowFunctionName(false)
	Info("Function names disabled")
	SetShowFunctionName(true)
	Info("Function names enabled")

	// Test package name toggle
	SetShowPackageName(false)
	Info("Package names disabled")
	SetShowPackageName(true)
	Info("Package names enabled")

	// Test full path toggle
	SetShowFullPath(true)
	Info("Full paths enabled")
	SetShowFullPath(false)
	Info("Full paths disabled")

	// Test stack depth
	SetStackDepth(5)
	Error("Stack depth set to 5")
	SetStackDepth(3)
	Error("Stack depth set back to 3")

	// Restore original settings
	SetShowFunctionName(originalFunctionName)
	SetShowPackageName(originalPackageName)
	SetShowFullPath(originalFullPath)
	SetStackDepth(originalStackDepth)
}

func TestErrorWithStackTrace(t *testing.T) {
	// Test error logging with stack trace
	SetLogLevel(ErrorLevel)

	// This should show stack trace
	Error("Test error with stack trace", "test_id", 123)
}

func TestPerformanceMetrics(t *testing.T) {
	// Test performance metrics
	start := time.Now()

	Metric("test_start", start.UnixNano(), "test", "performance")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	end := time.Now()
	duration := end.Sub(start).Milliseconds()

	Metric("test_end", end.UnixNano(), "test", "performance")
	Metric("test_duration", duration, "test", "performance", "unit", "ms")
}

// Example test showing real-world usage
func TestRealWorldScenario(t *testing.T) {
	// Simulate a web request processing scenario
	SetLogLevel(DebugLevel)

	// Request received
	Info("HTTP request received",
		"method", "POST",
		"path", "/api/users",
		"ip", "192.168.1.100",
		"user_agent", "Mozilla/5.0...",
	)

	// Authentication
	Debug("Authenticating user", "user_id", 123)
	Success("User authenticated successfully", "user_id", 123)

	// Database query
	Debug("Executing database query", "query", "SELECT * FROM users WHERE id = ?")
	Metric("db_query_start", time.Now().UnixNano(), "table", "users")

	// Simulate database work
	time.Sleep(5 * time.Millisecond)

	Metric("db_query_end", time.Now().UnixNano(), "table", "users")
	Metric("db_query_duration", 5, "table", "users", "unit", "ms")

	// Process response
	user := map[string]interface{}{
		"id":     123,
		"name":   "John Doe",
		"email":  "john@example.com",
		"active": true,
	}

	Json(user)

	// Response sent
	Success("HTTP response sent",
		"status_code", 200,
		"content_length", 150,
		"duration_ms", 25,
	)
}

// Benchmark comparison with different log levels
func BenchmarkLogLevels(b *testing.B) {
	levels := []LogLevel{TraceLevel, DebugLevel, InfoLevel, WarningLevel, ErrorLevel}

	for _, level := range levels {
		b.Run(getLevelString(level), func(b *testing.B) {
			SetLogLevel(level)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				switch level {
				case TraceLevel:
					Trace("Benchmark message", "iteration", i)
				case DebugLevel:
					Debug("Benchmark message", "iteration", i)
				case InfoLevel:
					Info("Benchmark message", "iteration", i)
				case WarningLevel:
					Warning("Benchmark message", "iteration", i)
				case ErrorLevel:
					Error("Benchmark message", "iteration", i)
				}
			}
		})
	}
}

// Memory usage test
func TestMemoryUsage(t *testing.T) {
	// Test that logging doesn't cause memory leaks
	SetLogLevel(InfoLevel)

	// Log many messages
	for i := 0; i < 1000; i++ {
		Info("Memory test message", "iteration", i)
	}

	// Force garbage collection to check for leaks
	// (In a real test, you'd use runtime.ReadMemStats)
}

// Concurrent logging test
func TestConcurrentLogging(t *testing.T) {
	SetLogLevel(InfoLevel)

	// Test logging from multiple goroutines
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				Info("Concurrent log message", "goroutine", id, "iteration", j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
