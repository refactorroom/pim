# 🎨 pim - Professional Go Logging Package

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/refactorrom/pim.svg)](https://pkg.go.dev/github.com/refactorrom/pim)
[![Go Report Card](https://goreportcard.com/badge/github.com/refactorrom/pim)](https://goreportcard.com/report/github.com/refactorrom/pim)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)

**✨ A modern, extensible logging package for Go with beautiful output, structured logging, and enterprise features**

[📖 Documentation](documents/) • [🚀 Examples](example/) • [🔧 API Reference](documents/api_reference.md) • [📚 Contributing](documents/lession_contribute/)

</div>

---

## 🌟 Features

<div align="center">

| 🎨 **Beautiful Output** | 🔧 **Structured Logging** | ⚡ **High Performance** | 🛡️ **Enterprise Ready** |
|------------------------|---------------------------|------------------------|------------------------|
| Colored themes & icons | JSON & key-value fields | Async buffered logging | Log rotation & retention |
| Custom templates | Context propagation | Zero-allocation paths | Multi-writer support |
| Rich formatting | Error wrapping | Graceful shutdown | Hooks & filtering |

</div>

### 🎯 Core Capabilities

- **🌈 Colored & Themed Output** - Customizable color themes, icons, and templates for beautiful terminal logs
- **📊 Multiple Log Levels** - Panic, Error, Warning, Info, Success, Debug, Trace, and custom levels
- **🔗 Structured Logging** - Arbitrary key-value/context fields, JSON logs, and context propagation
- **⚡ Async/Buffered Logging** - High-performance, non-blocking logging with background flush
- **📁 Log Rotation & Retention** - File size/time-based rotation, compression, and cleanup
- **🖥️ Multi-Writer Support** - Log to stdout, stderr, files, memory, remote endpoints, syslog
- **🎣 Hooks & Filtering** - Pre-output hooks for filtering, redaction, enrichment, metrics
- **📍 Caller/Source Info** - Configurable file, line, function, package, and stack traces
- **🌍 Localization** - Multi-language log messages, locale detection, and extensible catalogs
- **📈 Metrics & Analytics** - Log counts, error rates, and monitoring capabilities
- **🔄 Graceful Shutdown** - Ensures all logs are flushed on exit, panic, or signals

---

## 🚀 Quick Start

### Installation

```bash
go get github.com/refactorrom/pim
```

### Basic Usage

```go
package main

import "github.com/refactorrom/pim"

func main() {
    // 🎯 Simple logging
    pim.Info("Hello, world!")
    pim.Success("Operation completed!")
    pim.Error("Something went wrong: %v", err)

    // 🎨 Create a custom logger with beautiful output
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

    // 📊 Structured logging
    logger.Info("User login", pim.Fields{"user": "alice", "ip": "1.2.3.4"})
}
```

---

## 🎨 Modern Usage Examples

### 📊 Structured & Context Logging

```go
// Key-value logging
logger.InfoKV("Order placed", "order_id", 123, "amount", 99.99)

// Structured fields
logger.InfoWithFields("User login", map[string]interface{}{
    "user": "alice",
    "ip":   "1.2.3.4",
    "time": time.Now(),
})
```

### 🎛️ Dynamic Log Level Control

```go
// Set levels programmatically
logger.SetLevel(pim.DebugLevel)
logger.SetLevelFromString("trace")

// From environment
logger.SetLevelFromEnv("LOG_LEVEL")
```

### ⚡ Async Logging & Graceful Shutdown

```go
// Install exit handler for graceful shutdown
pim.InstallExitHandler()

// Create async logger
logger := pim.NewLoggerCore(pim.LoggerConfig{
    Async:         true,
    BufferSize:    1000,
    FlushInterval: 2 * time.Second,
})
defer logger.Close()
```

### 📁 Log Rotation & Retention

```go
rotation := pim.RotationConfig{
    MaxSize:   10 * 1024 * 1024, // 10MB
    MaxFiles:  5,
    Compress:  true,
    Retention: 30 * 24 * time.Hour, // 30 days
}

fileWriter, _ := pim.NewFileWriter("app.log", config, rotation)
logger.AddWriter(fileWriter)
```

### 🖥️ Multi-Writer & Custom Outputs

```go
// Multiple output destinations
logger.AddWriter(pim.NewConsoleWriter(config))
logger.AddWriter(pim.NewStderrWriter(config))
logger.AddWriter(pim.NewFileWriter("app.log", config, nil))
logger.AddWriter(pim.NewRemoteWriter(config, remoteCfg))
```

### 🎨 Theming & Formatting

```go
// Use built-in themes
logger.SetTheme("monokai")
logger.SetTheme("dracula")
logger.SetTheme("nord")

// Custom theme
customTheme := &pim.Theme{
    InfoColor:    pim.Color{Red: 0, Green: 255, Blue: 0},
    ErrorColor:   pim.Color{Red: 255, Green: 0, Blue: 0},
    WarningColor: pim.Color{Red: 255, Green: 255, Blue: 0},
}
logger.SetCustomTheme(customTheme)

// Custom templates
logger.RegisterTemplate("custom", "[{timestamp}] {level} {message}")
```

### 📍 Enhanced Caller/Source Info

```go
cfg := pim.NewCallerInfoConfig()
cfg.ShowFileLine = true
cfg.ShowFunctionName = true
cfg.ShowPackageName = true
cfg.StackDepth = 5
logger.SetCallerInfoConfig(cfg)
```

### 🔗 Error Wrapping & Stack Traces

```go
// Rich error types with stack traces
err := pim.NewError("failed to process", pim.WithStack())
logger.Error("Error occurred: %v", err)

// Error with context
err = pim.WrapError(err, "additional context")
```

### 🔍 Tracing/Telemetry Integration

```go
// Distributed tracing support
logger.WithTrace("trace-id").WithSpan("span-id").Info("Tracing event")

// Correlation IDs
logger.WithCorrelationID("corr-123").Info("Correlated log")
```

### 🎣 Hooks & Filtering

```go
// Add custom hooks
logger.AddHookFunc(func(entry pim.CoreLogEntry) (pim.CoreLogEntry, error) {
    // Skip debug logs in production
    if entry.Level == pim.DebugLevel {
        return entry, pim.ErrSkip
    }
    
    // Add timestamp to all logs
    entry.Fields["timestamp"] = time.Now().Unix()
    return entry, nil
})

// Metrics hook
logger.AddHookFunc(func(entry pim.CoreLogEntry) (pim.CoreLogEntry, error) {
    // Increment metrics
    metrics.Increment("logs." + entry.Level.String())
    return entry, nil
})
```

### 🌍 Localization/Internationalization

```go
// Create localized logger
loc := pim.NewLocalizedLogger(config, pim.Locale{Language: "en", Region: "US"})

// Use translation keys
loc.TInfo("app_started") // "Application started"

// Change locale
loc.SetLocale(pim.Locale{Language: "es"})
loc.TInfo("app_started") // "Aplicación iniciada"

// Custom messages
loc.AddCustomMessage("custom_welcome", "Welcome, {0}!")
loc.TInfo("custom_welcome", "Alice") // "Welcome, Alice!"
```

### 📈 Metrics & Analytics

```go
// Get logging metrics
metrics := logger.GetMetrics()
fmt.Printf("Total logs: %d\n", metrics.TotalLogs)
fmt.Printf("Error rate: %.2f%%\n", metrics.ErrorRate)

// Reset metrics
logger.ResetMetrics()
```

---

## 🧪 Testing & Examples

<div align="center">

**📁 [Comprehensive Examples](example/)**

| Example | Description |
|---------|-------------|
| [Basic Logger](example/basic_logger/) | Simple logging setup |
| [File Logging](example/file_logging_demo/) | File output with rotation |
| [Caller Info](example/caller_test/) | Stack trace and caller information |
| [JSON Logger](example/json_logger/) | Structured JSON logging |
| [Theming](example/theming/) | Custom colors and themes |
| [Localization](example/localization_demo/) | Multi-language support |
| [Hooks](example/hooks_demo/) | Custom hooks and filtering |
| [Full Demo](example/full_demo/) | Complete feature showcase |

</div>

### Running Examples

```bash
# Run all examples
cd example/basic_logger && go run basic_logger.go
cd example/file_logging_demo && go run main.go
cd example/theming && go run main.go

# Run tests
go test ./...
go test -cover ./...
```

---

## 📚 Documentation

<div align="center">

| 📖 **User Guides** | 🔧 **Developer Docs** | 📋 **API Reference** |
|-------------------|----------------------|---------------------|
| [Installation](documents/installation.md) | [Getting Started](documents/lession_contribute/01_getting_started.md) | [API Reference](documents/api_reference.md) |
| [Configuration](documents/configuration.md) | [Architecture](documents/lession_contribute/02_architecture.md) | [Migration Guide](documents/migration.md) |
| [Performance](documents/performance.md) | [Testing](documents/lession_contribute/04_writing_tests.md) | [Troubleshooting](documents/troubleshooting.md) |
| [Security](documents/SECURITY.md) | [Contributing](documents/lession_contribute/README.md) | [Examples](documents/lession_contribute/EXAMPLES.md) |

</div>

---

## 🔄  Compatibility

- **🚀 Modern APIs** - New ergonomic, structured APIs for better developer experience
- **📦 Easy Migration** - Simple import path changes with comprehensive migration guide

---

## 🤝 Contributing

We welcome contributions! See our [Contributing Guide](documents/lession_contribute/README.md) for:

- 🚀 [Getting Started](documents/lession_contribute/01_getting_started.md)
- 🏗️ [Architecture Overview](documents/lession_contribute/02_architecture.md)
- ⚙️ [Configuration System](documents/lession_contribute/03_configuration_system.md)
- 🧪 [Testing Guidelines](documents/lession_contribute/04_writing_tests.md)
- ✨ [Adding Features](documents/lession_contribute/05_adding_features.md)
- ⚡ [Performance Best Practices](documents/lession_contribute/06_performance_best_practices.md)
- 👥 [Community Guidelines](documents/lession_contribute/07_community_guidelines.md)

---

## 📄 License

<div align="center">

**MIT License** - see [LICENSE](LICENSE) for details

Made with ❤️ by The Refactor Room community

[![GitHub stars](https://img.shields.io/github/stars/refactorrom/pim?style=social)](https://github.com/refactorrom/pim)
[![GitHub forks](https://img.shields.io/github/forks/refactorrom/pim?style=social)](https://github.com/refactorrom/pim)

</div>