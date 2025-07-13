package pim

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// CallerInfoConfig defines configuration for caller information
type CallerInfoConfig struct {
	// Basic settings
	Enabled         bool `json:"enabled"`           // Enable/disable caller info
	ShowFile        bool `json:"show_file"`         // Show file name
	ShowLine        bool `json:"show_line"`         // Show line number
	ShowFunction    bool `json:"show_function"`     // Show function name
	ShowPackage     bool `json:"show_package"`      // Show package name
	ShowFullPath    bool `json:"show_full_path"`    // Show full file path
	ShowGoroutineID bool `json:"show_goroutine_id"` // Show goroutine ID

	// Depth control
	CallDepth    int `json:"call_depth"`     // Number of frames to skip (default: 2)
	StackDepth   int `json:"stack_depth"`    // Stack trace depth (default: 10)
	MaxCallDepth int `json:"max_call_depth"` // Maximum call depth to search
	MinCallDepth int `json:"min_call_depth"` // Minimum call depth to search

	// Formatting options
	Format         string `json:"format"`          // Custom format string
	Separator      string `json:"separator"`       // Separator between elements
	IncludeRuntime bool   `json:"include_runtime"` // Include runtime/internal calls
	IncludeTest    bool   `json:"include_test"`    // Include test files
	IncludeVendor  bool   `json:"include_vendor"`  // Include vendor packages

	// Filtering
	ExcludePatterns []string `json:"exclude_patterns"` // Regex patterns to exclude
	IncludePatterns []string `json:"include_patterns"` // Regex patterns to include
	PackageFilter   string   `json:"package_filter"`   // Package name filter

	// Performance
	CacheEnabled   bool `json:"cache_enabled"`   // Enable caching
	CacheSize      int  `json:"cache_size"`      // Cache size
	LazyEvaluation bool `json:"lazy_evaluation"` // Lazy evaluation of caller info
}

// CallerInfo contains detailed information about the calling function
type CallerInfo struct {
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Function    string `json:"function,omitempty"`
	Package     string `json:"package,omitempty"`
	FullPath    string `json:"full_path,omitempty"`
	GoroutineID string `json:"goroutine_id,omitempty"`
	CallDepth   int    `json:"call_depth,omitempty"`
	IsInternal  bool   `json:"is_internal,omitempty"`
	IsTest      bool   `json:"is_test,omitempty"`
	IsVendor    bool   `json:"is_vendor,omitempty"`
}

// CallerInfoFormatter handles formatting of caller information
type CallerInfoFormatter struct {
	config CallerInfoConfig
	cache  map[string]string // Simple cache for formatted strings
}

// NewCallerInfoConfig creates a new caller info configuration with sensible defaults
func NewCallerInfoConfig() CallerInfoConfig {
	return CallerInfoConfig{
		Enabled:         true,
		ShowFile:        true,
		ShowLine:        true,
		ShowFunction:    true,
		ShowPackage:     true,
		ShowFullPath:    false,
		ShowGoroutineID: true,
		CallDepth:       2,
		StackDepth:      10,
		MaxCallDepth:    20,
		MinCallDepth:    1,
		Format:          "",
		Separator:       ":",
		IncludeRuntime:  false,
		IncludeTest:     false,
		IncludeVendor:   false,
		CacheEnabled:    true,
		CacheSize:       1000,
		LazyEvaluation:  false,
		ExcludePatterns: []string{
			`^runtime\.`,
			`^reflect\.`,
			`^syscall\.`,
			`^internal/`,
		},
		IncludePatterns: []string{},
		PackageFilter:   "",
	}
}

// NewCallerInfoFormatter creates a new caller info formatter
func NewCallerInfoFormatter(config CallerInfoConfig) *CallerInfoFormatter {
	return &CallerInfoFormatter{
		config: config,
		cache:  make(map[string]string, config.CacheSize),
	}
}

// GetCallerInfo retrieves caller information with enhanced control
func (c *CallerInfoFormatter) GetCallerInfo(skip int) CallerInfo {
	if !c.config.Enabled {
		return CallerInfo{}
	}

	// Search for the first valid caller within the depth range
	for depth := skip + c.config.MinCallDepth; depth <= skip+c.config.MaxCallDepth; depth++ {
		pc, file, line, ok := runtime.Caller(depth)
		if !ok {
			continue
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		callerInfo := c.extractCallerInfo(pc, file, line, fn, depth)

		// Apply filters
		if c.shouldInclude(callerInfo) {
			return callerInfo
		}
	}

	return CallerInfo{}
}

// GetCallerInfoAtDepth retrieves caller information at a specific depth
func (c *CallerInfoFormatter) GetCallerInfoAtDepth(depth int) CallerInfo {
	if !c.config.Enabled {
		return CallerInfo{}
	}

	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		return CallerInfo{}
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return CallerInfo{}
	}

	return c.extractCallerInfo(pc, file, line, fn, depth)
}

// GetStackTrace retrieves a formatted stack trace with filtering
func (c *CallerInfoFormatter) GetStackTrace(skip int) []CallerInfo {
	if !c.config.Enabled {
		return []CallerInfo{}
	}

	var frames []CallerInfo
	for i := skip + c.config.MinCallDepth; i < skip+c.config.StackDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		callerInfo := c.extractCallerInfo(pc, file, line, fn, i)

		// Apply filters
		if c.shouldInclude(callerInfo) {
			frames = append(frames, callerInfo)
		}
	}

	return frames
}

// Format formats caller information according to configuration
func (c *CallerInfoFormatter) Format(info CallerInfo) string {
	if !c.config.Enabled {
		return ""
	}

	// Check cache first
	if c.config.CacheEnabled {
		cacheKey := fmt.Sprintf("%s:%d:%s:%s", info.File, info.Line, info.Function, info.Package)
		if cached, exists := c.cache[cacheKey]; exists {
			return cached
		}
	}

	var result string

	// Use custom format if provided
	if c.config.Format != "" {
		result = c.formatWithTemplate(info)
	} else {
		result = c.formatDefault(info)
	}

	// Cache the result
	if c.config.CacheEnabled && len(c.cache) < c.config.CacheSize {
		cacheKey := fmt.Sprintf("%s:%d:%s:%s", info.File, info.Line, info.Function, info.Package)
		c.cache[cacheKey] = result
	}

	return result
}

// FormatStackTrace formats a stack trace
func (c *CallerInfoFormatter) FormatStackTrace(frames []CallerInfo) string {
	if len(frames) == 0 {
		return ""
	}

	var lines []string
	for i, frame := range frames {
		indent := strings.Repeat("  ", i)
		formatted := c.Format(frame)
		line := fmt.Sprintf("%s↳ %s", indent, formatted)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// extractCallerInfo extracts detailed caller information
func (c *CallerInfoFormatter) extractCallerInfo(pc uintptr, file string, line int, fn *runtime.Func, depth int) CallerInfo {
	fullName := fn.Name()
	parts := strings.Split(fullName, ".")
	packageName := ""
	functionName := fullName

	if len(parts) > 1 {
		packageName = strings.Join(parts[:len(parts)-1], ".")
		functionName = parts[len(parts)-1]
	}

	// Determine file path
	fullPath := file
	if !c.config.ShowFullPath {
		file = filepath.Base(file)
	}

	// Get goroutine ID if enabled
	var goroutineID string
	if c.config.ShowGoroutineID {
		goroutineID = c.getGoroutineID()
	}

	// Determine file characteristics
	isInternal := strings.Contains(packageName, "runtime") || strings.Contains(packageName, "reflect") || strings.Contains(packageName, "syscall")
	isTest := strings.Contains(file, "_test.go") || strings.Contains(packageName, "test")
	isVendor := strings.Contains(fullPath, "/vendor/") || strings.Contains(fullPath, "\\vendor\\")

	return CallerInfo{
		File:        file,
		Line:        line,
		Function:    functionName,
		Package:     packageName,
		FullPath:    fullPath,
		GoroutineID: goroutineID,
		CallDepth:   depth,
		IsInternal:  isInternal,
		IsTest:      isTest,
		IsVendor:    isVendor,
	}
}

// shouldInclude determines if caller info should be included based on filters
func (c *CallerInfoFormatter) shouldInclude(info CallerInfo) bool {
	// Check runtime/internal exclusions
	if info.IsInternal && !c.config.IncludeRuntime {
		return false
	}

	// Check test exclusions
	if info.IsTest && !c.config.IncludeTest {
		return false
	}

	// Check vendor exclusions
	if info.IsVendor && !c.config.IncludeVendor {
		return false
	}

	// Check package filter
	if c.config.PackageFilter != "" && !strings.Contains(info.Package, c.config.PackageFilter) {
		return false
	}

	// Check exclude patterns
	for _, pattern := range c.config.ExcludePatterns {
		if matched, _ := regexp.MatchString(pattern, info.Package+"."+info.Function); matched {
			return false
		}
	}

	// Check include patterns (if any are specified, all must match)
	if len(c.config.IncludePatterns) > 0 {
		for _, pattern := range c.config.IncludePatterns {
			if matched, _ := regexp.MatchString(pattern, info.Package+"."+info.Function); !matched {
				return false
			}
		}
	}

	return true
}

// formatDefault formats caller info using default format
func (c *CallerInfoFormatter) formatDefault(info CallerInfo) string {
	var parts []string

	// Add file and line
	if c.config.ShowFile {
		parts = append(parts, info.File)
	}
	if c.config.ShowLine {
		parts = append(parts, fmt.Sprintf("L%d", info.Line))
	}

	// Add function and package
	if c.config.ShowFunction && info.Function != "" {
		if c.config.ShowPackage && info.Package != "" {
			parts = append(parts, fmt.Sprintf("%s.%s", info.Package, info.Function))
		} else {
			parts = append(parts, info.Function)
		}
	} else if c.config.ShowPackage && info.Package != "" {
		parts = append(parts, info.Package)
	}

	// Add goroutine ID
	if c.config.ShowGoroutineID && info.GoroutineID != "" {
		parts = append(parts, info.GoroutineID)
	}

	return strings.Join(parts, c.config.Separator)
}

// formatWithTemplate formats caller info using custom template
func (c *CallerInfoFormatter) formatWithTemplate(info CallerInfo) string {
	// Simple template replacement for now
	// Could be enhanced with a proper template engine
	result := c.config.Format

	// Replace placeholders
	replacements := map[string]string{
		"{file}":      info.File,
		"{line}":      fmt.Sprintf("%d", info.Line),
		"{function}":  info.Function,
		"{package}":   info.Package,
		"{fullpath}":  info.FullPath,
		"{goroutine}": info.GoroutineID,
		"{depth}":     fmt.Sprintf("%d", info.CallDepth),
		"{internal}":  fmt.Sprintf("%t", info.IsInternal),
		"{test}":      fmt.Sprintf("%t", info.IsTest),
		"{vendor}":    fmt.Sprintf("%t", info.IsVendor),
		"{file:line}": fmt.Sprintf("%s:%d", info.File, info.Line),
		"{pkg:func}":  fmt.Sprintf("%s.%s", info.Package, info.Function),
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// getGoroutineID returns the current goroutine ID
func (c *CallerInfoFormatter) getGoroutineID() string {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	id := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	return fmt.Sprintf("(goroutine %s)", id)
}

// ClearCache clears the formatter cache
func (c *CallerInfoFormatter) ClearCache() {
	c.cache = make(map[string]string, c.config.CacheSize)
}

// GetCacheStats returns cache statistics
func (c *CallerInfoFormatter) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"size":      len(c.cache),
		"max_size":  c.config.CacheSize,
		"enabled":   c.config.CacheEnabled,
		"hit_ratio": "N/A", // Could be enhanced with hit/miss tracking
	}
}

// SetConfig updates the formatter configuration
func (c *CallerInfoFormatter) SetConfig(config CallerInfoConfig) {
	c.config = config
	if c.config.CacheEnabled && len(c.cache) > c.config.CacheSize {
		c.ClearCache()
	}
}

// GetConfig returns the current configuration
func (c *CallerInfoFormatter) GetConfig() CallerInfoConfig {
	return c.config
}

// Convenience methods for common configurations

// NewMinimalCallerInfo creates a minimal caller info formatter
func NewMinimalCallerInfo() *CallerInfoFormatter {
	config := NewCallerInfoConfig()
	config.ShowPackage = false
	config.ShowGoroutineID = false
	config.Format = "{file}:{line}"
	return NewCallerInfoFormatter(config)
}

// NewDetailedCallerInfo creates a detailed caller info formatter
func NewDetailedCallerInfo() *CallerInfoFormatter {
	config := NewCallerInfoConfig()
	config.ShowFullPath = true
	config.IncludeRuntime = true
	config.IncludeTest = true
	config.IncludeVendor = true
	config.Format = "{file}:{line} {pkg:func} {goroutine}"
	return NewCallerInfoFormatter(config)
}

// NewProductionCallerInfo creates a production-optimized caller info formatter
func NewProductionCallerInfo() *CallerInfoFormatter {
	config := NewCallerInfoConfig()
	config.ShowPackage = false
	config.ShowGoroutineID = false
	config.IncludeRuntime = false
	config.IncludeTest = false
	config.IncludeVendor = false
	config.CacheEnabled = true
	config.CacheSize = 5000
	config.Format = "{file}:{line} {function}"
	return NewCallerInfoFormatter(config)
}
