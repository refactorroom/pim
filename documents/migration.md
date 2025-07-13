# Migration Guide

## Overview

This guide helps you migrate between versions of the CLG logging package and from other logging libraries.

## Version Migration

### From v1.x to v2.x

#### Breaking Changes

1. **File Logging Default Changed**
   - **v1.x**: File logging enabled by default
   - **v2.x**: File logging disabled by default

2. **Goroutine ID Default Changed**
   - **v1.x**: Goroutine ID shown by default
   - **v2.x**: Goroutine ID hidden by default

3. **Configuration API Changes**
   - New global configuration functions
   - Enhanced per-logger configuration

#### Migration Steps

**Step 1: Update Dependencies**
```bash
go get github.com/yourusername/clg@v2.0.0
go mod tidy
```

**Step 2: Update File Logging**
```go
// Old v1.x code
logger := pim.NewLogger()
// File logging was automatic

// New v2.x code
logger := pim.NewLogger().EnableFileLogging()
// Or set globally
pim.SetFileLogging(true)
logger := pim.NewLogger()
```

**Step 3: Update Goroutine ID**
```go
// Old v1.x code
logger := pim.NewLogger()
// Goroutine ID was shown by default

// New v2.x code - if you want goroutine ID
logger := pim.NewLogger().EnableGoroutineID()
// Or set globally
pim.SetGoroutineIDEnabled(true)
logger := pim.NewLogger()
```

**Step 4: Update Configuration**
```go
// Old v1.x code
logger := pim.NewLogger()
logger.SetSkipFrames(2)

// New v2.x code (same API, but with global options)
pim.SetSkipFrames(2) // Global default
logger := pim.NewLogger()
// Or per-logger
logger := pim.NewLogger().SetSkipFrames(2)
```

#### Compatibility Layer

For easier migration, you can create a compatibility wrapper:

```go
// compat.go - Wrapper for v1.x behavior
package main

import "github.com/yourusername/clg/pim"

func NewV1Logger() *pim.Logger {
    return pim.NewLogger().
        EnableFileLogging().      // v1.x default
        EnableGoroutineID()       // v1.x default
}

// Update your code gradually
func main() {
    logger := NewV1Logger() // Behaves like v1.x
    logger.Info("Migration message")
}
```

## Migrating from Other Logging Libraries

### From Standard `log` Package

**Old Code:**
```go
import "log"

func main() {
    log.Printf("Info message: %s", value)
    log.Println("Simple message")
}
```

**New Code:**
```go
import "github.com/yourusername/clg/pim"

func main() {
    logger := pim.NewLogger()
    logger.Info("Info message: %s", value)
    logger.Info("Simple message")
}
```

### From `logrus`

**Old Code:**
```go
import "github.com/sirupsen/logrus"

func main() {
    logrus.WithFields(logrus.Fields{
        "user_id": 123,
        "action": "login",
    }).Info("User logged in")
    
    logrus.SetLevel(logrus.DebugLevel)
    logrus.Debug("Debug message")
}
```

**New Code:**
```go
import "github.com/yourusername/clg/pim"

func main() {
    logger := pim.NewLogger()
    
    logger.WithFields(map[string]interface{}{
        "user_id": 123,
        "action": "login",
    }).Info("User logged in")
    
    logger.SetLevel(pim.LevelDebug)
    logger.Debug("Debug message")
}
```

### From `zap`

**Old Code:**
```go
import "go.uber.org/zap"

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    logger.Info("User logged in",
        zap.String("user_id", "123"),
        zap.String("action", "login"),
    )
}
```

**New Code:**
```go
import "github.com/yourusername/clg/pim"

func main() {
    logger := pim.NewLogger().EnableFileLogging()
    
    logger.WithFields(map[string]interface{}{
        "user_id": "123",
        "action": "login",
    }).Info("User logged in")
}
```

### From `slog` (Go 1.21+)

**Old Code:**
```go
import "log/slog"

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    logger.Info("User logged in",
        "user_id", 123,
        "action", "login",
    )
}
```

**New Code:**
```go
import "github.com/yourusername/clg/pim"

func main() {
    logger := pim.NewLogger()
    
    logger.WithFields(map[string]interface{}{
        "user_id": 123,
        "action": "login",
    }).Info("User logged in")
}
```

## Configuration Migration

### Environment-Based Migration

Create a migration helper that detects your current setup:

```go
package main

import (
    "os"
    "github.com/yourusername/clg/pim"
)

func createMigratedLogger() *pim.Logger {
    logger := pim.NewLogger()
    
    // Detect old v1.x usage patterns
    if os.Getenv("CLG_V1_COMPAT") == "true" {
        logger = logger.
            EnableFileLogging().
            EnableGoroutineID()
    }
    
    // Migrate from other loggers
    if os.Getenv("LOGRUS_LEVEL") != "" {
        // Convert logrus level to CLG level
        level := convertLogrusLevel(os.Getenv("LOGRUS_LEVEL"))
        logger.SetLevel(level)
    }
    
    return logger
}

func convertLogrusLevel(logrusLevel string) pim.Level {
    switch logrusLevel {
    case "debug":
        return pim.LevelDebug
    case "info":
        return pim.LevelInfo
    case "warn":
        return pim.LevelWarn
    case "error":
        return pim.LevelError
    default:
        return pim.LevelInfo
    }
}
```

### Gradual Migration Strategy

**Phase 1: Side-by-Side**
```go
import (
    "github.com/sirupsen/logrus" // Old
    "github.com/yourusername/clg/pim" // New
)

func main() {
    // Keep both during transition
    oldLogger := logrus.New()
    newLogger := pim.NewLogger()
    
    // New code uses CLG
    newLogger.Info("New feature implemented")
    
    // Old code still uses logrus (temporarily)
    oldLogger.Info("Legacy feature")
}
```

**Phase 2: Wrapper Functions**
```go
// Create wrapper functions for consistent interface
func LogInfo(msg string, fields map[string]interface{}) {
    if useNewLogger {
        globalLogger.WithFields(fields).Info(msg)
    } else {
        logrus.WithFields(logrus.Fields(fields)).Info(msg)
    }
}
```

**Phase 3: Complete Migration**
```go
// Remove old imports and use only CLG
import "github.com/yourusername/clg/pim"

var globalLogger = pim.NewLogger().EnableFileLogging()

func LogInfo(msg string, fields map[string]interface{}) {
    globalLogger.WithFields(fields).Info(msg)
}
```

## Testing Migration

### Unit Test Migration

**Old Tests:**
```go
func TestWithLogrus(t *testing.T) {
    var buf bytes.Buffer
    logrus.SetOutput(&buf)
    
    logrus.Info("test message")
    
    output := buf.String()
    assert.Contains(t, output, "test message")
}
```

**New Tests:**
```go
func TestWithCLG(t *testing.T) {
    var buf bytes.Buffer
    logger := pim.NewLogger().SetOutput(&buf)
    
    logger.Info("test message")
    
    output := buf.String()
    assert.Contains(t, output, "test message")
}
```

### Integration Test Migration

```go
func TestMigration(t *testing.T) {
    // Test that both old and new produce similar output
    var oldBuf, newBuf bytes.Buffer
    
    // Old logger setup
    oldLogger := logrus.New()
    oldLogger.SetOutput(&oldBuf)
    oldLogger.SetFormatter(&logrus.TextFormatter{
        DisableTimestamp: true,
    })
    
    // New logger setup
    newLogger := pim.NewLogger().
        SetOutput(&newBuf).
        DisableTimestamp()
    
    // Same message
    testMsg := "Migration test message"
    oldLogger.Info(testMsg)
    newLogger.Info(testMsg)
    
    // Compare essential content (ignore format differences)
    assert.Contains(t, oldBuf.String(), testMsg)
    assert.Contains(t, newBuf.String(), testMsg)
}
```

## Performance Migration

### Benchmarking Before/After

```go
func BenchmarkOldLogger(b *testing.B) {
    logger := logrus.New()
    logger.SetOutput(ioutil.Discard)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message")
    }
}

func BenchmarkNewLogger(b *testing.B) {
    logger := pim.NewLogger().SetOutput(ioutil.Discard)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message")
    }
}
```

### Memory Usage Comparison

```go
func compareMemoryUsage() {
    // Test old logger
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    oldLogger := logrus.New()
    for i := 0; i < 10000; i++ {
        oldLogger.Info("Memory test")
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    oldMemory := m2.Alloc - m1.Alloc
    
    // Test new logger
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    newLogger := pim.NewLogger()
    for i := 0; i < 10000; i++ {
        newLogger.Info("Memory test")
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    newMemory := m2.Alloc - m1.Alloc
    
    fmt.Printf("Old logger memory: %d bytes\n", oldMemory)
    fmt.Printf("New logger memory: %d bytes\n", newMemory)
    fmt.Printf("Improvement: %.2f%%\n", float64(oldMemory-newMemory)/float64(oldMemory)*100)
}
```

## Common Migration Issues

### Issue 1: Different Output Format

**Problem:** Log format changes after migration.

**Solution:**
```go
// Customize format to match old logger
logger := pim.NewLogger().
    SetTimeFormat("2006-01-02 15:04:05"). // Match old format
    DisableColors().                       // Match old behavior
    SetLevel(pim.LevelInfo)               // Match old level
```

### Issue 2: Missing Log Files

**Problem:** No log files created after migration.

**Solution:**
```go
// Explicitly enable file logging (disabled by default in v2.x)
logger := pim.NewLogger().EnableFileLogging()
```

### Issue 3: Performance Regression

**Problem:** Logging slower after migration.

**Solution:**
```go
// Optimize for performance
logger := pim.NewLogger().
    DisableCaller().          // Remove if not needed
    DisableGoroutineID().     // Remove if not needed
    EnableAsyncWriting(1000)  // Use async for file I/O
```

## Migration Checklist

- [ ] Update dependencies to CLG v2.x
- [ ] Enable file logging if needed
- [ ] Enable goroutine ID if needed
- [ ] Update configuration calls
- [ ] Test log output format
- [ ] Verify file creation behavior
- [ ] Run performance benchmarks
- [ ] Update unit tests
- [ ] Update integration tests
- [ ] Remove old logging dependencies

## Rollback Plan

If migration issues occur, you can quickly rollback:

```go
// Keep old logger as fallback
var useNewLogger = false // Feature flag

func getLogger() interface{} {
    if useNewLogger {
        return pim.NewLogger()
    } else {
        return logrus.New() // Fallback to old
    }
}
```

## Related Documentation

- [Configuration Guide](./CONFIGURATION.md)
- [Troubleshooting](./TROUBLESHOOTING.md)
- [Performance Guide](./PERFORMANCE.md)
- [API Reference](./api_reference.md)
