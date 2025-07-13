# Module 5: Adding New Features

## Overview

This module teaches you how to add new features to the pim package using the patterns and knowledge from previous modules. You'll learn the complete process from design to implementation to testing.

### What You'll Learn
- Feature design and planning process
- Implementation patterns and best practices
- Integration with existing architecture
- Comprehensive testing strategies
- Documentation and examples
- Common feature types and their patterns

## Feature Development Process

### 1. Planning Phase

#### Requirements Analysis
Before writing code, clearly define:
- **Purpose**: What problem does this feature solve?
- **Scope**: What should and shouldn't be included?
- **Interface**: How will users interact with this feature?
- **Performance**: What are the performance implications?
- **Compatibility**: Will this break existing functionality?

#### Design Considerations
- **Configuration**: Does this need configuration options?
- **Thread Safety**: Will this be accessed concurrently?
- **Error Handling**: How should errors be handled?
- **Testing**: How will this be tested?
- **Documentation**: What documentation is needed?

### 2. Implementation Strategy

#### Start Small
Begin with the minimal viable implementation:
1. Core functionality only
2. Basic configuration (if needed)
3. Simple tests
4. Basic documentation

#### Iterate and Expand
Add complexity gradually:
1. Advanced configuration options
2. Error handling and validation
3. Performance optimizations
4. Comprehensive tests
5. Examples and documentation

## Feature Implementation Examples

### Example 1: Adding Process ID Display

Let's walk through adding a feature to display process ID in log messages.

#### Step 1: Planning

**Purpose**: Help developers identify which process generated log messages in multi-process environments.

**Interface**: 
```go
SetShowProcessID(bool)
GetShowProcessID() bool
```

**Output**: `[PID:12345]` in log messages when enabled.

#### Step 2: Configuration

Add the configuration variable and functions:

```go
// In config.go
var showProcessID bool = false

// SetShowProcessID enables or disables process ID display in log messages.
// When enabled, log messages will include [PID:12345] showing the current process ID.
func SetShowProcessID(show bool) {
    configMutex.Lock()
    defer configMutex.Unlock()
    showProcessID = show
}

// GetShowProcessID returns whether process ID is currently displayed in log messages.
func GetShowProcessID() bool {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return showProcessID
}
```

#### Step 3: Core Implementation

Integrate with the output system:

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
    
    // Caller information
    if caller != "" {
        parts = append(parts, caller)
    }
    
    return strings.Join(parts, " ")
}
```

#### Step 4: Testing

Write comprehensive tests:

```go
func TestShowProcessID(t *testing.T) {
    originalShow := showProcessID
    defer func() { showProcessID = originalShow }()

    t.Run("DefaultValue", func(t *testing.T) {
        if GetShowProcessID() != false {
            t.Error("Expected ShowProcessID to default to false")
        }
    })

    t.Run("SetAndGet", func(t *testing.T) {
        SetShowProcessID(true)
        if !GetShowProcessID() {
            t.Error("Expected ShowProcessID to be true after setting")
        }

        SetShowProcessID(false)
        if GetShowProcessID() {
            t.Error("Expected ShowProcessID to be false after setting")
        }
    })
}

func TestProcessIDInOutput(t *testing.T) {
    originalShow := showProcessID
    defer func() { showProcessID = originalShow }()

    logger, output := TestLogger()

    t.Run("Disabled", func(t *testing.T) {
        output.Reset()
        SetShowProcessID(false)
        logger.Info("Test message")
        
        result := output.String()
        if strings.Contains(result, "[PID:") {
            t.Error("Expected no process ID when disabled")
        }
    })

    t.Run("Enabled", func(t *testing.T) {
        output.Reset()
        SetShowProcessID(true)
        logger.Info("Test message")
        
        result := output.String()
        expectedPID := fmt.Sprintf("[PID:%d]", os.Getpid())
        if !strings.Contains(result, expectedPID) {
            t.Errorf("Expected %s in output, got: %s", expectedPID, result)
        }
    })
}

func BenchmarkProcessID(b *testing.B) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    SetShowProcessID(true)
    defer SetShowProcessID(false)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message")
    }
}
```

#### Step 5: Documentation

Update relevant documentation:

```go
// SetShowProcessID enables or disables process ID display in log messages.
// When enabled, log messages will include [PID:12345] showing the current process ID.
// This is useful in multi-process environments to identify which process generated each log message.
//
// Default: false
// Performance: Minimal impact (single os.Getpid() call per message when enabled)
//
// Example:
//   pim.SetShowProcessID(true)
//   logger.Info("Application started")  // Output: 2023-01-15 10:30:45 [PID:12345] [INFO] Application started
func SetShowProcessID(show bool) {
    // Implementation
}
```

### Example 2: Adding Custom Log Levels

Let's add support for a custom "AUDIT" log level.

#### Step 1: Planning

**Purpose**: Provide a dedicated log level for audit trail messages.
**Position**: Between INFO and WARNING levels.
**Interface**: Same as existing levels with `logger.Audit()` method.

#### Step 2: Define the Level

```go
// In logger.go or logger_core.go
const (
    PanicLevel LogLevel = iota
    ErrorLevel
    WarningLevel
    AuditLevel    // New level between WARNING and INFO
    InfoLevel
    DebugLevel
    TraceLevel
)
```

#### Step 3: Update String Representation

```go
func (l LogLevel) String() string {
    switch l {
    case PanicLevel:
        return "PANIC"
    case ErrorLevel:
        return "ERROR"
    case WarningLevel:
        return "WARNING"
    case AuditLevel:
        return "AUDIT"  // New case
    case InfoLevel:
        return "INFO"
    case DebugLevel:
        return "DEBUG"
    case TraceLevel:
        return "TRACE"
    default:
        return "UNKNOWN"
    }
}
```

#### Step 4: Add Theming Support

```go
// In theming.go
var (
    AuditColor  = color.New(color.FgMagenta)
    AuditPrefix = ColoredString("[AUDIT]", AuditColor)
)
```

#### Step 5: Add Logger Method

```go
// In logger.go
func (l *Logger) Audit(message string, args ...interface{}) {
    if currentLogLevel >= AuditLevel {
        l.logMessage(AuditLevel, AuditPrefix, message, args...)
    }
}
```

#### Step 6: Testing

```go
func TestAuditLevel(t *testing.T) {
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    t.Run("AuditLevelConstant", func(t *testing.T) {
        if AuditLevel != 3 {
            t.Errorf("Expected AuditLevel to be 3, got %d", AuditLevel)
        }
    })

    t.Run("AuditLevelString", func(t *testing.T) {
        if AuditLevel.String() != "AUDIT" {
            t.Errorf("Expected AuditLevel.String() to be 'AUDIT', got '%s'", AuditLevel.String())
        }
    })

    t.Run("AuditLogging", func(t *testing.T) {
        logger, output := TestLogger()
        SetLogLevel(AuditLevel)
        
        logger.Audit("Audit message")
        result := output.String()
        
        if !strings.Contains(result, "[AUDIT]") {
            t.Error("Expected [AUDIT] prefix in output")
        }
        if !strings.Contains(result, "Audit message") {
            t.Error("Expected audit message in output")
        }
    })

    t.Run("AuditFiltering", func(t *testing.T) {
        logger, output := TestLogger()
        SetLogLevel(WarningLevel) // Should filter out audit messages
        
        logger.Audit("Should not appear")
        result := output.String()
        
        if len(strings.TrimSpace(result)) != 0 {
            t.Error("Expected no output when audit level is filtered")
        }
    })
}

func TestLevelOrdering(t *testing.T) {
    levels := []LogLevel{PanicLevel, ErrorLevel, WarningLevel, AuditLevel, InfoLevel, DebugLevel, TraceLevel}
    
    for i := 1; i < len(levels); i++ {
        if levels[i-1] >= levels[i] {
            t.Errorf("Level ordering incorrect: %v should be less than %v", levels[i-1], levels[i])
        }
    }
}
```

### Example 3: Adding Structured Logging Support

Let's add support for structured logging with key-value pairs.

#### Step 1: Planning

**Purpose**: Enable structured logging for better log parsing and analysis.
**Interface**: `logger.InfoKV(message, key1, value1, key2, value2, ...)`
**Output**: Include key-value pairs in a structured format.

#### Step 2: Define Structured Types

```go
// In logger.go or a new structured.go file
type LogFields map[string]interface{}

type StructuredLogger struct {
    *Logger
    fields LogFields
}

func (l *Logger) WithFields(fields LogFields) *StructuredLogger {
    return &StructuredLogger{
        Logger: l,
        fields: fields,
    }
}

func (sl *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
    newFields := make(LogFields)
    for k, v := range sl.fields {
        newFields[k] = v
    }
    newFields[key] = value
    
    return &StructuredLogger{
        Logger: sl.Logger,
        fields: newFields,
    }
}
```

#### Step 3: Implement Structured Methods

```go
func (sl *StructuredLogger) Info(message string, args ...interface{}) {
    sl.logStructured(InfoLevel, InfoPrefix, message, args...)
}

func (sl *StructuredLogger) Error(message string, args ...interface{}) {
    sl.logStructured(ErrorLevel, ErrorPrefix, message, args...)
}

func (sl *StructuredLogger) logStructured(level LogLevel, prefix, message string, args ...interface{}) {
    if currentLogLevel < level {
        return
    }
    
    formattedMessage := fmt.Sprintf(message, args...)
    
    // Build structured message
    var parts []string
    parts = append(parts, time.Now().Format("2006-01-02 15:04:05"))
    parts = append(parts, prefix)
    parts = append(parts, formattedMessage)
    
    // Add structured fields
    if len(sl.fields) > 0 {
        fieldsPart := sl.formatFields()
        parts = append(parts, fieldsPart)
    }
    
    finalMessage := strings.Join(parts, " ")
    sl.Logger.output.Write([]byte(finalMessage + "\n"))
}

func (sl *StructuredLogger) formatFields() string {
    if len(sl.fields) == 0 {
        return ""
    }
    
    var pairs []string
    for key, value := range sl.fields {
        pairs = append(pairs, fmt.Sprintf("%s=%v", key, value))
    }
    
    return fmt.Sprintf("[%s]", strings.Join(pairs, " "))
}
```

#### Step 4: Add Key-Value Methods

```go
func (l *Logger) InfoKV(message string, kvPairs ...interface{}) {
    l.logKV(InfoLevel, InfoPrefix, message, kvPairs...)
}

func (l *Logger) ErrorKV(message string, kvPairs ...interface{}) {
    l.logKV(ErrorLevel, ErrorPrefix, message, kvPairs...)
}

func (l *Logger) logKV(level LogLevel, prefix, message string, kvPairs ...interface{}) {
    if currentLogLevel < level {
        return
    }
    
    fields := parseKVPairs(kvPairs...)
    structuredLogger := l.WithFields(fields)
    structuredLogger.logStructured(level, prefix, message)
}

func parseKVPairs(pairs ...interface{}) LogFields {
    fields := make(LogFields)
    
    for i := 0; i < len(pairs); i += 2 {
        if i+1 < len(pairs) {
            key, ok := pairs[i].(string)
            if ok {
                fields[key] = pairs[i+1]
            }
        }
    }
    
    return fields
}
```

#### Step 5: Testing

```go
func TestStructuredLogging(t *testing.T) {
    logger, output := TestLogger()

    t.Run("WithFields", func(t *testing.T) {
        output.Reset()
        
        structuredLogger := logger.WithFields(LogFields{
            "userID": 12345,
            "action": "login",
        })
        
        structuredLogger.Info("User logged in")
        result := output.String()
        
        if !strings.Contains(result, "User logged in") {
            t.Error("Expected message in output")
        }
        if !strings.Contains(result, "userID=12345") {
            t.Error("Expected userID field in output")
        }
        if !strings.Contains(result, "action=login") {
            t.Error("Expected action field in output")
        }
    })

    t.Run("WithField", func(t *testing.T) {
        output.Reset()
        
        structuredLogger := logger.WithField("requestID", "abc123").WithField("method", "GET")
        structuredLogger.Info("Request processed")
        
        result := output.String()
        if !strings.Contains(result, "requestID=abc123") {
            t.Error("Expected requestID field in output")
        }
        if !strings.Contains(result, "method=GET") {
            t.Error("Expected method field in output")
        }
    })

    t.Run("KeyValuePairs", func(t *testing.T) {
        output.Reset()
        
        logger.InfoKV("Operation completed", "duration", "150ms", "status", "success")
        result := output.String()
        
        if !strings.Contains(result, "duration=150ms") {
            t.Error("Expected duration field in output")
        }
        if !strings.Contains(result, "status=success") {
            t.Error("Expected status field in output")
        }
    })
}

func TestKVPairsParsing(t *testing.T) {
    testCases := []struct {
        name     string
        input    []interface{}
        expected LogFields
    }{
        {
            name:     "even pairs",
            input:    []interface{}{"key1", "value1", "key2", "value2"},
            expected: LogFields{"key1": "value1", "key2": "value2"},
        },
        {
            name:     "odd pairs",
            input:    []interface{}{"key1", "value1", "key2"},
            expected: LogFields{"key1": "value1"},
        },
        {
            name:     "non-string keys",
            input:    []interface{}{123, "value1", "key2", "value2"},
            expected: LogFields{"key2": "value2"},
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := parseKVPairs(tc.input...)
            
            if len(result) != len(tc.expected) {
                t.Errorf("Expected %d fields, got %d", len(tc.expected), len(result))
            }
            
            for key, expectedValue := range tc.expected {
                if actualValue, exists := result[key]; !exists || actualValue != expectedValue {
                    t.Errorf("Expected %s=%v, got %v", key, expectedValue, actualValue)
                }
            }
        })
    }
}
```

## Advanced Feature Patterns

### Configuration-Driven Features

For features that need extensive configuration:

```go
// Example: Custom output formatters
type OutputFormat int

const (
    TextFormat OutputFormat = iota
    JSONFormat
    XMLFormat
)

var outputFormat OutputFormat = TextFormat

func SetOutputFormat(format OutputFormat) {
    if format < TextFormat || format > XMLFormat {
        return // Invalid format
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    outputFormat = format
}

func GetOutputFormat() OutputFormat {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return outputFormat
}

// Formatter interface
type MessageFormatter interface {
    Format(level LogLevel, message string, timestamp time.Time, fields LogFields) string
}

type TextFormatter struct{}
func (f TextFormatter) Format(level LogLevel, message string, timestamp time.Time, fields LogFields) string {
    // Text formatting implementation
}

type JSONFormatter struct{}
func (f JSONFormatter) Format(level LogLevel, message string, timestamp time.Time, fields LogFields) string {
    // JSON formatting implementation
}
```

### Plugin-Style Features

For extensible features:

```go
// Hook system for custom processing
type LogHook interface {
    Fire(level LogLevel, message string, fields LogFields) error
    Levels() []LogLevel
}

var hooks []LogHook
var hooksMutex sync.RWMutex

func AddHook(hook LogHook) {
    hooksMutex.Lock()
    defer hooksMutex.Unlock()
    hooks = append(hooks, hook)
}

func RemoveHook(hook LogHook) {
    hooksMutex.Lock()
    defer hooksMutex.Unlock()
    
    for i, h := range hooks {
        if h == hook {
            hooks = append(hooks[:i], hooks[i+1:]...)
            break
        }
    }
}

func fireHooks(level LogLevel, message string, fields LogFields) {
    hooksMutex.RLock()
    defer hooksMutex.RUnlock()
    
    for _, hook := range hooks {
        // Check if this hook handles this level
        for _, hookLevel := range hook.Levels() {
            if hookLevel == level {
                go hook.Fire(level, message, fields) // Fire asynchronously
                break
            }
        }
    }
}
```

### Performance-Critical Features

For features that might impact performance:

```go
// Example: Sampling for high-volume logging
type LogSampler struct {
    rate     float64
    counter  uint64
    interval uint64
}

func NewLogSampler(rate float64) *LogSampler {
    return &LogSampler{
        rate:     rate,
        interval: uint64(1.0 / rate),
    }
}

func (s *LogSampler) ShouldSample() bool {
    count := atomic.AddUint64(&s.counter, 1)
    return count%s.interval == 0
}

var sampler *LogSampler
var samplerMutex sync.RWMutex

func SetSamplingRate(rate float64) {
    if rate <= 0 || rate > 1 {
        return
    }
    
    samplerMutex.Lock()
    defer samplerMutex.Unlock()
    sampler = NewLogSampler(rate)
}

func DisableSampling() {
    samplerMutex.Lock()
    defer samplerMutex.Unlock()
    sampler = nil
}

// In logging methods, check sampling
func (l *Logger) Debug(message string, args ...interface{}) {
    if currentLogLevel < DebugLevel {
        return
    }
    
    // Check sampling for debug messages
    samplerMutex.RLock()
    s := sampler
    samplerMutex.RUnlock()
    
    if s != nil && !s.ShouldSample() {
        return
    }
    
    l.logMessage(DebugLevel, DebugPrefix, message, args...)
}
```

## Feature Integration Checklist

When adding a new feature, ensure:

### Core Implementation
- [ ] Feature follows established patterns
- [ ] Thread-safe access to shared state
- [ ] Proper error handling and validation
- [ ] Performance considerations addressed
- [ ] Memory usage optimized

### Configuration
- [ ] Configuration variables added with appropriate defaults
- [ ] Setter/getter functions with validation
- [ ] Thread-safe configuration access
- [ ] Configuration documented

### Testing
- [ ] Unit tests for core functionality
- [ ] Configuration tests (set/get, validation, defaults)
- [ ] Integration tests with existing features
- [ ] Performance benchmarks
- [ ] Thread safety tests
- [ ] Edge case testing

### Documentation
- [ ] Function documentation with examples
- [ ] README updates if needed
- [ ] Example programs demonstrating usage
- [ ] Performance impact documented

### Backward Compatibility
- [ ] No breaking changes to existing API
- [ ] Default behavior unchanged
- [ ] Existing tests still pass
- [ ] Migration guide if needed

## Common Pitfalls

### 1. Breaking Existing Functionality
```go
// ❌ BAD: Changing default behavior
var showTimestamp bool = false // This would break existing users

// ✅ GOOD: Adding new optional behavior
var showMilliseconds bool = false // New feature, disabled by default
```

### 2. Poor Performance Design
```go
// ❌ BAD: Expensive operations in hot path
func (l *Logger) Info(message string, args ...interface{}) {
    // Expensive operation happens even if message is filtered
    expensiveCallerInfo := getDetailedCallerInfo()
    
    if currentLogLevel < InfoLevel {
        return
    }
    // ... rest of logging
}

// ✅ GOOD: Early return before expensive operations
func (l *Logger) Info(message string, args ...interface{}) {
    if currentLogLevel < InfoLevel {
        return // Early return
    }
    
    // Only do expensive work if needed
    expensiveCallerInfo := getDetailedCallerInfo()
    // ... rest of logging
}
```

### 3. Inadequate Testing
```go
// ❌ BAD: Testing only happy path
func TestNewFeature(t *testing.T) {
    SetNewFeature(true)
    if !GetNewFeature() {
        t.Error("Feature should be enabled")
    }
}

// ✅ GOOD: Comprehensive testing
func TestNewFeature(t *testing.T) {
    originalState := getNewFeatureState()
    defer restoreNewFeatureState(originalState)

    // Test default value
    // Test valid inputs
    // Test invalid inputs
    // Test edge cases
    // Test integration with other features
    // Test performance impact
}
```

## Documentation and Examples

### Creating Example Programs

For each significant feature, create an example:

```go
// example/process_id_demo/main.go
package main

import (
    "time"
    "github.com/refactorrom/pim"
)

func main() {
    // Create logger
    logger := pim.NewLogger()
    
    // Show basic logging without process ID
    logger.Info("Starting application")
    
    // Enable process ID display
    pim.SetShowProcessID(true)
    logger.Info("Process ID now visible")
    
    // Show with different log levels
    logger.Error("Error with process ID")
    logger.Warning("Warning with process ID")
    
    // Simulate some work
    time.Sleep(100 * time.Millisecond)
    
    // Disable process ID
    pim.SetShowProcessID(false)
    logger.Info("Process ID hidden again")
}
```

### README Updates

Add feature documentation to README:

```markdown
## Process ID Display

The pim package can optionally display the process ID in log messages, which is useful in multi-process environments.

### Configuration

```go
// Enable process ID display
pim.SetShowProcessID(true)

// Check current setting
enabled := pim.GetShowProcessID()

// Disable process ID display
pim.SetShowProcessID(false)
```

### Example Output

```
2023-01-15 10:30:45 [PID:12345] [INFO] Application started
2023-01-15 10:30:46 [PID:12345] [ERROR] Database connection failed
```

### Performance Impact

Enabling process ID display has minimal performance impact (single `os.Getpid()` call per log message).
```

## Next Steps

After mastering feature development:
- **Module 6**: Performance Optimization and Best Practices
- **Module 7**: Contributing Guidelines and Pull Request Process
- Apply these patterns to implement your own features
- Help review and improve other contributors' features

Remember: Start with simple, well-tested features and gradually add complexity. Each feature should enhance the package's value while maintaining its performance and reliability characteristics.
