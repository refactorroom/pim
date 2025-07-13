# CLG Contribution Learning Path

Welcome to the comprehensive learning path for contributing to the CLG (Custom Logger for Go) package! This modular course will take you from beginner to expert contributor.

## Learning Modules

### 📚 **[Module 1: Getting Started](01_getting_started.md)**
- Package overview and key features
- Development environment setup
- Basic project structure
- Your first contribution workflow

### 🏗️ **[Module 2: Understanding Architecture](02_architecture.md)**
- Core components and their responsibilities
- Data flow through the logging system
- Configuration management patterns
- Extension points and interfaces

### ⚙️ **[Module 3: Configuration System](03_configuration_system.md)**
- Configuration architecture deep dive
- Thread safety in configuration
- Adding new configuration options
- Validation and error handling patterns

### 🧪 **[Module 4: Writing Effective Tests](04_writing_tests.md)**
- Testing strategies for logging systems
- Output capture and verification techniques
- File logging testing patterns
- Performance and benchmark testing
- Thread safety testing

### ✨ **[Module 5: Adding New Features](05_adding_features.md)**
- Feature design and planning process
- Implementation patterns and best practices
- Integration with existing architecture
- Comprehensive testing strategies
- Real-world examples

### ⚡ **[Module 6: Performance & Best Practices](06_performance_best_practices.md)**
- Performance optimization strategies
- Memory management techniques
- Advanced Go patterns for logging
- Profiling and benchmarking
- Production considerations

### 🤝 **[Module 7: Community Guidelines](07_community_guidelines.md)**
- Pull request process and best practices
- Code review guidelines
- Community interaction principles
- Project maintenance and governance
- Long-term sustainability

## Quick References

### 📋 **[Quick Reference Guide](QUICK_REFERENCE.md)**
- Commands cheat sheet
- Code patterns
- Common functions
- Testing patterns
- Debugging tips

### 💻 **[Practical Examples](EXAMPLES.md)**
- Step-by-step implementation examples
- Real-world scenarios
- Complete code samples
- Performance optimization examples

## Learning Path Recommendations

### 🆕 **New Contributors**
Start here if you're new to the project:
1. Module 1: Getting Started
2. Module 2: Understanding Architecture  
3. Module 4: Writing Effective Tests
4. Quick Reference Guide

### 🔧 **Feature Developers**
Focus on implementation if you want to add features:
1. Module 2: Understanding Architecture
2. Module 3: Configuration System
3. Module 5: Adding New Features
4. Practical Examples

### 🏃‍♂️ **Performance Focused**
Optimize for speed and efficiency:
1. Module 6: Performance & Best Practices
2. Module 4: Writing Effective Tests (benchmarking)
3. Practical Examples (optimization patterns)

### 👥 **Community Leaders**
Help build and maintain the community:
1. Module 7: Community Guidelines
2. All previous modules for technical understanding
3. Focus on mentoring and code review

## Prerequisites

### Required Knowledge
- Go programming language basics
- Git version control
- Basic understanding of logging concepts
- Command line proficiency

### Recommended Experience
- Go testing framework
- Concurrent programming concepts
- Open source contribution experience
- Performance profiling basics

## Module Structure

Each module follows a consistent structure:

### 📖 **Learning Objectives**
Clear goals for what you'll learn

### 🎯 **Core Concepts**
Key ideas and principles

### 💡 **Practical Examples**
Real code examples and implementations

### ✅ **Best Practices**
Do's and don'ts with explanations

### 🚀 **Next Steps**
What to learn next

## Getting Help

### 📚 Resources
- **Code**: Read existing implementations for patterns
- **Tests**: Look at test files for usage examples  
- **Examples**: Run and modify example programs
- **Documentation**: This complete lesson series

### 💬 Community Support
- Create issues for questions
- Join discussions in pull requests
- Ask for help in code reviews
- Participate in community discussions

### 🎯 Practice Projects
Try these to reinforce your learning:

1. **Beginner**: Add a simple configuration option
2. **Intermediate**: Implement a new output format
3. **Advanced**: Add performance monitoring features
4. **Expert**: Design a plugin system

## Success Metrics

Track your progress:

- [ ] Completed Module 1-3 (Foundation)
- [ ] Made first successful contribution
- [ ] Completed Module 4-5 (Implementation)  
- [ ] Added a significant feature
- [ ] Completed Module 6-7 (Advanced)
- [ ] Helped review others' contributions
- [ ] Mentored a new contributor

## Contributing to This Guide

This learning path is itself open source! Help improve it:

- Fix typos and unclear explanations
- Add more examples and use cases
- Suggest new modules or topics
- Share your learning experience
- Improve the organization and flow

## What's Next?

Ready to start? Choose your path:

### 🎯 **I'm New Here**
→ [Start with Module 1: Getting Started](01_getting_started.md)

### ⚡ **I Want Quick Reference**  
→ [Check the Quick Reference Guide](QUICK_REFERENCE.md)

### 💻 **I Learn by Examples**
→ [Browse Practical Examples](EXAMPLES.md)

### 🏗️ **I Want to Understand the Architecture**
→ [Jump to Module 2: Architecture](02_architecture.md)

## Overview

The CLG (Custom Logger for Go) package is a flexible, feature-rich logging library designed for Go applications. It provides structured logging, caller information, file logging, metrics, and extensive customization options.

### Key Features
- Multiple log levels (Panic, Error, Warning, Info, Debug, Trace)
- Configurable caller information (file, line, function, package)
- Optional file logging with rotation
- Color-coded console output
- JSON and structured logging
- Hooks and middleware support
- Localization support
- Performance metrics

## Package Architecture

### Core Files Structure
```
clg/
├── config.go           # Global configuration and settings
├── logger.go           # Main logging functionality
├── logger_core.go      # Core logger implementation
├── caller_info.go      # Caller information extraction
├── metrics.go          # Performance metrics and file logging
├── output.go           # Output formatting and destinations
├── writers.go          # File writers and rotation
├── serializer.go       # JSON and custom serialization
├── theming.go          # Color themes and formatting
├── localization.go     # Multi-language support
├── hooks.go            # Logging hooks and middleware
├── kvjson.go           # Key-value JSON utilities
├── json_util.go        # JSON utility functions
└── pim.go              # Package initialization and metadata
```

### Key Components

1. **Configuration System** (`config.go`)
   - Global settings management
   - Runtime configuration changes
   - Default values and validation

2. **Logger Core** (`logger.go`, `logger_core.go`)
   - Main logging interface
   - Message processing pipeline
   - Level-based filtering

3. **Caller Information** (`caller_info.go`)
   - Stack trace analysis
   - Function and file extraction
   - Configurable frame skipping

4. **File Management** (`metrics.go`, `writers.go`)
   - File logging capabilities
   - Log rotation and archiving
   - Performance monitoring

## Setting Up Development Environment

### Prerequisites
```bash
# Ensure you have Go 1.19+ installed
go version

# Clone the repository
git clone <repository-url>
cd clg

# Install dependencies
go mod tidy

# Run tests to verify setup
go test ./...
```

### Development Tools
```bash
# Install useful development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/rakyll/gotest@latest
```

### IDE Setup
- Install Go extension for VS Code
- Configure automatic imports and formatting
- Set up debugging configuration

## Understanding the Codebase

### Configuration System

The package uses a global configuration approach with thread-safe access:

```go
// Global configuration variables
var (
    currentLogLevel   LogLevel = InfoLevel
    showFileLine      bool     = true
    showGoroutineID   bool     = false  // Disabled by default
    callerSkipFrames  int      = 3      // Configurable frame skipping
    enableFileLogging bool     = false  // Disabled by default
)

// Configuration functions
func SetLogLevel(level LogLevel)
func SetShowFileLine(show bool)
func SetCallerSkipFrames(frames int)
func SetFileLogging(enabled bool)
```

### Logger Interface

The main logger interface provides methods for different log levels:

```go
type Logger interface {
    Info(message string, args ...interface{})
    Error(message string, args ...interface{})
    Warning(message string, args ...interface{})
    Debug(message string, args ...interface{})
    Trace(message string, args ...interface{})
    Panic(message string, args ...interface{})
}
```

### Caller Information System

The caller info system uses Go's runtime package to extract stack information:

```go
func getFileInfo(skipFrames int) (string, int, string, string) {
    // Extract caller information from stack
    pc, file, line, ok := runtime.Caller(skipFrames)
    if !ok {
        return "unknown", 0, "unknown", "unknown"
    }
    // Process and format information
    return processCallerInfo(pc, file, line)
}
```

## Writing Tests

### Test Organization

Tests are organized by functionality:
- `config_test.go` - Configuration testing
- `caller_info_test.go` - Caller information testing
- `file_logging_test.go` - File logging testing
- `writers_test.go` - File writers testing
- `metrics_test.go` - Metrics and performance testing

### Test Patterns

#### 1. Configuration Tests
```go
func TestSetLogLevel(t *testing.T) {
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    SetLogLevel(DebugLevel)
    if currentLogLevel != DebugLevel {
        t.Errorf("Expected log level %v, got %v", DebugLevel, currentLogLevel)
    }
}
```

#### 2. Feature Tests with Cleanup
```go
func TestFileLogging(t *testing.T) {
    // Setup
    originalEnabled := enableFileLogging
    defer func() { enableFileLogging = originalEnabled }()

    // Test
    SetFileLogging(true)
    // ... test file logging functionality

    // Cleanup happens automatically via defer
}
```

#### 3. Integration Tests
```go
func TestCallerInfoIntegration(t *testing.T) {
    // Test that caller info works correctly with different skip frames
    testCases := []struct {
        skipFrames int
        expectFunc string
    }{
        {3, "TestCallerInfoIntegration"},
        {4, "testing.tRunner"},
    }

    for _, tc := range testCases {
        SetCallerSkipFrames(tc.skipFrames)
        // ... verify caller info
    }
}
```

### Test Best Practices

1. **Always Clean Up**: Use `defer` to restore original state
2. **Test Defaults**: Verify default values are correct
3. **Test Edge Cases**: Include boundary conditions and error cases
4. **Use Table Tests**: For multiple similar test cases
5. **Mock External Dependencies**: Isolate units under test

## Adding New Features

### Step-by-Step Process

#### 1. Planning Phase
```go
// Example: Adding a new log level
// 1. Define the constant
const CustomLevel LogLevel = 6

// 2. Add to level checking functions
func isCustomLevel(level LogLevel) bool {
    return level == CustomLevel
}

// 3. Add formatting support
var CustomPrefix = ColoredString("[CUSTOM]", CustomColor)
```

#### 2. Implementation Phase
```go
// Add the main functionality
func (l *Logger) Custom(message string, args ...interface{}) {
    if currentLogLevel >= CustomLevel {
        l.logMessage(CustomLevel, CustomPrefix, message, args...)
    }
}
```

#### 3. Configuration Support
```go
// Add configuration if needed
var showCustomInfo bool = true

func SetShowCustomInfo(show bool) {
    showCustomInfo = show
}

func GetShowCustomInfo() bool {
    return showCustomInfo
}
```

#### 4. Testing
```go
func TestCustomLevel(t *testing.T) {
    // Test the new functionality
    originalLevel := currentLogLevel
    defer func() { currentLogLevel = originalLevel }()

    SetLogLevel(CustomLevel)
    
    // Create a test logger and verify behavior
    logger := NewLogger()
    logger.Custom("Test custom message")
    
    // Verify output, file logging, etc.
}
```

### Feature Integration Checklist

- [ ] Core functionality implemented
- [ ] Configuration options added
- [ ] Tests written and passing
- [ ] Documentation updated
- [ ] Examples created
- [ ] Backward compatibility maintained
- [ ] Performance impact assessed

## Best Practices

### Code Style

#### 1. Naming Conventions
```go
// Use clear, descriptive names
func SetCallerSkipFrames(frames int)  // Good
func SetCSF(f int)                    // Bad

// Use consistent prefixes for related functions
func SetFileLogging(enabled bool)
func GetFileLogging() bool
func SetShowFileLine(show bool)
func GetShowFileLine() bool
```

#### 2. Error Handling
```go
// Always handle errors appropriately
func writeToFile(data []byte) error {
    if !enableFileLogging {
        return nil // Silent skip, not an error
    }
    
    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return fmt.Errorf("failed to open log file: %w", err)
    }
    defer file.Close()
    
    _, err = file.Write(data)
    return err
}
```

#### 3. Thread Safety
```go
// Use mutexes for shared state
var (
    configMutex sync.RWMutex
    currentLogLevel LogLevel = InfoLevel
)

func SetLogLevel(level LogLevel) {
    configMutex.Lock()
    defer configMutex.Unlock()
    currentLogLevel = level
}

func GetLogLevel() LogLevel {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return currentLogLevel
}
```

### Performance Considerations

#### 1. Avoid Unnecessary Work
```go
// Check log level before expensive operations
func (l *Logger) Debug(message string, args ...interface{}) {
    if currentLogLevel < DebugLevel {
        return // Skip expensive formatting
    }
    
    // Only do expensive work if needed
    formattedMessage := fmt.Sprintf(message, args...)
    l.logMessage(DebugLevel, DebugPrefix, formattedMessage)
}
```

#### 2. Efficient String Building
```go
// Use strings.Builder for multiple concatenations
func buildLogMessage(parts ...string) string {
    var builder strings.Builder
    builder.Grow(estimateLength(parts)) // Pre-allocate
    
    for _, part := range parts {
        builder.WriteString(part)
    }
    
    return builder.String()
}
```

#### 3. Pool Expensive Objects
```go
// Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func getBuffer() []byte {
    return bufferPool.Get().([]byte)
}

func putBuffer(buf []byte) {
    bufferPool.Put(buf[:0])
}
```

## Common Contribution Patterns

### 1. Adding Configuration Options

```go
// Step 1: Add global variable
var newFeatureEnabled bool = false

// Step 2: Add setter/getter
func SetNewFeature(enabled bool) {
    configMutex.Lock()
    defer configMutex.Unlock()
    newFeatureEnabled = enabled
}

func GetNewFeature() bool {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return newFeatureEnabled
}

// Step 3: Add tests
func TestNewFeature(t *testing.T) {
    original := newFeatureEnabled
    defer func() { newFeatureEnabled = original }()
    
    SetNewFeature(true)
    if !GetNewFeature() {
        t.Error("Expected new feature to be enabled")
    }
}
```

### 2. Adding Output Formats

```go
// Step 1: Define format interface
type OutputFormatter interface {
    Format(level LogLevel, message string, caller CallerInfo) string
}

// Step 2: Implement formatter
type JSONFormatter struct{}

func (j *JSONFormatter) Format(level LogLevel, message string, caller CallerInfo) string {
    data := map[string]interface{}{
        "level":     level.String(),
        "message":   message,
        "timestamp": time.Now().UTC(),
        "caller":    caller,
    }
    
    result, _ := json.Marshal(data)
    return string(result)
}

// Step 3: Integrate with logger
func (l *Logger) SetFormatter(formatter OutputFormatter) {
    l.formatter = formatter
}
```

### 3. Adding Validation

```go
// Always validate input parameters
func SetStackDepth(depth int) {
    if depth < 1 || depth > 10 {
        // Invalid depth, ignore
        return
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    stackDepth = depth
}

func SetCallerSkipFrames(frames int) {
    if frames < 0 || frames > 15 {
        // Invalid frame count, ignore
        return
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    callerSkipFrames = frames
}
```

## Testing Your Changes

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test files
go test -run TestConfig config_test.go

# Run tests with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage in terminal
go tool cover -func=coverage.out
```

### Integration Testing

```bash
# Test with examples
cd example/basic_logger
go run basic_logger.go

cd ../file_logging_demo
go run main.go

cd ../caller_test
go run main.go
```

### Manual Testing Checklist

- [ ] All log levels work correctly
- [ ] File logging creates and writes to files
- [ ] Caller information is accurate
- [ ] Configuration changes take effect
- [ ] Colors work in terminal
- [ ] Performance is acceptable
- [ ] Memory usage is reasonable

## Documentation Guidelines

### Code Comments

```go
// SetCallerSkipFrames configures how many stack frames to skip when
// determining caller information. This is useful when wrapping the logger
// in other functions. Valid range is 0-15.
//
// Example:
//   SetCallerSkipFrames(3) // Skip 3 frames
//   logger.Info("message") // Will show caller 3 frames up
func SetCallerSkipFrames(frames int) {
    if frames < 0 || frames > 15 {
        return
    }
    
    configMutex.Lock()
    defer configMutex.Unlock()
    callerSkipFrames = frames
}
```

### README Updates

When adding features, update relevant sections:

```markdown
## New Feature

Description of the feature and its benefits.

### Usage

```go
// Example usage
logger := NewLogger()
logger.SetNewFeature(true)
logger.Info("Using new feature")
```

### Configuration

- `SetNewFeature(enabled bool)` - Enable/disable the feature
- `GetNewFeature() bool` - Check if feature is enabled
```

### Example Programs

Create examples demonstrating new features:

```go
// example/new_feature_demo/main.go
package main

import "github.com/your-org/clg"

func main() {
    // Demonstrate the new feature
    pim.SetNewFeature(true)
    
    logger := pim.NewLogger()
    logger.Info("Demonstrating new feature")
    
    // Show configuration options
    pim.SetNewFeature(false)
    logger.Info("Feature now disabled")
}
```

## Conclusion

Contributing to the CLG package involves understanding its architecture, following established patterns, writing comprehensive tests, and maintaining backward compatibility. The key principles are:

1. **Safety First**: Always clean up resources and handle errors
2. **Test Everything**: Write tests before and after changes
3. **Follow Patterns**: Use established patterns for consistency
4. **Document Changes**: Update docs and examples
5. **Performance Matters**: Consider the impact of changes

By following this guide, you'll be able to contribute effectively to the CLG logging package and help improve its functionality for all users.

### Quick Start Checklist for Contributors

- [ ] Fork and clone the repository
- [ ] Set up development environment
- [ ] Read through existing code
- [ ] Write tests for your changes
- [ ] Implement your feature
- [ ] Run all tests and ensure they pass
- [ ] Update documentation
- [ ] Create examples if needed
- [ ] Submit pull request with clear description

Happy contributing! 🚀
