# Module 2: Understanding CLG Architecture

## Overview

This module provides a deep dive into the CLG package architecture, explaining how different components work together and interact.

### What You'll Learn
- Core components and their responsibilities
- Data flow through the logging system
- Configuration management
- Extension points and interfaces

## Core Components

### 1. Configuration System (`config.go`)

The configuration system manages global settings that affect all logger instances.

#### Global Variables
```go
var (
    currentLogLevel   LogLevel = InfoLevel  // Minimum level to log
    showFileLine      bool     = true       // Show file:line info
    showGoroutineID   bool     = false      // Show goroutine ID
    showFunctionName  bool     = true       // Show function name
    showPackageName   bool     = true       // Show package name
    showFullPath      bool     = false      // Show full file path
    stackDepth        int      = 3          // Stack trace depth
    callerSkipFrames  int      = 3          // Frames to skip
    enableFileLogging bool     = false      // File logging enabled
)
```

#### Thread Safety
```go
var configMutex sync.RWMutex

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

### 2. Logger Core (`logger.go`, `logger_core.go`)

The logger core handles the main logging interface and message processing.

#### Logger Structure
```go
type Logger struct {
    config LoggerConfig
    output io.Writer
    // Additional fields for specific configurations
}

type LoggerConfig struct {
    CallerSkipFrames int
    ShowFileLine     bool
    ShowGoroutineID  bool
    // Other configuration options
}
```

#### Message Processing Pipeline
```go
func (l *Logger) Info(message string, args ...interface{}) {
    // 1. Check log level
    if currentLogLevel < InfoLevel {
        return
    }
    
    // 2. Format message
    formattedMessage := fmt.Sprintf(message, args...)
    
    // 3. Get caller information
    callerInfo := l.getCallerInfo()
    
    // 4. Build final message
    finalMessage := l.buildMessage(InfoLevel, InfoPrefix, formattedMessage, callerInfo)
    
    // 5. Output message
    l.writeMessage(finalMessage)
}
```

### 3. Caller Information (`caller_info.go`)

Extracts and formats caller information from the call stack.

#### Stack Analysis
```go
func getFileInfo(skipFrames int) (string, int, string, string) {
    // Get caller information from runtime
    pc, file, line, ok := runtime.Caller(skipFrames)
    if !ok {
        return "unknown", 0, "unknown", "unknown"
    }
    
    // Extract function information
    fn := runtime.FuncForPC(pc)
    var funcName, packageName string
    if fn != nil {
        fullName := fn.Name()
        funcName, packageName = splitFunctionName(fullName)
    }
    
    // Process file path
    if !showFullPath {
        file = filepath.Base(file)
    }
    
    return file, line, funcName, packageName
}
```

#### Configurable Frame Skipping
```go
// Different contexts require different skip counts
func (l *Logger) getCallerInfo() string {
    skipFrames := callerSkipFrames
    if l.config.CallerSkipFrames != 0 {
        skipFrames = l.config.CallerSkipFrames
    }
    
    file, line, function, pkg := getFileInfo(skipFrames)
    return formatCallerInfo(file, line, function, pkg)
}
```

### 4. Output System (`output.go`)

Handles message formatting and output destinations.

#### Message Formatting
```go
func formatLogMessage(level LogLevel, prefix, message, caller string) string {
    var parts []string
    
    // Add timestamp
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    parts = append(parts, timestamp)
    
    // Add log level prefix
    parts = append(parts, prefix)
    
    // Add goroutine ID if enabled
    if showGoroutineID {
        gid := getGoroutineID()
        parts = append(parts, fmt.Sprintf("[G:%d]", gid))
    }
    
    // Add message
    parts = append(parts, message)
    
    // Add caller information
    if caller != "" {
        parts = append(parts, caller)
    }
    
    return strings.Join(parts, " ")
}
```

### 5. File Management (`metrics.go`, `writers.go`)

Handles file logging, rotation, and performance monitoring.

#### File Logging
```go
func writeToLogFile(data []byte) error {
    if !enableFileLogging {
        return nil
    }
    
    fileMutex.Lock()
    defer fileMutex.Unlock()
    
    // Open or create log file
    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return fmt.Errorf("failed to open log file: %w", err)
    }
    defer file.Close()
    
    // Write data
    _, err = file.Write(data)
    return err
}
```

#### Log Rotation
```go
func rotateLogFile() error {
    if !shouldRotate() {
        return nil
    }
    
    // Create backup filename with timestamp
    timestamp := time.Now().Format("2006-01-02_15-04-05")
    backupFile := fmt.Sprintf("%s.%s.gz", logFile, timestamp)
    
    // Compress and move current log
    return compressAndMove(logFile, backupFile)
}
```

## Data Flow

### 1. Message Creation
```
User Code
    ↓
logger.Info("message", args...)
    ↓
Level Check (currentLogLevel >= InfoLevel)
    ↓
Message Formatting (fmt.Sprintf)
```

### 2. Caller Information
```
Runtime Stack Analysis
    ↓
runtime.Caller(skipFrames)
    ↓
Extract Function/File/Line
    ↓
Format Caller String
```

### 3. Message Assembly
```
Timestamp Generation
    ↓
Level Prefix (e.g., "[INFO]")
    ↓
Optional Goroutine ID
    ↓
Formatted Message
    ↓
Caller Information
    ↓
Final Message String
```

### 4. Output Distribution
```
Final Message
    ↓
Console Output (io.Writer)
    ↓
File Output (if enabled)
    ↓
Hooks/Middleware (if configured)
```

## Configuration Architecture

### Global vs Instance Configuration

#### Global Configuration
- Affects all logger instances
- Thread-safe with RWMutex
- Changed via Set*/Get* functions
- Examples: log level, file logging, caller info

#### Instance Configuration
- Specific to individual logger instances
- Overrides global settings when specified
- Useful for specialized loggers

```go
// Global configuration
pim.SetLogLevel(pim.InfoLevel)
pim.SetFileLogging(true)

// Instance-specific configuration
logger := pim.NewLogger()
logger.SetCallerSkipFrames(5) // Override global setting
```

### Configuration Precedence
1. Instance-specific configuration (highest priority)
2. Global configuration
3. Default values (lowest priority)

## Extension Points

### 1. Custom Output Writers
```go
// Implement io.Writer interface
type CustomWriter struct {
    // Custom fields
}

func (w *CustomWriter) Write(p []byte) (n int, err error) {
    // Custom output logic
    return len(p), nil
}

// Use with logger
logger := pim.NewLoggerWithOutput(&CustomWriter{})
```

### 2. Hooks and Middleware
```go
// Hook interface
type LogHook interface {
    Fire(level LogLevel, message string, fields map[string]interface{}) error
}

// Example hook implementation
type DatabaseHook struct {
    db *sql.DB
}

func (h *DatabaseHook) Fire(level LogLevel, message string, fields map[string]interface{}) error {
    // Store log in database
    return h.insertLog(level, message, fields)
}
```

### 3. Custom Formatters
```go
// Formatter interface
type MessageFormatter interface {
    Format(level LogLevel, message string, timestamp time.Time, caller CallerInfo) string
}

// JSON formatter example
type JSONFormatter struct{}

func (f *JSONFormatter) Format(level LogLevel, message string, timestamp time.Time, caller CallerInfo) string {
    data := map[string]interface{}{
        "level":     level.String(),
        "message":   message,
        "timestamp": timestamp.UTC().Format(time.RFC3339),
        "caller":    caller,
    }
    
    jsonBytes, _ := json.Marshal(data)
    return string(jsonBytes)
}
```

## Component Interactions

### 1. Configuration Changes
```
SetLogLevel(DebugLevel)
    ↓
configMutex.Lock()
    ↓
currentLogLevel = DebugLevel
    ↓
configMutex.Unlock()
    ↓
All subsequent log calls use new level
```

### 2. File Logging Initialization
```
SetFileLogging(true)
    ↓
enableFileLogging = true
    ↓
Next log message triggers file creation
    ↓
File opened with appropriate permissions
    ↓
Message written to both console and file
```

### 3. Caller Information Resolution
```
logger.Info("message")
    ↓
getCallerInfo() called
    ↓
runtime.Caller(callerSkipFrames)
    ↓
Extract file, line, function info
    ↓
Format according to configuration
    ↓
Include in final message
```

## Performance Considerations

### 1. Level Checking
- Early return for filtered log levels
- Minimal overhead for disabled levels
- No expensive operations until level check passes

### 2. Caller Information
- Most expensive operation in logging
- Only performed when showFileLine is true
- Cached when possible

### 3. File Operations
- Buffered writes when possible
- Minimal file handle operations
- Background rotation when needed

### 4. Memory Management
- Buffer pooling for frequently allocated objects
- Efficient string building
- Minimal allocations in hot paths

## Architecture Benefits

### 1. Modularity
- Clear separation of concerns
- Easy to modify individual components
- Minimal interdependencies

### 2. Configurability
- Fine-grained control over behavior
- Global and instance-level configuration
- Runtime reconfiguration support

### 3. Performance
- Efficient hot paths
- Minimal overhead for disabled features
- Optimized string operations

### 4. Extensibility
- Hook system for custom processing
- Custom output destinations
- Pluggable formatters

## Common Architecture Patterns

### 1. Configuration Pattern
```go
// Always use mutex for thread safety
var someMutex sync.RWMutex
var someConfig SomeType

func SetSomeConfig(value SomeType) {
    someMutex.Lock()
    defer someMutex.Unlock()
    someConfig = value
}

func GetSomeConfig() SomeType {
    someMutex.RLock()
    defer someMutex.RUnlock()
    return someConfig
}
```

### 2. Factory Pattern
```go
// Different logger types for different needs
func NewLogger() *Logger {
    return &Logger{
        output: os.Stdout,
        config: getDefaultConfig(),
    }
}

func NewLoggerWithOutput(output io.Writer) *Logger {
    return &Logger{
        output: output,
        config: getDefaultConfig(),
    }
}

func NewFileLogger(filename string) (*Logger, error) {
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, err
    }
    
    return &Logger{
        output: file,
        config: getDefaultConfig(),
    }, nil
}
```

### 3. Builder Pattern
```go
// Fluent configuration interface
type LoggerBuilder struct {
    config LoggerConfig
    output io.Writer
}

func NewLoggerBuilder() *LoggerBuilder {
    return &LoggerBuilder{
        config: getDefaultConfig(),
        output: os.Stdout,
    }
}

func (b *LoggerBuilder) WithOutput(output io.Writer) *LoggerBuilder {
    b.output = output
    return b
}

func (b *LoggerBuilder) WithCallerSkipFrames(frames int) *LoggerBuilder {
    b.config.CallerSkipFrames = frames
    return b
}

func (b *LoggerBuilder) Build() *Logger {
    return &Logger{
        output: b.output,
        config: b.config,
    }
}

// Usage
logger := NewLoggerBuilder().
    WithOutput(file).
    WithCallerSkipFrames(5).
    Build()
```

## Next Steps

Now that you understand the architecture:
- **Module 3**: Deep dive into the configuration system
- **Module 4**: Learn advanced testing techniques
- **Module 5**: Implement new features using these patterns

Understanding this architecture will help you make informed decisions when contributing to the CLG package and ensure your changes fit well with the existing design.
