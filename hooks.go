package pim

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// HookResult represents the result of a hook processing
type HookResult struct {
	Entry    CoreLogEntry
	Filtered bool // If true, the entry should be filtered out
	Modified bool // If true, the entry was modified
	Error    error
}

// HookType defines the type of hook
type HookType int

const (
	HookTypeFilter HookType = iota
	HookTypeRedact
	HookTypeEnrich
	HookTypeTransform
	HookTypeMetrics
	HookTypeCustom
)

// HookConfig holds configuration for a hook
type HookConfig struct {
	Type        HookType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	Priority    int      `json:"priority"` // Lower numbers = higher priority
}

// FilterConfig holds configuration for filtering hooks
type FilterConfig struct {
	HookConfig
	Levels     []LogLevel                `json:"levels,omitempty"`     // Filter by log levels
	Messages   []string                  `json:"messages,omitempty"`   // Filter by message patterns
	Files      []string                  `json:"files,omitempty"`      // Filter by file patterns
	Functions  []string                  `json:"functions,omitempty"`  // Filter by function patterns
	Packages   []string                  `json:"packages,omitempty"`   // Filter by package patterns
	Conditions map[string]interface{}    `json:"conditions,omitempty"` // Filter by context conditions
	Regex      map[string]*regexp.Regexp `json:"-"`                    // Compiled regex patterns
	CustomFunc func(CoreLogEntry) bool   `json:"-"`                    // Custom filter function
}

// RedactConfig holds configuration for redaction hooks
type RedactConfig struct {
	HookConfig
	Fields      []string                        `json:"fields,omitempty"`   // Fields to redact
	Patterns    map[string]string               `json:"patterns,omitempty"` // Regex patterns to redact
	Replacement string                          `json:"replacement"`        // Replacement string
	CustomFunc  func(CoreLogEntry) CoreLogEntry `json:"-"`                  // Custom redaction function
}

// EnrichConfig holds configuration for enrichment hooks
type EnrichConfig struct {
	HookConfig
	Fields      map[string]interface{}                    `json:"fields,omitempty"` // Static fields to add
	DynamicFunc func(CoreLogEntry) map[string]interface{} `json:"-"`                // Dynamic fields function
	Context     context.Context                           `json:"-"`                // Context for dynamic enrichment
}

// TransformConfig holds configuration for transformation hooks
type TransformConfig struct {
	HookConfig
	MessageFunc func(string) string             `json:"-"` // Message transformation
	LevelFunc   func(LogLevel) LogLevel         `json:"-"` // Level transformation
	CustomFunc  func(CoreLogEntry) CoreLogEntry `json:"-"` // Custom transformation
}

// MetricsConfig holds configuration for metrics hooks
type MetricsConfig struct {
	HookConfig
	Counters   map[string]int           `json:"counters,omitempty"` // Counters to track
	Timers     map[string]time.Duration `json:"timers,omitempty"`   // Timers to track
	CustomFunc func(CoreLogEntry)       `json:"-"`                  // Custom metrics function
	mu         sync.RWMutex             `json:"-"`                  // Mutex for thread safety
}

// EnhancedLogHook extends the basic LogHook interface with additional capabilities
type EnhancedLogHook interface {
	LogHook
	GetConfig() HookConfig
	GetType() HookType
	IsEnabled() bool
	SetEnabled(enabled bool)
	GetPriority() int
	SetPriority(priority int)
}

// FilterHook implements filtering functionality
type FilterHook struct {
	config FilterConfig
}

// NewFilterHook creates a new filter hook
func NewFilterHook(config FilterConfig) *FilterHook {
	return &FilterHook{config: config}
}

// Process implements LogHook interface
func (h *FilterHook) Process(entry CoreLogEntry) (CoreLogEntry, error) {
	if !h.config.Enabled {
		return entry, nil
	}

	// Check if entry should be filtered
	if h.shouldFilter(entry) {
		return CoreLogEntry{}, fmt.Errorf("entry filtered by hook: %s", h.config.Name)
	}

	return entry, nil
}

// shouldFilter determines if an entry should be filtered
func (h *FilterHook) shouldFilter(entry CoreLogEntry) bool {
	// Check custom function first
	if h.config.CustomFunc != nil {
		return h.config.CustomFunc(entry)
	}

	// Check levels
	if len(h.config.Levels) > 0 {
		found := false
		for _, level := range h.config.Levels {
			if entry.Level == level {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check message patterns
	if len(h.config.Messages) > 0 {
		found := false
		for _, pattern := range h.config.Messages {
			if strings.Contains(entry.Message, pattern) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check file patterns
	if len(h.config.Files) > 0 {
		found := false
		for _, pattern := range h.config.Files {
			if strings.Contains(entry.File, pattern) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check function patterns
	if len(h.config.Functions) > 0 {
		found := false
		for _, pattern := range h.config.Functions {
			if strings.Contains(entry.Function, pattern) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check package patterns
	if len(h.config.Packages) > 0 {
		found := false
		for _, pattern := range h.config.Packages {
			if strings.Contains(entry.Package, pattern) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check context conditions
	if len(h.config.Conditions) > 0 && entry.Context != nil {
		for key, expectedValue := range h.config.Conditions {
			if actualValue, exists := entry.Context[key]; !exists || actualValue != expectedValue {
				return true
			}
		}
	}

	return false
}

// GetConfig implements EnhancedLogHook interface
func (h *FilterHook) GetConfig() HookConfig {
	return h.config.HookConfig
}

// GetType implements EnhancedLogHook interface
func (h *FilterHook) GetType() HookType {
	return h.config.Type
}

// IsEnabled implements EnhancedLogHook interface
func (h *FilterHook) IsEnabled() bool {
	return h.config.Enabled
}

// SetEnabled implements EnhancedLogHook interface
func (h *FilterHook) SetEnabled(enabled bool) {
	h.config.Enabled = enabled
}

// GetPriority implements EnhancedLogHook interface
func (h *FilterHook) GetPriority() int {
	return h.config.Priority
}

// SetPriority implements EnhancedLogHook interface
func (h *FilterHook) SetPriority(priority int) {
	h.config.Priority = priority
}

// RedactHook implements redaction functionality
type RedactHook struct {
	config RedactConfig
}

// NewRedactHook creates a new redaction hook
func NewRedactHook(config RedactConfig) *RedactHook {
	return &RedactHook{config: config}
}

// Process implements LogHook interface
func (h *RedactHook) Process(entry CoreLogEntry) (CoreLogEntry, error) {
	if !h.config.Enabled {
		return entry, nil
	}

	// Use custom function if provided
	if h.config.CustomFunc != nil {
		return h.config.CustomFunc(entry), nil
	}

	// Redact fields in context
	if entry.Context != nil {
		for _, field := range h.config.Fields {
			if _, exists := entry.Context[field]; exists {
				entry.Context[field] = h.config.Replacement
			}
		}
	}

	// Redact patterns in message
	if len(h.config.Patterns) > 0 {
		for field, pattern := range h.config.Patterns {
			if regex, err := regexp.Compile(pattern); err == nil {
				if field == "message" {
					entry.Message = regex.ReplaceAllString(entry.Message, h.config.Replacement)
				} else if entry.Context != nil {
					if value, exists := entry.Context[field]; exists {
						if str, ok := value.(string); ok {
							entry.Context[field] = regex.ReplaceAllString(str, h.config.Replacement)
						}
					}
				}
			}
		}
	}

	return entry, nil
}

// GetConfig implements EnhancedLogHook interface
func (h *RedactHook) GetConfig() HookConfig {
	return h.config.HookConfig
}

// GetType implements EnhancedLogHook interface
func (h *RedactHook) GetType() HookType {
	return h.config.Type
}

// IsEnabled implements EnhancedLogHook interface
func (h *RedactHook) IsEnabled() bool {
	return h.config.Enabled
}

// SetEnabled implements EnhancedLogHook interface
func (h *RedactHook) SetEnabled(enabled bool) {
	h.config.Enabled = enabled
}

// GetPriority implements EnhancedLogHook interface
func (h *RedactHook) GetPriority() int {
	return h.config.Priority
}

// SetPriority implements EnhancedLogHook interface
func (h *RedactHook) SetPriority(priority int) {
	h.config.Priority = priority
}

// EnrichHook implements enrichment functionality
type EnrichHook struct {
	config EnrichConfig
}

// NewEnrichHook creates a new enrichment hook
func NewEnrichHook(config EnrichConfig) *EnrichHook {
	return &EnrichHook{config: config}
}

// Process implements LogHook interface
func (h *EnrichHook) Process(entry CoreLogEntry) (CoreLogEntry, error) {
	if !h.config.Enabled {
		return entry, nil
	}

	// Initialize context if needed
	if entry.Context == nil {
		entry.Context = make(map[string]interface{})
	}

	// Add static fields
	for key, value := range h.config.Fields {
		entry.Context[key] = value
	}

	// Add dynamic fields
	if h.config.DynamicFunc != nil {
		dynamicFields := h.config.DynamicFunc(entry)
		for key, value := range dynamicFields {
			entry.Context[key] = value
		}
	}

	return entry, nil
}

// GetConfig implements EnhancedLogHook interface
func (h *EnrichHook) GetConfig() HookConfig {
	return h.config.HookConfig
}

// GetType implements EnhancedLogHook interface
func (h *EnrichHook) GetType() HookType {
	return h.config.Type
}

// IsEnabled implements EnhancedLogHook interface
func (h *EnrichHook) IsEnabled() bool {
	return h.config.Enabled
}

// SetEnabled implements EnhancedLogHook interface
func (h *EnrichHook) SetEnabled(enabled bool) {
	h.config.Enabled = enabled
}

// GetPriority implements EnhancedLogHook interface
func (h *EnrichHook) GetPriority() int {
	return h.config.Priority
}

// SetPriority implements EnhancedLogHook interface
func (h *EnrichHook) SetPriority(priority int) {
	h.config.Priority = priority
}

// TransformHook implements transformation functionality
type TransformHook struct {
	config TransformConfig
}

// NewTransformHook creates a new transformation hook
func NewTransformHook(config TransformConfig) *TransformHook {
	return &TransformHook{config: config}
}

// Process implements LogHook interface
func (t *TransformHook) Process(entry CoreLogEntry) (CoreLogEntry, error) {
	if !t.config.Enabled {
		return entry, nil
	}

	// Use custom function if provided
	if t.config.CustomFunc != nil {
		return t.config.CustomFunc(entry), nil
	}

	// Transform message
	if t.config.MessageFunc != nil {
		entry.Message = t.config.MessageFunc(entry.Message)
	}

	// Transform level
	if t.config.LevelFunc != nil {
		entry.Level = t.config.LevelFunc(entry.Level)
		entry.LevelString = getLevelString(entry.Level)
	}

	return entry, nil
}

// GetConfig implements EnhancedLogHook interface
func (t *TransformHook) GetConfig() HookConfig {
	return t.config.HookConfig
}

// GetType implements EnhancedLogHook interface
func (t *TransformHook) GetType() HookType {
	return t.config.Type
}

// IsEnabled implements EnhancedLogHook interface
func (t *TransformHook) IsEnabled() bool {
	return t.config.Enabled
}

// SetEnabled implements EnhancedLogHook interface
func (t *TransformHook) SetEnabled(enabled bool) {
	t.config.Enabled = enabled
}

// GetPriority implements EnhancedLogHook interface
func (t *TransformHook) GetPriority() int {
	return t.config.Priority
}

// SetPriority implements EnhancedLogHook interface
func (t *TransformHook) SetPriority(priority int) {
	t.config.Priority = priority
}

// MetricsHook implements metrics functionality
type MetricsHook struct {
	config MetricsConfig
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook(config MetricsConfig) *MetricsHook {
	if config.Counters == nil {
		config.Counters = make(map[string]int)
	}
	if config.Timers == nil {
		config.Timers = make(map[string]time.Duration)
	}
	return &MetricsHook{config: config}
}

// Process implements LogHook interface
func (m *MetricsHook) Process(entry CoreLogEntry) (CoreLogEntry, error) {
	if !m.config.Enabled {
		return entry, nil
	}

	m.config.mu.Lock()
	defer m.config.mu.Unlock()

	// Use custom function if provided
	if m.config.CustomFunc != nil {
		m.config.CustomFunc(entry)
		return entry, nil
	}

	// Count by level
	levelKey := fmt.Sprintf("level_%s", strings.ToLower(entry.LevelString))
	m.config.Counters[levelKey]++

	// Count total
	m.config.Counters["total"]++

	// Count by service
	if entry.ServiceName != "" {
		serviceKey := fmt.Sprintf("service_%s", entry.ServiceName)
		m.config.Counters[serviceKey]++
	}

	return entry, nil
}

// GetMetrics returns current metrics
func (m *MetricsHook) GetMetrics() map[string]interface{} {
	m.config.mu.RLock()
	defer m.config.mu.RUnlock()

	metrics := make(map[string]interface{})

	// Copy counters
	counters := make(map[string]int)
	for k, v := range m.config.Counters {
		counters[k] = v
	}
	metrics["counters"] = counters

	// Copy timers
	timers := make(map[string]time.Duration)
	for k, v := range m.config.Timers {
		timers[k] = v
	}
	metrics["timers"] = timers

	return metrics
}

// ResetMetrics resets all metrics
func (m *MetricsHook) ResetMetrics() {
	m.config.mu.Lock()
	defer m.config.mu.Unlock()

	m.config.Counters = make(map[string]int)
	m.config.Timers = make(map[string]time.Duration)
}

// GetConfig implements EnhancedLogHook interface
func (m *MetricsHook) GetConfig() HookConfig {
	return m.config.HookConfig
}

// GetType implements EnhancedLogHook interface
func (m *MetricsHook) GetType() HookType {
	return m.config.Type
}

// IsEnabled implements EnhancedLogHook interface
func (m *MetricsHook) IsEnabled() bool {
	return m.config.Enabled
}

// SetEnabled implements EnhancedLogHook interface
func (m *MetricsHook) SetEnabled(enabled bool) {
	m.config.Enabled = enabled
}

// GetPriority implements EnhancedLogHook interface
func (m *MetricsHook) GetPriority() int {
	return m.config.Priority
}

// SetPriority implements EnhancedLogHook interface
func (m *MetricsHook) SetPriority(priority int) {
	m.config.Priority = priority
}

// HookManager manages all hooks with priority ordering
type HookManager struct {
	hooks   []EnhancedLogHook
	mu      sync.RWMutex
	enabled bool
}

// NewHookManager creates a new hook manager
func NewHookManager() *HookManager {
	return &HookManager{
		hooks:   make([]EnhancedLogHook, 0),
		enabled: true,
	}
}

// AddHook adds a hook to the manager
func (hm *HookManager) AddHook(hook EnhancedLogHook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.hooks = append(hm.hooks, hook)
	hm.sortHooks() // Sort by priority
}

// RemoveHook removes a hook by name
func (hm *HookManager) RemoveHook(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for i, hook := range hm.hooks {
		if hook.GetConfig().Name == name {
			hm.hooks = append(hm.hooks[:i], hm.hooks[i+1:]...)
			break
		}
	}
}

// GetHook returns a hook by name
func (hm *HookManager) GetHook(name string) EnhancedLogHook {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	for _, hook := range hm.hooks {
		if hook.GetConfig().Name == name {
			return hook
		}
	}
	return nil
}

// GetHooksByType returns all hooks of a specific type
func (hm *HookManager) GetHooksByType(hookType HookType) []EnhancedLogHook {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var hooks []EnhancedLogHook
	for _, hook := range hm.hooks {
		if hook.GetType() == hookType && hook.IsEnabled() {
			hooks = append(hooks, hook)
		}
	}
	return hooks
}

// ProcessHooks processes all enabled hooks in priority order
func (hm *HookManager) ProcessHooks(entry CoreLogEntry) (CoreLogEntry, error) {
	if !hm.enabled {
		return entry, nil
	}

	hm.mu.RLock()
	hooks := make([]EnhancedLogHook, len(hm.hooks))
	copy(hooks, hm.hooks)
	hm.mu.RUnlock()

	for _, hook := range hooks {
		if hook.IsEnabled() {
			if modifiedEntry, err := hook.Process(entry); err != nil {
				// If error is a filter error, return empty entry
				if strings.Contains(err.Error(), "filtered by hook") {
					return CoreLogEntry{}, err
				}
				// For other errors, continue processing but log the error
				fmt.Fprintf(os.Stderr, "Hook error (%s): %v\n", hook.GetConfig().Name, err)
			} else {
				entry = modifiedEntry
			}
		}
	}

	return entry, nil
}

// sortHooks sorts hooks by priority (lower number = higher priority)
func (hm *HookManager) sortHooks() {
	// Simple bubble sort for small number of hooks
	for i := 0; i < len(hm.hooks)-1; i++ {
		for j := 0; j < len(hm.hooks)-i-1; j++ {
			if hm.hooks[j].GetPriority() > hm.hooks[j+1].GetPriority() {
				hm.hooks[j], hm.hooks[j+1] = hm.hooks[j+1], hm.hooks[j]
			}
		}
	}
}

// SetEnabled enables or disables the hook manager
func (hm *HookManager) SetEnabled(enabled bool) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.enabled = enabled
}

// IsEnabled returns whether the hook manager is enabled
func (hm *HookManager) IsEnabled() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.enabled
}

// GetHookCount returns the number of hooks
func (hm *HookManager) GetHookCount() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.hooks)
}

// Convenience functions for creating common hooks

// NewSensitiveDataRedactHook creates a hook to redact sensitive data
func NewSensitiveDataRedactHook() *RedactHook {
	return NewRedactHook(RedactConfig{
		HookConfig: HookConfig{
			Type:        HookTypeRedact,
			Name:        "sensitive_data_redact",
			Description: "Redacts sensitive data like passwords, tokens, and keys",
			Enabled:     true,
			Priority:    10,
		},
		Fields: []string{
			"password", "passwd", "pwd",
			"token", "access_token", "refresh_token",
			"key", "secret", "api_key", "private_key",
			"credit_card", "cc_number", "ssn",
		},
		Patterns: map[string]string{
			"password": `(?i)(password|passwd|pwd)\s*[:=]\s*\S+`,
			"token":    `(?i)(token|key|secret)\s*[:=]\s*\S+`,
			"email":    `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
		},
		Replacement: "[REDACTED]",
	})
}

// NewRequestIDEnrichHook creates a hook to add request IDs
func NewRequestIDEnrichHook() *EnrichHook {
	return NewEnrichHook(EnrichConfig{
		HookConfig: HookConfig{
			Type:        HookTypeEnrich,
			Name:        "request_id_enrich",
			Description: "Adds request ID to all log entries",
			Enabled:     true,
			Priority:    20,
		},
		DynamicFunc: func(entry CoreLogEntry) map[string]interface{} {
			// Generate a simple request ID if not present
			if entry.RequestID == "" {
				return map[string]interface{}{
					"request_id": fmt.Sprintf("req_%d", time.Now().UnixNano()),
				}
			}
			return nil
		},
	})
}

// NewDebugFilterHook creates a hook to filter debug messages in production
func NewDebugFilterHook() *FilterHook {
	return NewFilterHook(FilterConfig{
		HookConfig: HookConfig{
			Type:        HookTypeFilter,
			Name:        "debug_filter",
			Description: "Filters debug and trace messages in production",
			Enabled:     true,
			Priority:    5,
		},
		Levels: []LogLevel{DebugLevel, TraceLevel},
	})
}
