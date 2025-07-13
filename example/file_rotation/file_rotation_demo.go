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
	fmt.Println("=== Log Rotation and Retention Demo ===\n")

	// Create a temporary directory for log files
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Configure rotation settings
	rotationConfig := pim.RotationConfig{
		MaxSize:         1024 * 10,        // 10KB max file size
		MaxAge:          24 * time.Hour,   // Keep files for 24 hours
		MaxFiles:        5,                // Keep max 5 files
		Compress:        true,             // Compress old files
		RotateTime:      1 * time.Minute,  // Rotate every minute for demo
		CleanupInterval: 30 * time.Second, // Cleanup every 30 seconds for demo
		VerboseCleanup:  true,             // Show cleanup operations
	}

	// Create logger configuration
	config := pim.LoggerConfig{
		Level:            pim.InfoLevel,
		EnableJSON:       false,
		TimestampFormat:  "2006-01-02 15:04:05",
		ShowFunctionName: true,
		ShowPackageName:  true,
		ServiceName:      "rotation-demo",
	}

	// Create file writer with rotation
	fileWriter, err := pim.NewFileWriter(
		filepath.Join(logDir, "app.log"),
		config,
		rotationConfig,
	)
	if err != nil {
		log.Fatalf("Failed to create file writer: %v", err)
	}
	defer fileWriter.Close()

	// Create console writer for immediate feedback
	consoleWriter := pim.NewConsoleWriter(config)

	// Create multi-writer to write to both console and file
	multiWriter := pim.NewMultiWriter(consoleWriter, fileWriter)

	// Create logger core with multi-writer
	logger := pim.NewLoggerCore(config)
	logger.AddWriter(multiWriter)

	fmt.Println("Starting log rotation demo...")
	fmt.Printf("Rotation config: MaxSize=%d bytes, MaxAge=%v, MaxFiles=%d, Compress=%v\n",
		rotationConfig.MaxSize, rotationConfig.MaxAge, rotationConfig.MaxFiles, rotationConfig.Compress)
	fmt.Printf("Log files will be written to: %s\n\n", logDir)

	// Simulate application logging with different scenarios
	demoScenarios(logger)

	fmt.Println("\n=== Demo completed ===")
	fmt.Println("Check the logs directory to see rotated files:")
	fmt.Printf("  Directory: %s\n", logDir)
	
	// List log files
	if files, err := filepath.Glob(filepath.Join(logDir, "*.log*")); err == nil {
		fmt.Println("  Files found:")
		for _, file := range files {
			if stat, err := os.Stat(file); err == nil {
				fmt.Printf("    %s (%d bytes, %s)\n", filepath.Base(file), stat.Size(), stat.ModTime().Format("15:04:05"))
			}
		}
	}
}

func demoScenarios(logger *pim.LoggerCore) {
	// Scenario 1: Normal application startup
	logger.InfoWithFields("Application started", map[string]interface{}{
		"version": "1.0.0",
		"port":    8080,
	})

	// Scenario 2: User authentication
	logger.InfoWithFields("User authenticated", map[string]interface{}{
		"user_id": "user123",
		"method":  "password",
		"ip":      "192.168.1.100",
	})

	// Scenario 3: Database operations
	logger.InfoWithFields("Database query executed", map[string]interface{}{
		"query":    "SELECT * FROM users WHERE id = ?",
		"duration": "15ms",
		"rows":     1,
	})

	// Scenario 4: API requests
	logger.InfoWithFields("API request processed", map[string]interface{}{
		"method":   "GET",
		"path":     "/api/users",
		"status":   200,
		"duration": "45ms",
	})

	// Scenario 5: Warning messages
	logger.WarningWithFields("High memory usage detected", map[string]interface{}{
		"memory_usage": "85%",
		"threshold":    "80%",
	})

	// Scenario 6: Error handling
	logger.ErrorWithFields("Database connection failed", map[string]interface{}{
		"error":   "connection timeout",
		"retries": 3,
		"host":    "db.example.com",
	})

	// Scenario 7: Performance metrics
	logger.InfoWithFields("Performance metrics", map[string]interface{}{
		"cpu_usage":    "45%",
		"memory_usage": "60%",
		"disk_usage":   "30%",
		"active_users": 150,
	})

	// Scenario 8: Business events
	logger.InfoWithFields("Order placed", map[string]interface{}{
		"order_id": "ORD-12345",
		"user_id":  "user123",
		"amount":   99.99,
		"currency": "USD",
		"items":    3,
	})

	// Scenario 9: System events
	logger.InfoWithFields("Backup completed", map[string]interface{}{
		"backup_id": "backup-2024-01-15",
		"size":      "2.5GB",
		"duration":  "15m30s",
		"status":    "success",
	})

	// Scenario 10: Security events
	logger.WarningWithFields("Failed login attempt", map[string]interface{}{
		"username": "admin",
		"ip":       "192.168.1.200",
		"reason":   "invalid password",
		"attempts": 5,
	})

	// Wait a bit to see rotation in action
	fmt.Println("\nWaiting for log rotation...")
	time.Sleep(2 * time.Minute)

	// Continue with more logs to trigger rotation
	for i := 1; i <= 20; i++ {
		logger.InfoWithFields("Additional log entry", map[string]interface{}{
			"iteration": i,
			"timestamp": time.Now().Unix(),
		})
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for cleanup to run
	fmt.Println("\nWaiting for cleanup operations...")
	time.Sleep(1 * time.Minute)
} 