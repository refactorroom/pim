package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"encoding/json"

	pim "github.com/refactorroom/pim"
)

// Example struct for JSON output
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// Example function to demonstrate function name tracking
func processUser(user User) error {
	pim.Info("Processing user", "user_id", user.ID, "name", user.Name)

	if user.ID <= 0 {
		return errors.New("invalid user ID")
	}

	pim.Success("User processed successfully", "user_id", user.ID)
	return nil
}

// Example function to demonstrate nested calls and stack traces
func handleUserRequest(userID int) error {
	pim.Debug("Handling user request", "user_id", userID)

	user := User{
		ID:       userID,
		Name:     "John Doe",
		Email:    "john@example.com",
		IsActive: true,
	}

	if err := processUser(user); err != nil {
		pim.Error("Failed to process user", "user_id", userID, "error", err)
		return err
	}

	return nil
}

// Example function to demonstrate performance metrics
func simulateDatabaseQuery() {
	pim.Metric("database_query_start", time.Now().UnixNano(), "operation", "select_users")

	// Simulate database query
	time.Sleep(100 * time.Millisecond)

	pim.Metric("database_query_end", time.Now().UnixNano(), "operation", "select_users")
	pim.Metric("database_query_duration", 100, "operation", "select_users", "unit", "ms")
}

// Example function to demonstrate different log levels
func demonstrateLogLevels() {
	pim.Trace("This is a trace message - most detailed debugging")
	pim.Debug("This is a debug message - detailed debugging info")
	pim.Info("This is an info message - general operational info")
	pim.Success("This is a success message - operation completed")
	pim.Warning("This is a warning message - potentially harmful situation")
	pim.Error("This is an error message - error condition occurred")

	// Uncomment the next line to see panic behavior
	// pim.Panic("This is a panic message - system is unusable")
}

// Example function to demonstrate JSON output
func demonstrateJSONOutput() {
	user := User{
		ID:       1,
		Name:     "Alice Johnson",
		Email:    "alice@example.com",
		IsActive: true,
	}

	pim.Info("User data:")
	pim.Json(user)

	// Demonstrate different JSON options
	pim.Json(user, pim.JsonOptions{Colored: false, Indent: 4})
}

// Example function to demonstrate key-value formatting
func demonstrateKeyValueFormatting() {
	data := map[string]interface{}{
		"request_id": "req_12345",
		"user_id":    123,
		"action":     "login",
		"timestamp":  time.Now().Format(time.RFC3339),
		"success":    true,
		"ip_address": "192.168.1.100",
	}

	pim.Info("Request data:")
	pim.KeyValue(data)

	pim.Info("Request data (inline):")
	pim.KeyValueInline(data)
}

// Example function to demonstrate call info and stack traces
func demonstrateCallInfo() {
	// Get current call information
	callInfo := pim.GetCallInfo()
	pim.Info("Current call info",
		"file", callInfo.File,
		"line", callInfo.Line,
		"function", callInfo.Function,
		"package", callInfo.Package,
	)

	// Get current stack trace
	stackTrace := pim.GetStackTrace()
	pim.Info("Current stack trace has", "frames", len(stackTrace))

	for i, frame := range stackTrace {
		pim.Debug("Stack frame",
			"index", i,
			"file", frame.File,
			"line", frame.Line,
			"function", frame.Function,
			"package", frame.Package,
		)
	}
}

// Example function to demonstrate configuration options
func demonstrateConfiguration() {
	pim.Info("=== Configuration Demo ===")

	// Show current configuration
	pim.Info("Current configuration:")
	pim.SetShowFunctionName(true)
	pim.SetShowPackageName(true)
	pim.SetShowFullPath(false)
	pim.SetStackDepth(5)

	pim.Info("Function names and package names enabled")

	// Disable function names
	pim.SetShowFunctionName(false)
	pim.Info("Function names disabled")

	// Re-enable function names
	pim.SetShowFunctionName(true)
	pim.Info("Function names re-enabled")

	// Show full paths
	pim.SetShowFullPath(true)
	pim.Info("Full paths enabled")

	// Disable full paths
	pim.SetShowFullPath(false)
	pim.Info("Full paths disabled")
}

// Custom callback function for log processing
func customLogCallback(entry pim.LogEntry, level pim.LogLevel) error {
	fmt.Printf("🔔 CALLBACK: Received log entry - Level: %s, Message: %s\n", entry.Level, entry.Message)
	// You can implement custom logic here like:
	// - Send to external monitoring service
	// - Store in custom database
	// - Trigger alerts
	// - Transform log data
	return nil
}

// Simple HTTP server to receive webhook logs
func startWebhookServer() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		fmt.Printf("🌐 WEBHOOK: Received %d logs from service: %s\n",
			int(payload["batch_size"].(float64)),
			payload["service_name"])

		// Process the logs
		if logs, ok := payload["logs"].([]interface{}); ok {
			for _, logEntry := range logs {
				if log, ok := logEntry.(map[string]interface{}); ok {
					fmt.Printf("  📝 Log: [%s] %s\n", log["level"], log["message"])
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "received"}`))
	})

	go func() {
		fmt.Println("🌐 Starting webhook server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()
}

func demonstrateEnhancedStamper() {
	fmt.Println("\n=== Enhanced Stamper Features Demo ===")

	// 1. Configure webhook
	webhookConfig := &pim.WebhookConfig{
		URL:        "http://localhost:8080/webhook",
		Method:     "POST",
		Headers:    map[string]string{"Authorization": "Bearer secret-token"},
		Timeout:    10 * time.Second,
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
		BatchSize:  5,
		BatchDelay: 2 * time.Second,
	}

	// 2. Set up stamper configuration
	stamperConfig := pim.StamperConfig{
		Enabled:     true,
		FileLogging: true,
		Webhook:     webhookConfig,
		Callback:    customLogCallback,
	}

	// 3. Configure the stamper
	pim.ConfigureStamper(stamperConfig)

	pim.Info("Enhanced stamper configured with webhook and callback")

	// 4. Demonstrate different log levels with webhook delivery
	pim.Info("This log will be sent to file, webhook, and callback")
	pim.Warning("Warning logs are also processed by all outputs")
	pim.Error("Error logs trigger all delivery methods")
	pim.Debug("Debug logs are batched and sent together")

	// 5. Demonstrate stamper control
	pim.Info("Disabling stamper temporarily...")
	pim.EnableStamper(false)
	pim.Info("This log won't be processed by stamper")
	pim.Warning("Neither will this warning")

	pim.Info("Re-enabling stamper...")
	pim.EnableStamper(true)
	pim.Info("Stamper is back online!")

	// 6. Demonstrate file logging control
	pim.Info("Disabling file logging...")
	pim.EnableFileLogging(false)
	pim.Info("This log will go to webhook and callback, but not to file")

	pim.Info("Re-enabling file logging...")
	pim.EnableFileLogging(true)
	pim.Info("File logging is back online!")

	// 7. Demonstrate webhook configuration changes
	pim.Info("Updating webhook configuration...")
	newWebhookConfig := &pim.WebhookConfig{
		URL:        "http://localhost:8080/webhook",
		Method:     "POST",
		Headers:    map[string]string{"Authorization": "Bearer new-token"},
		Timeout:    15 * time.Second,
		RetryCount: 5,
		RetryDelay: 2 * time.Second,
		BatchSize:  10,
		BatchDelay: 5 * time.Second,
	}
	pim.SetWebhookConfig(newWebhookConfig)
	pim.Info("Webhook configuration updated with new settings")

	// 8. Demonstrate callback changes
	pim.Info("Updating callback function...")
	pim.SetLogCallback(func(entry pim.LogEntry, level pim.LogLevel) error {
		fmt.Printf("🔄 NEW CALLBACK: [%s] %s\n", entry.Level, entry.Message)
		return nil
	})
	pim.Info("This log uses the new callback function")

	// 9. Wait for webhook batch to be sent
	pim.Info("Waiting for webhook batch to be sent...")
	time.Sleep(6 * time.Second)

	pim.Success("Enhanced stamper demo completed!")
}

func main() {
	// Initialize file logging
	err := pim.InitializeFileLogging(".", "pim-example")
	if err != nil {
		pim.Error("Failed to initialize logging", err)
		return
	}
	defer pim.CloseLogFiles()

	// Start webhook server for enhanced stamper demo
	startWebhookServer()

	pim.Init("Starting PIM example application")

	// Set log level to show all messages
	pim.SetLogLevel(pim.TraceLevel)

	pim.Info("=== PIM Enhanced Features Demo ===")

	// Demonstrate basic logging
	pim.Info("Basic logging demonstration")
	demonstrateLogLevels()

	// Demonstrate function tracking
	pim.Info("=== Function Tracking Demo ===")
	handleUserRequest(123)
	handleUserRequest(-1) // This will cause an error

	// Demonstrate JSON output
	pim.Info("=== JSON Output Demo ===")
	demonstrateJSONOutput()

	// Demonstrate key-value formatting
	pim.Info("=== Key-Value Formatting Demo ===")
	demonstrateKeyValueFormatting()

	// Demonstrate call info and stack traces
	pim.Info("=== Call Info and Stack Traces Demo ===")
	demonstrateCallInfo()

	// Demonstrate configuration options
	demonstrateConfiguration()

	// Demonstrate performance metrics
	pim.Info("=== Performance Metrics Demo ===")
	simulateDatabaseQuery()

	// Demonstrate enhanced stamper features
	demonstrateEnhancedStamper()

	// Demonstrate print functions
	pim.Info("=== Print Functions Demo ===")
	pim.Print("This is a print statement")
	pim.Println("This is a println statement")
	pim.Printf("This is a printf statement with %s", "formatting")

	// Demonstrate color utilities
	pim.Info("=== Color Utilities Demo ===")
	pim.Red.Println("This is red text")
	pim.Green.Println("This is green text")
	pim.Blue.Println("This is blue text")
	pim.Yellow.Println("This is yellow text")
	pim.Cyan.Println("This is cyan text")
	pim.Purple.Println("This is purple text")

	pim.Bold.Println("This is bold text")
	pim.Underline.Println("This is underlined text")

	pim.SuccessColor.Println("This is success colored text")
	pim.ErrorColor.Println("This is error colored text")
	pim.WarningColor.Println("This is warning colored text")
	pim.InfoColor.Println("This is info colored text")

	pim.Success("PIM example completed successfully!")

	fmt.Println("\n=== Log Files Generated ===")
	fmt.Println("Check the .log/pim/ directory for structured log files:")
	fmt.Println("- panic.jaeger.json")
	fmt.Println("- error.jaeger.json")
	fmt.Println("- warning.jaeger.json")
	fmt.Println("- info.jaeger.json")
	fmt.Println("- debug.jaeger.json")
	fmt.Println("- trace.jaeger.json")
}
