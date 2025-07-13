package pim

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestFilterHook(t *testing.T) {
	// Test basic filtering by level
	hook := NewFilterHook(FilterConfig{
		HookConfig: HookConfig{
			Type:     HookTypeFilter,
			Name:     "test_filter",
			Enabled:  true,
			Priority: 10,
		},
		Levels: []LogLevel{DebugLevel, TraceLevel},
	})

	entry := CoreLogEntry{
		Level:   DebugLevel,
		Message: "Debug message",
	}

	// Should be filtered out
	result, err := hook.Process(entry)
	if err == nil {
		t.Error("Expected error for filtered entry")
	}
	if !strings.Contains(err.Error(), "filtered by hook") {
		t.Errorf("Expected filter error, got: %v", err)
	}

	// Should not be filtered
	entry.Level = InfoLevel
	result, err = hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Message != "Debug message" {
		t.Errorf("Expected message to be unchanged, got: %s", result.Message)
	}
}

func TestFilterHookByMessage(t *testing.T) {
	hook := NewFilterHook(FilterConfig{
		HookConfig: HookConfig{
			Type:     HookTypeFilter,
			Name:     "message_filter",
			Enabled:  true,
			Priority: 10,
		},
		Messages: []string{"health check", "ping"},
	})

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "Health check completed",
	}

	// Should be filtered out
	_, err := hook.Process(entry)
	if err == nil {
		t.Error("Expected error for filtered entry")
	}

	// Should not be filtered
	entry.Message = "User login successful"
	_, err = hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFilterHookByContext(t *testing.T) {
	hook := NewFilterHook(FilterConfig{
		HookConfig: HookConfig{
			Type:     HookTypeFilter,
			Name:     "context_filter",
			Enabled:  true,
			Priority: 10,
		},
		Conditions: map[string]interface{}{
			"user_id": 12345,
			"status":  "active",
		},
	})

	entry := CoreLogEntry{
		Level: InfoLevel,
		Context: map[string]interface{}{
			"user_id": 12345,
			"status":  "active",
		},
	}

	// Should not be filtered (matches conditions)
	_, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should be filtered (doesn't match conditions)
	entry.Context["status"] = "inactive"
	_, err = hook.Process(entry)
	if err == nil {
		t.Error("Expected error for filtered entry")
	}
}

func TestRedactHook(t *testing.T) {
	hook := NewRedactHook(RedactConfig{
		HookConfig: HookConfig{
			Type:     HookTypeRedact,
			Name:     "test_redact",
			Enabled:  true,
			Priority: 10,
		},
		Fields:      []string{"password", "token"},
		Replacement: "[REDACTED]",
	})

	entry := CoreLogEntry{
		Level: InfoLevel,
		Context: map[string]interface{}{
			"user_id":  12345,
			"password": "secret123",
			"token":    "abc123",
			"email":    "user@example.com",
		},
	}

	result, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that sensitive fields are redacted
	if result.Context["password"] != "[REDACTED]" {
		t.Errorf("Expected password to be redacted, got: %v", result.Context["password"])
	}
	if result.Context["token"] != "[REDACTED]" {
		t.Errorf("Expected token to be redacted, got: %v", result.Context["token"])
	}

	// Check that non-sensitive fields are unchanged
	if result.Context["user_id"] != 12345 {
		t.Errorf("Expected user_id to be unchanged, got: %v", result.Context["user_id"])
	}
	if result.Context["email"] != "user@example.com" {
		t.Errorf("Expected email to be unchanged, got: %v", result.Context["email"])
	}
}

func TestRedactHookWithPatterns(t *testing.T) {
	hook := NewRedactHook(RedactConfig{
		HookConfig: HookConfig{
			Type:     HookTypeRedact,
			Name:     "pattern_redact",
			Enabled:  true,
			Priority: 10,
		},
		Patterns: map[string]string{
			"email": `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
		},
		Replacement: "[EMAIL]",
	})

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "User email: user@example.com",
		Context: map[string]interface{}{
			"email": "admin@company.com",
		},
	}

	result, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that email in message is redacted
	if !strings.Contains(result.Message, "[EMAIL]") {
		t.Errorf("Expected email in message to be redacted, got: %s", result.Message)
	}

	// Check that email in context is redacted
	if result.Context["email"] != "[EMAIL]" {
		t.Errorf("Expected email in context to be redacted, got: %v", result.Context["email"])
	}
}

func TestEnrichHook(t *testing.T) {
	hook := NewEnrichHook(EnrichConfig{
		HookConfig: HookConfig{
			Type:     HookTypeEnrich,
			Name:     "test_enrich",
			Enabled:  true,
			Priority: 10,
		},
		Fields: map[string]interface{}{
			"environment": "test",
			"version":     "1.0.0",
		},
	})

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "Test message",
		Context: map[string]interface{}{
			"user_id": 12345,
		},
	}

	result, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that static fields are added
	if result.Context["environment"] != "test" {
		t.Errorf("Expected environment to be added, got: %v", result.Context["environment"])
	}
	if result.Context["version"] != "1.0.0" {
		t.Errorf("Expected version to be added, got: %v", result.Context["version"])
	}

	// Check that existing fields are preserved
	if result.Context["user_id"] != 12345 {
		t.Errorf("Expected user_id to be preserved, got: %v", result.Context["user_id"])
	}
}

func TestEnrichHookDynamic(t *testing.T) {
	hook := NewEnrichHook(EnrichConfig{
		HookConfig: HookConfig{
			Type:     HookTypeEnrich,
			Name:     "dynamic_enrich",
			Enabled:  true,
			Priority: 10,
		},
		DynamicFunc: func(entry CoreLogEntry) map[string]interface{} {
			return map[string]interface{}{
				"timestamp_unix": entry.Timestamp.Unix(),
				"level_numeric":  int(entry.Level),
			}
		},
	})

	entry := CoreLogEntry{
		Level:     InfoLevel,
		Message:   "Test message",
		Timestamp: time.Now(),
	}

	result, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that dynamic fields are added
	if result.Context["timestamp_unix"] == nil {
		t.Error("Expected timestamp_unix to be added")
	}
	if result.Context["level_numeric"] != int(InfoLevel) {
		t.Errorf("Expected level_numeric to be %d, got: %v", int(InfoLevel), result.Context["level_numeric"])
	}
}

func TestTransformHook(t *testing.T) {
	hook := NewTransformHook(TransformConfig{
		HookConfig: HookConfig{
			Type:     HookTypeTransform,
			Name:     "test_transform",
			Enabled:  true,
			Priority: 10,
		},
		MessageFunc: func(message string) string {
			return "[TRANSFORMED] " + message
		},
	})

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "Original message",
	}

	result, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.Message != "[TRANSFORMED] Original message" {
		t.Errorf("Expected transformed message, got: %s", result.Message)
	}
}

func TestTransformHookLevel(t *testing.T) {
	hook := NewTransformHook(TransformConfig{
		HookConfig: HookConfig{
			Type:     HookTypeTransform,
			Name:     "level_transform",
			Enabled:  true,
			Priority: 10,
		},
		LevelFunc: func(level LogLevel) LogLevel {
			if level == InfoLevel {
				return WarningLevel
			}
			return level
		},
	})

	entry := CoreLogEntry{
		Level:       InfoLevel,
		LevelString: "INFO",
		Message:     "Test message",
	}

	result, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.Level != WarningLevel {
		t.Errorf("Expected level to be WarningLevel, got: %v", result.Level)
	}
	if result.LevelString != "WARNING" {
		t.Errorf("Expected level string to be WARNING, got: %s", result.LevelString)
	}
}

func TestMetricsHook(t *testing.T) {
	hook := NewMetricsHook(MetricsConfig{
		HookConfig: HookConfig{
			Type:     HookTypeMetrics,
			Name:     "test_metrics",
			Enabled:  true,
			Priority: 10,
		},
	})

	// Process some entries
	entries := []CoreLogEntry{
		{Level: InfoLevel, LevelString: "INFO", Message: "Info message"},
		{Level: WarningLevel, LevelString: "WARNING", Message: "Warning message"},
		{Level: ErrorLevel, LevelString: "ERROR", Message: "Error message"},
		{Level: InfoLevel, LevelString: "INFO", Message: "Another info message"},
	}

	for _, entry := range entries {
		_, err := hook.Process(entry)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Check metrics
	metrics := hook.GetMetrics()
	counters, ok := metrics["counters"].(map[string]int)
	if !ok {
		t.Fatal("Expected counters in metrics")
	}

	if counters["total"] != 4 {
		t.Errorf("Expected total count to be 4, got: %d", counters["total"])
	}
	if counters["level_info"] != 2 {
		t.Errorf("Expected info count to be 2, got: %d", counters["level_info"])
	}
	if counters["level_warning"] != 1 {
		t.Errorf("Expected warning count to be 1, got: %d", counters["level_warning"])
	}
	if counters["level_error"] != 1 {
		t.Errorf("Expected error count to be 1, got: %d", counters["level_error"])
	}
}

func TestHookManager(t *testing.T) {
	manager := NewHookManager()

	// Add hooks
	filterHook := NewFilterHook(FilterConfig{
		HookConfig: HookConfig{
			Type:     HookTypeFilter,
			Name:     "test_filter",
			Enabled:  true,
			Priority: 10,
		},
		Levels: []LogLevel{DebugLevel},
	})

	redactHook := NewRedactHook(RedactConfig{
		HookConfig: HookConfig{
			Type:     HookTypeRedact,
			Name:     "test_redact",
			Enabled:  true,
			Priority: 20,
		},
		Fields:      []string{"password"},
		Replacement: "[REDACTED]",
	})

	manager.AddHook(filterHook)
	manager.AddHook(redactHook)

	// Test processing
	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "Test message",
		Context: map[string]interface{}{
			"password": "secret123",
		},
	}

	result, err := manager.ProcessHooks(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that redaction worked
	if result.Context["password"] != "[REDACTED]" {
		t.Errorf("Expected password to be redacted, got: %v", result.Context["password"])
	}

	// Test filtering
	filterEntry := CoreLogEntry{
		Level:   DebugLevel,
		Message: "Debug message",
	}

	_, err = manager.ProcessHooks(filterEntry)
	if err == nil {
		t.Error("Expected error for filtered entry")
	}
}

func TestHookManagerPriority(t *testing.T) {
	manager := NewHookManager()

	// Add hooks with different priorities
	hook1 := NewEnrichHook(EnrichConfig{
		HookConfig: HookConfig{
			Type:     HookTypeEnrich,
			Name:     "hook1",
			Enabled:  true,
			Priority: 30, // Lower priority (runs first)
		},
		Fields: map[string]interface{}{
			"field1": "value1",
		},
	})

	hook2 := NewEnrichHook(EnrichConfig{
		HookConfig: HookConfig{
			Type:     HookTypeEnrich,
			Name:     "hook2",
			Enabled:  true,
			Priority: 10, // Higher priority (runs second)
		},
		Fields: map[string]interface{}{
			"field2": "value2",
		},
	})

	manager.AddHook(hook1)
	manager.AddHook(hook2)

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "Test message",
	}

	result, err := manager.ProcessHooks(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Both fields should be present
	if result.Context["field1"] != "value1" {
		t.Errorf("Expected field1 to be present, got: %v", result.Context["field1"])
	}
	if result.Context["field2"] != "value2" {
		t.Errorf("Expected field2 to be present, got: %v", result.Context["field2"])
	}
}

func TestHookManagerDisable(t *testing.T) {
	manager := NewHookManager()

	hook := NewFilterHook(FilterConfig{
		HookConfig: HookConfig{
			Type:     HookTypeFilter,
			Name:     "test_filter",
			Enabled:  true,
			Priority: 10,
		},
		Levels: []LogLevel{DebugLevel},
	})

	manager.AddHook(hook)

	// Test with enabled hook
	entry := CoreLogEntry{
		Level:   DebugLevel,
		Message: "Debug message",
	}

	_, err := manager.ProcessHooks(entry)
	if err == nil {
		t.Error("Expected error for filtered entry")
	}

	// Disable hook
	manager.SetEnabled(false)

	// Test with disabled hook manager
	_, err = manager.ProcessHooks(entry)
	if err != nil {
		t.Errorf("Unexpected error with disabled manager: %v", err)
	}

	// Re-enable hook manager
	manager.SetEnabled(true)

	// Test with re-enabled hook manager
	_, err = manager.ProcessHooks(entry)
	if err == nil {
		t.Error("Expected error for filtered entry after re-enabling")
	}
}

func TestLoggerCoreWithEnhancedHooks(t *testing.T) {
	config := LoggerConfig{
		Level:            InfoLevel,
		ServiceName:      "test-service",
		EnableConsole:    false, // Disable console for testing
		PropagateContext: true,
	}

	logger := NewLoggerCore(config)
	defer logger.Close()

	// Add enhanced hooks
	logger.AddSensitiveDataRedactHook()
	logger.AddMetricsHook()

	// Test logging with sensitive data
	logger.Info("User login", "user_id", 12345, "password", "secret123")

	// Check metrics
	metrics := logger.GetMetrics()
	counters, ok := metrics["counters"].(map[string]int)
	if !ok {
		t.Fatal("Expected counters in metrics")
	}

	if counters["total"] != 1 {
		t.Errorf("Expected total count to be 1, got: %d", counters["total"])
	}

	// Test hook management
	if logger.GetHookCount() < 2 {
		t.Errorf("Expected at least 2 hooks, got: %d", logger.GetHookCount())
	}

	// Test getting hooks by type
	redactHooks := logger.GetEnhancedHooksByType(HookTypeRedact)
	if len(redactHooks) == 0 {
		t.Error("Expected at least one redact hook")
	}

	metricsHooks := logger.GetEnhancedHooksByType(HookTypeMetrics)
	if len(metricsHooks) == 0 {
		t.Error("Expected at least one metrics hook")
	}
}

func TestCustomHookFunctions(t *testing.T) {
	// Test custom filter function
	customFilter := NewFilterHook(FilterConfig{
		HookConfig: HookConfig{
			Type:     HookTypeFilter,
			Name:     "custom_filter",
			Enabled:  true,
			Priority: 10,
		},
		CustomFunc: func(entry CoreLogEntry) bool {
			return strings.Contains(entry.Message, "secret")
		},
	})

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "This contains secret information",
	}

	_, err := customFilter.Process(entry)
	if err == nil {
		t.Error("Expected error for filtered entry")
	}

	entry.Message = "This is public information"
	_, err = customFilter.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test custom redaction function
	customRedact := NewRedactHook(RedactConfig{
		HookConfig: HookConfig{
			Type:     HookTypeRedact,
			Name:     "custom_redact",
			Enabled:  true,
			Priority: 10,
		},
		CustomFunc: func(entry CoreLogEntry) CoreLogEntry {
			if entry.Context != nil {
				if _, exists := entry.Context["sensitive"]; exists {
					entry.Context["sensitive"] = "[CUSTOM_REDACTED]"
				}
			}
			return entry
		},
	})

	entry = CoreLogEntry{
		Level: InfoLevel,
		Context: map[string]interface{}{
			"sensitive": "secret_value",
			"public":    "public_value",
		},
	}

	result, err := customRedact.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.Context["sensitive"] != "[CUSTOM_REDACTED]" {
		t.Errorf("Expected sensitive to be custom redacted, got: %v", result.Context["sensitive"])
	}
	if result.Context["public"] != "public_value" {
		t.Errorf("Expected public to be unchanged, got: %v", result.Context["public"])
	}
}

func TestHookErrorHandling(t *testing.T) {
	// Test hook that returns error
	errorHook := LogHookFunc(func(entry CoreLogEntry) (CoreLogEntry, error) {
		return CoreLogEntry{}, fmt.Errorf("hook error")
	})

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "Test message",
	}

	_, err := errorHook.Process(entry)
	if err == nil {
		t.Error("Expected error from hook")
	}
	if err.Error() != "hook error" {
		t.Errorf("Expected 'hook error', got: %v", err)
	}
}

func TestHookConfigValidation(t *testing.T) {
	// Test hook with invalid regex pattern
	hook := NewRedactHook(RedactConfig{
		HookConfig: HookConfig{
			Type:     HookTypeRedact,
			Name:     "invalid_regex",
			Enabled:  true,
			Priority: 10,
		},
		Patterns: map[string]string{
			"invalid": "[invalid regex",
		},
		Replacement: "[REDACTED]",
	})

	entry := CoreLogEntry{
		Level:   InfoLevel,
		Message: "Test message",
	}

	// Should not panic and should handle invalid regex gracefully
	result, err := hook.Process(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Message != "Test message" {
		t.Errorf("Expected message to be unchanged, got: %s", result.Message)
	}
}
