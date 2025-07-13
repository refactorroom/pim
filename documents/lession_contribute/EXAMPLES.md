# Practical Examples for Contributors

This document provides hands-on examples for common contribution scenarios to the CLG logging package.

## Table of Contents
1. [Adding a New Log Level](#adding-a-new-log-level)
2. [Adding Configuration Options](#adding-configuration-options)
3. [Creating Custom Output Formatters](#creating-custom-output-formatters)
4. [Adding Validation and Error Handling](#adding-validation-and-error-handling)
5. [Writing Comprehensive Tests](#writing-comprehensive-tests)
6. [Performance Optimization Examples](#performance-optimization-examples)

## Adding a New Log Level

### Scenario
You want to add a "FATAL" log level that sits between ERROR and PANIC.

### Step 1: Define the Log Level
```go
// In logger.go or logger_core.go
const (
    PanicLevel LogLevel = iota
    FatalLevel          // New level
    ErrorLevel
    WarningLevel
    InfoLevel
    DebugLevel
    TraceLevel
)
```

### Step 2: Add String Representation
```go
// Update the String() method for LogLevel
func (l LogLevel) String() string {
    switch l {
    case PanicLevel:
        return "PANIC"
    case FatalLevel:
        return "FATAL"  // New case
    case ErrorLevel:
        return "ERROR"
    // ... rest of cases
    default:
        return "UNKNOWN"
    }
}
```

### Step 3: Add Formatting Support
```go
// In theming.go or output.go
var (
    FatalPrefix = ColoredString("[FATAL]", FatalColor)
    FatalColor  = Red // or define a new color
)
```

### Step 4: Add Logger Method
```go
// In logger.go
func (l *Logger) Fatal(message string, args ...interface{}) {
    if currentLogLevel >= FatalLevel {
        l.logMessage(FatalLevel, FatalPrefix, message, args...)
    }
}
```

### Step 5: Write Tests
```go
// In a new test file or existing one
func TestFatalLevel(t *testing.T) {
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    // Test that Fatal logs when level is appropriate
    SetLogLevel(FatalLevel)
    
    var output strings.Builder
    logger := NewLoggerWithOutput(&output)
    logger.Fatal("Fatal error occurred")
    
    result := output.String()
    if !strings.Contains(result, "[FATAL]") {
        t.Error("Expected [FATAL] prefix in output")
    }
    if !strings.Contains(result, "Fatal error occurred") {
        t.Error("Expected message in output")
    }
}

func TestFatalLevelFiltering(t *testing.T) {
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    // Test that Fatal doesn't log when level is too low
    SetLogLevel(PanicLevel)
    
    var output strings.Builder
    logger := NewLoggerWithOutput(&output)
    logger.Fatal("Should not appear")
    
    if output.Len() != 0 {
        t.Error("Expected no output when log level filters out Fatal")
    }
}
```

## Adding Configuration Options

### Scenario
You want to add a configuration option to show/hide process ID in log messages.

### Step 1: Add Global Variable
```go
// In config.go
var (
    showProcessID bool = false
    configMutex   sync.RWMutex
)
```

### Step 2: Add Configuration Functions
```go
// SetShowProcessID enables or disables process ID in log messages
func SetShowProcessID(show bool) {
    configMutex.Lock()
    defer configMutex.Unlock()
    showProcessID = show
}

// GetShowProcessID returns whether process ID is shown in log messages
func GetShowProcessID() bool {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return showProcessID
}
```

### Step 3: Integrate with Log Formatting
```go
// In logger.go or output.go
func formatLogMessage(level LogLevel, prefix, message, caller string) string {
    var parts []string
    
    // Add timestamp
    parts = append(parts, time.Now().Format("2006-01-02 15:04:05"))
    
    // Add process ID if enabled
    if GetShowProcessID() {
        parts = append(parts, fmt.Sprintf("[PID:%d]", os.Getpid()))
    }
    
    // Add other components
    parts = append(parts, prefix, message)
    
    if caller != "" {
        parts = append(parts, caller)
    }
    
    return strings.Join(parts, " ")
}
```

### Step 4: Write Configuration Tests
```go
func TestShowProcessID(t *testing.T) {
    originalShow := showProcessID
    defer func() { showProcessID = originalShow }()

    // Test default value
    if GetShowProcessID() != false {
        t.Error("Expected ShowProcessID to default to false")
    }

    // Test setting to true
    SetShowProcessID(true)
    if !GetShowProcessID() {
        t.Error("Expected ShowProcessID to be true after setting")
    }

    // Test setting to false
    SetShowProcessID(false)
    if GetShowProcessID() {
        t.Error("Expected ShowProcessID to be false after setting")
    }
}

func TestProcessIDInOutput(t *testing.T) {
    originalShow := showProcessID
    defer func() { showProcessID = originalShow }()

    var output strings.Builder
    logger := NewLoggerWithOutput(&output)

    // Test without process ID
    SetShowProcessID(false)
    logger.Info("Test message")
    result1 := output.String()
    
    if strings.Contains(result1, "[PID:") {
        t.Error("Expected no process ID in output when disabled")
    }

    // Test with process ID
    output.Reset()
    SetShowProcessID(true)
    logger.Info("Test message")
    result2 := output.String()
    
    if !strings.Contains(result2, "[PID:") {
        t.Error("Expected process ID in output when enabled")
    }
}
```

## Creating Custom Output Formatters

### Scenario
You want to create a structured formatter that outputs logs in a specific format.

### Step 1: Define Formatter Interface
```go
// In a new file formatters.go or in output.go
type OutputFormatter interface {
    Format(level LogLevel, message string, callerInfo CallerInfo, timestamp time.Time) string
}

type CallerInfo struct {
    File     string
    Line     int
    Function string
    Package  string
}
```

### Step 2: Implement Custom Formatter
```go
// JSONFormatter outputs logs in JSON format
type JSONFormatter struct {
    IncludeTimestamp bool
    IncludeCaller    bool
}

func NewJSONFormatter() *JSONFormatter {
    return &JSONFormatter{
        IncludeTimestamp: true,
        IncludeCaller:    true,
    }
}

func (j *JSONFormatter) Format(level LogLevel, message string, callerInfo CallerInfo, timestamp time.Time) string {
    logEntry := map[string]interface{}{
        "level":   level.String(),
        "message": message,
    }

    if j.IncludeTimestamp {
        logEntry["timestamp"] = timestamp.UTC().Format(time.RFC3339)
    }

    if j.IncludeCaller && callerInfo.File != "" {
        logEntry["caller"] = map[string]interface{}{
            "file":     callerInfo.File,
            "line":     callerInfo.Line,
            "function": callerInfo.Function,
            "package":  callerInfo.Package,
        }
    }

    jsonBytes, err := json.Marshal(logEntry)
    if err != nil {
        // Fallback to simple format if JSON marshaling fails
        return fmt.Sprintf(`{"level":"%s","message":"%s","error":"json_marshal_failed"}`, 
            level.String(), message)
    }

    return string(jsonBytes)
}

// KeyValueFormatter outputs logs in key=value format
type KeyValueFormatter struct {
    Separator string
}

func NewKeyValueFormatter() *KeyValueFormatter {
    return &KeyValueFormatter{Separator: " "}
}

func (k *KeyValueFormatter) Format(level LogLevel, message string, callerInfo CallerInfo, timestamp time.Time) string {
    var parts []string
    
    parts = append(parts, fmt.Sprintf("level=%s", level.String()))
    parts = append(parts, fmt.Sprintf("time=%s", timestamp.Format(time.RFC3339)))
    parts = append(parts, fmt.Sprintf("msg=%q", message))
    
    if callerInfo.File != "" {
        parts = append(parts, fmt.Sprintf("caller=%s:%d", callerInfo.File, callerInfo.Line))
    }
    
    return strings.Join(parts, k.Separator)
}
```

### Step 3: Integrate with Logger
```go
// Add formatter support to Logger struct
type Logger struct {
    output    io.Writer
    formatter OutputFormatter
    config    LoggerConfig
}

func (l *Logger) SetFormatter(formatter OutputFormatter) {
    l.formatter = formatter
}

func (l *Logger) logMessage(level LogLevel, prefix, message string, args ...interface{}) {
    if currentLogLevel < level {
        return
    }

    formattedMessage := fmt.Sprintf(message, args...)
    
    // Get caller info
    callerInfo := CallerInfo{}
    if GetShowFileLine() {
        file, line, function, pkg := getFileInfo(GetCallerSkipFrames())
        callerInfo = CallerInfo{
            File:     file,
            Line:     line,
            Function: function,
            Package:  pkg,
        }
    }

    // Use custom formatter if set
    var output string
    if l.formatter != nil {
        output = l.formatter.Format(level, formattedMessage, callerInfo, time.Now())
    } else {
        // Use default formatting
        output = l.defaultFormat(level, prefix, formattedMessage, callerInfo)
    }

    l.output.Write([]byte(output + "\n"))
}
```

### Step 4: Write Formatter Tests
```go
func TestJSONFormatter(t *testing.T) {
    formatter := NewJSONFormatter()
    
    callerInfo := CallerInfo{
        File:     "test.go",
        Line:     42,
        Function: "TestFunction",
        Package:  "main",
    }
    
    timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
    output := formatter.Format(InfoLevel, "test message", callerInfo, timestamp)
    
    // Parse the JSON to verify structure
    var parsed map[string]interface{}
    err := json.Unmarshal([]byte(output), &parsed)
    if err != nil {
        t.Fatalf("Failed to parse JSON output: %v", err)
    }
    
    // Check required fields
    if parsed["level"] != "INFO" {
        t.Errorf("Expected level 'INFO', got %v", parsed["level"])
    }
    
    if parsed["message"] != "test message" {
        t.Errorf("Expected message 'test message', got %v", parsed["message"])
    }
    
    // Check caller info
    caller, ok := parsed["caller"].(map[string]interface{})
    if !ok {
        t.Error("Expected caller info to be an object")
    } else {
        if caller["file"] != "test.go" {
            t.Errorf("Expected file 'test.go', got %v", caller["file"])
        }
        if caller["line"] != float64(42) { // JSON numbers are float64
            t.Errorf("Expected line 42, got %v", caller["line"])
        }
    }
}

func TestKeyValueFormatter(t *testing.T) {
    formatter := NewKeyValueFormatter()
    
    callerInfo := CallerInfo{File: "test.go", Line: 42}
    timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
    output := formatter.Format(InfoLevel, "test message", callerInfo, timestamp)
    
    expected := `level=INFO time=2023-01-01T12:00:00Z msg="test message" caller=test.go:42`
    if output != expected {
        t.Errorf("Expected:\n%s\nGot:\n%s", expected, output)
    }
}
```

## Adding Validation and Error Handling

### Scenario
You want to add robust validation for configuration parameters.

### Step 1: Define Validation Functions
```go
// validation.go
package main

import (
    "fmt"
    "path/filepath"
    "strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
    Parameter string
    Value     interface{}
    Message   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error for %s=%v: %s", e.Parameter, e.Value, e.Message)
}

// ValidateLogLevel checks if a log level is valid
func ValidateLogLevel(level LogLevel) error {
    if level < PanicLevel || level > TraceLevel {
        return &ValidationError{
            Parameter: "logLevel",
            Value:     level,
            Message:   "must be between PanicLevel and TraceLevel",
        }
    }
    return nil
}

// ValidateCallerSkipFrames checks if skip frames value is reasonable
func ValidateCallerSkipFrames(frames int) error {
    if frames < 0 {
        return &ValidationError{
            Parameter: "callerSkipFrames",
            Value:     frames,
            Message:   "cannot be negative",
        }
    }
    if frames > 20 {
        return &ValidationError{
            Parameter: "callerSkipFrames",
            Value:     frames,
            Message:   "cannot exceed 20 (likely too high)",
        }
    }
    return nil
}

// ValidateLogFile checks if a log file path is valid
func ValidateLogFile(path string) error {
    if path == "" {
        return &ValidationError{
            Parameter: "logFile",
            Value:     path,
            Message:   "cannot be empty",
        }
    }
    
    // Check if directory exists or can be created
    dir := filepath.Dir(path)
    if dir != "." {
        // Try to create directory if it doesn't exist
        if err := os.MkdirAll(dir, 0755); err != nil {
            return &ValidationError{
                Parameter: "logFile",
                Value:     path,
                Message:   fmt.Sprintf("cannot create directory: %v", err),
            }
        }
    }
    
    // Check for invalid characters
    if strings.ContainsAny(filepath.Base(path), `<>:"|?*`) {
        return &ValidationError{
            Parameter: "logFile",
            Value:     path,
            Message:   "contains invalid characters",
        }
    }
    
    return nil
}
```

### Step 2: Update Configuration Functions with Validation
```go
// SetLogLevelSafe sets the log level with validation
func SetLogLevelSafe(level LogLevel) error {
    if err := ValidateLogLevel(level); err != nil {
        return err
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    currentLogLevel = level
    return nil
}

// SetCallerSkipFramesSafe sets caller skip frames with validation
func SetCallerSkipFramesSafe(frames int) error {
    if err := ValidateCallerSkipFrames(frames); err != nil {
        return err
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    callerSkipFrames = frames
    return nil
}

// SetLogFileSafe sets the log file with validation
func SetLogFileSafe(path string) error {
    if err := ValidateLogFile(path); err != nil {
        return err
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    logFile = path
    return nil
}
```

### Step 3: Write Validation Tests
```go
func TestValidateLogLevel(t *testing.T) {
    testCases := []struct {
        level     LogLevel
        shouldErr bool
        name      string
    }{
        {PanicLevel, false, "valid panic level"},
        {InfoLevel, false, "valid info level"},
        {TraceLevel, false, "valid trace level"},
        {LogLevel(-1), true, "invalid negative level"},
        {LogLevel(100), true, "invalid high level"},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateLogLevel(tc.level)
            if tc.shouldErr && err == nil {
                t.Error("Expected validation error but got none")
            }
            if !tc.shouldErr && err != nil {
                t.Errorf("Expected no error but got: %v", err)
            }
        })
    }
}

func TestSetLogLevelSafe(t *testing.T) {
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    // Test valid level
    err := SetLogLevelSafe(DebugLevel)
    if err != nil {
        t.Errorf("Expected no error for valid level, got: %v", err)
    }
    if currentLogLevel != DebugLevel {
        t.Error("Log level was not set correctly")
    }

    // Test invalid level
    err = SetLogLevelSafe(LogLevel(-1))
    if err == nil {
        t.Error("Expected validation error for invalid level")
    }
    if currentLogLevel != DebugLevel {
        t.Error("Log level should not have changed after validation error")
    }
}

func TestValidateCallerSkipFrames(t *testing.T) {
    testCases := []struct {
        frames    int
        shouldErr bool
        name      string
    }{
        {0, false, "zero frames"},
        {5, false, "normal frames"},
        {20, false, "maximum frames"},
        {-1, true, "negative frames"},
        {21, true, "too many frames"},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateCallerSkipFrames(tc.frames)
            if tc.shouldErr && err == nil {
                t.Error("Expected validation error but got none")
            }
            if !tc.shouldErr && err != nil {
                t.Errorf("Expected no error but got: %v", err)
            }
        })
    }
}
```

## Writing Comprehensive Tests

### Scenario
You want to create a comprehensive test suite for the file logging feature.

### Complete Test Example
```go
// file_logging_comprehensive_test.go
package main

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

func TestFileLoggingComprehensive(t *testing.T) {
    // Setup: Create temporary directory for test logs
    tempDir, err := ioutil.TempDir("", "clg_test_logs")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir) // Cleanup

    // Store original state
    originalFileLogging := enableFileLogging
    originalLogFile := logFile
    defer func() {
        enableFileLogging = originalFileLogging
        logFile = originalLogFile
    }()

    t.Run("FileLoggingDisabledByDefault", func(t *testing.T) {
        if GetFileLogging() {
            t.Error("File logging should be disabled by default")
        }
    })

    t.Run("EnableFileLogging", func(t *testing.T) {
        logPath := filepath.Join(tempDir, "test.log")
        
        SetLogFile(logPath)
        SetFileLogging(true)
        
        if !GetFileLogging() {
            t.Error("File logging should be enabled")
        }
        
        // Test that logger creates the file
        logger := NewLogger()
        logger.Info("Test message for file logging")
        
        // Wait a moment for file operations
        time.Sleep(100 * time.Millisecond)
        
        // Check if file was created
        if _, err := os.Stat(logPath); os.IsNotExist(err) {
            t.Error("Log file was not created")
        }
        
        // Check file contents
        content, err := ioutil.ReadFile(logPath)
        if err != nil {
            t.Fatalf("Failed to read log file: %v", err)
        }
        
        contentStr := string(content)
        if !strings.Contains(contentStr, "Test message for file logging") {
            t.Error("Expected message not found in log file")
        }
        if !strings.Contains(contentStr, "[INFO]") {
            t.Error("Expected log level not found in log file")
        }
    })

    t.Run("DisableFileLogging", func(t *testing.T) {
        logPath := filepath.Join(tempDir, "disabled_test.log")
        
        SetLogFile(logPath)
        SetFileLogging(false)
        
        logger := NewLogger()
        logger.Info("This should not go to file")
        
        // Wait a moment
        time.Sleep(100 * time.Millisecond)
        
        // File should not exist or be empty
        if _, err := os.Stat(logPath); err == nil {
            content, _ := ioutil.ReadFile(logPath)
            if len(content) > 0 {
                t.Error("Log file should be empty when file logging is disabled")
            }
        }
    })

    t.Run("FileLoggingWithDifferentLevels", func(t *testing.T) {
        logPath := filepath.Join(tempDir, "levels_test.log")
        
        SetLogFile(logPath)
        SetFileLogging(true)
        SetLogLevel(TraceLevel)
        
        logger := NewLogger()
        logger.Panic("Panic message")
        logger.Error("Error message")
        logger.Warning("Warning message")
        logger.Info("Info message")
        logger.Debug("Debug message")
        logger.Trace("Trace message")
        
        time.Sleep(200 * time.Millisecond)
        
        content, err := ioutil.ReadFile(logPath)
        if err != nil {
            t.Fatalf("Failed to read log file: %v", err)
        }
        
        contentStr := string(content)
        expectedMessages := []string{
            "[PANIC]", "Panic message",
            "[ERROR]", "Error message",
            "[WARNING]", "Warning message",
            "[INFO]", "Info message",
            "[DEBUG]", "Debug message",
            "[TRACE]", "Trace message",
        }
        
        for _, expected := range expectedMessages {
            if !strings.Contains(contentStr, expected) {
                t.Errorf("Expected '%s' not found in log file", expected)
            }
        }
    })

    t.Run("FileLoggingWithLevelFiltering", func(t *testing.T) {
        logPath := filepath.Join(tempDir, "filtering_test.log")
        
        SetLogFile(logPath)
        SetFileLogging(true)
        SetLogLevel(WarningLevel) // Only warning and above
        
        logger := NewLogger()
        logger.Error("Should appear")
        logger.Warning("Should appear")
        logger.Info("Should not appear")
        logger.Debug("Should not appear")
        
        time.Sleep(100 * time.Millisecond)
        
        content, err := ioutil.ReadFile(logPath)
        if err != nil {
            t.Fatalf("Failed to read log file: %v", err)
        }
        
        contentStr := string(content)
        
        // Should contain
        if !strings.Contains(contentStr, "Should appear") {
            t.Error("Expected messages not found in log file")
        }
        
        // Should not contain
        if strings.Contains(contentStr, "Should not appear") {
            t.Error("Filtered messages found in log file")
        }
    })

    t.Run("ConcurrentFileLogging", func(t *testing.T) {
        logPath := filepath.Join(tempDir, "concurrent_test.log")
        
        SetLogFile(logPath)
        SetFileLogging(true)
        
        // Start multiple goroutines writing to the log
        done := make(chan bool, 10)
        
        for i := 0; i < 10; i++ {
            go func(id int) {
                logger := NewLogger()
                for j := 0; j < 5; j++ {
                    logger.Info("Goroutine %d message %d", id, j)
                }
                done <- true
            }(i)
        }
        
        // Wait for all goroutines to complete
        for i := 0; i < 10; i++ {
            <-done
        }
        
        time.Sleep(200 * time.Millisecond)
        
        content, err := ioutil.ReadFile(logPath)
        if err != nil {
            t.Fatalf("Failed to read log file: %v", err)
        }
        
        // Count the number of log lines
        lines := strings.Split(string(content), "\n")
        nonEmptyLines := 0
        for _, line := range lines {
            if strings.TrimSpace(line) != "" {
                nonEmptyLines++
            }
        }
        
        expectedLines := 50 // 10 goroutines * 5 messages each
        if nonEmptyLines != expectedLines {
            t.Errorf("Expected %d log lines, got %d", expectedLines, nonEmptyLines)
        }
    })
}
```

## Performance Optimization Examples

### Scenario
You want to optimize the logging performance for high-throughput applications.

### Buffer Pool Optimization
```go
// performance.go
package main

import (
    "bytes"
    "sync"
)

// Pre-allocated buffer pool for reducing allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

// getBuffer gets a buffer from the pool
func getBuffer() *bytes.Buffer {
    return bufferPool.Get().(*bytes.Buffer)
}

// putBuffer returns a buffer to the pool
func putBuffer(buf *bytes.Buffer) {
    buf.Reset()
    bufferPool.Put(buf)
}

// Optimized log message formatting
func (l *Logger) logMessageOptimized(level LogLevel, prefix, message string, args ...interface{}) {
    if currentLogLevel < level {
        return // Early return for filtered levels
    }

    // Get buffer from pool
    buf := getBuffer()
    defer putBuffer(buf)

    // Build message efficiently
    buf.WriteString(time.Now().Format("2006-01-02 15:04:05"))
    buf.WriteByte(' ')
    buf.WriteString(prefix)
    buf.WriteByte(' ')
    
    if len(args) > 0 {
        fmt.Fprintf(buf, message, args...)
    } else {
        buf.WriteString(message)
    }

    // Add caller info if needed
    if GetShowFileLine() {
        file, line, _, _ := getFileInfo(GetCallerSkipFrames())
        fmt.Fprintf(buf, " [%s:%d]", file, line)
    }

    buf.WriteByte('\n')

    // Write to output
    l.output.Write(buf.Bytes())

    // Write to file if enabled
    if GetFileLogging() {
        writeToLogFile(buf.Bytes())
    }
}
```

### Benchmark Tests
```go
// benchmark_test.go
package main

import (
    "io/ioutil"
    "testing"
)

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

func BenchmarkLoggerConcurrent(b *testing.B) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            logger.Info("Concurrent benchmark message")
        }
    })
}
```

### Memory Optimization
```go
// memory_optimization.go
package main

import (
    "strings"
    "unsafe"
)

// stringBuilder is a memory-efficient string builder
type stringBuilder struct {
    parts []string
    size  int
}

func newStringBuilder(estimatedSize int) *stringBuilder {
    return &stringBuilder{
        parts: make([]string, 0, 8), // Pre-allocate for common case
        size:  0,
    }
}

func (sb *stringBuilder) add(s string) {
    sb.parts = append(sb.parts, s)
    sb.size += len(s)
}

func (sb *stringBuilder) build() string {
    if len(sb.parts) == 0 {
        return ""
    }
    if len(sb.parts) == 1 {
        return sb.parts[0]
    }
    
    // Pre-allocate the result string with known size
    result := make([]byte, 0, sb.size)
    for _, part := range sb.parts {
        result = append(result, part...)
    }
    
    // Convert to string without additional allocation
    return *(*string)(unsafe.Pointer(&result))
}

// Optimized message formatting using string builder
func formatMessageOptimized(level LogLevel, prefix, message, caller string) string {
    builder := newStringBuilder(len(prefix) + len(message) + len(caller) + 32)
    
    builder.add(time.Now().Format("2006-01-02 15:04:05"))
    builder.add(" ")
    builder.add(prefix)
    builder.add(" ")
    builder.add(message)
    
    if caller != "" {
        builder.add(" ")
        builder.add(caller)
    }
    
    return builder.build()
}
```

These examples demonstrate common patterns and best practices for contributing to the CLG logging package. Each example includes complete implementation details, comprehensive tests, and follows the established patterns in the codebase.

Remember to:
1. Always include tests for new functionality
2. Follow the existing code style and patterns
3. Consider performance implications
4. Validate input parameters
5. Maintain backward compatibility
6. Document your changes clearly
