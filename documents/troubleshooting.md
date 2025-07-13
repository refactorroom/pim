# Troubleshooting Guide

## Common Issues and Solutions

### Installation Issues

#### "Package not found"
```
go: github.com/yourusername/clg@latest: reading github.com/yourusername/clg/go.mod: 404 Not Found
```

**Solutions:**
1. Check the correct import path
2. Verify Go module is initialized: `go mod init your-project`
3. Ensure you have internet connectivity
4. Try: `go clean -modcache` and retry

#### "Permission denied"
```
permission denied: cannot create directory
```

**Solutions:**
```bash
# Linux/macOS
sudo chown -R $USER:$USER $GOPATH
chmod -R 755 $GOPATH

# Windows (Run as Administrator)
icacls %GOPATH% /grant %USERNAME%:F /T
```

### Configuration Issues

#### File Logging Not Working

**Problem:** Logs not appearing in files despite configuration.

**Diagnosis:**
```go
// Check if file logging is enabled
if pim.GetFileLogging() {
    fmt.Println("File logging is globally enabled")
} else {
    fmt.Println("File logging is globally disabled")
}

// Check logger configuration
logger := pim.NewLogger()
if logger.IsFileLoggingEnabled() {
    fmt.Println("Logger has file logging enabled")
} else {
    fmt.Println("Logger has file logging disabled")
}
```

**Solutions:**
```go
// Solution 1: Enable globally
pim.SetFileLogging(true)

// Solution 2: Enable per logger
logger := pim.NewLogger().EnableFileLogging()

// Solution 3: Check file permissions
logger.SetLogFile("./logs/app.log") // Use writable directory
```

#### Incorrect Caller Information

**Problem:** Caller info shows wrong file/line numbers.

**Symptoms:**
```
2025-07-13 10:30:45 [INFO] [wrong_file.go:123] Message
```

**Solutions:**
```go
// Adjust skip frames
logger.SetSkipFrames(1) // Try different values: 0, 1, 2, 3

// For wrapper functions, increase skip frames
func MyLog(msg string) {
    logger.SetSkipFrames(3) // Skip wrapper + CLG internals
    logger.Info(msg)
}
```

#### Missing Goroutine ID

**Problem:** Goroutine ID not showing when expected.

**Solution:**
```go
// Enable goroutine ID (disabled by default)
logger.EnableGoroutineID()

// Or globally
pim.SetGoroutineIDEnabled(true)
```

### Runtime Issues

#### High Memory Usage

**Problem:** Application consuming too much memory.

**Diagnosis:**
```go
import (
    "runtime"
    "fmt"
)

func checkMemory() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("Memory usage: %d KB\n", m.Alloc/1024)
}
```

**Solutions:**
```go
// 1. Disable expensive features in production
logger := pim.NewLogger().
    DisableCaller().           // Saves stack trace overhead
    DisableGoroutineID().      // Saves goroutine lookup
    SetLevel(pim.LevelInfo)    // Skip debug logs

// 2. Use structured logging instead of string formatting
// Bad
logger.Info("User " + userID + " action " + action)

// Good
logger.WithFields(map[string]interface{}{
    "user_id": userID,
    "action": action,
}).Info("User action")

// 3. Guard expensive operations
if logger.IsDebugEnabled() {
    result := expensiveOperation()
    logger.Debug("Result: %v", result)
}
```

#### Slow Performance

**Problem:** Logging operations taking too long.

**Diagnosis:**
```go
import "time"

func benchmarkLogging() {
    logger := pim.NewLogger()
    
    start := time.Now()
    for i := 0; i < 1000; i++ {
        logger.Info("Test message %d", i)
    }
    duration := time.Since(start)
    
    fmt.Printf("1000 logs took: %v\n", duration)
    fmt.Printf("Per log: %v\n", duration/1000)
}
```

**Solutions:**
```go
// 1. Use async writing for file output
logger := pim.NewLogger().
    EnableFileLogging().
    EnableAsyncWriting(1000)

// 2. Increase buffer sizes
logger.SetBufferSize(8192) // 8KB buffer

// 3. Optimize file rotation
logger.EnableRotation(100 * 1024 * 1024) // Larger files
```

### File System Issues

#### "File not found" or "Access denied"

**Problem:** Cannot write to log files.

**Solutions:**
```go
// 1. Use absolute paths
logger.SetLogFile("/var/log/myapp/app.log")

// 2. Create directory first
os.MkdirAll("./logs", 0755)
logger.SetLogFile("./logs/app.log")

// 3. Check permissions
info, err := os.Stat("./logs")
if err != nil {
    fmt.Printf("Directory error: %v\n", err)
} else {
    fmt.Printf("Directory permissions: %v\n", info.Mode())
}
```

#### Disk Space Issues

**Problem:** Log files consuming too much disk space.

**Solutions:**
```go
// 1. Enable rotation
logger.EnableRotation(50 * 1024 * 1024) // 50MB files
logger.SetRotationCount(10) // Keep only 10 files

// 2. Monitor disk usage
func checkDiskSpace(logDir string) {
    files, _ := filepath.Glob(filepath.Join(logDir, "*.log*"))
    var totalSize int64
    
    for _, file := range files {
        info, _ := os.Stat(file)
        totalSize += info.Size()
    }
    
    fmt.Printf("Total log size: %d MB\n", totalSize/(1024*1024))
}

// 3. Clean old logs
func cleanOldLogs(logDir string, days int) {
    cutoff := time.Now().AddDate(0, 0, -days)
    files, _ := filepath.Glob(filepath.Join(logDir, "*.log*"))
    
    for _, file := range files {
        info, _ := os.Stat(file)
        if info.ModTime().Before(cutoff) {
            os.Remove(file)
        }
    }
}
```

### Concurrency Issues

#### Race Conditions

**Problem:** Panic or data races in concurrent code.

**Detection:**
```bash
# Run with race detector
go run -race main.go
go test -race ./...
```

**Solutions:**
```go
// 1. Use separate loggers per goroutine
func worker(id int, jobs <-chan Job) {
    logger := pim.NewLogger().Clone() // Each worker gets own logger
    
    for job := range jobs {
        logger.WithFields(map[string]interface{}{
            "worker": id,
            "job": job.ID,
        }).Info("Processing job")
    }
}

// 2. Use mutex for shared logger (if necessary)
type SafeLogger struct {
    mu     sync.Mutex
    logger *pim.Logger
}

func (sl *SafeLogger) Info(msg string, args ...interface{}) {
    sl.mu.Lock()
    defer sl.mu.Unlock()
    sl.logger.Info(msg, args...)
}
```

#### Deadlocks

**Problem:** Application hanging during logging.

**Solutions:**
```go
// 1. Use timeouts for file operations
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 2. Avoid circular dependencies
// Don't log from inside log handlers

// 3. Use separate goroutines for file I/O
logger := pim.NewLogger().
    EnableFileLogging().
    EnableAsyncWriting(1000) // Non-blocking writes
```

### Testing Issues

#### Tests Creating Unwanted Files

**Problem:** Test runs create log files in the project directory.

**Solutions:**
```go
// 1. Use temp directories in tests
func TestLogging(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "test_logs")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    logger := pim.NewLogger().
        EnableFileLogging().
        SetLogFile(filepath.Join(tempDir, "test.log"))
    
    logger.Info("Test message")
}

// 2. Disable file logging in tests
func TestLogic(t *testing.T) {
    logger := pim.NewLogger() // File logging disabled by default
    
    // Test your logic without file output
}

// 3. Use test-specific configuration
func newTestLogger() *pim.Logger {
    return pim.NewLogger().
        DisableColors(). // Consistent output
        SetLevel(pim.LevelDebug) // Show all logs in tests
}
```

#### Flaky Tests

**Problem:** Tests pass/fail inconsistently.

**Solutions:**
```go
// 1. Ensure deterministic output
logger := pim.NewLogger().
    DisableGoroutineID(). // Non-deterministic
    SetTimeFormat("2006-01-02 15:04:05") // Fixed format

// 2. Use buffer for testing output
var buf bytes.Buffer
logger := pim.NewLogger().SetOutput(&buf)

logger.Info("Test message")
output := buf.String()
assert.Contains(t, output, "Test message")

// 3. Wait for async operations
logger := pim.NewLogger().EnableFileLogging()
logger.Info("Message")
logger.Sync() // Wait for all writes to complete
```

## Debugging Tools

### Verbose Logging

```go
// Enable debug level for troubleshooting
logger := pim.NewLogger().SetLevel(pim.LevelDebug)

// Log configuration state
logger.Debug("File logging enabled: %v", logger.IsFileLoggingEnabled())
logger.Debug("Skip frames: %d", logger.GetSkipFrames())
logger.Debug("Goroutine ID enabled: %v", logger.IsGoroutineIDEnabled())
```

### Performance Profiling

```go
import (
    _ "net/http/pprof"
    "net/http"
)

// Enable profiling endpoint
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Then visit: http://localhost:6060/debug/pprof/
```

### Log Analysis

```bash
# Analyze log patterns
grep "ERROR" app.log | head -10

# Count log levels
grep -c "INFO" app.log
grep -c "ERROR" app.log
grep -c "DEBUG" app.log

# Find performance issues
grep "slow" app.log | tail -20
```

## Getting Help

### Before Reporting Issues

1. **Check this troubleshooting guide**
2. **Review the configuration** - most issues are configuration-related
3. **Test with minimal reproduction** - isolate the problem
4. **Check Go version compatibility** - ensure Go 1.19+

### Reporting Issues

Include this information:

```go
// Version information
fmt.Printf("Go version: %s\n", runtime.Version())
fmt.Printf("OS: %s\n", runtime.GOOS)
fmt.Printf("Arch: %s\n", runtime.GOARCH)

// CLG configuration
fmt.Printf("File logging: %v\n", pim.GetFileLogging())
fmt.Printf("Skip frames: %d\n", pim.GetSkipFrames())
fmt.Printf("Goroutine ID: %v\n", pim.GetGoroutineIDEnabled())
```

### Minimal Reproduction

```go
package main

import "github.com/yourusername/clg/pim"

func main() {
    logger := pim.NewLogger().EnableFileLogging()
    logger.Info("Test message")
    // Describe what you expected vs what happened
}
```

## Related Documentation

- [Configuration Guide](./CONFIGURATION.md)
- [Performance Guide](./PERFORMANCE.md)
- [API Reference](./api_reference.md)
- [Examples](./lession_contribute/EXAMPLES.md)
