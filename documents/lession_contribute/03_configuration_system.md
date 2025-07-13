# Module 3: Configuration System Deep Dive

## Overview

This module focuses on the CLG package's configuration system, teaching you how to work with existing configuration options and add new ones safely and effectively.

### What You'll Learn
- Configuration architecture and patterns
- Thread safety in configuration
- Adding new configuration options
- Validation and error handling
- Testing configuration changes

## Configuration Architecture

### Global Configuration Model

CLG uses a global configuration model where settings affect all logger instances unless overridden at the instance level.

#### Core Configuration Variables
```go
// In config.go
var (
    // Logging behavior
    currentLogLevel   LogLevel = InfoLevel
    
    // Caller information display
    showFileLine      bool = true
    showGoroutineID   bool = false  // Disabled by default
    showFunctionName  bool = true
    showPackageName   bool = true
    showFullPath      bool = false
    
    // Stack trace configuration
    stackDepth        int = 3
    callerSkipFrames  int = 3
    
    // File logging
    enableFileLogging bool = false  // Disabled by default
    logFile           string = "app.log"
    
    // Thread safety
    configMutex sync.RWMutex
)
```

### Configuration Function Patterns

#### Setter Functions
```go
func SetLogLevel(level LogLevel) {
    configMutex.Lock()
    defer configMutex.Unlock()
    currentLogLevel = level
}

func SetShowFileLine(show bool) {
    configMutex.Lock()
    defer configMutex.Unlock()
    showFileLine = show
}

func SetCallerSkipFrames(frames int) {
    if frames < 0 || frames > 15 {
        return // Validation: ignore invalid values
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    callerSkipFrames = frames
}
```

#### Getter Functions
```go
func GetLogLevel() LogLevel {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return currentLogLevel
}

func GetShowFileLine() bool {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return showFileLine
}

func GetCallerSkipFrames() int {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return callerSkipFrames
}
```

## Thread Safety Principles

### Why Thread Safety Matters

In concurrent applications, multiple goroutines may access configuration simultaneously:
```go
// Goroutine 1
go func() {
    pim.SetLogLevel(pim.DebugLevel)
    logger.Debug("Debug message from goroutine 1")
}()

// Goroutine 2
go func() {
    level := pim.GetLogLevel()
    logger.Info("Current level: %v", level)
}()

// Main goroutine
pim.SetFileLogging(true)
logger.Info("File logging enabled")
```

### Read-Write Mutex Pattern

#### Why RWMutex?
- Multiple concurrent reads are safe and efficient
- Writes require exclusive access
- Read operations are much more frequent than writes

```go
var configMutex sync.RWMutex

// Read operations (frequent) - allow concurrency
func GetLogLevel() LogLevel {
    configMutex.RLock()           // Shared lock
    defer configMutex.RUnlock()   // Release shared lock
    return currentLogLevel        // Safe concurrent read
}

// Write operations (infrequent) - require exclusivity
func SetLogLevel(level LogLevel) {
    configMutex.Lock()            // Exclusive lock
    defer configMutex.Unlock()    // Release exclusive lock
    currentLogLevel = level       // Safe exclusive write
}
```

### Atomic Operations Alternative

For simple types, you might consider atomic operations:
```go
import "sync/atomic"

var logLevel int64 = int64(InfoLevel)

func SetLogLevelAtomic(level LogLevel) {
    atomic.StoreInt64(&logLevel, int64(level))
}

func GetLogLevelAtomic() LogLevel {
    return LogLevel(atomic.LoadInt64(&logLevel))
}
```

However, RWMutex is preferred in CLG for consistency and when protecting multiple related variables.

## Adding New Configuration Options

### Step-by-Step Process

#### 1. Define the Variable
```go
// Add to config.go with appropriate default value
var showProcessID bool = false
```

#### 2. Create Setter Function
```go
// SetShowProcessID enables or disables process ID display in log messages.
// When enabled, logs will include [PID:12345] in the output.
func SetShowProcessID(show bool) {
    configMutex.Lock()
    defer configMutex.Unlock()
    showProcessID = show
}
```

#### 3. Create Getter Function
```go
// GetShowProcessID returns whether process ID is currently shown in log messages.
func GetShowProcessID() bool {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return showProcessID
}
```

#### 4. Integration with Logging
```go
// In output.go or logger.go
func formatLogMessage(level LogLevel, prefix, message, caller string) string {
    var parts []string
    
    // Timestamp
    parts = append(parts, time.Now().Format("2006-01-02 15:04:05"))
    
    // Process ID (new feature)
    if GetShowProcessID() {
        parts = append(parts, fmt.Sprintf("[PID:%d]", os.Getpid()))
    }
    
    // Level prefix
    parts = append(parts, prefix)
    
    // Message
    parts = append(parts, message)
    
    // Caller info
    if caller != "" {
        parts = append(parts, caller)
    }
    
    return strings.Join(parts, " ")
}
```

#### 5. Add Tests
```go
func TestShowProcessID(t *testing.T) {
    // Save original state
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
    
    // Verify the actual PID is shown
    expectedPID := fmt.Sprintf("[PID:%d]", os.Getpid())
    if !strings.Contains(result2, expectedPID) {
        t.Errorf("Expected %s in output, got: %s", expectedPID, result2)
    }
}
```

## Input Validation

### Validation Patterns

#### Range Validation
```go
func SetStackDepth(depth int) {
    if depth < 1 || depth > 10 {
        // Invalid depth - ignore silently or return error
        return
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    stackDepth = depth
}
```

#### Enum Validation
```go
func SetLogLevel(level LogLevel) {
    if level < PanicLevel || level > TraceLevel {
        // Invalid log level - ignore
        return
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    currentLogLevel = level
}
```

#### String Validation
```go
func SetLogFile(filename string) {
    if filename == "" {
        // Empty filename - ignore
        return
    }
    
    // Check if directory exists or can be created
    dir := filepath.Dir(filename)
    if dir != "." {
        if err := os.MkdirAll(dir, 0755); err != nil {
            // Cannot create directory - ignore
            return
        }
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    logFile = filename
}
```

### Error-Returning Validation

For cases where you want to inform the caller of validation errors:

```go
// ValidationError represents a configuration validation error
type ValidationError struct {
    Parameter string
    Value     interface{}
    Message   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error for %s=%v: %s", 
        e.Parameter, e.Value, e.Message)
}

// SetCallerSkipFramesSafe validates input and returns an error if invalid
func SetCallerSkipFramesSafe(frames int) error {
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
            Message:   "cannot exceed 20",
        }
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    callerSkipFrames = frames
    return nil
}
```

## Configuration Groups

### Related Configuration Management

Sometimes you need to set multiple related configuration options atomically:

```go
// CallerConfig groups related caller information settings
type CallerConfig struct {
    ShowFileLine     bool
    ShowFunctionName bool
    ShowPackageName  bool
    ShowFullPath     bool
    SkipFrames       int
}

// SetCallerConfig atomically updates all caller-related settings
func SetCallerConfig(config CallerConfig) error {
    // Validate all settings first
    if config.SkipFrames < 0 || config.SkipFrames > 15 {
        return fmt.Errorf("invalid skip frames: %d", config.SkipFrames)
    }
    
    // Apply all changes atomically
    configMutex.Lock()
    defer configMutex.Unlock()
    
    showFileLine = config.ShowFileLine
    showFunctionName = config.ShowFunctionName
    showPackageName = config.ShowPackageName
    showFullPath = config.ShowFullPath
    callerSkipFrames = config.SkipFrames
    
    return nil
}

// GetCallerConfig returns current caller configuration
func GetCallerConfig() CallerConfig {
    configMutex.RLock()
    defer configMutex.RUnlock()
    
    return CallerConfig{
        ShowFileLine:     showFileLine,
        ShowFunctionName: showFunctionName,
        ShowPackageName:  showPackageName,
        ShowFullPath:     showFullPath,
        SkipFrames:       callerSkipFrames,
    }
}
```

## Configuration Testing Patterns

### Test Structure Template

```go
func TestConfigurationOption(t *testing.T) {
    // 1. Save original state
    originalValue := globalConfigVariable
    defer func() { globalConfigVariable = originalValue }()

    // 2. Test default value
    defaultValue := GetConfigurationOption()
    expectedDefault := someExpectedValue
    if defaultValue != expectedDefault {
        t.Errorf("Expected default value %v, got %v", expectedDefault, defaultValue)
    }

    // 3. Test setting values
    testValues := []ValueType{value1, value2, value3}
    for _, testValue := range testValues {
        SetConfigurationOption(testValue)
        if GetConfigurationOption() != testValue {
            t.Errorf("Expected %v, got %v", testValue, GetConfigurationOption())
        }
    }

    // 4. Test validation (if applicable)
    invalidValues := []ValueType{invalidValue1, invalidValue2}
    for _, invalidValue := range invalidValues {
        beforeValue := GetConfigurationOption()
        SetConfigurationOption(invalidValue)
        afterValue := GetConfigurationOption()
        if beforeValue != afterValue {
            t.Errorf("Invalid value %v should not change configuration", invalidValue)
        }
    }
}
```

### Comprehensive Test Example

```go
func TestFileLoggingConfiguration(t *testing.T) {
    // Save original state
    originalEnabled := enableFileLogging
    originalFile := logFile
    defer func() {
        enableFileLogging = originalEnabled
        logFile = originalFile
    }()

    t.Run("DefaultValues", func(t *testing.T) {
        // File logging should be disabled by default
        if GetFileLogging() != false {
            t.Error("Expected file logging to be disabled by default")
        }
    })

    t.Run("EnableDisable", func(t *testing.T) {
        // Test enabling
        SetFileLogging(true)
        if !GetFileLogging() {
            t.Error("Expected file logging to be enabled")
        }

        // Test disabling  
        SetFileLogging(false)
        if GetFileLogging() {
            t.Error("Expected file logging to be disabled")
        }
    })

    t.Run("FileNameConfiguration", func(t *testing.T) {
        testFiles := []string{
            "test.log",
            "logs/app.log",
            "/tmp/myapp.log",
        }

        for _, testFile := range testFiles {
            SetLogFile(testFile)
            if GetLogFile() != testFile {
                t.Errorf("Expected log file %s, got %s", testFile, GetLogFile())
            }
        }
    })

    t.Run("ValidationLogFile", func(t *testing.T) {
        originalFile := GetLogFile()
        
        // Test empty filename (should be ignored)
        SetLogFile("")
        if GetLogFile() != originalFile {
            t.Error("Empty filename should not change log file")
        }
    })
}
```

### Thread Safety Testing

```go
func TestConfigurationThreadSafety(t *testing.T) {
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    // Start multiple goroutines modifying configuration
    const numGoroutines = 100
    const numOperations = 100

    var wg sync.WaitGroup
    wg.Add(numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                // Mix of read and write operations
                if j%2 == 0 {
                    SetLogLevel(LogLevel(j % 6)) // Write
                } else {
                    _ = GetLogLevel() // Read
                }
            }
        }(i)
    }

    wg.Wait()

    // Verify configuration is still in valid state
    level := GetLogLevel()
    if level < PanicLevel || level > TraceLevel {
        t.Errorf("Invalid log level after concurrent access: %v", level)
    }
}
```

## Best Practices

### 1. Default Values
- Choose sensible defaults that work for most use cases
- Document why specific defaults were chosen
- Consider backward compatibility when changing defaults

```go
// Good defaults with reasoning
var (
    showGoroutineID   bool = false  // Performance: goroutine lookup is expensive
    enableFileLogging bool = false  // Safety: don't create files without explicit request
    callerSkipFrames  int  = 3      // Empirical: works for most wrapper layers
)
```

### 2. Validation Strategy
- Fail silently for invalid values (ignore)
- Return errors for functions that need error handling
- Log validation failures if appropriate

```go
// Silent validation (preferred for simple setters)
func SetStackDepth(depth int) {
    if depth < 1 || depth > 10 {
        return // Ignore invalid values
    }
    // ... set value
}

// Error-returning validation (for complex operations)
func SetLogFileWithValidation(filename string) error {
    if err := validateLogFile(filename); err != nil {
        return fmt.Errorf("invalid log file: %w", err)
    }
    // ... set value
    return nil
}
```

### 3. Documentation
- Document the purpose and valid range of each option
- Provide examples of when to use different settings
- Explain performance implications

```go
// SetCallerSkipFrames configures how many stack frames to skip when
// determining caller information. This is useful when wrapping the logger
// in other functions.
//
// Valid range: 0-15
// Default: 3 (works for most cases)
// Performance: Higher values require more stack walking
//
// Example:
//   SetCallerSkipFrames(5) // Skip 5 frames to show actual caller
//   logger.Info("message") // Will show caller 5 frames up the stack
func SetCallerSkipFrames(frames int) {
    // Implementation
}
```

### 4. Testing
- Always test default values
- Test validation boundaries
- Include thread safety tests for critical paths
- Test integration with actual logging output

## Common Pitfalls

### 1. Race Conditions
```go
// ❌ BAD: Not thread-safe
func GetLogLevelUnsafe() LogLevel {
    return currentLogLevel // Race condition!
}

// ✅ GOOD: Thread-safe
func GetLogLevel() LogLevel {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return currentLogLevel
}
```

### 2. Validation Inconsistency
```go
// ❌ BAD: Inconsistent validation
func SetFrames1(frames int) {
    if frames < 0 { return }  // Only checks negative
    callerSkipFrames = frames
}

func SetFrames2(frames int) {
    if frames < 0 || frames > 20 { return }  // Different validation
    callerSkipFrames = frames
}

// ✅ GOOD: Consistent validation
func validateFrames(frames int) bool {
    return frames >= 0 && frames <= 15
}

func SetFrames(frames int) {
    if !validateFrames(frames) { return }
    callerSkipFrames = frames
}
```

### 3. Poor Test Cleanup
```go
// ❌ BAD: Doesn't restore state
func TestBadCleanup(t *testing.T) {
    SetLogLevel(DebugLevel)
    // Test logic...
    // Original state not restored!
}

// ✅ GOOD: Proper cleanup
func TestGoodCleanup(t *testing.T) {
    original := GetLogLevel()
    defer SetLogLevel(original)
    
    SetLogLevel(DebugLevel)
    // Test logic...
    // Original state restored automatically
}
```

## Next Module

In **Module 4: Writing Effective Tests**, you'll learn advanced testing techniques specifically for logging systems, including output capture, file testing, and performance testing.

The configuration system is the foundation of CLG's flexibility. Understanding these patterns will help you add new features that integrate seamlessly with the existing architecture.
