# pim Contributor Quick Reference

## Quick Start Commands

```bash
# Clone and setup
git clone <repo-url>
cd pim
go mod tidy

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific tests
go test -run TestConfig

# Run benchmarks
go test -bench=.

# Build examples
cd example/basic_logger && go run basic_logger.go
```

## Code Patterns Cheat Sheet

### Adding Configuration
```go
// 1. Add global variable
var newFeature bool = false

// 2. Add functions
func SetNewFeature(enabled bool) { /* thread-safe implementation */ }
func GetNewFeature() bool { /* thread-safe implementation */ }

// 3. Add tests
func TestNewFeature(t *testing.T) { /* test default, set, get */ }
```

### Adding Log Level
```go
// 1. Add constant
const NewLevel LogLevel = X

// 2. Add String() case
case NewLevel: return "NEW"

// 3. Add logger method
func (l *Logger) New(message string, args ...interface{}) { /* implementation */ }

// 4. Add tests
func TestNewLevel(t *testing.T) { /* test logging and filtering */ }
```

### Writing Tests
```go
func TestFeature(t *testing.T) {
    // Save original state
    original := globalVar
    defer func() { globalVar = original }()

    // Test logic
    SetFeature(true)
    if !GetFeature() {
        t.Error("Feature should be enabled")
    }
}
```

## File Structure Quick Reference

| File | Purpose |
|------|---------|
| `config.go` | Global configuration and settings |
| `logger.go` | Main logging interface and methods |
| `logger_core.go` | Core logger implementation |
| `caller_info.go` | Stack trace and caller info |
| `metrics.go` | File logging and performance |
| `output.go` | Output formatting |
| `writers.go` | File writers and rotation |
| `theming.go` | Colors and themes |
| `*_test.go` | Test files |

## Common Functions Reference

### Configuration
```go
SetLogLevel(level LogLevel)
GetLogLevel() LogLevel
SetShowFileLine(show bool)
GetShowFileLine() bool
SetCallerSkipFrames(frames int)
GetCallerSkipFrames() int
SetFileLogging(enabled bool)
GetFileLogging() bool
SetShowGoroutineID(show bool)
GetShowGoroutineID() bool
```

### Logger Methods
```go
logger.Info(message string, args ...interface{})
logger.Error(message string, args ...interface{})
logger.Warning(message string, args ...interface{})
logger.Debug(message string, args ...interface{})
logger.Trace(message string, args ...interface{})
logger.Panic(message string, args ...interface{})
```

### Utility Functions
```go
NewLogger() *Logger
NewLoggerWithOutput(output io.Writer) *Logger
getFileInfo(skipFrames int) (file, line, function, package)
```

## Test Patterns

### Basic Test
```go
func TestBasic(t *testing.T) {
    // Arrange
    original := globalVar
    defer func() { globalVar = original }()
    
    // Act
    SetValue(newValue)
    
    // Assert
    if GetValue() != newValue {
        t.Errorf("Expected %v, got %v", newValue, GetValue())
    }
}
```

### Table Test
```go
func TestMultiple(t *testing.T) {
    testCases := []struct {
        name     string
        input    interface{}
        expected interface{}
    }{
        {"case1", input1, expected1},
        {"case2", input2, expected2},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := FunctionUnderTest(tc.input)
            if result != tc.expected {
                t.Errorf("Expected %v, got %v", tc.expected, result)
            }
        })
    }
}
```

### Output Capture Test
```go
func TestOutput(t *testing.T) {
    var output strings.Builder
    logger := NewLoggerWithOutput(&output)
    
    logger.Info("test message")
    
    result := output.String()
    if !strings.Contains(result, "test message") {
        t.Error("Expected message not found in output")
    }
}
```

### File Test
```go
func TestFileLogging(t *testing.T) {
    tempDir, err := ioutil.TempDir("", "test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tempDir)
    
    logFile := filepath.Join(tempDir, "test.log")
    SetLogFile(logFile)
    SetFileLogging(true)
    
    logger := NewLogger()
    logger.Info("test")
    
    // Check file exists and contains expected content
}
```

## Common Mistakes to Avoid

❌ **Don't do this:**
```go
// Forgetting to restore state
func TestBad(t *testing.T) {
    SetLogLevel(DebugLevel) // This affects other tests!
    // ... test logic
}

// Not using thread-safe access
func SetConfig(value bool) {
    globalConfig = value // Race condition!
}

// Not validating input
func SetFrames(frames int) {
    skipFrames = frames // Could be negative!
}
```

✅ **Do this instead:**
```go
// Always restore state
func TestGood(t *testing.T) {
    original := currentLogLevel
    defer func() { currentLogLevel = original }()
    
    SetLogLevel(DebugLevel)
    // ... test logic
}

// Use mutexes for thread safety
func SetConfig(value bool) {
    configMutex.Lock()
    defer configMutex.Unlock()
    globalConfig = value
}

// Validate input
func SetFrames(frames int) {
    if frames < 0 || frames > 20 {
        return // or return error
    }
    skipFrames = frames
}
```

## Debugging Tips

### Debug Configuration
```go
// Add temporary debug logging
fmt.Printf("DEBUG: currentLogLevel=%v, enableFileLogging=%v\n", 
    currentLogLevel, enableFileLogging)
```

### Debug Tests
```go
// Use t.Logf for test debugging
t.Logf("Output: %s", output.String())
t.Logf("File contents: %s", string(fileContents))
```

### Debug File Operations
```go
// Check file operations
if _, err := os.Stat(logFile); err != nil {
    t.Logf("File stat error: %v", err)
}

// Check file permissions
info, _ := os.Stat(logFile)
t.Logf("File mode: %v", info.Mode())
```

## Performance Tips

### Avoid Allocations
```go
// Good: Use buffer pool
buf := getBuffer()
defer putBuffer(buf)

// Bad: Multiple string concatenations
result := prefix + " " + message + " " + suffix
```

### Early Returns
```go
// Good: Early return for filtered levels
func (l *Logger) Debug(message string, args ...interface{}) {
    if currentLogLevel < DebugLevel {
        return // Skip expensive operations
    }
    // ... rest of logging
}
```

### Efficient String Building
```go
// Good: Pre-calculate capacity
builder := strings.Builder{}
builder.Grow(estimatedSize)

// Good: Use WriteString for known strings
builder.WriteString(prefix)
```

## Release Checklist

Before submitting a contribution:

- [ ] All tests pass (`go test ./...`)
- [ ] Code follows existing patterns
- [ ] New features have tests
- [ ] Configuration changes are thread-safe
- [ ] Input validation is included
- [ ] Documentation is updated
- [ ] Examples are provided if needed
- [ ] Performance impact is considered
- [ ] Backward compatibility is maintained

## Getting Help

1. **Read the code**: Start with existing similar features
2. **Run examples**: See how features work in practice
3. **Write tests first**: Clarifies requirements
4. **Small changes**: Make incremental improvements
5. **Ask questions**: Better to ask than guess

## Useful Commands

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Test with race detection
go test -race ./...

# Profile tests
go test -cpuprofile=cpu.prof -memprofile=mem.prof

# Clean up
go mod tidy
go clean -testcache
```
