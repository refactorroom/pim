# Module 1: Getting Started with pim Contributions

## Overview

The pim package is a flexible, feature-rich logging library designed for Go applications. This module covers the basics of getting started with contributing to the project.

### What You'll Learn
- Package overview and key features
- Development environment setup
- Basic project structure
- First contribution workflow

## Package Overview

### Key Features
- Multiple log levels (Panic, Error, Warning, Info, Debug, Trace)
- Configurable caller information (file, line, function, package)
- Optional file logging with rotation
- Color-coded console output
- JSON and structured logging
- Hooks and middleware support
- Localization support
- Performance metrics

### Philosophy
The pim package follows these principles:
- **Performance First**: Minimal overhead for production use
- **Developer Friendly**: Easy to use and debug
- **Configurable**: Extensive customization options
- **Thread Safe**: Safe for concurrent use
- **Backward Compatible**: Changes don't break existing code

## Setting Up Development Environment

### Prerequisites
```bash
# Ensure you have Go 1.19+ installed
go version

# Verify Go installation
go env GOPATH
go env GOROOT
```

### Clone and Setup
```bash
# Clone the repository
git clone <repository-url>
cd pim

# Install dependencies
go mod tidy

# Verify everything works
go test ./...
```

### Development Tools
Install these helpful tools for Go development:

```bash
# Code formatting and imports
go install golang.org/x/tools/cmd/goimports@latest

# Linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Enhanced testing
go install github.com/rakyll/gotest@latest

# Documentation
go install golang.org/x/tools/cmd/godoc@latest
```

### IDE Setup (VS Code)
1. Install the Go extension
2. Configure settings:
```json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.coverOnSave": true
}
```

## Project Structure Overview

### Core Files
```
pim/
├── config.go           # Global configuration
├── logger.go           # Main logging interface
├── logger_core.go      # Core implementation
├── caller_info.go      # Stack trace analysis
├── metrics.go          # File logging & metrics
├── output.go           # Output formatting
├── writers.go          # File operations
├── theming.go          # Colors and themes
├── localization.go     # Multi-language support
├── hooks.go            # Middleware hooks
├── serializer.go       # JSON serialization
└── pim.go              # Package metadata
```

### Test Files
```
├── config_test.go      # Configuration tests
├── caller_info_test.go # Caller info tests
├── file_logging_test.go # File logging tests
├── writers_test.go     # File writer tests
├── metrics_test.go     # Performance tests
└── *_test.go           # Other test files
```

### Examples
```
example/
├── basic_logger/       # Simple usage
├── file_logging_demo/  # File logging
├── caller_test/        # Caller info demo
├── full_demo/          # Complete features
└── */                  # Other examples
```

## First Steps

### 1. Explore the Code
```bash
# Look at the main interface
cat logger.go | head -50

# Check configuration options
cat config.go | grep "func Set"

# See test patterns
cat config_test.go | head -30
```

### 2. Run Examples
```bash
# Try basic logging
cd example/basic_logger
go run basic_logger.go

# Test file logging
cd ../file_logging_demo
go run main.go

# Check caller information
cd ../caller_test
go run main.go
```

### 3. Run Tests
```bash
# Run all tests
cd ../../  # Back to root
go test ./...

# Run with coverage
go test -cover ./...

# Run specific tests
go test -run TestConfig ./...
```

## Understanding the Basics

### How Logging Works
```go
// 1. Create a logger
logger := pim.NewLogger()

// 2. Configure if needed
pim.SetLogLevel(pim.DebugLevel)
pim.SetFileLogging(true)

// 3. Log messages
logger.Info("Application started")
logger.Debug("Debug information")
logger.Error("Something went wrong")
```

### Configuration System
```go
// Global configuration affects all loggers
pim.SetLogLevel(pim.InfoLevel)
pim.SetShowFileLine(true)
pim.SetCallerSkipFrames(3)
pim.SetFileLogging(false)

// Check current settings
level := pim.GetLogLevel()
fileLogging := pim.GetFileLogging()
```

### Key Concepts

#### Log Levels
- **Panic**: Severe errors, application cannot continue
- **Error**: Error conditions that need attention
- **Warning**: Potentially harmful situations
- **Info**: General information about application flow
- **Debug**: Detailed information for debugging
- **Trace**: Very detailed information, including function calls

#### Caller Information
The package can show where log messages come from:
```
2025-07-13 10:30:45 [INFO] Application started [main.go:15]
```

#### File Logging
Logs can be written to files with automatic rotation:
- Disabled by default
- Configurable location and rotation
- Thread-safe file operations

## Your First Contribution

### 1. Pick a Small Task
Good first contributions:
- Fix typos in documentation
- Add a simple test case
- Improve code comments
- Add a configuration option

### 2. Make the Change
```bash
# Create a new branch
git checkout -b fix-typo-in-readme

# Make your changes
nano README.md

# Test your changes
go test ./...
```

### 3. Test Everything
```bash
# Run all tests
go test ./...

# Check formatting
go fmt ./...

# Run linter
golangci-lint run
```

### 4. Submit for Review
```bash
# Commit your changes
git add .
git commit -m "Fix typo in README documentation"

# Push to your fork
git push origin fix-typo-in-readme

# Create pull request on GitHub
```

## Common Beginner Tasks

### 1. Add a Test Case
```go
func TestNewFeature(t *testing.T) {
    // Save original state
    original := someGlobalVar
    defer func() { someGlobalVar = original }()
    
    // Test the feature
    SetSomeFeature(true)
    if !GetSomeFeature() {
        t.Error("Feature should be enabled")
    }
}
```

### 2. Improve Documentation
- Add examples to function comments
- Fix typos and grammar
- Clarify confusing explanations
- Add missing documentation

### 3. Add Configuration Options
```go
// Add a new boolean configuration
var showProcessID bool = false

func SetShowProcessID(show bool) {
    configMutex.Lock()
    defer configMutex.Unlock()
    showProcessID = show
}

func GetShowProcessID() bool {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return showProcessID
}
```

## Getting Help

### Resources
- **Code**: Read existing similar features
- **Tests**: Look at test files for patterns
- **Examples**: Run and modify example programs
- **Documentation**: This lesson series

### Best Practices
- Start small and simple
- Always write tests
- Follow existing patterns
- Ask questions early
- Test thoroughly

## Next Steps

After completing this module, you should:
- Have a working development environment
- Understand the basic project structure
- Know how to run tests and examples
- Be ready for your first contribution

### Continue Learning
- **Module 2**: Understanding the Architecture
- **Module 3**: Configuration System Deep Dive
- **Module 4**: Writing Effective Tests
- **Module 5**: Adding New Features

## Quick Reference

### Essential Commands
```bash
# Development
go test ./...           # Run all tests
go fmt ./...           # Format code
go mod tidy           # Clean dependencies

# Testing
go test -v ./...      # Verbose tests
go test -cover ./...  # With coverage
go test -run TestName # Specific test

# Quality
golangci-lint run     # Lint code
go vet ./...         # Static analysis
```

### Key Files to Know
- `config.go` - Configuration functions
- `logger.go` - Main logging interface
- `*_test.go` - Test files
- `example/` - Working examples

Congratulations! You're now ready to start contributing to the pim package. Move on to Module 2 to dive deeper into the architecture.
