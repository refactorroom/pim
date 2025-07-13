# CLG Package API Reference

## Overview

This document provides a comprehensive API reference for the CLG (Custom Logger for Go) package, including all public functions, types, and configuration options.

## Table of Contents
1. [Core Types](#core-types)
2. [Logger Interface](#logger-interface)
3. [Configuration Functions](#configuration-functions)
4. [Utility Functions](#utility-functions)
5. [Constants](#constants)
6. [Examples](#examples)

## Core Types

### LogLevel

```go
type LogLevel int
```

Represents the severity level of log messages.

**Constants:**
```go
const (
    PanicLevel   LogLevel = 0  // Highest priority, application cannot continue
    ErrorLevel   LogLevel = 1  // Error conditions
    WarningLevel LogLevel = 2  // Warning conditions
    InfoLevel    LogLevel = 3  // Informational messages (default)
    DebugLevel   LogLevel = 4  // Debug-level messages
    TraceLevel   LogLevel = 5  // Very detailed trace information
)
```

**Methods:**
```go
func (l LogLevel) String() string
```
Returns the string representation of the log level.

### Logger

```go
type Logger struct {
    // Contains filtered or unexported fields
}
```

The main logging struct that provides all logging functionality.

### LoggerConfig

```go
type LoggerConfig struct {
    CallerSkipFrames int  // Number of stack frames to skip for caller info
    ShowFileLine     bool // Whether to show file and line information
    ShowGoroutineID  bool // Whether to show goroutine ID
}
```

Configuration specific to individual logger instances.

## Logger Interface

### Creating Loggers

#### NewLogger
```go
func NewLogger() *Logger
```
Creates a new logger with default settings that outputs to stdout.

**Example:**
```go
logger := pim.NewLogger()
logger.Info("Hello, World!")
```

#### NewLoggerWithOutput
```go
func NewLoggerWithOutput(output io.Writer) *Logger
```
Creates a new logger that outputs to the specified writer.

**Parameters:**
- `output`: Any type that implements `io.Writer`

**Example:**
```go
var buffer strings.Builder
logger := pim.NewLoggerWithOutput(&buffer)
logger.Info("This goes to the buffer")
```

### Logging Methods

#### Info
```go
func (l *Logger) Info(message string, args ...interface{})
```
Logs an informational message.

**Parameters:**
- `message`: Format string for the log message
- `args`: Optional arguments for string formatting

**Example:**
```go
logger.Info("User %s logged in at %v", username, time.Now())
```

#### Error
```go
func (l *Logger) Error(message string, args ...interface{})
```
Logs an error message.

**Example:**
```go
logger.Error("Database connection failed: %v", err)
```

#### Warning
```go
func (l *Logger) Warning(message string, args ...interface{})
```
Logs a warning message.

**Example:**
```go
logger.Warning("Deprecated function called: %s", funcName)
```

#### Debug
```go
func (l *Logger) Debug(message string, args ...interface{})
```
Logs a debug message (only shown when log level is DebugLevel or higher).

**Example:**
```go
logger.Debug("Processing request with ID: %s", requestID)
```

#### Trace
```go
func (l *Logger) Trace(message string, args ...interface{})
```
Logs a trace message (only shown when log level is TraceLevel).

**Example:**
```go
logger.Trace("Entering function: %s", getFunctionName())
```

#### Panic
```go
func (l *Logger) Panic(message string, args ...interface{})
```
Logs a panic message and then calls `panic()`.

**Example:**
```go
logger.Panic("Critical system failure: %v", criticalError)
```

### Specialized Logging

#### Success
```go
func (l *Logger) Success(message string, args ...interface{})
```
Logs a success message with green color formatting.

#### Init
```go
func (l *Logger) Init(message string, args ...interface{})
```
Logs an initialization message, typically used during application startup.

#### Config
```go
func (l *Logger) Config(message string, args ...interface{})
```
Logs configuration-related messages.

#### Data
```go
func (l *Logger) Data(message string, args ...interface{})
```
Logs data-related operations.

#### Model
```go
func (l *Logger) Model(message string, args ...interface{})
```
Logs model or database-related operations.

#### Json
```go
func (l *Logger) Json(message string, args ...interface{})
```
Logs JSON-formatted messages.

## Configuration Functions

### Log Level Configuration

#### SetLogLevel
```go
func SetLogLevel(level LogLevel)
```
Sets the global minimum log level. Messages below this level will be filtered out.

**Parameters:**
- `level`: The minimum log level to display

**Example:**
```go
pim.SetLogLevel(pim.DebugLevel) // Show debug and trace messages
pim.SetLogLevel(pim.ErrorLevel) // Only show errors and panics
```

#### GetLogLevel
```go
func GetLogLevel() LogLevel
```
Returns the current global log level.

**Example:**
```go
currentLevel := pim.GetLogLevel()
fmt.Printf("Current log level: %s\n", currentLevel.String())
```

### Caller Information Configuration

#### SetShowFileLine
```go
func SetShowFileLine(show bool)
```
Enables or disables display of file name and line number in log messages.

**Parameters:**
- `show`: Whether to show file and line information

**Default:** `true`

**Example:**
```go
pim.SetShowFileLine(true)
logger.Info("This will show file:line")
// Output: 2025-07-13 10:30:45 [INFO] This will show file:line [main.go:15]

pim.SetShowFileLine(false)
logger.Info("This will not show file:line")
// Output: 2025-07-13 10:30:45 [INFO] This will not show file:line
```

#### GetShowFileLine
```go
func GetShowFileLine() bool
```
Returns whether file and line information is currently displayed.

#### SetShowGoroutineID
```go
func SetShowGoroutineID(show bool)
```
Enables or disables display of goroutine ID in log messages.

**Parameters:**
- `show`: Whether to show goroutine ID

**Default:** `false`

**Example:**
```go
pim.SetShowGoroutineID(true)
logger.Info("This will show goroutine ID")
// Output: 2025-07-13 10:30:45 [G:1] [INFO] This will show goroutine ID
```

#### GetShowGoroutineID
```go
func GetShowGoroutineID() bool
```
Returns whether goroutine ID is currently displayed.

#### SetShowFunctionName
```go
func SetShowFunctionName(show bool)
```
Enables or disables display of function name in caller information.

#### GetShowFunctionName
```go
func GetShowFunctionName() bool
```
Returns whether function name is displayed in caller information.

#### SetShowPackageName
```go
func SetShowPackageName(show bool)
```
Enables or disables display of package name in caller information.

#### GetShowPackageName
```go
func GetShowPackageName() bool
```
Returns whether package name is displayed in caller information.

#### SetShowFullPath
```go
func SetShowFullPath(show bool)
```
Enables or disables display of full file path (vs just filename) in caller information.

**Default:** `false` (shows only filename)

#### GetShowFullPath
```go
func GetShowFullPath() bool
```
Returns whether full file path is displayed.

#### SetCallerSkipFrames
```go
func SetCallerSkipFrames(frames int)
```
Sets the number of stack frames to skip when determining caller information.

**Parameters:**
- `frames`: Number of frames to skip (0-15)

**Default:** `3`

**Use Case:** When wrapping the logger in other functions, you may need to skip additional frames to show the actual caller.

**Example:**
```go
func myLoggingWrapper(message string) {
    pim.SetCallerSkipFrames(4) // Skip one extra frame
    logger := pim.NewLogger()
    logger.Info(message)
}
```

#### GetCallerSkipFrames
```go
func GetCallerSkipFrames() int
```
Returns the current number of stack frames being skipped.

### File Logging Configuration

#### SetFileLogging
```go
func SetFileLogging(enabled bool)
```
Enables or disables file logging.

**Parameters:**
- `enabled`: Whether to enable file logging

**Default:** `false`

**Example:**
```go
pim.SetFileLogging(true)
logger.Info("This will appear in both console and file")
```

#### GetFileLogging
```go
func GetFileLogging() bool
```
Returns whether file logging is currently enabled.

#### SetLogFile
```go
func SetLogFile(filename string)
```
Sets the path for the log file.

**Parameters:**
- `filename`: Path to the log file

**Default:** `"app.log"`

**Example:**
```go
pim.SetLogFile("logs/application.log")
pim.SetFileLogging(true)
```

#### GetLogFile
```go
func GetLogFile() string
```
Returns the current log file path.

### Stack Trace Configuration

#### SetStackDepth
```go
func SetStackDepth(depth int)
```
Sets the depth of stack trace information to collect.

**Parameters:**
- `depth`: Stack depth (1-10)

**Default:** `3`

#### GetStackDepth
```go
func GetStackDepth() int
```
Returns the current stack depth setting.

### Color Configuration

#### EnableColors
```go
func EnableColors()
```
Enables colored output in the terminal.

#### DisableColors
```go
func DisableColors()
```
Disables colored output (useful for file logging or non-terminal outputs).

## Utility Functions

### String Formatting

#### ColoredString
```go
func ColoredString(text string, color *color.Color) string
```
Returns a colored version of the input string.

**Parameters:**
- `text`: The text to colorize
- `color`: The color to apply

**Example:**
```go
redText := pim.ColoredString("Error occurred", pim.Red)
fmt.Println(redText)
```

#### ColoredFormat
```go
func ColoredFormat(format string, color *color.Color, args ...interface{}) string
```
Returns a formatted and colored string.

**Example:**
```go
coloredMsg := pim.ColoredFormat("User %s has %d points", pim.Green, username, points)
```

## Constants

### Color Constants

```go
const (
    ColorReset   = "\033[0m"
    ColorBlack   = "\033[30m"
    ColorRed     = "\033[31m"
    ColorGreen   = "\033[32m"
    ColorYellow  = "\033[33m"
    ColorBlue    = "\033[34m"
    ColorPurple  = "\033[35m"
    ColorCyan    = "\033[36m"
    ColorWhite   = "\033[37m"
    ColorGray    = "\033[90m"
)
```

### Predefined Colors

```go
var (
    Black     = color.New(color.FgBlack)
    Red       = color.New(color.FgRed)
    Green     = color.New(color.FgGreen)
    Yellow    = color.New(color.FgYellow)
    Blue      = color.New(color.FgBlue)
    Purple    = color.New(color.FgMagenta)
    Cyan      = color.New(color.FgCyan)
    White     = color.New(color.FgWhite)
    Gray      = color.New(color.FgHiBlack)
)
```

### Log Level Prefixes

```go
var (
    InfoPrefix    string // "[INFO]"
    ErrorPrefix   string // "[ERROR]"
    WarningPrefix string // "[WARNING]"
    DebugPrefix   string // "[DEBUG]"
    TracePrefix   string // "[TRACE]"
    PanicPrefix   string // "[PANIC]"
    SuccessPrefix string // "[SUCCESS]"
)
```

## Examples

### Basic Usage

```go
package main

import "github.com/your-org/clg"

func main() {
    // Create a logger
    logger := pim.NewLogger()
    
    // Basic logging
    logger.Info("Application started")
    logger.Error("Something went wrong")
    logger.Debug("Debug information")
}
```

### Configuration Example

```go
package main

import "github.com/your-org/clg"

func main() {
    // Configure logging
    pim.SetLogLevel(pim.DebugLevel)
    pim.SetShowFileLine(true)
    pim.SetShowGoroutineID(true)
    pim.SetFileLogging(true)
    pim.SetLogFile("logs/app.log")
    
    // Create logger
    logger := pim.NewLogger()
    
    // Log messages
    logger.Info("Configuration complete")
    logger.Debug("Debug mode enabled")
}
```

### Custom Output Example

```go
package main

import (
    "os"
    "github.com/your-org/clg"
)

func main() {
    // Log to file
    file, err := os.Create("custom.log")
    if err != nil {
        panic(err)
    }
    defer file.Close()
    
    logger := pim.NewLoggerWithOutput(file)
    logger.Info("This goes to the file")
}
```

### Wrapper Function Example

```go
package main

import "github.com/your-org/clg"

func logWithContext(context string, message string, args ...interface{}) {
    // Skip one extra frame to show the actual caller
    pim.SetCallerSkipFrames(4)
    
    logger := pim.NewLogger()
    fullMessage := fmt.Sprintf("[%s] %s", context, message)
    logger.Info(fullMessage, args...)
    
    // Reset to default
    pim.SetCallerSkipFrames(3)
}

func main() {
    logWithContext("AUTH", "User %s logged in", "john_doe")
    // Output will show main() as the caller, not logWithContext()
}
```

### Performance-Conscious Logging

```go
package main

import "github.com/your-org/clg"

func main() {
    logger := pim.NewLogger()
    
    // Check log level before expensive operations
    if pim.GetLogLevel() >= pim.DebugLevel {
        expensiveData := computeExpensiveDebugInfo()
        logger.Debug("Debug data: %v", expensiveData)
    }
    
    // Simple logging is always fast
    logger.Info("Fast log message")
}

func computeExpensiveDebugInfo() interface{} {
    // Expensive computation here
    return "expensive data"
}
```

## Thread Safety

All configuration functions are thread-safe and can be called from multiple goroutines simultaneously. Logger instances are also thread-safe for concurrent logging operations.

## Performance Notes

- **Filtered log levels**: Minimal overhead when messages are filtered out
- **Caller information**: Adds overhead due to stack walking
- **File logging**: Additional I/O overhead
- **String formatting**: Only performed if message will be logged

## Error Handling

The CLG package is designed to be robust and will not panic under normal circumstances. File logging errors are handled gracefully and will fall back to console output when possible.

---

*This API reference covers all public interfaces of the CLG package. For implementation details and contribution guidelines, see the [contribution modules](../lession_contribute/).*
