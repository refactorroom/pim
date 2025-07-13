package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/refactorroom/pim"
)

func main() {
	fmt.Println("=== Enhanced Hooks Demo ===\n")

	// Create logger with enhanced hooks
	config := pim.LoggerConfig{
		Level:            pim.InfoLevel,
		ServiceName:      "hooks-demo",
		EnableConsole:    true,
		EnableColors:     true,
		PropagateContext: true,
	}

	logger := pim.NewLoggerCore(config)
	defer logger.Close()

	// Example 1: Basic hooks demonstration
	fmt.Println("1. Basic Hooks Demonstration:")
	logger.AddSensitiveDataRedactHook()
	logger.AddRequestIDEnrichHook()
	logger.AddMetricsHook()

	logger.Info("User login attempt", "user_id", 12345, "password", "secret123", "token", "abc123")
	logger.Info("API request processed", "endpoint", "/api/users", "method", "POST", "status", 200)

	// Example 2: Custom filtering hooks
	fmt.Println("\n2. Custom Filtering Hooks:")

	// Filter by level
	levelFilter := pim.NewFilterHook(pim.FilterConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeFilter,
			Name:        "level_filter",
			Description: "Filter out debug messages",
			Enabled:     true,
			Priority:    5,
		},
		Levels: []pim.LogLevel{pim.DebugLevel, pim.TraceLevel},
	})
	logger.AddEnhancedHook(levelFilter)

	// Filter by message pattern
	messageFilter := pim.NewFilterHook(pim.FilterConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeFilter,
			Name:        "message_filter",
			Description: "Filter out health check messages",
			Enabled:     true,
			Priority:    10,
		},
		Messages: []string{"health check", "ping"},
	})
	logger.AddEnhancedHook(messageFilter)

	logger.Debug("This debug message will be filtered")
	logger.Info("Health check completed")
	logger.Info("User data processed")
	logger.Trace("This trace message will be filtered")

	// Example 3: Advanced redaction hooks
	fmt.Println("\n3. Advanced Redaction Hooks:")

	// Custom redaction hook
	customRedact := pim.NewRedactHook(pim.RedactConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeRedact,
			Name:        "custom_redact",
			Description: "Custom redaction patterns",
			Enabled:     true,
			Priority:    15,
		},
		Fields: []string{"ssn", "credit_card", "phone"},
		Patterns: map[string]string{
			"email":       `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			"phone":       `\b\d{3}-\d{3}-\d{4}\b`,
			"credit_card": `\b\d{4}-\d{4}-\d{4}-\d{4}\b`,
		},
		Replacement: "[REDACTED]",
	})
	logger.AddEnhancedHook(customRedact)

	logger.Info("User registration",
		"email", "user@example.com",
		"phone", "555-123-4567",
		"ssn", "123-45-6789",
		"credit_card", "1234-5678-9012-3456",
		"address", "123 Main St",
	)

	// Example 4: Enrichment hooks
	fmt.Println("\n4. Enrichment Hooks:")

	// Static enrichment
	staticEnrich := pim.NewEnrichHook(pim.EnrichConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeEnrich,
			Name:        "static_enrich",
			Description: "Add static fields to all logs",
			Enabled:     true,
			Priority:    20,
		},
		Fields: map[string]interface{}{
			"environment": "production",
			"version":     "1.0.0",
			"region":      "us-west-2",
		},
	})
	logger.AddEnhancedHook(staticEnrich)

	// Dynamic enrichment
	dynamicEnrich := pim.NewEnrichHook(pim.EnrichConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeEnrich,
			Name:        "dynamic_enrich",
			Description: "Add dynamic fields based on context",
			Enabled:     true,
			Priority:    25,
		},
		DynamicFunc: func(entry pim.CoreLogEntry) map[string]interface{} {
			fields := make(map[string]interface{})

			// Add timestamp in different formats
			fields["timestamp_unix"] = entry.Timestamp.Unix()
			fields["timestamp_rfc3339"] = entry.Timestamp.Format(time.RFC3339)

			// Add log level info
			fields["level_numeric"] = int(entry.Level)
			fields["is_error"] = entry.Level <= pim.ErrorLevel

			// Add caller info if available
			if entry.File != "" {
				fields["caller_file"] = entry.File
				fields["caller_line"] = entry.Line
			}

			return fields
		},
	})
	logger.AddEnhancedHook(dynamicEnrich)

	logger.Info("Application started")
	logger.Error("Database connection failed")

	// Example 5: Transformation hooks
	fmt.Println("\n5. Transformation Hooks:")

	// Message transformation
	messageTransform := pim.NewTransformHook(pim.TransformConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeTransform,
			Name:        "message_transform",
			Description: "Transform message content",
			Enabled:     true,
			Priority:    30,
		},
		MessageFunc: func(message string) string {
			// Add prefix for certain messages
			if strings.Contains(strings.ToLower(message), "error") {
				return "[ERROR] " + message
			}
			if strings.Contains(strings.ToLower(message), "warning") {
				return "[WARNING] " + message
			}
			return message
		},
	})
	logger.AddEnhancedHook(messageTransform)

	// Level transformation
	levelTransform := pim.NewTransformHook(pim.TransformConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeTransform,
			Name:        "level_transform",
			Description: "Transform log levels based on conditions",
			Enabled:     true,
			Priority:    35,
		},
		LevelFunc: func(level pim.LogLevel) pim.LogLevel {
			// Promote certain info messages to warnings
			if level == pim.InfoLevel {
				// This would be based on message content in a real implementation
				return pim.WarningLevel
			}
			return level
		},
	})
	logger.AddEnhancedHook(levelTransform)

	logger.Info("This message will be transformed")
	logger.Info("This is an error condition")
	logger.Warning("This warning will be processed")

	// Example 6: Custom hooks with complex logic
	fmt.Println("\n6. Custom Hooks with Complex Logic:")

	// Custom filter with complex conditions
	complexFilter := pim.NewFilterHook(pim.FilterConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeFilter,
			Name:        "complex_filter",
			Description: "Complex filtering logic",
			Enabled:     true,
			Priority:    40,
		},
		CustomFunc: func(entry pim.CoreLogEntry) bool {
			// Filter out logs from specific packages in production
			if strings.Contains(entry.Package, "debug") && entry.Level <= pim.DebugLevel {
				return true // Filter out
			}

			// Filter out logs with certain context values
			if entry.Context != nil {
				if userID, exists := entry.Context["user_id"]; exists {
					if id, ok := userID.(int); ok && id == 99999 {
						return true // Filter out test user
					}
				}
			}

			return false // Don't filter
		},
	})
	logger.AddEnhancedHook(complexFilter)

	// Custom redaction with complex patterns
	complexRedact := pim.NewRedactHook(pim.RedactConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeRedact,
			Name:        "complex_redact",
			Description: "Complex redaction patterns",
			Enabled:     true,
			Priority:    45,
		},
		CustomFunc: func(entry pim.CoreLogEntry) pim.CoreLogEntry {
			// Redact sensitive patterns in message
			patterns := map[string]string{
				`password\s*[:=]\s*\S+`: "[PASSWORD]",
				`token\s*[:=]\s*\S+`:    "[TOKEN]",
				`key\s*[:=]\s*\S+`:      "[KEY]",
			}

			for pattern, replacement := range patterns {
				if regex, err := regexp.Compile(pattern); err == nil {
					entry.Message = regex.ReplaceAllString(entry.Message, replacement)
				}
			}

			// Redact sensitive context fields
			if entry.Context != nil {
				sensitiveFields := []string{"password", "token", "secret", "key", "auth"}
				for _, field := range sensitiveFields {
					if _, exists := entry.Context[field]; exists {
						entry.Context[field] = "[REDACTED]"
					}
				}
			}

			return entry
		},
	})
	logger.AddEnhancedHook(complexRedact)

	logger.Info("User login: password=secret123 token=abc123")
	logger.Info("API call", "user_id", 99999, "password", "test123", "token", "xyz789")

	// Example 7: Metrics and monitoring
	fmt.Println("\n7. Metrics and Monitoring:")

	// Generate some logs to collect metrics
	for i := 0; i < 10; i++ {
		logger.Info("Processing request", "request_id", fmt.Sprintf("req_%d", i))
		if i%3 == 0 {
			logger.Warning("Rate limit approaching", "requests", i)
		}
		if i%5 == 0 {
			logger.Error("Database timeout", "attempt", i)
		}
	}

	// Display metrics
	metrics := logger.GetMetrics()
	fmt.Printf("\nCollected Metrics:\n")
	if counters, ok := metrics["counters"].(map[string]int); ok {
		for key, value := range counters {
			fmt.Printf("  %s: %d\n", key, value)
		}
	}

	// Example 8: Hook management
	fmt.Println("\n8. Hook Management:")

	// List all hooks
	fmt.Printf("Total hooks: %d\n", logger.GetHookCount())

	// Get hooks by type
	filterHooks := logger.GetEnhancedHooksByType(pim.HookTypeFilter)
	fmt.Printf("Filter hooks: %d\n", len(filterHooks))

	redactHooks := logger.GetEnhancedHooksByType(pim.HookTypeRedact)
	fmt.Printf("Redact hooks: %d\n", len(redactHooks))

	enrichHooks := logger.GetEnhancedHooksByType(pim.HookTypeEnrich)
	fmt.Printf("Enrich hooks: %d\n", len(enrichHooks))

	// Disable specific hook
	if hook := logger.GetEnhancedHook("message_filter"); hook != nil {
		hook.SetEnabled(false)
		fmt.Println("Disabled message_filter hook")
	}

	logger.Info("Health check completed") // This should now be logged

	// Re-enable hook
	if hook := logger.GetEnhancedHook("message_filter"); hook != nil {
		hook.SetEnabled(true)
		fmt.Println("Re-enabled message_filter hook")
	}

	logger.Info("Health check completed") // This should be filtered again

	// Example 9: Context-aware hooks
	fmt.Println("\n9. Context-Aware Hooks:")

	// Set global context
	logger.SetContext("deployment_id", "deploy_12345")
	logger.SetContext("service_version", "2.1.0")

	// Create context-aware enrichment hook
	contextEnrich := pim.NewEnrichHook(pim.EnrichConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeEnrich,
			Name:        "context_enrich",
			Description: "Enrich based on global context",
			Enabled:     true,
			Priority:    50,
		},
		DynamicFunc: func(entry pim.CoreLogEntry) map[string]interface{} {
			fields := make(map[string]interface{})

			// Add correlation ID if not present
			if entry.TraceID == "" {
				fields["correlation_id"] = fmt.Sprintf("corr_%d", time.Now().UnixNano())
			}

			// Add session info if user context exists
			if entry.Context != nil {
				if userID, exists := entry.Context["user_id"]; exists {
					fields["session_id"] = fmt.Sprintf("session_%v", userID)
				}
			}

			return fields
		},
	})
	logger.AddEnhancedHook(contextEnrich)

	logger.Info("User action", "user_id", 12345, "action", "profile_update")
	logger.Info("System event", "event", "backup_completed")

	// Example 10: Performance monitoring
	fmt.Println("\n10. Performance Monitoring:")

	// Create performance monitoring hook
	perfHook := pim.NewMetricsHook(pim.MetricsConfig{
		HookConfig: pim.HookConfig{
			Type:        pim.HookTypeMetrics,
			Name:        "performance_monitor",
			Description: "Monitor log performance",
			Enabled:     true,
			Priority:    100,
		},
		CustomFunc: func(entry pim.CoreLogEntry) {
			// In a real implementation, this would track timing, throughput, etc.
			fmt.Printf("  [PERF] Log entry processed: %s (level: %s)\n",
				entry.Message, entry.LevelString)
		},
	})
	logger.AddEnhancedHook(perfHook)

	logger.Info("Performance test message 1")
	logger.Warning("Performance test message 2")
	logger.Error("Performance test message 3")

	// Final summary
	fmt.Println("\n=== Hook Demo Summary ===")
	fmt.Printf("Total hooks configured: %d\n", logger.GetHookCount())
	fmt.Printf("Hook manager enabled: %t\n", logger.IsHookManagerEnabled())

	// Show final metrics
	finalMetrics := logger.GetMetrics()
	if counters, ok := finalMetrics["counters"].(map[string]int); ok {
		fmt.Printf("Total log entries processed: %d\n", counters["total"])
	}

	fmt.Println("\nDemo completed successfully!")
}
