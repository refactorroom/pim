# pim

[![Go Reference](https://pkg.go.dev/badge/github.com/refactorroom/pim.svg)](https://pkg.go.dev/github.com/refactorroom/pim)

**pim** is a professional, extensible logging package for Go. It supports colored and themed output, structured/context logging, JSON formatting, async/buffered logging, log rotation, multi-writer, hooks, metrics, tracing, localization, and much more.

See full API documentation and examples at [pkg.go.dev/github.com/refactorroom/pim](https://pkg.go.dev/github.com/refactorroom/pim).

## Comprehensive Features
- **Colored & Themed Output:** Customizable color themes, icons, and templates for beautiful terminal logs.
- **Multiple Log Levels:** Panic, Error, Warning, Info, Success, Debug, Trace, and custom levels.
- **Structured Logging:** Arbitrary key-value/context fields, JSON logs, and context propagation.
- **Async/Buffered Logging:** High-performance, non-blocking logging with background flush and graceful shutdown.
- **Log Rotation & Retention:** File size/time-based rotation, max file count, compression, and cleanup.
- **Multi-Writer & Custom Outputs:** Log to stdout, stderr, files, memory, remote endpoints, syslog, and more.
- **Hooks & Filtering:** Pre-output hooks for filtering, redaction, enrichment, metrics, and custom logic.
- **Rich Formatting & Theming:** Customizable color themes, timestamp/message templates, and user-defined formats.
- **Caller/Source Info:** Configurable file, line, function, package, and stack trace inclusion.
- **Error Wrapping & Stack Traces:** Rich error types, stack trace capture, and formatting.
- **Tracing/Telemetry Integration:** Trace/span/correlation IDs, context propagation, and distributed tracing support.
- **Localization/Internationalization:** Multi-language log messages, locale detection, and extensible catalogs.
- **Metrics & Analytics:** Log counts, error rates, and metrics for monitoring and analytics.
- **Graceful Shutdown:** Ensures all logs are flushed on exit, panic, or signals.

## Getting Started

```go
import "github.com/refactorroom/pim"

func main() {
    // Basic usage (global logger)
    pim.Info("Hello, world!")
    pim.Success("Operation completed!")
    pim.Error("Something went wrong: %v", err)

    // Advanced: Create a custom logger core
    config := pim.LoggerConfig{
        Level:         pim.InfoLevel,
        EnableConsole: true,
        EnableColors:  true,
        Async:         true,
        BufferSize:    1000,
        FlushInterval: 2 * time.Second,
        ThemeName:     "monokai",
    }
    logger := pim.NewLoggerCore(config)
    defer logger.Close()

    logger.Info("Structured log", pim.Fields{"user": "alice", "action": "login"})
}
```

## Modern Usage Examples

### Structured & Context Logging
```go
logger.InfoWithFields("User login", map[string]interface{}{"user": "alice", "ip": "1.2.3.4"})
logger.InfoKV("Order placed", "order_id", 123, "amount", 99.99)
```

### Dynamic Log Level Control
```go
logger.SetLevel(pim.DebugLevel)
logger.SetLevelFromString("trace")
logger.SetLevelFromEnv("LOG_LEVEL")
```

### Async Logging & Graceful Shutdown
```go
pim.InstallExitHandler() // Flushes all logs on exit/panic
logger := pim.NewLoggerCore(pim.LoggerConfig{Async: true, ...})
defer logger.Close()
```

### Log Rotation & Retention
```go
rotation := pim.RotationConfig{MaxSize: 10*1024*1024, MaxFiles: 5, Compress: true}
fileWriter, _ := pim.NewFileWriter("app.log", config, rotation)
logger.AddWriter(fileWriter)
```

### Multi-Writer & Custom Outputs
```go
logger.AddWriter(pim.NewConsoleWriter(config))
logger.AddWriter(pim.NewStderrWriter(config))
logger.AddWriter(pim.NewRemoteWriter(config, remoteCfg))
```

### Theming & Formatting
```go
logger.SetTheme("monokai")
logger.SetCustomTheme(&pim.Theme{...})
logger.RegisterTemplate("custom", "[{timestamp}] {level} {message}")
```

### Enhanced Caller/Source Info
```go
cfg := pim.NewCallerInfoConfig()
cfg.ShowFileLine = true
cfg.ShowFunctionName = true
cfg.StackDepth = 5
logger.SetCallerInfoConfig(cfg)
```

### Error Wrapping & Stack Traces
```go
err := pim.NewError("failed to process", pim.WithStack())
logger.Error("Error occurred: %v", err)
```

### Tracing/Telemetry Integration
```go
logger.WithTrace("trace-id").WithSpan("span-id").Info("Tracing event")
```

### Hooks & Filtering
```go
logger.AddHookFunc(func(entry pim.CoreLogEntry) (pim.CoreLogEntry, error) {
    if entry.Level == pim.DebugLevel { return entry, pim.ErrSkip }
    return entry, nil
})
```

### Localization/Internationalization
```go
loc := pim.NewLocalizedLogger(config, pim.Locale{Language: "en", Region: "US"})
loc.TInfo("app_started")
loc.SetLocale(pim.Locale{Language: "es"})
loc.TInfo("app_started") // "Aplicación iniciada"
loc.AddCustomMessage("custom_welcome", "Welcome, {0}!")
loc.TInfo("custom_welcome", "Alice")
```

### Metrics & Analytics
```go
metrics := logger.GetMetrics()
logger.ResetMetrics()
```

## Testing & Examples
- See the [example/](example/) directory for comprehensive demos of every feature.
- Run `go test` for full test coverage (see `*_test.go` files).

## Migration & Compatibility
- `pim` is the successor to `clg`, with a modern, extensible architecture.
- All legacy APIs are supported, but new code should use the ergonomic, structured APIs shown above.

## License
MIT