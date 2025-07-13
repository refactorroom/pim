package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/refactorroom/pim"
)

func main() {
	fmt.Println("=== Custom Output Destinations Demo ===\n")

	// Create a temporary directory for log files
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Create logger configuration
	config := pim.LoggerConfig{
		Level:            pim.InfoLevel,
		EnableJSON:       false,
		TimestampFormat:  "2006-01-02 15:04:05",
		ShowFunctionName: true,
		ShowPackageName:  true,
		ServiceName:      "outputs-demo",
	}

	// Create various output destinations
	writers := createOutputDestinations(config, logDir)

	// Create multi-writer to write to all destinations
	multiWriter := pim.NewMultiWriter(writers...)

	// Create logger core with multi-writer
	logger := pim.NewLoggerCore(config)
	logger.AddWriter(multiWriter)

	fmt.Println("Starting custom outputs demo...")
	fmt.Printf("Writing to %d different output destinations\n\n", len(writers))

	// Demonstrate different output destinations
	demoOutputDestinations(logger)

	// Show buffer contents
	showBufferContents(writers)

	fmt.Println("\n=== Demo completed ===")
	fmt.Println("Check the logs directory to see files created by different writers:")
	fmt.Printf("  Directory: %s\n", logDir)

	// List log files
	if files, err := filepath.Glob(filepath.Join(logDir, "*")); err == nil {
		fmt.Println("  Files found:")
		for _, file := range files {
			if stat, err := os.Stat(file); err == nil {
				fmt.Printf("    %s (%d bytes, %s)\n", filepath.Base(file), stat.Size(), stat.ModTime().Format("15:04:05"))
			}
		}
	}
}

func createOutputDestinations(config pim.LoggerConfig, logDir string) []pim.LogWriter {
	var writers []pim.LogWriter

	// 1. Console writer (stdout)
	consoleWriter := pim.NewConsoleWriter(config)
	writers = append(writers, consoleWriter)
	fmt.Println("✓ Console writer (stdout)")

	// 2. Stderr writer
	stderrWriter := pim.NewStderrWriter(config)
	writers = append(writers, stderrWriter)
	fmt.Println("✓ Stderr writer")

	// 3. File writer with rotation
	rotationConfig := pim.RotationConfig{
		MaxSize:         1024 * 5,         // 5KB max file size
		MaxAge:          1 * time.Hour,    // Keep files for 1 hour
		MaxFiles:        3,                // Keep max 3 files
		Compress:        true,             // Compress old files
		RotateTime:      30 * time.Second, // Rotate every 30 seconds for demo
		CleanupInterval: 10 * time.Second, // Cleanup every 10 seconds for demo
		VerboseCleanup:  true,             // Show cleanup operations
	}

	fileWriter, err := pim.NewFileWriter(
		filepath.Join(logDir, "app.log"),
		config,
		rotationConfig,
	)
	if err != nil {
		log.Printf("Failed to create file writer: %v", err)
	} else {
		writers = append(writers, fileWriter)
		fmt.Println("✓ File writer with rotation")
	}

	// 4. Buffer writer (in-memory)
	bufferWriter := pim.NewBufferWriter(config, 100)
	writers = append(writers, bufferWriter)
	fmt.Println("✓ Buffer writer (in-memory, 100 entries)")

	// 5. Conditional writer (only errors and warnings)
	errorFileWriter, err := pim.NewFileWriter(
		filepath.Join(logDir, "errors.log"),
		config,
		pim.RotationConfig{},
	)
	if err != nil {
		log.Printf("Failed to create error file writer: %v", err)
	} else {
		conditionalWriter := pim.NewConditionalWriter(errorFileWriter, func(entry pim.CoreLogEntry) bool {
			return entry.Level <= pim.WarningLevel // Only errors and warnings
		})
		writers = append(writers, conditionalWriter)
		fmt.Println("✓ Conditional writer (errors and warnings only)")
	}

	// 6. Rate-limited writer (max 10 logs per second)
	rateLimitedWriter := pim.NewRateLimitedWriter(
		pim.NewConsoleWriter(config),
		10, // max 10 logs per second
	)
	writers = append(writers, rateLimitedWriter)
	fmt.Println("✓ Rate-limited writer (10 logs/second)")

	// 7. Remote writer (simulated - would normally point to a real endpoint)
	remoteConfig := pim.RemoteWriterConfig{
		Endpoint:      "http://localhost:8080/logs", // Simulated endpoint
		Headers:       map[string]string{"X-API-Key": "demo-key"},
		Timeout:       5 * time.Second,
		BatchSize:     10,
		BatchDelay:    2 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
	}

	// Note: This will fail in demo since there's no server, but shows the concept
	remoteWriter := pim.NewRemoteWriter(config, remoteConfig)
	writers = append(writers, remoteWriter)
	fmt.Println("✓ Remote writer (HTTP endpoint - will fail in demo)")

	// 8. Null writer (discards all logs - useful for testing)
	nullWriter := pim.NewNullWriter()
	writers = append(writers, nullWriter)
	fmt.Println("✓ Null writer (discards all logs)")

	// 9. Syslog writer (Unix systems only)
	// Note: This will fail on Windows, but shows the concept
	syslogWriter := pim.NewSyslogWriter(config, "pim-demo")
	writers = append(writers, syslogWriter)
	fmt.Println("✓ Syslog writer (Unix systems only)")

	return writers
}

func demoOutputDestinations(logger *pim.LoggerCore) {
	// Scenario 1: Normal application startup
	logger.InfoWithFields("Application started", map[string]interface{}{
		"version": "1.0.0",
		"port":    8080,
		"outputs": "multiple",
	})

	// Scenario 2: User authentication (info level)
	logger.InfoWithFields("User authenticated", map[string]interface{}{
		"user_id": "user123",
		"method":  "password",
		"ip":      "192.168.1.100",
		"success": true,
	})

	// Scenario 3: Warning message (should go to conditional writer)
	logger.WarningWithFields("High memory usage detected", map[string]interface{}{
		"memory_usage": "85%",
		"threshold":    "80%",
		"action":       "monitor",
	})

	// Scenario 4: Error message (should go to conditional writer)
	logger.ErrorWithFields("Database connection failed", map[string]interface{}{
		"error":    "connection timeout",
		"retries":  3,
		"host":     "db.example.com",
		"severity": "high",
	})

	// Scenario 5: Performance metrics
	logger.InfoWithFields("Performance metrics", map[string]interface{}{
		"cpu_usage":     "45%",
		"memory_usage":  "60%",
		"disk_usage":    "30%",
		"active_users":  150,
		"response_time": "120ms",
	})

	// Scenario 6: Business events
	logger.InfoWithFields("Order placed", map[string]interface{}{
		"order_id": "ORD-12345",
		"user_id":  "user123",
		"amount":   99.99,
		"currency": "USD",
		"items":    3,
		"status":   "pending",
	})

	// Scenario 7: System events
	logger.InfoWithFields("Backup completed", map[string]interface{}{
		"backup_id": "backup-2024-01-15",
		"size":      "2.5GB",
		"duration":  "15m30s",
		"status":    "success",
		"location":  "s3://backups/",
	})

	// Scenario 8: Security events (warning level)
	logger.WarningWithFields("Failed login attempt", map[string]interface{}{
		"username": "admin",
		"ip":       "192.168.1.200",
		"reason":   "invalid password",
		"attempts": 5,
		"blocked":  true,
	})

	// Scenario 9: Debug information
	logger.DebugWithFields("Cache miss", map[string]interface{}{
		"cache_key": "user:123:profile",
		"source":    "database",
		"duration":  "5ms",
	})

	// Scenario 10: Trace information
	logger.TraceWithFields("Function call", map[string]interface{}{
		"function": "processOrder",
		"params":   map[string]interface{}{"order_id": "ORD-12345"},
		"duration": "2ms",
	})

	// Wait a bit to see rate limiting and batching in action
	fmt.Println("\nWaiting for rate limiting and batching...")
	time.Sleep(3 * time.Second)

	// Continue with more logs to trigger rotation and rate limiting
	for i := 1; i <= 15; i++ {
		logger.InfoWithFields("Additional log entry", map[string]interface{}{
			"iteration": i,
			"timestamp": time.Now().Unix(),
			"sequence":  fmt.Sprintf("seq-%03d", i),
		})
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for cleanup and batching
	fmt.Println("\nWaiting for cleanup and batching operations...")
	time.Sleep(3 * time.Second)
}

func showBufferContents(writers []pim.LogWriter) {
	fmt.Println("\n=== Buffer Contents ===")

	// Find buffer writer and show its contents
	for _, writer := range writers {
		if bufferWriter, ok := writer.(*pim.BufferWriter); ok {
			entries := bufferWriter.GetBuffer()
			fmt.Printf("Buffer contains %d entries:\n", len(entries))

			for i, entry := range entries {
				if i >= 5 { // Show only first 5 entries
					fmt.Printf("  ... and %d more entries\n", len(entries)-5)
					break
				}
				fmt.Printf("  [%s] %s: %s\n",
					entry.Timestamp.Format("15:04:05"),
					entry.LevelString,
					entry.Message)
			}
			break
		}
	}
}
