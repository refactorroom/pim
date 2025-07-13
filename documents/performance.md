# Performance Guide

## Overview

This guide covers performance considerations, benchmarks, and optimization techniques for the pim logging package.

## Performance Characteristics

### Memory Usage

pim is designed for minimal memory allocation:

- **Zero allocation** for disabled log levels
- **Minimal allocation** for enabled logs
- **Pool-based** buffer reuse for formatting
- **Lazy evaluation** of expensive operations

### Throughput Benchmarks

Typical performance on modern hardware:

| Operation | Throughput | Allocation |
|-----------|------------|------------|
| Disabled Level | ~1B ops/sec | 0 B/op |
| Console Only | ~10M ops/sec | 24 B/op |
| File Only | ~5M ops/sec | 48 B/op |
| Console + File | ~3M ops/sec | 72 B/op |
| With Caller Info | ~2M ops/sec | 120 B/op |

## Optimization Strategies

### 1. Level-Based Filtering

```go
// Efficient - checked before any work
logger.Debug("expensive operation: %v", expensiveFunction())

// Even better - guard expensive operations
if logger.IsDebugEnabled() {
    result := expensiveFunction()
    logger.Debug("expensive operation: %v", result)
}
```

### 2. Structured Logging

```go
// Less efficient - string formatting
logger.Info("User %s logged in from %s", userID, ipAddress)

// More efficient - structured fields
logger.WithFields(map[string]interface{}{
    "user_id": userID,
    "ip_address": ipAddress,
}).Info("User logged in")
```

### 3. Batch Operations

```go
// For high-volume logging, batch related operations
func logUserActions(actions []UserAction) {
    logger := pim.NewLogger().EnableFileLogging()
    
    for _, action := range actions {
        logger.WithFields(map[string]interface{}{
            "user_id": action.UserID,
            "action": action.Type,
            "timestamp": action.Timestamp,
        }).Info("User action")
    }
}
```

### 4. Buffer Pool Usage

```go
// pim automatically uses buffer pools, but you can help:
// Reuse logger instances rather than creating new ones
var globalLogger = pim.NewLogger()

func someFunction() {
    // Use global logger instead of creating new ones
    globalLogger.Info("Operation completed")
}
```

## Memory Optimization

### Reduce Allocations

```go
// Avoid string concatenation in log messages
// Bad
logger.Info("User " + userID + " performed " + action)

// Good
logger.Info("User %s performed %s", userID, action)

// Better
logger.WithFields(map[string]interface{}{
    "user_id": userID,
    "action": action,
}).Info("User performed action")
```

### Control Goroutine ID

```go
// Goroutine ID adds overhead - use only when needed
logger.DisableGoroutineID() // Default behavior

// Enable only for debugging
if debugMode {
    logger.EnableGoroutineID()
}
```

### Optimize Caller Information

```go
// Caller info has overhead - tune skip frames
logger.SetSkipFrames(1) // Minimal overhead

// Disable in production if not needed
if production {
    logger.DisableCaller()
}
```

## File I/O Optimization

### Buffered Writing

```go
// pim uses buffered writers by default
// But you can tune buffer sizes for your use case
logger := pim.NewLogger().
    EnableFileLogging().
    SetBufferSize(8192) // 8KB buffer
```

### Rotation Strategy

```go
// Optimize rotation for your I/O patterns
logger.EnableRotation(50 * 1024 * 1024) // 50MB files
logger.SetRotationCount(10) // Keep 10 files

// For high-volume logging
logger.EnableRotation(100 * 1024 * 1024) // Larger files
logger.SetRotationCount(20) // More files
```

### Asynchronous Logging

```go
// For maximum throughput, use background writers
logger := pim.NewLogger().
    EnableFileLogging().
    EnableAsyncWriting(1000) // 1000 message buffer
```

## Concurrency Optimization

### Per-Goroutine Loggers

```go
// For high-concurrency applications
func workerPool(jobs <-chan Job) {
    // Each worker gets its own logger
    logger := globalLogger.Clone()
    
    for job := range jobs {
        logger.WithFields(map[string]interface{}{
            "job_id": job.ID,
            "worker": runtime.NumGoroutine(),
        }).Info("Processing job")
        
        processJob(job)
    }
}
```

### Lock-Free Patterns

```go
// Use structured fields to avoid lock contention
type RequestLogger struct {
    base *pim.Logger
    requestID string
}

func (rl *RequestLogger) Info(msg string) {
    rl.base.WithFields(map[string]interface{}{
        "request_id": rl.requestID,
    }).Info(msg)
}
```

## Profiling and Monitoring

### Built-in Metrics

```go
// Enable performance metrics
logger := pim.NewLogger().EnableMetrics()

// Get performance stats
stats := logger.GetStats()
fmt.Printf("Messages logged: %d\n", stats.MessageCount)
fmt.Printf("Average latency: %v\n", stats.AverageLatency)
```

### Custom Profiling

```go
import (
    "runtime"
    "time"
)

func benchmarkLogging() {
    logger := pim.NewLogger()
    
    // Measure memory before
    var m1 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Perform logging operations
    start := time.Now()
    for i := 0; i < 100000; i++ {
        logger.Info("Test message %d", i)
    }
    duration := time.Since(start)
    
    // Measure memory after
    var m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    fmt.Printf("Duration: %v\n", duration)
    fmt.Printf("Memory used: %d bytes\n", m2.Alloc-m1.Alloc)
    fmt.Printf("Operations/sec: %.0f\n", float64(100000)/duration.Seconds())
}
```

## Production Optimizations

### Configuration for High-Volume

```go
func newProductionLogger() *pim.Logger {
    return pim.NewLogger().
        EnableFileLogging().
        SetLogFile("/var/log/app/app.log").
        EnableRotation(200 * 1024 * 1024). // 200MB rotation
        SetRotationCount(50). // Keep 50 files
        EnableAsyncWriting(2000). // Large async buffer
        DisableCaller(). // Reduce overhead
        DisableGoroutineID(). // Reduce overhead
        SetLevel(pim.LevelInfo) // Skip debug logs
}
```

### Load Testing

```go
func loadTest() {
    logger := newProductionLogger()
    
    // Simulate high load
    var wg sync.WaitGroup
    workers := runtime.NumCPU()
    
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            for j := 0; j < 10000; j++ {
                logger.WithFields(map[string]interface{}{
                    "worker": workerID,
                    "iteration": j,
                }).Info("Load test message")
            }
        }(i)
    }
    
    wg.Wait()
}
```

## Benchmarking

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem

# Run specific benchmarks
go test -bench=BenchmarkLogger -benchmem

# Profile memory
go test -bench=. -memprofile=mem.prof

# Profile CPU
go test -bench=. -cpuprofile=cpu.prof
```

### Benchmark Results

```go
func BenchmarkLoggerInfo(b *testing.B) {
    logger := pim.NewLogger()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message %d", i)
    }
}

func BenchmarkLoggerWithFields(b *testing.B) {
    logger := pim.NewLogger()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        logger.WithFields(map[string]interface{}{
            "iteration": i,
            "benchmark": true,
        }).Info("Benchmark message")
    }
}
```

## Performance Anti-Patterns

### Avoid These Patterns

```go
// DON'T: Create new loggers frequently
func handleRequest() {
    logger := pim.NewLogger() // Bad - creates new logger each time
    logger.Info("Handling request")
}

// DON'T: String concatenation in log messages
logger.Info("Request " + requestID + " from " + userAgent) // Bad

// DON'T: Expensive operations in log parameters
logger.Debug("Result: %v", expensiveCalculation()) // Bad - always runs

// DON'T: Ignore log levels
logger.Debug("Debug info") // Bad if debug is disabled but still allocates
```

### Do These Instead

```go
// DO: Reuse logger instances
var requestLogger = pim.NewLogger()

func handleRequest() {
    requestLogger.Info("Handling request")
}

// DO: Use formatting
logger.Info("Request %s from %s", requestID, userAgent)

// DO: Guard expensive operations
if logger.IsDebugEnabled() {
    logger.Debug("Result: %v", expensiveCalculation())
}

// DO: Use appropriate log levels
logger.Info("Important information")  // Always shown
logger.Debug("Debug information")     // Only in debug mode
```

## Related Documentation

- [Configuration Guide](./CONFIGURATION.md)
- [Best Practices](./lession_contribute/06_performance_best_practices.md)
- [API Reference](./api_reference.md)
