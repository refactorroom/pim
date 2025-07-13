# Module 4: Writing Effective Tests

## Overview

Testing is crucial for maintaining the reliability and performance of the pim package. This module covers comprehensive testing strategies, patterns, and best practices specific to logging systems.

### What You'll Learn
- Testing strategies for logging systems
- Output capture and verification techniques
- File logging testing patterns
- Performance and benchmark testing
- Mock and stub techniques
- Thread safety testing

## Testing Philosophy

### Why Testing Matters in Logging

Logging systems have unique testing challenges:
- **Output verification**: Checking what's written to console/files
- **Performance critical**: Logging should have minimal overhead
- **Thread safety**: Multiple goroutines logging simultaneously
- **Configuration changes**: Runtime behavior modifications
- **File operations**: Creating, writing, rotating log files

### Test Categories

1. **Unit Tests**: Individual functions and methods
2. **Integration Tests**: Component interactions
3. **Output Tests**: Verifying log output format and content
4. **File Tests**: File creation, writing, and rotation
5. **Performance Tests**: Benchmarks and resource usage
6. **Thread Safety Tests**: Concurrent access patterns

## Test Structure and Organization

### File Organization

Tests should be organized by functionality:
```
config_test.go           # Configuration testing
logger_test.go           # Core logging functionality
caller_info_test.go      # Caller information extraction
file_logging_test.go     # File operations
metrics_test.go          # Performance metrics
output_test.go           # Output formatting
writers_test.go          # File writers
integration_test.go      # End-to-end testing
```

### Test Function Naming

Use descriptive names that indicate what's being tested:
```go
// Good names
func TestSetLogLevel(t *testing.T)
func TestFileLoggingWithRotation(t *testing.T)
func TestCallerInfoWithCustomSkipFrames(t *testing.T)
func TestConcurrentConfigurationAccess(t *testing.T)

// Poor names
func TestConfig(t *testing.T)
func TestLogging(t *testing.T)
func TestFile(t *testing.T)
```

## Basic Testing Patterns

### State Management Pattern

Always save and restore global state:
```go
func TestLogLevel(t *testing.T) {
    // Save original state
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    // Test logic
    SetLogLevel(DebugLevel)
    if GetLogLevel() != DebugLevel {
        t.Errorf("Expected DebugLevel, got %v", GetLogLevel())
    }
}
```

### Table-Driven Tests

Use table-driven tests for multiple similar cases:
```go
func TestLogLevelValidation(t *testing.T) {
    testCases := []struct {
        name          string
        inputLevel    LogLevel
        expectedLevel LogLevel
        shouldChange  bool
    }{
        {"valid panic level", PanicLevel, PanicLevel, true},
        {"valid info level", InfoLevel, InfoLevel, true},
        {"valid trace level", TraceLevel, TraceLevel, true},
        {"invalid negative level", LogLevel(-1), InfoLevel, false},
        {"invalid high level", LogLevel(100), InfoLevel, false},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Save state
            originalLevel := GetLogLevel()
            defer SetLogLevel(originalLevel)

            // Set to known state
            SetLogLevel(InfoLevel)

            // Test
            SetLogLevel(tc.inputLevel)
            actualLevel := GetLogLevel()

            if tc.shouldChange {
                if actualLevel != tc.expectedLevel {
                    t.Errorf("Expected level %v, got %v", tc.expectedLevel, actualLevel)
                }
            } else {
                if actualLevel != InfoLevel {
                    t.Errorf("Level should not have changed from InfoLevel, got %v", actualLevel)
                }
            }
        })
    }
}
```

### Subtests for Organization

Use subtests to group related test cases:
```go
func TestFileLogging(t *testing.T) {
    t.Run("EnableDisable", func(t *testing.T) {
        // Test enabling and disabling file logging
    })
    
    t.Run("FileCreation", func(t *testing.T) {
        // Test that log files are created
    })
    
    t.Run("ContentVerification", func(t *testing.T) {
        // Test that correct content is written
    })
    
    t.Run("Rotation", func(t *testing.T) {
        // Test log rotation functionality
    })
}
```

## Output Capture Testing

### Console Output Capture

Capture and verify console output:
```go
func TestConsoleOutput(t *testing.T) {
    // Create a buffer to capture output
    var output strings.Builder
    
    // Create logger with custom output
    logger := NewLoggerWithOutput(&output)
    
    // Generate log message
    logger.Info("Test message with %s", "parameter")
    
    // Verify output
    result := output.String()
    expectedContents := []string{
        "[INFO]",
        "Test message with parameter",
        time.Now().Format("2006-01-02"), // Check date portion
    }
    
    for _, expected := range expectedContents {
        if !strings.Contains(result, expected) {
            t.Errorf("Expected '%s' in output, got: %s", expected, result)
        }
    }
}
```

### Multiple Output Destinations

Test output to multiple destinations:
```go
func TestMultipleOutputs(t *testing.T) {
    // Create multiple output destinations
    var consoleOutput strings.Builder
    var fileOutput strings.Builder
    
    // Create multi-writer
    multiWriter := io.MultiWriter(&consoleOutput, &fileOutput)
    logger := NewLoggerWithOutput(multiWriter)
    
    // Log message
    logger.Info("Test message")
    
    // Verify both outputs received the message
    consoleResult := consoleOutput.String()
    fileResult := fileOutput.String()
    
    if consoleResult != fileResult {
        t.Error("Console and file outputs should be identical")
    }
    
    if !strings.Contains(consoleResult, "Test message") {
        t.Error("Expected message not found in output")
    }
}
```

### Format Verification

Test specific output formats:
```go
func TestOutputFormat(t *testing.T) {
    originalShowFileLine := showFileLine
    originalShowGoroutineID := showGoroutineID
    defer func() {
        showFileLine = originalShowFileLine
        showGoroutineID = originalShowGoroutineID
    }()

    var output strings.Builder
    logger := NewLoggerWithOutput(&output)

    t.Run("WithCallerInfo", func(t *testing.T) {
        output.Reset()
        SetShowFileLine(true)
        SetShowGoroutineID(false)
        
        logger.Info("Test message")
        result := output.String()
        
        // Should contain file:line info
        if !strings.Contains(result, ".go:") {
            t.Error("Expected caller info (file:line) in output")
        }
        
        // Should not contain goroutine ID
        if strings.Contains(result, "[G:") {
            t.Error("Should not contain goroutine ID when disabled")
        }
    })

    t.Run("WithGoroutineID", func(t *testing.T) {
        output.Reset()
        SetShowFileLine(false)
        SetShowGoroutineID(true)
        
        logger.Info("Test message")
        result := output.String()
        
        // Should contain goroutine ID
        if !strings.Contains(result, "[G:") {
            t.Error("Expected goroutine ID in output")
        }
        
        // Should not contain file info
        if strings.Contains(result, ".go:") {
            t.Error("Should not contain caller info when disabled")
        }
    })
}
```

## File Testing Patterns

### Temporary Directory Setup

Use temporary directories for file tests:
```go
func TestFileLogging(t *testing.T) {
    // Create temporary directory
    tempDir, err := ioutil.TempDir("", "pim_test")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir) // Cleanup

    // Save original file logging state
    originalEnabled := enableFileLogging
    originalLogFile := logFile
    defer func() {
        enableFileLogging = originalEnabled
        logFile = originalLogFile
    }()

    // Configure file logging
    logPath := filepath.Join(tempDir, "test.log")
    SetLogFile(logPath)
    SetFileLogging(true)

    // Test file logging
    logger := NewLogger()
    logger.Info("Test message for file")

    // Wait for file operations to complete
    time.Sleep(100 * time.Millisecond)

    // Verify file was created
    if _, err := os.Stat(logPath); os.IsNotExist(err) {
        t.Error("Log file was not created")
    }

    // Verify file contents
    content, err := ioutil.ReadFile(logPath)
    if err != nil {
        t.Fatalf("Failed to read log file: %v", err)
    }

    contentStr := string(content)
    if !strings.Contains(contentStr, "Test message for file") {
        t.Error("Expected message not found in log file")
    }
}
```

### File Content Verification

Verify specific file content:
```go
func TestFileContentFormat(t *testing.T) {
    tempDir, err := ioutil.TempDir("", "pim_test")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Setup
    originalEnabled := enableFileLogging
    originalLogFile := logFile
    defer func() {
        enableFileLogging = originalEnabled
        logFile = originalLogFile
    }()

    logPath := filepath.Join(tempDir, "content_test.log")
    SetLogFile(logPath)
    SetFileLogging(true)

    // Log different levels
    logger := NewLogger()
    logger.Info("Info message")
    logger.Error("Error message")
    logger.Debug("Debug message")

    time.Sleep(100 * time.Millisecond)

    // Read and verify content
    content, err := ioutil.ReadFile(logPath)
    if err != nil {
        t.Fatalf("Failed to read log file: %v", err)
    }

    lines := strings.Split(string(content), "\n")
    
    // Remove empty lines
    var nonEmptyLines []string
    for _, line := range lines {
        if strings.TrimSpace(line) != "" {
            nonEmptyLines = append(nonEmptyLines, line)
        }
    }

    expectedLines := 3 // Info, Error, Debug (assuming debug level is enabled)
    if len(nonEmptyLines) != expectedLines {
        t.Errorf("Expected %d log lines, got %d", expectedLines, len(nonEmptyLines))
    }

    // Verify each line contains expected elements
    for i, line := range nonEmptyLines {
        // Check timestamp format
        if !strings.Contains(line, time.Now().Format("2006-01-02")) {
            t.Errorf("Line %d missing timestamp: %s", i, line)
        }
        
        // Check log level
        hasLevel := strings.Contains(line, "[INFO]") || 
                   strings.Contains(line, "[ERROR]") || 
                   strings.Contains(line, "[DEBUG]")
        if !hasLevel {
            t.Errorf("Line %d missing log level: %s", i, line)
        }
    }
}
```

### File Rotation Testing

Test log file rotation:
```go
func TestFileRotation(t *testing.T) {
    tempDir, err := ioutil.TempDir("", "pim_rotation_test")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Setup
    originalEnabled := enableFileLogging
    originalLogFile := logFile
    originalMaxSize := maxLogFileSize
    defer func() {
        enableFileLogging = originalEnabled
        logFile = originalLogFile
        maxLogFileSize = originalMaxSize
    }()

    logPath := filepath.Join(tempDir, "rotation_test.log")
    SetLogFile(logPath)
    SetFileLogging(true)
    maxLogFileSize = 1024 // Small size to trigger rotation

    logger := NewLogger()

    // Write enough data to trigger rotation
    for i := 0; i < 100; i++ {
        logger.Info("This is a long message that will fill up the log file and trigger rotation - message %d", i)
    }

    time.Sleep(200 * time.Millisecond)

    // Check if rotation occurred
    files, err := filepath.Glob(filepath.Join(tempDir, "*.log*"))
    if err != nil {
        t.Fatalf("Failed to list log files: %v", err)
    }

    if len(files) < 2 {
        t.Errorf("Expected at least 2 files after rotation, got %d", len(files))
    }

    // Verify at least one compressed file exists
    hasCompressed := false
    for _, file := range files {
        if strings.HasSuffix(file, ".gz") {
            hasCompressed = true
            break
        }
    }
    
    if !hasCompressed {
        t.Error("Expected at least one compressed (rotated) log file")
    }
}
```

## Performance Testing

### Benchmarking Logging Operations

Create benchmarks for critical paths:
```go
func BenchmarkLogger(b *testing.B) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message %d", i)
    }
}

func BenchmarkLoggerWithCallerInfo(b *testing.B) {
    originalShow := showFileLine
    SetShowFileLine(true)
    defer func() { showFileLine = originalShow }()
    
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message %d", i)
    }
}

func BenchmarkLoggerFiltered(b *testing.B) {
    originalLevel := currentLogLevel
    SetLogLevel(ErrorLevel) // Filter out Info messages
    defer func() { currentLogLevel = originalLevel }()
    
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("This message should be filtered")
    }
}
```

### Memory Allocation Testing

Test memory allocations:
```go
func BenchmarkLoggerAllocs(b *testing.B) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        logger.Info("Allocation test message")
    }
}

func TestLoggerAllocations(t *testing.T) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    // Measure allocations
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Perform logging operations
    for i := 0; i < 1000; i++ {
        logger.Info("Allocation test message %d", i)
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    allocations := m2.TotalAlloc - m1.TotalAlloc
    allocationsPerOp := allocations / 1000
    
    // Set reasonable threshold (adjust based on requirements)
    maxAllocationsPerOp := uint64(200) // bytes
    if allocationsPerOp > maxAllocationsPerOp {
        t.Errorf("Too many allocations per log operation: %d bytes (max: %d)", 
            allocationsPerOp, maxAllocationsPerOp)
    }
}
```

### Concurrent Performance Testing

Test performance under concurrent load:
```go
func BenchmarkLoggerConcurrent(b *testing.B) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            logger.Info("Concurrent benchmark message")
        }
    })
}

func TestConcurrentLoggingPerformance(t *testing.T) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    const numGoroutines = 10
    const messagesPerGoroutine = 1000
    
    start := time.Now()
    
    var wg sync.WaitGroup
    wg.Add(numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            defer wg.Done()
            for j := 0; j < messagesPerGoroutine; j++ {
                logger.Info("Goroutine %d message %d", id, j)
            }
        }(i)
    }
    
    wg.Wait()
    elapsed := time.Since(start)
    
    totalMessages := numGoroutines * messagesPerGoroutine
    messagesPerSecond := float64(totalMessages) / elapsed.Seconds()
    
    // Set reasonable performance threshold
    minMessagesPerSecond := 10000.0
    if messagesPerSecond < minMessagesPerSecond {
        t.Errorf("Performance too slow: %.2f messages/sec (min: %.2f)", 
            messagesPerSecond, minMessagesPerSecond)
    }
    
    t.Logf("Concurrent logging performance: %.2f messages/sec", messagesPerSecond)
}
```

## Thread Safety Testing

### Race Condition Detection

Test for race conditions:
```go
func TestConfigurationRaceConditions(t *testing.T) {
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    const numGoroutines = 50
    const numOperations = 100

    var wg sync.WaitGroup
    wg.Add(numGoroutines)

    // Start multiple goroutines performing concurrent operations
    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                switch j % 4 {
                case 0:
                    SetLogLevel(DebugLevel)
                case 1:
                    SetLogLevel(InfoLevel)
                case 2:
                    _ = GetLogLevel()
                case 3:
                    SetLogLevel(ErrorLevel)
                }
            }
        }(i)
    }

    wg.Wait()

    // Verify final state is valid
    finalLevel := GetLogLevel()
    validLevels := []LogLevel{DebugLevel, InfoLevel, ErrorLevel}
    isValid := false
    for _, valid := range validLevels {
        if finalLevel == valid {
            isValid = true
            break
        }
    }
    
    if !isValid {
        t.Errorf("Invalid final log level after concurrent access: %v", finalLevel)
    }
}
```

### Deadlock Prevention Testing

Test for potential deadlocks:
```go
func TestNoDeadlocks(t *testing.T) {
    // Create a timeout to detect deadlocks
    done := make(chan bool, 1)
    
    go func() {
        // Perform operations that could potentially deadlock
        for i := 0; i < 100; i++ {
            SetLogLevel(DebugLevel)
            _ = GetLogLevel()
            SetFileLogging(true)
            _ = GetFileLogging()
            SetCallerSkipFrames(3)
            _ = GetCallerSkipFrames()
        }
        done <- true
    }()
    
    select {
    case <-done:
        // Test completed successfully
    case <-time.After(5 * time.Second):
        t.Fatal("Test timed out - possible deadlock detected")
    }
}
```

## Mock and Stub Techniques

### Mock File System

Mock file operations for testing:
```go
type MockFileSystem struct {
    files map[string][]byte
    mutex sync.RWMutex
}

func NewMockFileSystem() *MockFileSystem {
    return &MockFileSystem{
        files: make(map[string][]byte),
    }
}

func (mfs *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
    mfs.mutex.Lock()
    defer mfs.mutex.Unlock()
    
    mfs.files[filename] = make([]byte, len(data))
    copy(mfs.files[filename], data)
    return nil
}

func (mfs *MockFileSystem) ReadFile(filename string) ([]byte, error) {
    mfs.mutex.RLock()
    defer mfs.mutex.RUnlock()
    
    data, exists := mfs.files[filename]
    if !exists {
        return nil, os.ErrNotExist
    }
    
    result := make([]byte, len(data))
    copy(result, data)
    return result, nil
}

// Test using mock file system
func TestFileLoggingWithMock(t *testing.T) {
    mockFS := NewMockFileSystem()
    
    // Inject mock file system (would require refactoring to support dependency injection)
    // This is a simplified example showing the concept
    
    logger := NewLogger()
    logger.Info("Test message")
    
    // Verify mock file system received the data
    data, err := mockFS.ReadFile("app.log")
    if err != nil {
        t.Errorf("Expected file to be written: %v", err)
    }
    
    if !strings.Contains(string(data), "Test message") {
        t.Error("Expected message not found in mock file")
    }
}
```

### Time Mocking

Mock time for consistent testing:
```go
type MockTimeProvider struct {
    currentTime time.Time
}

func (mtp *MockTimeProvider) Now() time.Time {
    return mtp.currentTime
}

func (mtp *MockTimeProvider) SetTime(t time.Time) {
    mtp.currentTime = t
}

// Test with mocked time
func TestTimestampFormat(t *testing.T) {
    mockTime := NewMockTimeProvider()
    testTime := time.Date(2023, 1, 15, 10, 30, 45, 0, time.UTC)
    mockTime.SetTime(testTime)
    
    // Would require refactoring to inject time provider
    var output strings.Builder
    logger := NewLoggerWithOutput(&output)
    logger.Info("Test message")
    
    result := output.String()
    expectedTimestamp := "2023-01-15 10:30:45"
    if !strings.Contains(result, expectedTimestamp) {
        t.Errorf("Expected timestamp %s in output: %s", expectedTimestamp, result)
    }
}
```

## Test Utilities and Helpers

### Test Logger Factory

Create helper functions for common test scenarios:
```go
// TestLogger creates a logger with captured output for testing
func TestLogger() (*Logger, *strings.Builder) {
    var output strings.Builder
    logger := NewLoggerWithOutput(&output)
    return logger, &output
}

// TestLoggerWithFile creates a logger with file output in a temp directory
func TestLoggerWithFile(t *testing.T) (*Logger, string, func()) {
    tempDir, err := ioutil.TempDir("", "pim_test")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    
    logPath := filepath.Join(tempDir, "test.log")
    
    // Save original state
    originalEnabled := enableFileLogging
    originalLogFile := logFile
    
    // Setup file logging
    SetLogFile(logPath)
    SetFileLogging(true)
    
    logger := NewLogger()
    
    cleanup := func() {
        enableFileLogging = originalEnabled
        logFile = originalLogFile
        os.RemoveAll(tempDir)
    }
    
    return logger, logPath, cleanup
}

// Usage example
func TestWithHelpers(t *testing.T) {
    t.Run("ConsoleOutput", func(t *testing.T) {
        logger, output := TestLogger()
        logger.Info("Test message")
        
        if !strings.Contains(output.String(), "Test message") {
            t.Error("Expected message not found")
        }
    })
    
    t.Run("FileOutput", func(t *testing.T) {
        logger, logPath, cleanup := TestLoggerWithFile(t)
        defer cleanup()
        
        logger.Info("File test message")
        time.Sleep(100 * time.Millisecond)
        
        content, err := ioutil.ReadFile(logPath)
        if err != nil {
            t.Fatalf("Failed to read log file: %v", err)
        }
        
        if !strings.Contains(string(content), "File test message") {
            t.Error("Expected message not found in file")
        }
    })
}
```

### Configuration Test Helpers

```go
// SaveConfig saves current configuration state
type SavedConfig struct {
    LogLevel        LogLevel
    ShowFileLine    bool
    ShowGoroutineID bool
    FileLogging     bool
    LogFile         string
    CallerSkipFrames int
}

func SaveConfiguration() SavedConfig {
    return SavedConfig{
        LogLevel:         GetLogLevel(),
        ShowFileLine:     GetShowFileLine(),
        ShowGoroutineID:  GetShowGoroutineID(),
        FileLogging:      GetFileLogging(),
        LogFile:          GetLogFile(),
        CallerSkipFrames: GetCallerSkipFrames(),
    }
}

func (sc SavedConfig) Restore() {
    SetLogLevel(sc.LogLevel)
    SetShowFileLine(sc.ShowFileLine)
    SetShowGoroutineID(sc.ShowGoroutineID)
    SetFileLogging(sc.FileLogging)
    SetLogFile(sc.LogFile)
    SetCallerSkipFrames(sc.CallerSkipFrames)
}

// Usage
func TestWithConfigHelper(t *testing.T) {
    saved := SaveConfiguration()
    defer saved.Restore()
    
    // Modify configuration for test
    SetLogLevel(DebugLevel)
    SetFileLogging(true)
    
    // Test logic here
    
    // Configuration automatically restored by defer
}
```

## Test Coverage and Quality

### Measuring Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out
```

### Coverage Goals

Aim for high coverage in critical areas:
- Configuration functions: 100%
- Core logging paths: >95%
- Error handling: >90%
- Edge cases: >80%

### Test Quality Checklist

- [ ] All public functions have tests
- [ ] Error conditions are tested
- [ ] Default values are verified
- [ ] Thread safety is tested
- [ ] Performance is benchmarked
- [ ] File operations are tested
- [ ] Output format is verified
- [ ] Configuration changes are tested
- [ ] Integration scenarios are covered
- [ ] Edge cases are handled

## Next Module

In **Module 5: Adding New Features**, you'll learn how to implement new functionality using the testing patterns and architectural knowledge from previous modules.

Effective testing is essential for maintaining the reliability and performance of the pim package. The patterns and techniques in this module will help you write comprehensive, maintainable tests that catch issues early and ensure the logging system works correctly under all conditions.
