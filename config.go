package pim

import (
	"github.com/fatih/color"
)

// Color constants for direct ANSI usage when needed
const (
	ColorString  = "\033[32m" // Green for strings
	ColorNumber  = "\033[36m" // Cyan for numbers
	ColorBool    = "\033[33m" // Yellow for booleans
	ColorNull    = "\033[90m" // Gray for null
	ColorKey     = "\033[34m" // Blue for keys
	ColorComment = "\033[90m" // Gray for comments (same as null for consistency)

	// Reset
	ColorReset = "\033[0m"

	// Regular Colors
	ColorBlack  = "\033[30m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"

	// High Intensity
	ColorHiBlack  = "\033[90m"
	ColorHiRed    = "\033[91m"
	ColorHiGreen  = "\033[92m"
	ColorHiYellow = "\033[93m"
	ColorHiBlue   = "\033[94m"
	ColorHiPurple = "\033[95m"
	ColorHiCyan   = "\033[96m"
	ColorHiWhite  = "\033[97m"
)

// Predefined color styles
var (
	// Basic colors
	Black  = color.New(color.FgBlack)
	Red    = color.New(color.FgRed)
	Green  = color.New(color.FgGreen)
	Yellow = color.New(color.FgYellow)
	Blue   = color.New(color.FgBlue)
	Purple = color.New(color.FgMagenta)
	Cyan   = color.New(color.FgCyan)
	White  = color.New(color.FgWhite)
	Gray   = color.New(color.FgHiBlack)

	// High intensity colors
	HiBlack  = color.New(color.FgHiBlack)
	HiRed    = color.New(color.FgHiRed)
	HiGreen  = color.New(color.FgHiGreen)
	HiYellow = color.New(color.FgHiYellow)
	HiBlue   = color.New(color.FgHiBlue)
	HiPurple = color.New(color.FgHiMagenta)
	HiCyan   = color.New(color.FgHiCyan)
	HiWhite  = color.New(color.FgHiWhite)

	// Special styles
	Bold      = color.New(color.Bold)
	Underline = color.New(color.Underline)
	Italic    = color.New(color.Italic)

	// Semantic colors for specific data types
	StringColor  = Green
	NumberColor  = Cyan
	BoolColor    = Yellow
	NullColor    = Gray
	KeyColor     = Blue
	ErrorColor   = HiRed
	SuccessColor = HiGreen
	WarningColor = HiYellow
	InfoColor    = HiBlue
	DebugColor   = HiPurple
	TraceColor   = HiCyan
	DateColor    = Cyan.Add(color.Bold)
	IDColor      = Purple.Add(color.Bold)
	TagColor     = Green.Add(color.Bold)
	URLColor     = Blue.Add(color.Underline)
)

// Log prefixes using the color package
var (
	// Standard log prefixes
	InfoPrefix    = Cyan.Sprint("ℹ️  INFO     ")
	SuccessPrefix = Green.Sprint("✅ SUCCESS  ")
	InitPrefix    = Blue.Sprint("🚀 INIT     ")
	ConfigPrefix  = Purple.Sprint("⚙️  CONFIG   ")
	WarningPrefix = Yellow.Sprint("⚠️  WARNING  ")
	ErrorPrefix   = Red.Sprint("❌ ERROR    ")

	// Debug log prefixes
	DebugPrefix  = White.Sprint("🔍 DEBUG    ")
	TracePrefix  = HiBlue.Sprint("📍 TRACE    ")
	PanicPrefix  = HiRed.Add(color.Bold).Sprint("💥 PANIC    ")
	MetricPrefix = HiPurple.Sprint("📊 METRIC   ")

	// JSON and data prefixes
	JsonPrefix  = HiCyan.Sprint("🔍 JSON     ")
	DataPrefix  = HiGreen.Sprint("📝 DATA     ")
	ModelPrefix = HiBlue.Sprint("📊 MODEL    ")
)

// LogLevel type for controlling log output
type LogLevel int

const (
	// PanicLevel logs and then calls panic()
	PanicLevel LogLevel = iota
	// ErrorLevel indicates error conditions
	ErrorLevel
	// WarningLevel indicates potentially harmful situations
	WarningLevel
	// InfoLevel indicates general operational information
	InfoLevel
	// DebugLevel indicates detailed debug information
	DebugLevel
	// TraceLevel indicates the most detailed debugging information
	TraceLevel
)

var (
	currentLogLevel   = InfoLevel
	showFileLine      = true
	showGoroutineID   = false
	showFunctionName  = true
	stackDepth        = 3
	showPackageName   = true
	showFullPath      = false
	callerSkipFrames  = 3     // Number of frames to skip to find the actual caller
	enableFileLogging = false // File logging disabled by default
)

// Helper functions for colored output
func ColoredString(s string, c *color.Color) string {
	return c.Sprint(s)
}

func ColoredFormat(format string, c *color.Color, a ...interface{}) string {
	return c.Sprintf(format, a...)
}

// SetLogLevel sets the current logging level
func SetLogLevel(level LogLevel) {
	currentLogLevel = level
}

// SetShowFileLine enables/disables file and line number in logs
func SetShowFileLine(show bool) {
	showFileLine = show
}

// SetShowGoroutineID enables/disables goroutine ID in logs
func SetShowGoroutineID(show bool) {
	showGoroutineID = show
}

// Disable/Enable colors globally
func DisableColors() {
	color.NoColor = true
}

func EnableColors() {
	color.NoColor = false
}

// Enhanced configuration functions
func SetShowFunctionName(show bool) {
	showFunctionName = show
}

func SetShowPackageName(show bool) {
	showPackageName = show
}

func SetShowFullPath(show bool) {
	showFullPath = show
}

func SetStackDepth(depth int) {
	if depth > 0 && depth <= 10 {
		stackDepth = depth
	}
}

// SetCallerSkipFrames sets the number of frames to skip when determining caller info
func SetCallerSkipFrames(skip int) {
	if skip >= 0 && skip <= 15 { // Allow 0-15 frames to skip
		callerSkipFrames = skip
	}
}

// GetCallerSkipFrames returns the current number of frames to skip
func GetCallerSkipFrames() int {
	return callerSkipFrames
}

// SetFileLogging enables/disables file logging
func SetFileLogging(enabled bool) {
	enableFileLogging = enabled
}

// GetFileLogging returns whether file logging is enabled
func GetFileLogging() bool {
	return enableFileLogging
}

// EnableFileLoggingWithPath enables file logging and initializes it with the given path
// Note: This function requires the metrics package to be fully loaded
// Use SetFileLogging(true) and then call pim.InitializeFileLogging() separately
/*
func EnableFileLoggingWithPath(baseDir, serviceName string) error {
	enableFileLogging = true
	return InitializeFileLogging(baseDir, serviceName)
}
*/
