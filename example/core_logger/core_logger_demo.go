package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/refactorroom/pim"
)

func main() {
	fmt.Println("=== Extensible Logger Core Demo ===\n")

	// Create a new logger with custom configuration
	config := pim.LoggerConfig{
		Level:            pim.InfoLevel,
		ServiceName:      "core-demo",
		TimestampFormat:  "2006-01-02 15:04:05.000",
		ShowFileLine:     true,
		ShowFunctionName: true,
		ShowPackageName:  true,
		ShowGoroutineID:  true,
		ShowFullPath:     false,
		StackDepth:       3,
		EnableColors:     true,
		EnableJSON:       false,
		EnableConsole:    true,
		Async:            false,
		BufferSize:       1000,
		FlushInterval:    5 * time.Second,
		EnableSampling:   false,
		SampleRate:       1.0,
		PropagateContext: true,
	}

	logger := pim.NewLoggerCore(config)
	defer logger.Close()

	// Example 1: Basic logging
	fmt.Println("1. Basic Logging:")
	logger.Info("This is a basic info message")
	logger.Success("Operation completed successfully")
	logger.Warning("This is a warning message")
	logger.Error("This is an error message")

	// Example 2: Logging with context
	fmt.Println("\n2. Logging with Context:")
	logger.LogWithContext(pim.InfoLevel, pim.InfoPrefix, "User action performed",
		map[string]interface{}{
			"user_id":    12345,
			"action":     "login",
			"ip_address": "192.168.1.100",
			"user_agent": "Mozilla/5.0...",
		})

	// Example 2b: Structured logging with InfoWithFields
	fmt.Println("\n2b. Structured Logging with InfoWithFields:")
	logger.InfoWithFields("User updated profile", map[string]interface{}{
		"user_id": 12345,
		"fields":  []string{"email", "avatar"},
		"success": true,
	})

	// Example 2c: Structured logging with InfoKV
	fmt.Println("\n2c. Structured Logging with InfoKV:")
	logger.InfoKV("User performed action", "user_id", 12345, "action", "logout", "ip", "10.0.0.1")

	// Example 3: Using WithField and WithFields
	fmt.Println("\n3. Using WithField and WithFields:")
	userLogger := logger.WithField("user_id", 12345)
	userLogger.Info("User logged in")

	requestLogger := logger.WithFields(map[string]interface{}{
		"request_id": "req_12345",
		"method":     "POST",
		"endpoint":   "/api/users",
	})
	requestLogger.Info("API request received")

	// Example 4: Adding hooks
	fmt.Println("\n4. Adding Hooks:")

	// Hook to add request ID to all logs
	logger.AddHookFunc(func(entry pim.CoreLogEntry) (pim.CoreLogEntry, error) {
		if entry.Context == nil {
			entry.Context = make(map[string]interface{})
		}
		entry.Context["request_id"] = "auto_generated_123"
		return entry, nil
	})

	// Hook to redact sensitive information
	logger.AddHookFunc(func(entry pim.CoreLogEntry) (pim.CoreLogEntry, error) {
		if entry.Context != nil {
			if _, exists := entry.Context["password"]; exists {
				entry.Context["password"] = "[REDACTED]"
			}
			if _, exists := entry.Context["token"]; exists {
				entry.Context["token"] = "[REDACTED]"
			}
		}
		return entry, nil
	})

	logger.Info("Processing user data", "user_id", 123, "password", "secret123", "token", "abc123")

	// Example 5: Adding multiple writers
	fmt.Println("\n5. Adding Multiple Writers:")

	// Add a file writer
	fileWriter, err := pim.NewFileWriter("demo.log", config, pim.RotationConfig{})
	if err != nil {
		logger.Error("Failed to create file writer", err)
	} else {
		logger.AddWriter(fileWriter)
		logger.Info("This message will be written to both console and file")
	}

	// Example 6: Setting global context
	fmt.Println("\n6. Setting Global Context:")
	logger.SetContext("service_version", "1.0.0")
	logger.SetContext("environment", "development")
	logger.SetContext("deployment_id", "deploy_12345")

	logger.Info("Application started")
	logger.Info("Configuration loaded")

	// Example 7: Logging with stack traces
	fmt.Println("\n7. Logging with Stack Traces:")
	logger.LogWithStackTrace(pim.ErrorLevel, pim.ErrorPrefix, "An error occurred during processing")

	// Example 8: Sampling demonstration
	fmt.Println("\n8. Sampling Demonstration:")

	// Create a new logger with sampling enabled
	samplingConfig := config
	samplingConfig.EnableSampling = true
	samplingConfig.SampleRate = 0.3 // Only log 30% of messages

	samplingLogger := pim.NewLoggerCore(samplingConfig)
	defer samplingLogger.Close()

	fmt.Println("Logging 10 messages with 30% sampling rate:")
	for i := 0; i < 10; i++ {
		samplingLogger.Info("High-frequency message", "sequence", i)
	}

	// Example 8b: Rate-based sampling demonstration
	fmt.Println("\n8b. Rate-based Sampling Demonstration:")
	rateConfig := config
	rateConfig.EnableSampling = true
	rateConfig.SampleRate = 0.0
	rateLogger := pim.NewLoggerCore(rateConfig)
	defer rateLogger.Close()
	rateLogger.SetSamplingByLevel(map[pim.LogLevel]pim.SamplingConfig{
		pim.InfoLevel: {EnableSampling: true, Rate: 3}, // Log every 3rd info message
	})
	fmt.Println("Logging 10 info messages with rate-based sampling (every 3rd):")
	for i := 1; i <= 10; i++ {
		rateLogger.InfoKV("Rate-sampled info message", "sequence", i)
	}

	// Example 8c: Per-level sampling demonstration
	fmt.Println("\n8c. Per-level Sampling Demonstration:")
	perLevelConfig := config
	perLevelConfig.EnableSampling = false
	perLevelLogger := pim.NewLoggerCore(perLevelConfig)
	defer perLevelLogger.Close()
	perLevelLogger.SetSamplingByLevel(map[pim.LogLevel]pim.SamplingConfig{
		pim.InfoLevel:  {EnableSampling: true, SampleRate: 0.2}, // 20% of info
		pim.DebugLevel: {EnableSampling: true, Rate: 2},         // every 2nd debug
	})
	fmt.Println("Logging 10 info and 10 debug messages with per-level sampling:")
	for i := 1; i <= 10; i++ {
		perLevelLogger.InfoKV("Per-level sampled info", "sequence", i)
		perLevelLogger.DebugKV("Per-level sampled debug", "sequence", i)
	}

	// Example 9: JSON output
	fmt.Println("\n9. JSON Output:")
	jsonConfig := config
	jsonConfig.EnableJSON = true

	jsonLogger := pim.NewLoggerCore(jsonConfig)
	defer jsonLogger.Close()

	jsonLogger.LogWithContext(pim.InfoLevel, pim.InfoPrefix, "JSON formatted log entry",
		map[string]interface{}{
			"user_id":   12345,
			"action":    "data_export",
			"file_size": 1024,
			"completed": true,
		})

	// Example 9b: JSON output with structured fields
	fmt.Println("\n9b. JSON Output with Structured Fields:")
	jsonLogger.InfoWithFields("Export completed", map[string]interface{}{
		"user_id":  12345,
		"file":     "report.csv",
		"size":     2048,
		"duration": 1.23,
		"success":  true,
	})
	jsonLogger.InfoKV("User deleted account", "user_id", 54321, "reason", "requested by user", "timestamp", time.Now().Format(time.RFC3339))

	// Example 10: Custom hook for metrics
	fmt.Println("\n10. Custom Hook for Metrics:")

	// Hook to count log entries by level
	levelCounts := make(map[pim.LogLevel]int)

	logger.AddHookFunc(func(entry pim.CoreLogEntry) (pim.CoreLogEntry, error) {
		levelCounts[entry.Level]++
		return entry, nil
	})

	logger.Info("First message")
	logger.Warning("Second message")
	logger.Error("Third message")
	logger.Info("Fourth message")

	fmt.Printf("Log level counts: %+v\n", levelCounts)

	// Example 11: Asynchronous logging demonstration
	fmt.Println("\n11. Asynchronous Logging Demonstration:")

	// Create async logger with small buffer to demonstrate buffering
	asyncConfig := config
	asyncConfig.Async = true
	asyncConfig.BufferSize = 5
	asyncConfig.FlushInterval = 2 * time.Second

	asyncLogger := pim.NewLoggerCore(asyncConfig)
	defer asyncLogger.Close()

	fmt.Println("Logging 10 messages with async buffering (buffer size: 5):")
	for i := 1; i <= 10; i++ {
		asyncLogger.InfoKV("Async message", "sequence", i, "timestamp", time.Now().Format("15:04:05.000"))
		time.Sleep(100 * time.Millisecond) // Simulate some work
	}

	// Demonstrate flush
	fmt.Println("Flushing remaining messages...")
	asyncLogger.Flush()
	fmt.Println("Flush completed")

	// Example 12: Graceful shutdown demonstration
	fmt.Println("\n12. Graceful Shutdown Demonstration:")

	// Create another async logger
	shutdownLogger := pim.NewLoggerCore(asyncConfig)

	fmt.Println("Logging messages before shutdown...")
	for i := 1; i <= 5; i++ {
		shutdownLogger.InfoKV("Pre-shutdown message", "sequence", i)
	}

	fmt.Println("Closing logger (graceful shutdown)...")
	shutdownLogger.Close()
	fmt.Println("Logger closed successfully")

	// Example 13: Log rotation and retention demonstration
	fmt.Println("\n13. Log Rotation and Retention Demonstration:")

	// Create rotation config
	rotationConfig := pim.RotationConfig{
		MaxSize:    1024,           // 1KB max file size
		MaxAge:     24 * time.Hour, // Keep files for 24 hours
		MaxFiles:   5,              // Keep max 5 files
		Compress:   false,          // Don't compress for demo
		RotateTime: 0,              // No time-based rotation for demo
	}

	// Create file writer with rotation
	fileWriter, err = pim.NewFileWriter("rotated.log", config, rotationConfig)
	if err != nil {
		logger.Error("Failed to create file writer with rotation", err)
	} else {
		logger.AddWriter(fileWriter)

		fmt.Println("Logging messages to trigger rotation (1KB max size):")
		largeMessage := strings.Repeat("This is a large log message to trigger rotation. ", 20)
		for i := 1; i <= 10; i++ {
			logger.InfoKV("Rotated log message", "sequence", i, "message", largeMessage)
		}

		fmt.Println("Check the 'rotated.log' file and rotated files in the directory")
	}

	// Example 14: Time-based rotation demonstration
	fmt.Println("\n14. Time-based Rotation Demonstration:")

	timeRotationConfig := pim.RotationConfig{
		MaxSize:    0,             // No size limit
		MaxAge:     1 * time.Hour, // Keep files for 1 hour
		MaxFiles:   3,             // Keep max 3 files
		Compress:   false,
		RotateTime: 10 * time.Second, // Rotate every 10 seconds for demo
	}

	timeFileWriter, err := pim.NewFileWriter("time-rotated.log", config, timeRotationConfig)
	if err != nil {
		logger.Error("Failed to create time-based file writer", err)
	} else {
		logger.AddWriter(timeFileWriter)

		fmt.Println("Logging messages with time-based rotation (every 10 seconds):")
		for i := 1; i <= 5; i++ {
			logger.InfoKV("Time-rotated message", "sequence", i, "timestamp", time.Now().Format("15:04:05"))
			time.Sleep(3 * time.Second) // Wait 3 seconds between messages
		}

		fmt.Println("Check the 'time-rotated.log' file and time-based rotated files")
	}

	fmt.Println("\n=== Extensible Logger Core Demo Complete ===")
}
