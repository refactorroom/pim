package pim

import (
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestLogLevelConstants(t *testing.T) {
	// Test that log levels have correct values
	if PanicLevel != 0 {
		t.Errorf("Expected PanicLevel to be 0, got %d", PanicLevel)
	}
	if ErrorLevel != 1 {
		t.Errorf("Expected ErrorLevel to be 1, got %d", ErrorLevel)
	}
	if WarningLevel != 2 {
		t.Errorf("Expected WarningLevel to be 2, got %d", WarningLevel)
	}
	if InfoLevel != 3 {
		t.Errorf("Expected InfoLevel to be 3, got %d", InfoLevel)
	}
	if DebugLevel != 4 {
		t.Errorf("Expected DebugLevel to be 4, got %d", DebugLevel)
	}
	if TraceLevel != 5 {
		t.Errorf("Expected TraceLevel to be 5, got %d", TraceLevel)
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are defined and not empty
	colorConstants := map[string]string{
		"ColorString":  ColorString,
		"ColorNumber":  ColorNumber,
		"ColorBool":    ColorBool,
		"ColorNull":    ColorNull,
		"ColorKey":     ColorKey,
		"ColorComment": ColorComment,
		"ColorReset":   ColorReset,
		"ColorBlack":   ColorBlack,
		"ColorRed":     ColorRed,
		"ColorGreen":   ColorGreen,
		"ColorYellow":  ColorYellow,
		"ColorBlue":    ColorBlue,
		"ColorPurple":  ColorPurple,
		"ColorCyan":    ColorCyan,
		"ColorWhite":   ColorWhite,
		"ColorGray":    ColorGray,
	}

	for name, value := range colorConstants {
		if value == "" {
			t.Errorf("Expected %s to be non-empty", name)
		}
		if !strings.HasPrefix(value, "\033[") {
			t.Errorf("Expected %s to start with ANSI escape sequence, got %s", name, value)
		}
	}
}

func TestColorVariables(t *testing.T) {
	// Test that color variables are initialized
	colors := []*color.Color{
		Black, Red, Green, Yellow, Blue, Purple, Cyan, White, Gray,
		HiBlack, HiRed, HiGreen, HiYellow, HiBlue, HiPurple, HiCyan, HiWhite,
		Bold, Underline, Italic,
		StringColor, NumberColor, BoolColor, NullColor, KeyColor, ErrorColor,
		SuccessColor, WarningColor, InfoColor, DebugColor, TraceColor,
		DateColor, IDColor, TagColor, URLColor,
	}

	for i, c := range colors {
		if c == nil {
			t.Errorf("Color variable at index %d is nil", i)
		}
	}
}

func TestLogPrefixes(t *testing.T) {
	prefixes := map[string]string{
		"InfoPrefix":    InfoPrefix,
		"SuccessPrefix": SuccessPrefix,
		"InitPrefix":    InitPrefix,
		"ConfigPrefix":  ConfigPrefix,
		"WarningPrefix": WarningPrefix,
		"ErrorPrefix":   ErrorPrefix,
		"DebugPrefix":   DebugPrefix,
		"TracePrefix":   TracePrefix,
		"PanicPrefix":   PanicPrefix,
		"MetricPrefix":  MetricPrefix,
		"JsonPrefix":    JsonPrefix,
		"DataPrefix":    DataPrefix,
		"ModelPrefix":   ModelPrefix,
	}

	for name, prefix := range prefixes {
		if prefix == "" {
			t.Errorf("Expected %s to be non-empty", name)
		}
		// Each prefix should contain some text
		if len(prefix) < 5 {
			t.Errorf("Expected %s to have reasonable length, got %d characters", name, len(prefix))
		}
	}
}

func TestColoredString(t *testing.T) {
	testText := "Hello, World!"
	result := ColoredString(testText, Red)

	if result == "" {
		t.Error("Expected ColoredString to return non-empty string")
	}

	// When colors are disabled, should return original text
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = originalNoColor }()

	result = ColoredString(testText, Red)
	if result != testText {
		t.Errorf("Expected ColoredString to return original text when colors disabled, got %s", result)
	}
}

func TestColoredFormat(t *testing.T) {
	result := ColoredFormat("Hello, %s!", Blue, "World")

	if result == "" {
		t.Error("Expected ColoredFormat to return non-empty string")
	}

	// When colors are disabled, should return formatted text
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = originalNoColor }()

	result = ColoredFormat("Hello, %s!", Blue, "World")
	if result != "Hello, World!" {
		t.Errorf("Expected ColoredFormat to return formatted text when colors disabled, got %s", result)
	}
}

func TestSetLogLevel(t *testing.T) {
	originalLevel := currentLogLevel
	defer func() { currentLogLevel = originalLevel }()

	// Test setting each log level
	levels := []LogLevel{PanicLevel, ErrorLevel, WarningLevel, InfoLevel, DebugLevel, TraceLevel}

	for _, level := range levels {
		SetLogLevel(level)
		if currentLogLevel != level {
			t.Errorf("Expected currentLogLevel to be %v, got %v", level, currentLogLevel)
		}
	}
}

func TestSetShowFileLine(t *testing.T) {
	originalShow := showFileLine
	defer func() { showFileLine = originalShow }()

	// Test enabling
	SetShowFileLine(true)
	if !showFileLine {
		t.Error("Expected showFileLine to be true after SetShowFileLine(true)")
	}

	// Test disabling
	SetShowFileLine(false)
	if showFileLine {
		t.Error("Expected showFileLine to be false after SetShowFileLine(false)")
	}
}

func TestSetShowGoroutineID(t *testing.T) {
	originalShow := showGoroutineID
	defer func() { showGoroutineID = originalShow }()

	// Test enabling
	SetShowGoroutineID(true)
	if !showGoroutineID {
		t.Error("Expected showGoroutineID to be true after SetShowGoroutineID(true)")
	}

	// Test disabling
	SetShowGoroutineID(false)
	if showGoroutineID {
		t.Error("Expected showGoroutineID to be false after SetShowGoroutineID(false)")
	}
}

func TestSetShowFunctionName(t *testing.T) {
	originalShow := showFunctionName
	defer func() { showFunctionName = originalShow }()

	// Test enabling
	SetShowFunctionName(true)
	if !showFunctionName {
		t.Error("Expected showFunctionName to be true after SetShowFunctionName(true)")
	}

	// Test disabling
	SetShowFunctionName(false)
	if showFunctionName {
		t.Error("Expected showFunctionName to be false after SetShowFunctionName(false)")
	}
}

func TestSetShowPackageName(t *testing.T) {
	originalShow := showPackageName
	defer func() { showPackageName = originalShow }()

	// Test enabling
	SetShowPackageName(true)
	if !showPackageName {
		t.Error("Expected showPackageName to be true after SetShowPackageName(true)")
	}

	// Test disabling
	SetShowPackageName(false)
	if showPackageName {
		t.Error("Expected showPackageName to be false after SetShowPackageName(false)")
	}
}

func TestSetShowFullPath(t *testing.T) {
	originalShow := showFullPath
	defer func() { showFullPath = originalShow }()

	// Test enabling
	SetShowFullPath(true)
	if !showFullPath {
		t.Error("Expected showFullPath to be true after SetShowFullPath(true)")
	}

	// Test disabling
	SetShowFullPath(false)
	if showFullPath {
		t.Error("Expected showFullPath to be false after SetShowFullPath(false)")
	}
}

func TestSetStackDepth(t *testing.T) {
	originalDepth := stackDepth
	defer func() { stackDepth = originalDepth }()

	// Test valid depths
	validDepths := []int{1, 5, 10}
	for _, depth := range validDepths {
		SetStackDepth(depth)
		if stackDepth != depth {
			t.Errorf("Expected stackDepth to be %d, got %d", depth, stackDepth)
		}
	}

	// Test invalid depths (should not change stackDepth)
	currentDepth := stackDepth
	invalidDepths := []int{0, -1, 11, 100}
	for _, depth := range invalidDepths {
		SetStackDepth(depth)
		if stackDepth != currentDepth {
			t.Errorf("Expected stackDepth to remain %d for invalid input %d, got %d", currentDepth, depth, stackDepth)
		}
	}
}

func TestSetCallerSkipFrames(t *testing.T) {
	originalSkip := callerSkipFrames
	defer func() { callerSkipFrames = originalSkip }()

	// Test valid skip values
	validSkips := []int{0, 3, 8, 15}
	for _, skip := range validSkips {
		SetCallerSkipFrames(skip)
		if callerSkipFrames != skip {
			t.Errorf("Expected callerSkipFrames to be %d, got %d", skip, callerSkipFrames)
		}

		// Test getter function
		if GetCallerSkipFrames() != skip {
			t.Errorf("Expected GetCallerSkipFrames() to return %d, got %d", skip, GetCallerSkipFrames())
		}
	}

	// Test invalid skip values (should not change callerSkipFrames)
	currentSkip := callerSkipFrames
	invalidSkips := []int{-1, 16, 100, -10}
	for _, skip := range invalidSkips {
		SetCallerSkipFrames(skip)
		if callerSkipFrames != currentSkip {
			t.Errorf("Expected callerSkipFrames to remain %d for invalid input %d, got %d", currentSkip, skip, callerSkipFrames)
		}
	}
}

func TestGetCallerSkipFrames(t *testing.T) {
	originalSkip := callerSkipFrames
	defer func() { callerSkipFrames = originalSkip }()

	// Test that getter returns current value
	testValue := 5
	callerSkipFrames = testValue

	result := GetCallerSkipFrames()
	if result != testValue {
		t.Errorf("Expected GetCallerSkipFrames() to return %d, got %d", testValue, result)
	}
}

func TestDefaultValues(t *testing.T) {
	// Test that default values are as expected
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"currentLogLevel", currentLogLevel, InfoLevel},
		{"showFileLine", showFileLine, true},
		{"showGoroutineID", showGoroutineID, false}, // Updated default
		{"showFunctionName", showFunctionName, true},
		{"stackDepth", stackDepth, 3},
		{"showPackageName", showPackageName, true},
		{"showFullPath", showFullPath, false},
		{"callerSkipFrames", callerSkipFrames, 3},
		{"enableFileLogging", enableFileLogging, false}, // File logging disabled by default
	}

	for _, test := range tests {
		if test.actual != test.expected {
			t.Errorf("Expected default %s to be %v, got %v", test.name, test.expected, test.actual)
		}
	}
}

func TestSetFileLogging(t *testing.T) {
	originalEnabled := enableFileLogging
	defer func() { enableFileLogging = originalEnabled }()

	// Test enabling
	SetFileLogging(true)
	if !enableFileLogging {
		t.Error("Expected enableFileLogging to be true after SetFileLogging(true)")
	}
	if !GetFileLogging() {
		t.Error("Expected GetFileLogging() to return true after SetFileLogging(true)")
	}

	// Test disabling
	SetFileLogging(false)
	if enableFileLogging {
		t.Error("Expected enableFileLogging to be false after SetFileLogging(false)")
	}
	if GetFileLogging() {
		t.Error("Expected GetFileLogging() to return false after SetFileLogging(false)")
	}
}

func TestGetFileLogging(t *testing.T) {
	originalEnabled := enableFileLogging
	defer func() { enableFileLogging = originalEnabled }()

	// Test that getter returns current value
	enableFileLogging = true
	result := GetFileLogging()
	if !result {
		t.Error("Expected GetFileLogging() to return true when enableFileLogging is true")
	}

	enableFileLogging = false
	result = GetFileLogging()
	if result {
		t.Error("Expected GetFileLogging() to return false when enableFileLogging is false")
	}
}

func TestDisableEnableColors(t *testing.T) {
	originalNoColor := color.NoColor
	defer func() { color.NoColor = originalNoColor }()

	// Test disabling colors
	DisableColors()
	if !color.NoColor {
		t.Error("Expected color.NoColor to be true after DisableColors()")
	}

	// Test enabling colors
	EnableColors()
	if color.NoColor {
		t.Error("Expected color.NoColor to be false after EnableColors()")
	}
}

func TestColorIntegration(t *testing.T) {
	// Test that colors work with prefixes
	originalNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = originalNoColor }()

	// Test that each prefix contains color information when colors are enabled
	prefixes := []string{InfoPrefix, SuccessPrefix, ErrorPrefix, WarningPrefix}

	for _, prefix := range prefixes {
		if len(prefix) < 10 { // Should be longer when colors are included
			t.Errorf("Expected prefix to be longer when colors are enabled, got length %d", len(prefix))
		}
	}

	// Test color consistency
	redText := Red.Sprint("test")
	blueText := Blue.Sprint("test")

	if redText == blueText {
		t.Error("Expected different colors to produce different output")
	}
}

func BenchmarkColoredString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ColoredString("benchmark text", Red)
	}
}

func BenchmarkColoredFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ColoredFormat("benchmark %s %d", Green, "text", i)
	}
}

func BenchmarkSetLogLevel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SetLogLevel(InfoLevel)
	}
}

func BenchmarkSetCallerSkipFrames(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SetCallerSkipFrames(3)
	}
}

func BenchmarkGetCallerSkipFrames(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetCallerSkipFrames()
	}
}

func BenchmarkSetFileLogging(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SetFileLogging(true)
	}
}

func BenchmarkGetFileLogging(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFileLogging()
	}
}
