# Configuration Guide

## Overview

CLG provides flexible configuration options that can be set globally or per-logger instance. This guide covers all available configuration options and best practices.

## Global Configuration

### File Logging

File logging is **disabled by default** to prevent unexpected file creation.

```go
import "github.com/yourusername/clg/pim"

// Enable file logging globally
pim.SetFileLogging(true)

// Check current setting
enabled := pim.GetFileLogging()
```

### Caller Information

#### Global Skip Frames
```go
// Set global skip frames (affects all new loggers)
pim.SetSkipFrames(2) // Skip 2 frames

// Get current setting
frames := pim.GetSkipFrames()
```

#### Goroutine ID Display
```go
// Enable/disable goroutine ID (disabled by default)
pim.SetGoroutineIDEnabled(true)

// Check current setting
enabled := pim.GetGoroutineIDEnabled()
```

## Per-Logger Configuration

### Creating Configured Loggers

```go
// Basic logger with defaults
logger := pim.NewLogger()

// Logger with custom configuration
logger := pim.NewLogger().
    EnableFileLogging().
    SetSkipFrames(3).
    EnableGoroutineID()
```

### File Output Configuration

```go
// Enable file logging for this logger
logger.EnableFileLogging()

// Set custom log file path
logger.SetLogFile("custom/path/app.log")

// Enable file rotation
logger.EnableRotation(10 * 1024 * 1024) // 10MB rotation
```

### Caller Information Configuration

```go
// Set skip frames for this logger
logger.SetSkipFrames(2)

// Enable caller information
logger.EnableCaller()

// Enable goroutine ID
logger.EnableGoroutineID()
```

## Configuration Options Reference

### File Logging Options

| Method | Description | Default |
|--------|-------------|---------|
| `EnableFileLogging()` | Enable file output | Disabled |
| `DisableFileLogging()` | Disable file output | - |
| `SetLogFile(path)` | Set custom file path | "app.log" |
| `EnableRotation(size)` | Enable file rotation | Disabled |
| `SetRotationCount(count)` | Max rotated files | 5 |

### Caller Information Options

| Method | Description | Default |
|--------|-------------|---------|
| `EnableCaller()` | Show caller info | Enabled |
| `DisableCaller()` | Hide caller info | - |
| `SetSkipFrames(n)` | Skip N stack frames | 2 |
| `EnableGoroutineID()` | Show goroutine ID | Disabled |
| `DisableGoroutineID()` | Hide goroutine ID | - |

### Output Formatting Options

| Method | Description | Default |
|--------|-------------|---------|
| `SetTimeFormat(format)` | Custom time format | RFC3339 |
| `EnableColors()` | Colored output | Auto |
| `DisableColors()` | No colored output | - |
| `SetTheme(theme)` | Custom color theme | Default |

## Best Practices

### Production Configuration

```go
// Production logger setup
logger := pim.NewLogger().
    EnableFileLogging().
    SetLogFile("/var/log/myapp/app.log").
    EnableRotation(100 * 1024 * 1024). // 100MB
    SetRotationCount(10).
    DisableColors().
    SetSkipFrames(1)
```

### Development Configuration

```go
// Development logger setup
logger := pim.NewLogger().
    EnableCaller().
    EnableGoroutineID().
    EnableColors().
    SetSkipFrames(2)
// File logging disabled by default
```

### Testing Configuration

```go
// Test logger (minimal output)
logger := pim.NewLogger().
    DisableCaller().
    DisableGoroutineID().
    DisableColors().
    SetLevel(pim.LevelError) // Only errors
```

## Environment-Based Configuration

### Using Environment Variables

```go
import "os"

func configureLogger() *pim.Logger {
    logger := pim.NewLogger()
    
    // Configure based on environment
    if os.Getenv("ENV") == "production" {
        logger.EnableFileLogging().
               SetLogFile("/var/log/app/app.log").
               DisableColors()
    }
    
    if os.Getenv("DEBUG") == "true" {
        logger.EnableGoroutineID().
               SetSkipFrames(1)
    }
    
    return logger
}
```

### Configuration File Support

```go
import (
    "encoding/json"
    "os"
)

type LogConfig struct {
    FileLogging  bool   `json:"file_logging"`
    LogFile      string `json:"log_file"`
    SkipFrames   int    `json:"skip_frames"`
    GoroutineID  bool   `json:"goroutine_id"`
    Colors       bool   `json:"colors"`
}

func loadConfig(filename string) (*LogConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config LogConfig
    err = json.Unmarshal(data, &config)
    return &config, err
}

func applyConfig(logger *pim.Logger, config *LogConfig) {
    if config.FileLogging {
        logger.EnableFileLogging()
        if config.LogFile != "" {
            logger.SetLogFile(config.LogFile)
        }
    }
    
    logger.SetSkipFrames(config.SkipFrames)
    
    if config.GoroutineID {
        logger.EnableGoroutineID()
    }
    
    if !config.Colors {
        logger.DisableColors()
    }
}
```

## Migration Guide

### From v1.x to v2.x

```go
// Old way (v1.x)
logger := pim.NewLogger()
// File logging was enabled by default

// New way (v2.x)
logger := pim.NewLogger().EnableFileLogging()
// File logging is now opt-in
```

### Updating Global Defaults

```go
// Set new defaults at application startup
func init() {
    pim.SetFileLogging(true)     // Enable globally if needed
    pim.SetSkipFrames(3)         // Adjust for your wrapper
    pim.SetGoroutineIDEnabled(false) // Explicit setting
}
```

## Troubleshooting

### Common Configuration Issues

#### File Permission Errors
```go
// Check directory permissions
logger.SetLogFile("./logs/app.log") // Use relative path
```

#### Incorrect Caller Information
```go
// Adjust skip frames if caller info is wrong
logger.SetSkipFrames(1) // Try different values
```

#### Missing Log Output
```go
// Check if file logging is enabled when expected
if pim.GetFileLogging() {
    logger.Info("File logging is enabled")
} else {
    logger.Info("File logging is disabled - enable with EnableFileLogging()")
}
```

## Related Documentation

- [API Reference](./api_reference.md)
- [Examples](./lession_contribute/EXAMPLES.md)
- [Performance Best Practices](./lession_contribute/06_performance_best_practices.md)
