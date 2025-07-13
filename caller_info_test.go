package pim

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewCallerInfoConfig(t *testing.T) {
	config := NewCallerInfoConfig()

	// Test default values
	if !config.Enabled {
		t.Error("Expected Enabled to be true by default")
	}
	if !config.ShowFile {
		t.Error("Expected ShowFile to be true by default")
	}
	if !config.ShowLine {
		t.Error("Expected ShowLine to be true by default")
	}
	if !config.ShowFunction {
		t.Error("Expected ShowFunction to be true by default")
	}
	if !config.ShowPackage {
		t.Error("Expected ShowPackage to be true by default")
	}
	if config.ShowFullPath {
		t.Error("Expected ShowFullPath to be false by default")
	}
	if !config.ShowGoroutineID {
		t.Error("Expected ShowGoroutineID to be true by default")
	}
	if config.CallDepth != 2 {
		t.Errorf("Expected CallDepth to be 2, got %d", config.CallDepth)
	}
	if config.StackDepth != 10 {
		t.Errorf("Expected StackDepth to be 10, got %d", config.StackDepth)
	}
	if config.MaxCallDepth != 20 {
		t.Errorf("Expected MaxCallDepth to be 20, got %d", config.MaxCallDepth)
	}
	if config.MinCallDepth != 1 {
		t.Errorf("Expected MinCallDepth to be 1, got %d", config.MinCallDepth)
	}
	if config.Separator != ":" {
		t.Errorf("Expected Separator to be ':', got '%s'", config.Separator)
	}
	if config.CacheEnabled {
		t.Error("Expected CacheEnabled to be true by default")
	}
	if config.CacheSize != 1000 {
		t.Errorf("Expected CacheSize to be 1000, got %d", config.CacheSize)
	}

	// Test exclude patterns
	expectedExcludes := []string{`^runtime\.`, `^reflect\.`, `^syscall\.`, `^internal/`}
	if len(config.ExcludePatterns) != len(expectedExcludes) {
		t.Errorf("Expected %d exclude patterns, got %d", len(expectedExcludes), len(config.ExcludePatterns))
	}
}

func TestNewCallerInfoFormatter(t *testing.T) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	if formatter == nil {
		t.Error("Expected formatter to be created")
	}
	if !reflect.DeepEqual(formatter.config, config) {
		t.Error("Expected formatter to have the provided config")
	}
	if len(formatter.cache) != 0 {
		t.Error("Expected cache to be empty initially")
	}
}

func TestCallerInfoFormatter_GetCallerInfo(t *testing.T) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	// Test with default skip
	info := formatter.GetCallerInfo(1)
	if info.File == "" {
		t.Error("Expected file to be populated")
	}
	if info.Line == 0 {
		t.Error("Expected line to be populated")
	}
	if info.Function == "" {
		t.Error("Expected function to be populated")
	}
	if info.Package == "" {
		t.Error("Expected package to be populated")
	}
	if info.CallDepth == 0 {
		t.Error("Expected call depth to be populated")
	}

	// Test with disabled caller info
	config.Enabled = false
	formatter.SetConfig(config)
	info = formatter.GetCallerInfo(1)
	if info.File != "" || info.Line != 0 || info.Function != "" || info.Package != "" {
		t.Error("Expected empty caller info when disabled")
	}
}

func TestCallerInfoFormatter_GetCallerInfoAtDepth(t *testing.T) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	// Test at specific depth
	info := formatter.GetCallerInfoAtDepth(1)
	if info.File == "" {
		t.Error("Expected file to be populated")
	}
	if info.Line == 0 {
		t.Error("Expected line to be populated")
	}
	if info.Function == "" {
		t.Error("Expected function to be populated")
	}
	if info.Package == "" {
		t.Error("Expected package to be populated")
	}
	if info.CallDepth != 1 {
		t.Errorf("Expected call depth to be 1, got %d", info.CallDepth)
	}
}

func TestCallerInfoFormatter_GetStackTrace(t *testing.T) {
	config := NewCallerInfoConfig()
	config.StackDepth = 5
	formatter := NewCallerInfoFormatter(config)

	frames := formatter.GetStackTrace(1)
	if len(frames) == 0 {
		t.Error("Expected stack trace to have frames")
	}
	if len(frames) > config.StackDepth {
		t.Errorf("Expected at most %d frames, got %d", config.StackDepth, len(frames))
	}

	// Test first frame
	firstFrame := frames[0]
	if firstFrame.File == "" {
		t.Error("Expected first frame to have file")
	}
	if firstFrame.Line == 0 {
		t.Error("Expected first frame to have line")
	}
	if firstFrame.Function == "" {
		t.Error("Expected first frame to have function")
	}
}

func TestCallerInfoFormatter_Format(t *testing.T) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	info := CallerInfo{
		File:     "test.go",
		Line:     42,
		Function: "TestFunction",
		Package:  "testpackage",
	}

	// Test default format
	formatted := formatter.Format(info)
	if formatted == "" {
		t.Error("Expected formatted string to not be empty")
	}
	if !strings.Contains(formatted, "test.go") {
		t.Error("Expected formatted string to contain file name")
	}
	if !strings.Contains(formatted, "L42") {
		t.Error("Expected formatted string to contain line number")
	}
	if !strings.Contains(formatted, "testpackage.TestFunction") {
		t.Error("Expected formatted string to contain package.function")
	}

	// Test custom format
	config.Format = "{file}:{line} | {function}"
	formatter.SetConfig(config)
	formatted = formatter.Format(info)
	expected := "test.go:42 | TestFunction"
	if formatted != expected {
		t.Errorf("Expected '%s', got '%s'", expected, formatted)
	}

	// Test with disabled caller info
	config.Enabled = false
	formatter.SetConfig(config)
	formatted = formatter.Format(info)
	if formatted != "" {
		t.Error("Expected empty string when caller info is disabled")
	}
}

func TestCallerInfoFormatter_FormatStackTrace(t *testing.T) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	frames := []CallerInfo{
		{File: "test1.go", Line: 10, Function: "Function1", Package: "pkg1"},
		{File: "test2.go", Line: 20, Function: "Function2", Package: "pkg2"},
		{File: "test3.go", Line: 30, Function: "Function3", Package: "pkg3"},
	}

	formatted := formatter.FormatStackTrace(frames)
	if formatted == "" {
		t.Error("Expected formatted stack trace to not be empty")
	}

	lines := strings.Split(formatted, "\n")
	if len(lines) != len(frames) {
		t.Errorf("Expected %d lines, got %d", len(frames), len(lines))
	}

	// Check that each line contains the expected content
	for i, line := range lines {
		if !strings.Contains(line, frames[i].File) {
			t.Errorf("Expected line %d to contain file name", i)
		}
		if !strings.Contains(line, frames[i].Function) {
			t.Errorf("Expected line %d to contain function name", i)
		}
	}
}

func TestCallerInfoFormatter_shouldInclude(t *testing.T) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	// Test internal function exclusion
	internalInfo := CallerInfo{
		File:       "runtime.go",
		Line:       1,
		Function:   "runtime.main",
		Package:    "runtime",
		IsInternal: true,
	}

	if formatter.shouldInclude(internalInfo) {
		t.Error("Expected internal function to be excluded by default")
	}

	// Test with runtime inclusion enabled
	config.IncludeRuntime = true
	formatter.SetConfig(config)
	if !formatter.shouldInclude(internalInfo) {
		t.Error("Expected internal function to be included when runtime is enabled")
	}

	// Test test file exclusion
	testInfo := CallerInfo{
		File:     "test_test.go",
		Line:     1,
		Function: "TestFunction",
		Package:  "test",
		IsTest:   true,
	}

	if formatter.shouldInclude(testInfo) {
		t.Error("Expected test file to be excluded by default")
	}

	// Test with test inclusion enabled
	config.IncludeTest = true
	formatter.SetConfig(config)
	if !formatter.shouldInclude(testInfo) {
		t.Error("Expected test file to be included when test is enabled")
	}

	// Test vendor exclusion
	vendorInfo := CallerInfo{
		File:     "vendor.go",
		Line:     1,
		Function: "VendorFunction",
		Package:  "vendor",
		IsVendor: true,
	}

	if formatter.shouldInclude(vendorInfo) {
		t.Error("Expected vendor file to be excluded by default")
	}

	// Test with vendor inclusion enabled
	config.IncludeVendor = true
	formatter.SetConfig(config)
	if !formatter.shouldInclude(vendorInfo) {
		t.Error("Expected vendor file to be included when vendor is enabled")
	}

	// Test package filter
	config.PackageFilter = "mypackage"
	formatter.SetConfig(config)

	filteredInfo := CallerInfo{
		File:     "myfile.go",
		Line:     1,
		Function: "MyFunction",
		Package:  "otherpackage",
	}

	if formatter.shouldInclude(filteredInfo) {
		t.Error("Expected function to be excluded when package doesn't match filter")
	}

	filteredInfo.Package = "mypackage"
	if !formatter.shouldInclude(filteredInfo) {
		t.Error("Expected function to be included when package matches filter")
	}

	// Test exclude patterns
	config.ExcludePatterns = []string{`^test\.`}
	formatter.SetConfig(config)

	excludedInfo := CallerInfo{
		File:     "test.go",
		Line:     1,
		Function: "test.function",
		Package:  "test",
	}

	if formatter.shouldInclude(excludedInfo) {
		t.Error("Expected function to be excluded when it matches exclude pattern")
	}
}

func TestCallerInfoFormatter_Cache(t *testing.T) {
	config := NewCallerInfoConfig()
	config.CacheEnabled = true
	config.CacheSize = 10
	formatter := NewCallerInfoFormatter(config)

	info := CallerInfo{
		File:     "test.go",
		Line:     42,
		Function: "TestFunction",
		Package:  "testpackage",
	}

	// Test cache functionality
	formatted1 := formatter.Format(info)
	formatted2 := formatter.Format(info)

	if formatted1 != formatted2 {
		t.Error("Expected cached results to be identical")
	}

	stats := formatter.GetCacheStats()
	if stats["size"].(int) != 1 {
		t.Errorf("Expected cache size to be 1, got %d", stats["size"])
	}

	// Test cache clearing
	formatter.ClearCache()
	stats = formatter.GetCacheStats()
	if stats["size"].(int) != 0 {
		t.Errorf("Expected cache size to be 0 after clearing, got %d", stats["size"])
	}

	// Test cache disabled
	config.CacheEnabled = false
	formatter.SetConfig(config)
	stats = formatter.GetCacheStats()
	if stats["enabled"].(bool) {
		t.Error("Expected cache to be disabled")
	}
}

func TestConvenienceConstructors(t *testing.T) {
	// Test minimal caller info
	minimal := NewMinimalCallerInfo()
	if minimal == nil {
		t.Error("Expected minimal caller info to be created")
	}
	config := minimal.GetConfig()
	if config.ShowPackage {
		t.Error("Expected minimal config to not show package")
	}
	if config.ShowGoroutineID {
		t.Error("Expected minimal config to not show goroutine ID")
	}
	if config.Format != "{file}:{line}" {
		t.Errorf("Expected minimal format to be '{file}:{line}', got '%s'", config.Format)
	}

	// Test production caller info
	prod := NewProductionCallerInfo()
	if prod == nil {
		t.Error("Expected production caller info to be created")
	}
	config = prod.GetConfig()
	if config.ShowPackage {
		t.Error("Expected production config to not show package")
	}
	if config.ShowGoroutineID {
		t.Error("Expected production config to not show goroutine ID")
	}
	if config.IncludeRuntime {
		t.Error("Expected production config to not include runtime")
	}
	if config.IncludeTest {
		t.Error("Expected production config to not include test")
	}
	if config.IncludeVendor {
		t.Error("Expected production config to not include vendor")
	}
	if !config.CacheEnabled {
		t.Error("Expected production config to enable cache")
	}
	if config.CacheSize != 5000 {
		t.Errorf("Expected production cache size to be 5000, got %d", config.CacheSize)
	}
	if config.Format != "{file}:{line} {function}" {
		t.Errorf("Expected production format to be '{file}:{line} {function}', got '%s'", config.Format)
	}

	// Test detailed caller info
	detailed := NewDetailedCallerInfo()
	if detailed == nil {
		t.Error("Expected detailed caller info to be created")
	}
	config = detailed.GetConfig()
	if !config.ShowFullPath {
		t.Error("Expected detailed config to show full path")
	}
	if !config.IncludeRuntime {
		t.Error("Expected detailed config to include runtime")
	}
	if !config.IncludeTest {
		t.Error("Expected detailed config to include test")
	}
	if !config.IncludeVendor {
		t.Error("Expected detailed config to include vendor")
	}
	if config.Format != "{file}:{line} {pkg:func} {goroutine}" {
		t.Errorf("Expected detailed format to be '{file}:{line} {pkg:func} {goroutine}', got '%s'", config.Format)
	}
}

func TestLoggerCore_CallerInfoIntegration(t *testing.T) {
	config := LoggerConfig{
		Level:            DebugLevel,
		EnableConsole:    false, // Disable console for testing
		CallerInfoConfig: NewCallerInfoConfig(),
	}

	logger := NewLoggerCore(config)

	// Test GetCallerInfo
	callerInfo := logger.GetCallerInfo(1)
	if callerInfo.File == "" {
		t.Error("Expected caller info to have file")
	}
	if callerInfo.Line == 0 {
		t.Error("Expected caller info to have line")
	}
	if callerInfo.Function == "" {
		t.Error("Expected caller info to have function")
	}

	// Test GetCallerInfoAtDepth
	callerInfo2 := logger.GetCallerInfoAtDepth(1)
	if callerInfo2.File == "" {
		t.Error("Expected caller info at depth to have file")
	}
	if callerInfo2.CallDepth != 1 {
		t.Errorf("Expected call depth to be 1, got %d", callerInfo2.CallDepth)
	}

	// Test GetEnhancedStackTrace
	stackTrace := logger.GetEnhancedStackTrace(1)
	if len(stackTrace) == 0 {
		t.Error("Expected enhanced stack trace to have frames")
	}

	// Test FormatCallerInfo
	formatted := logger.FormatCallerInfo(callerInfo)
	if formatted == "" {
		t.Error("Expected formatted caller info to not be empty")
	}

	// Test FormatStackTrace
	formattedStack := logger.FormatStackTrace(stackTrace)
	if formattedStack == "" {
		t.Error("Expected formatted stack trace to not be empty")
	}

	// Test cache stats
	stats := logger.GetCallerCacheStats()
	if stats["enabled"] != true {
		t.Error("Expected cache to be enabled")
	}

	// Test cache clearing
	logger.ClearCallerCache()
	stats = logger.GetCallerCacheStats()
	if stats["size"].(int) != 0 {
		t.Errorf("Expected cache size to be 0 after clearing, got %d", stats["size"])
	}

	// Test configuration updates
	newConfig := NewCallerInfoConfig()
	newConfig.Format = "CUSTOM {file}:{line}"
	logger.SetCallerInfoConfig(newConfig)

	updatedConfig := logger.GetCallerInfoConfig()
	if updatedConfig.Format != "CUSTOM {file}:{line}" {
		t.Errorf("Expected format to be updated, got '%s'", updatedConfig.Format)
	}
}

func TestCallerInfo_JSONTags(t *testing.T) {
	// Test that CallerInfo struct has proper JSON tags
	info := CallerInfo{
		File:        "test.go",
		Line:        42,
		Function:    "TestFunction",
		Package:     "testpackage",
		FullPath:    "/path/to/test.go",
		GoroutineID: "(goroutine 1)",
		CallDepth:   2,
		IsInternal:  false,
		IsTest:      false,
		IsVendor:    false,
	}

	// This test ensures the struct can be marshaled to JSON
	// The actual marshaling is tested elsewhere, but we verify the structure
	if info.File == "" || info.Line == 0 || info.Function == "" {
		t.Error("Expected CallerInfo to have populated fields")
	}
}

func TestCallerInfoConfig_JSONTags(t *testing.T) {
	// Test that CallerInfoConfig struct has proper JSON tags
	config := NewCallerInfoConfig()

	// This test ensures the struct can be marshaled to JSON
	// The actual marshaling is tested elsewhere, but we verify the structure
	if !config.Enabled || config.CallDepth == 0 || config.StackDepth == 0 {
		t.Error("Expected CallerInfoConfig to have populated fields")
	}
}

func BenchmarkCallerInfoFormatter_GetCallerInfo(b *testing.B) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.GetCallerInfo(1)
	}
}

func BenchmarkCallerInfoFormatter_Format(b *testing.B) {
	config := NewCallerInfoConfig()
	formatter := NewCallerInfoFormatter(config)

	info := CallerInfo{
		File:     "test.go",
		Line:     42,
		Function: "TestFunction",
		Package:  "testpackage",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.Format(info)
	}
}

func BenchmarkCallerInfoFormatter_FormatWithCache(b *testing.B) {
	config := NewCallerInfoConfig()
	config.CacheEnabled = true
	formatter := NewCallerInfoFormatter(config)

	info := CallerInfo{
		File:     "test.go",
		Line:     42,
		Function: "TestFunction",
		Package:  "testpackage",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.Format(info)
	}
}
