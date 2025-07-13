# Module 6: Performance and Best Practices

## Overview

This module focuses on performance optimization, best practices, and advanced techniques for maintaining high-quality code in the CLG package.

### What You'll Learn
- Performance optimization strategies
- Memory management techniques
- Code quality best practices
- Advanced Go patterns for logging
- Profiling and benchmarking
- Production deployment considerations

## Performance Fundamentals

### Why Performance Matters in Logging

Logging should have minimal impact on application performance:
- **Hot Path Optimization**: Logging calls are frequent and should be fast
- **Memory Efficiency**: Avoid unnecessary allocations
- **CPU Efficiency**: Minimize computational overhead
- **I/O Efficiency**: Optimize file and network operations
- **Scalability**: Performance should remain good under load

### Performance Hierarchy

1. **Filtered Log Levels** (fastest): Early return without processing
2. **Simple Messages**: Basic string formatting
3. **Caller Information**: Stack walking overhead
4. **File Operations**: Disk I/O overhead
5. **Complex Formatting**: JSON serialization, complex layouts

## Optimization Strategies

### 1. Early Returns and Level Checking

Always check log level before expensive operations:

```go
// ✅ GOOD: Early return pattern
func (l *Logger) Debug(message string, args ...interface{}) {
    if currentLogLevel < DebugLevel {
        return // Fastest possible path for filtered messages
    }
    
    // Only do expensive work if message will be logged
    formattedMessage := fmt.Sprintf(message, args...)
    callerInfo := l.getCallerInfo()
    l.writeMessage(DebugLevel, DebugPrefix, formattedMessage, callerInfo)
}

// ❌ BAD: Work done before level check
func (l *Logger) Debug(message string, args ...interface{}) {
    formattedMessage := fmt.Sprintf(message, args...)  // Wasted work
    callerInfo := l.getCallerInfo()                     // Expensive!
    
    if currentLogLevel < DebugLevel {
        return
    }
    
    l.writeMessage(DebugLevel, DebugPrefix, formattedMessage, callerInfo)
}
```

### 2. String Building Optimization

Use efficient string building techniques:

```go
// ✅ GOOD: Pre-allocated builder with estimated capacity
func buildLogMessage(parts []string) string {
    totalLen := 0
    for _, part := range parts {
        totalLen += len(part)
    }
    totalLen += len(parts) - 1 // For separators
    
    var builder strings.Builder
    builder.Grow(totalLen) // Pre-allocate
    
    for i, part := range parts {
        if i > 0 {
            builder.WriteByte(' ')
        }
        builder.WriteString(part)
    }
    
    return builder.String()
}

// ❌ BAD: Multiple string concatenations
func buildLogMessage(parts []string) string {
    result := ""
    for i, part := range parts {
        if i > 0 {
            result += " "  // Creates new string each time
        }
        result += part     // Creates new string each time
    }
    return result
}

// ❌ BAD: Using fmt.Sprintf for simple concatenation
func buildLogMessage(parts []string) string {
    return fmt.Sprintf("%s %s %s %s", parts[0], parts[1], parts[2], parts[3])
}
```

### 3. Buffer Pool Pattern

Implement buffer pooling for frequently allocated objects:

```go
// Buffer pool for reducing allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

func getBuffer() *bytes.Buffer {
    return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer) {
    buf.Reset()
    bufferPool.Put(buf)
}

// Usage in logging
func (l *Logger) formatMessage(level LogLevel, message string) string {
    buf := getBuffer()
    defer putBuffer(buf)
    
    // Build message in buffer
    buf.WriteString(time.Now().Format("2006-01-02 15:04:05"))
    buf.WriteByte(' ')
    buf.WriteString(getLevelPrefix(level))
    buf.WriteByte(' ')
    buf.WriteString(message)
    
    return buf.String()
}
```

### 4. Atomic Operations for Simple State

Use atomic operations for frequently accessed simple state:

```go
import "sync/atomic"

// For simple boolean flags
var enableDebugLogging int64 // 0 = false, 1 = true

func SetDebugLogging(enabled bool) {
    var value int64
    if enabled {
        value = 1
    }
    atomic.StoreInt64(&enableDebugLogging, value)
}

func GetDebugLogging() bool {
    return atomic.LoadInt64(&enableDebugLogging) == 1
}

// For simple numeric values
var logCounter uint64

func incrementLogCounter() {
    atomic.AddUint64(&logCounter, 1)
}

func getLogCount() uint64 {
    return atomic.LoadUint64(&logCounter)
}
```

### 5. Conditional Compilation for Debug Features

Use build tags for debug features:

```go
// debug.go - only included in debug builds
//go:build debug

package pim

var debugMode = true

func debugLog(message string) {
    fmt.Printf("DEBUG: %s\n", message)
}

// release.go - only included in release builds  
//go:build !debug

package pim

var debugMode = false

func debugLog(message string) {
    // No-op in release builds
}
```

## Memory Management

### 1. Avoiding Memory Leaks

Be careful with goroutines and resources:

```go
// ✅ GOOD: Proper resource cleanup
func (l *Logger) startFileWriter() {
    l.stopChan = make(chan struct{})
    l.wg.Add(1)
    
    go func() {
        defer l.wg.Done()
        defer func() {
            if l.file != nil {
                l.file.Close()
            }
        }()
        
        for {
            select {
            case message := <-l.messageChan:
                l.writeToFile(message)
            case <-l.stopChan:
                return
            }
        }
    }()
}

func (l *Logger) stop() {
    close(l.stopChan)
    l.wg.Wait()
}

// ❌ BAD: Goroutine leak
func (l *Logger) startFileWriter() {
    go func() {
        for message := range l.messageChan {
            l.writeToFile(message)
        }
        // No way to stop this goroutine!
    }()
}
```

### 2. Slice and Map Management

Manage growing data structures carefully:

```go
// ✅ GOOD: Bounded slice with cleanup
type RecentLogs struct {
    logs []string
    mu   sync.Mutex
    max  int
}

func (r *RecentLogs) Add(log string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.logs = append(r.logs, log)
    
    // Keep only recent logs
    if len(r.logs) > r.max {
        copy(r.logs, r.logs[len(r.logs)-r.max:])
        r.logs = r.logs[:r.max]
    }
}

// ❌ BAD: Unbounded growth
type RecentLogs struct {
    logs []string
    mu   sync.Mutex
}

func (r *RecentLogs) Add(log string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.logs = append(r.logs, log) // Grows indefinitely!
}
```

### 3. Interface Optimization

Use specific interfaces instead of empty interface when possible:

```go
// ✅ GOOD: Specific interface
type StringWriter interface {
    WriteString(string) (int, error)
}

func writeLog(w StringWriter, message string) {
    w.WriteString(message) // No allocations
}

// ❌ LESS OPTIMAL: Empty interface
func writeLog(w interface{}, message string) {
    if writer, ok := w.(StringWriter); ok {
        writer.WriteString(message) // Requires type assertion
    }
}
```

## Advanced Patterns

### 1. Fast Path / Slow Path Pattern

Optimize common cases:

```go
func (l *Logger) logMessage(level LogLevel, prefix, message string, args ...interface{}) {
    if currentLogLevel < level {
        return
    }
    
    // Fast path: no formatting needed
    if len(args) == 0 && !showFileLine && !enableFileLogging {
        l.fastLog(prefix, message)
        return
    }
    
    // Slow path: full processing
    l.slowLog(level, prefix, message, args...)
}

func (l *Logger) fastLog(prefix, message string) {
    // Optimized for common case
    timestamp := time.Now().Format("15:04:05")
    fmt.Printf("%s %s %s\n", timestamp, prefix, message)
}

func (l *Logger) slowLog(level LogLevel, prefix, message string, args ...interface{}) {
    // Full feature processing
    var formattedMessage string
    if len(args) > 0 {
        formattedMessage = fmt.Sprintf(message, args...)
    } else {
        formattedMessage = message
    }
    
    // Add caller info if needed
    var callerInfo string
    if showFileLine {
        callerInfo = l.getCallerInfo()
    }
    
    // Build final message
    finalMessage := l.buildCompleteMessage(level, prefix, formattedMessage, callerInfo)
    
    // Output to console
    l.output.Write([]byte(finalMessage + "\n"))
    
    // Output to file if enabled
    if enableFileLogging {
        l.writeToFile(finalMessage)
    }
}
```

### 2. Lazy Evaluation Pattern

Defer expensive operations until needed:

```go
// Caller info provider with lazy evaluation
type CallerInfo struct {
    skipFrames int
    file       string
    line       int
    function   string
    computed   bool
}

func newCallerInfo(skipFrames int) *CallerInfo {
    return &CallerInfo{
        skipFrames: skipFrames,
        computed:   false,
    }
}

func (c *CallerInfo) File() string {
    c.compute()
    return c.file
}

func (c *CallerInfo) Line() int {
    c.compute()
    return c.line
}

func (c *CallerInfo) Function() string {
    c.compute()
    return c.function
}

func (c *CallerInfo) compute() {
    if c.computed {
        return
    }
    
    pc, file, line, ok := runtime.Caller(c.skipFrames)
    if !ok {
        c.file = "unknown"
        c.line = 0
        c.function = "unknown"
    } else {
        c.file = filepath.Base(file)
        c.line = line
        if fn := runtime.FuncForPC(pc); fn != nil {
            c.function = fn.Name()
        } else {
            c.function = "unknown"
        }
    }
    
    c.computed = true
}

// Usage
func (l *Logger) logWithCaller(level LogLevel, message string) {
    callerInfo := newCallerInfo(3)
    
    // Caller info is only computed if actually used
    if showFileLine {
        fmt.Printf("%s [%s:%d] %s\n", message, callerInfo.File(), callerInfo.Line(), message)
    } else {
        fmt.Printf("%s %s\n", getLevelPrefix(level), message)
    }
}
```

### 3. Batch Processing Pattern

Batch operations for better performance:

```go
type BatchLogger struct {
    messages chan string
    batch    []string
    flush    chan struct{}
    maxSize  int
    interval time.Duration
}

func NewBatchLogger(maxSize int, interval time.Duration) *BatchLogger {
    bl := &BatchLogger{
        messages: make(chan string, 1000),
        batch:    make([]string, 0, maxSize),
        flush:    make(chan struct{}),
        maxSize:  maxSize,
        interval: interval,
    }
    
    go bl.processBatch()
    return bl
}

func (bl *BatchLogger) Log(message string) {
    select {
    case bl.messages <- message:
    default:
        // Channel full, drop message or handle differently
    }
}

func (bl *BatchLogger) processBatch() {
    ticker := time.NewTicker(bl.interval)
    defer ticker.Stop()
    
    for {
        select {
        case message := <-bl.messages:
            bl.batch = append(bl.batch, message)
            
            if len(bl.batch) >= bl.maxSize {
                bl.flushBatch()
            }
            
        case <-ticker.C:
            if len(bl.batch) > 0 {
                bl.flushBatch()
            }
            
        case <-bl.flush:
            bl.flushBatch()
            return
        }
    }
}

func (bl *BatchLogger) flushBatch() {
    if len(bl.batch) == 0 {
        return
    }
    
    // Write all messages at once
    combined := strings.Join(bl.batch, "\n") + "\n"
    os.Stdout.Write([]byte(combined))
    
    // Reset batch
    bl.batch = bl.batch[:0]
}
```

## Profiling and Benchmarking

### 1. CPU Profiling

Profile CPU usage to find bottlenecks:

```go
// Example benchmark with profiling
func BenchmarkLogging(b *testing.B) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message %d", i)
    }
}

// Run with profiling:
// go test -bench=. -cpuprofile=cpu.prof
// go tool pprof cpu.prof
```

### 2. Memory Profiling

Track memory allocations:

```go
func BenchmarkMemoryUsage(b *testing.B) {
    logger := NewLoggerWithOutput(ioutil.Discard)
    
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        logger.Info("Memory benchmark %d", i)
    }
    
    b.StopTimer()
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    allocsPerOp := (m2.TotalAlloc - m1.TotalAlloc) / uint64(b.N)
    b.ReportMetric(float64(allocsPerOp), "allocs/op")
}
```

### 3. Benchmark Comparison

Compare different implementations:

```go
func BenchmarkStringConcat(b *testing.B) {
    parts := []string{"2023-01-15", "10:30:45", "[INFO]", "Test message"}
    
    b.Run("PlusOperator", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = parts[0] + " " + parts[1] + " " + parts[2] + " " + parts[3]
        }
    })
    
    b.Run("Sprintf", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = fmt.Sprintf("%s %s %s %s", parts[0], parts[1], parts[2], parts[3])
        }
    })
    
    b.Run("StringsJoin", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = strings.Join(parts, " ")
        }
    })
    
    b.Run("StringsBuilder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            builder.Grow(50) // Estimated capacity
            for j, part := range parts {
                if j > 0 {
                    builder.WriteByte(' ')
                }
                builder.WriteString(part)
            }
            _ = builder.String()
        }
    })
}
```

## Code Quality Best Practices

### 1. Error Handling

Handle errors appropriately:

```go
// ✅ GOOD: Proper error handling with context
func (l *Logger) writeToFile(message string) error {
    if l.file == nil {
        return fmt.Errorf("log file not initialized")
    }
    
    n, err := l.file.WriteString(message + "\n")
    if err != nil {
        return fmt.Errorf("failed to write to log file: %w", err)
    }
    
    if n != len(message)+1 {
        return fmt.Errorf("incomplete write: wrote %d bytes, expected %d", n, len(message)+1)
    }
    
    return nil
}

// Handle the error appropriately
func (l *Logger) Info(message string, args ...interface{}) {
    // ... message processing ...
    
    if enableFileLogging {
        if err := l.writeToFile(finalMessage); err != nil {
            // Log to stderr or handle appropriately
            fmt.Fprintf(os.Stderr, "Failed to write log to file: %v\n", err)
        }
    }
}
```

### 2. Resource Management

Manage resources properly:

```go
// ✅ GOOD: Proper resource management
type FileLogger struct {
    file *os.File
    mu   sync.Mutex
}

func NewFileLogger(filename string) (*FileLogger, error) {
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, fmt.Errorf("failed to open log file: %w", err)
    }
    
    return &FileLogger{file: file}, nil
}

func (fl *FileLogger) Write(data []byte) (int, error) {
    fl.mu.Lock()
    defer fl.mu.Unlock()
    
    if fl.file == nil {
        return 0, fmt.Errorf("file logger closed")
    }
    
    return fl.file.Write(data)
}

func (fl *FileLogger) Close() error {
    fl.mu.Lock()
    defer fl.mu.Unlock()
    
    if fl.file == nil {
        return nil
    }
    
    err := fl.file.Close()
    fl.file = nil
    return err
}
```

### 3. Configuration Validation

Validate configuration thoroughly:

```go
type LogConfig struct {
    Level           LogLevel
    ShowFileLine    bool
    CallerSkipFrames int
    OutputFile      string
    MaxFileSize     int64
    MaxBackups      int
}

func (c *LogConfig) Validate() error {
    if c.Level < PanicLevel || c.Level > TraceLevel {
        return fmt.Errorf("invalid log level: %d", c.Level)
    }
    
    if c.CallerSkipFrames < 0 || c.CallerSkipFrames > 20 {
        return fmt.Errorf("invalid caller skip frames: %d (must be 0-20)", c.CallerSkipFrames)
    }
    
    if c.OutputFile != "" {
        dir := filepath.Dir(c.OutputFile)
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("cannot create log directory: %w", err)
        }
    }
    
    if c.MaxFileSize < 1024 {
        return fmt.Errorf("max file size too small: %d (minimum 1024 bytes)", c.MaxFileSize)
    }
    
    if c.MaxBackups < 0 {
        return fmt.Errorf("max backups cannot be negative: %d", c.MaxBackups)
    }
    
    return nil
}

func ApplyConfig(config *LogConfig) error {
    if err := config.Validate(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    SetLogLevel(config.Level)
    SetShowFileLine(config.ShowFileLine)
    SetCallerSkipFrames(config.CallerSkipFrames)
    
    if config.OutputFile != "" {
        SetLogFile(config.OutputFile)
        SetFileLogging(true)
    }
    
    return nil
}
```

## Production Considerations

### 1. Performance Monitoring

Monitor logging performance in production:

```go
var (
    logMessagesTotal   uint64
    logErrorsTotal     uint64
    logDurationTotal   int64  // nanoseconds
    logDurationSamples uint64
)

func recordLogMetrics(duration time.Duration, hadError bool) {
    atomic.AddUint64(&logMessagesTotal, 1)
    atomic.AddInt64(&logDurationTotal, duration.Nanoseconds())
    atomic.AddUint64(&logDurationSamples, 1)
    
    if hadError {
        atomic.AddUint64(&logErrorsTotal, 1)
    }
}

func GetLogMetrics() map[string]interface{} {
    messages := atomic.LoadUint64(&logMessagesTotal)
    errors := atomic.LoadUint64(&logErrorsTotal)
    totalDuration := atomic.LoadInt64(&logDurationTotal)
    samples := atomic.LoadUint64(&logDurationSamples)
    
    var avgDuration float64
    if samples > 0 {
        avgDuration = float64(totalDuration) / float64(samples)
    }
    
    return map[string]interface{}{
        "messages_total":    messages,
        "errors_total":      errors,
        "avg_duration_ns":   avgDuration,
        "error_rate":        float64(errors) / float64(messages),
    }
}
```

### 2. Graceful Degradation

Handle failure scenarios gracefully:

```go
func (l *Logger) safeWrite(message string) {
    defer func() {
        if r := recover(); r != nil {
            // Logging failed, write to stderr as fallback
            fmt.Fprintf(os.Stderr, "LOGGING PANIC: %v\nOriginal message: %s\n", r, message)
        }
    }()
    
    // Try primary output
    if err := l.writeToOutput(message); err != nil {
        // Primary failed, try fallback
        if err := l.writeToFallback(message); err != nil {
            // Both failed, write to stderr
            fmt.Fprintf(os.Stderr, "LOGGING ERROR: %v\nOriginal message: %s\n", err, message)
        }
    }
}

func (l *Logger) writeToOutput(message string) error {
    if l.output == nil {
        return fmt.Errorf("no output configured")
    }
    
    _, err := l.output.Write([]byte(message + "\n"))
    return err
}

func (l *Logger) writeToFallback(message string) error {
    // Fallback to os.Stdout
    _, err := os.Stdout.Write([]byte(message + "\n"))
    return err
}
```

### 3. Resource Limits

Implement resource limits:

```go
type RateLimitedLogger struct {
    *Logger
    limiter *rate.Limiter
    dropped uint64
}

func NewRateLimitedLogger(logger *Logger, rps int) *RateLimitedLogger {
    return &RateLimitedLogger{
        Logger:  logger,
        limiter: rate.NewLimiter(rate.Limit(rps), rps),
    }
}

func (rl *RateLimitedLogger) Info(message string, args ...interface{}) {
    if !rl.limiter.Allow() {
        atomic.AddUint64(&rl.dropped, 1)
        
        // Periodically log that messages are being dropped
        if atomic.LoadUint64(&rl.dropped)%1000 == 0 {
            rl.Logger.Warning("Rate limit exceeded, dropped %d messages", 
                atomic.LoadUint64(&rl.dropped))
        }
        return
    }
    
    rl.Logger.Info(message, args...)
}

func (rl *RateLimitedLogger) GetDroppedCount() uint64 {
    return atomic.LoadUint64(&rl.dropped)
}
```

## Performance Testing Framework

Create a comprehensive performance testing framework:

```go
// performance_test.go
package main

import (
    "fmt"
    "runtime"
    "testing"
    "time"
)

type PerformanceResult struct {
    Name           string
    MessagesPerSec float64
    AllocsPerOp    uint64
    BytesPerOp     uint64
    Duration       time.Duration
}

func RunPerformanceTest(name string, iterations int, testFunc func()) PerformanceResult {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    start := time.Now()
    testFunc()
    duration := time.Since(start)
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    return PerformanceResult{
        Name:           name,
        MessagesPerSec: float64(iterations) / duration.Seconds(),
        AllocsPerOp:    (m2.TotalAlloc - m1.TotalAlloc) / uint64(iterations),
        BytesPerOp:     (m2.Mallocs - m1.Mallocs) / uint64(iterations),
        Duration:       duration,
    }
}

func TestPerformanceSuite(t *testing.T) {
    const iterations = 100000
    
    tests := []struct {
        name string
        test func()
    }{
        {
            name: "Simple Logging",
            test: func() {
                logger := NewLoggerWithOutput(ioutil.Discard)
                for i := 0; i < iterations; i++ {
                    logger.Info("Simple message")
                }
            },
        },
        {
            name: "Formatted Logging",
            test: func() {
                logger := NewLoggerWithOutput(ioutil.Discard)
                for i := 0; i < iterations; i++ {
                    logger.Info("Message %d with %s", i, "formatting")
                }
            },
        },
        {
            name: "Filtered Logging",
            test: func() {
                SetLogLevel(ErrorLevel)
                defer SetLogLevel(InfoLevel)
                
                logger := NewLoggerWithOutput(ioutil.Discard)
                for i := 0; i < iterations; i++ {
                    logger.Info("This message is filtered")
                }
            },
        },
    }
    
    for _, test := range tests {
        result := RunPerformanceTest(test.name, iterations, test.test)
        
        t.Logf("Performance Test: %s", result.Name)
        t.Logf("  Messages/sec: %.2f", result.MessagesPerSec)
        t.Logf("  Allocs/op:    %d", result.AllocsPerOp)
        t.Logf("  Bytes/op:     %d", result.BytesPerOp)
        t.Logf("  Duration:     %v", result.Duration)
        t.Logf("")
        
        // Set performance thresholds
        if result.MessagesPerSec < 50000 {
            t.Errorf("Performance too slow for %s: %.2f msg/sec", result.Name, result.MessagesPerSec)
        }
        
        if result.AllocsPerOp > 10 {
            t.Errorf("Too many allocations for %s: %d allocs/op", result.Name, result.AllocsPerOp)
        }
    }
}
```

## Summary

Performance and quality are critical for logging libraries. Key takeaways:

1. **Optimize the Hot Path**: Make filtered log levels as fast as possible
2. **Manage Memory**: Use pools, avoid leaks, minimize allocations
3. **Profile Regularly**: Measure performance and find bottlenecks
4. **Handle Failures Gracefully**: Logging should never bring down your application
5. **Monitor in Production**: Track performance metrics in real deployments

## Next Module

In the final module, you'll learn about:
- **Module 7**: Contributing Guidelines and Community Practices
- Pull request process and code review
- Community interaction and maintenance
- Long-term project health

Master these performance principles to ensure the CLG package remains fast, reliable, and production-ready.
