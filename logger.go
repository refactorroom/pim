// Package pim provides an enhanced logging system for Go applications.
// It supports colored output, multiple log levels, JSON formatting, performance metrics,
// and advanced features like stack traces, function/package tracking, and webhook/callback delivery.
//
// For full documentation and examples, see https://pkg.go.dev/github.com/refactorroom/pim
package pim

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CallInfo contains information about the calling function
type CallInfo struct {
	File     string
	Line     int
	Function string
	Package  string
}

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Package  string `json:"package"`
}

// getCallInfo retrieves detailed information about the calling function
func getCallInfo(skip int) CallInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return CallInfo{}
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return CallInfo{}
	}

	// Extract package and function names
	fullName := fn.Name()
	parts := strings.Split(fullName, ".")
	packageName := ""
	functionName := fullName

	if len(parts) > 1 {
		packageName = strings.Join(parts[:len(parts)-1], ".")
		functionName = parts[len(parts)-1]
	}

	// Handle file path
	if !showFullPath {
		file = filepath.Base(file)
	}

	return CallInfo{
		File:     file,
		Line:     line,
		Function: functionName,
		Package:  packageName,
	}
}

// getFileInfo returns formatted file and line information
func getFileInfo() string {
	if !showFileLine {
		return ""
	}

	callInfo := getCallInfo(3) // Skip 3 frames to get to the caller
	if callInfo.File == "" {
		return ""
	}

	parts := []string{callInfo.File}
	if showFunctionName && callInfo.Function != "" {
		if showPackageName && callInfo.Package != "" {
			parts = append(parts, fmt.Sprintf("%s.%s", callInfo.Package, callInfo.Function))
		} else {
			parts = append(parts, callInfo.Function)
		}
	}
	parts = append(parts, fmt.Sprintf("L%d", callInfo.Line))

	return strings.Join(parts, ":")
}

// getGoroutineID returns the current goroutine ID
func getGoroutineID() string {
	if !showGoroutineID {
		return ""
	}
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	id := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	return fmt.Sprintf("(goroutine %s)", id)
}

// getStackTrace returns a formatted stack trace
func getStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	for i := skip; i < skip+stackDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		fullName := fn.Name()
		parts := strings.Split(fullName, ".")
		packageName := ""
		functionName := fullName

		if len(parts) > 1 {
			packageName = strings.Join(parts[:len(parts)-1], ".")
			functionName = parts[len(parts)-1]
		}

		if !showFullPath {
			file = filepath.Base(file)
		}

		frames = append(frames, StackFrame{
			File:     file,
			Line:     line,
			Function: functionName,
			Package:  packageName,
		})
	}

	return frames
}

// formatStackTrace formats stack trace for pim output
func formatStackTrace(frames []StackFrame) string {
	if len(frames) == 0 {
		return ""
	}

	var lines []string
	for i, frame := range frames {
		indent := strings.Repeat("  ", i)
		parts := []string{frame.File}

		if showFunctionName && frame.Function != "" {
			if showPackageName && frame.Package != "" {
				parts = append(parts, fmt.Sprintf("%s.%s", frame.Package, frame.Function))
			} else {
				parts = append(parts, frame.Function)
			}
		}
		parts = append(parts, fmt.Sprintf("L%d", frame.Line))

		line := fmt.Sprintf("%s↳ %s", indent, strings.Join(parts, ":"))
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func LogWithTimestamp(Prefix, msg string, level LogLevel) {
	if level > currentLogLevel {
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05.000 UTC")
	fileInfo := getFileInfo()
	goroutineInfo := getGoroutineID()

	// Construct the log message
	logMsg := fmt.Sprintf("%s [%s]", Prefix, timestamp)
	if fileInfo != "" {
		logMsg += fmt.Sprintf(" [%s]", fileInfo)
	}
	if goroutineInfo != "" {
		logMsg += fmt.Sprintf(" %s", goroutineInfo)
	}
	logMsg += fmt.Sprintf(" %s", msg)

	fmt.Println(logMsg)

	// File output
	entry := LogEntry{
		Timestamp:   timestamp,
		Level:       getLevelString(level),
		Message:     msg,
		File:        fileInfo,
		GoroutineID: goroutineInfo,
	}

	if err := writeLogEntry(entry, level); err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
	}
}

// LogWithStackTrace logs a message with stack trace information
func LogWithStackTrace(Prefix, msg string, level LogLevel) {
	if level > currentLogLevel {
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05.000 UTC")
	fileInfo := getFileInfo()
	goroutineInfo := getGoroutineID()
	stackTrace := getStackTrace(4) // Skip 4 frames to get meaningful stack

	// Construct the log message
	logMsg := fmt.Sprintf("%s [%s]", Prefix, timestamp)
	if fileInfo != "" {
		logMsg += fmt.Sprintf(" [%s]", fileInfo)
	}
	if goroutineInfo != "" {
		logMsg += fmt.Sprintf(" %s", goroutineInfo)
	}
	logMsg += fmt.Sprintf(" %s", msg)

	fmt.Println(logMsg)

	// Print stack trace if available
	if len(stackTrace) > 0 {
		stackStr := formatStackTrace(stackTrace)
		if stackStr != "" {
			fmt.Println(stackStr)
		}
	}

	// File output
	entry := LogEntry{
		Timestamp:   timestamp,
		Level:       getLevelString(level),
		Message:     msg,
		File:        fileInfo,
		GoroutineID: goroutineInfo,
	}

	if err := writeLogEntry(entry, level); err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
	}
}

func Info(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(InfoPrefix, logMsg, InfoLevel)
}

func Success(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(SuccessPrefix, logMsg, InfoLevel)
}

func Init(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(InitPrefix, logMsg, InfoLevel)
}

func Config(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(ConfigPrefix, logMsg, InfoLevel)
}

func Warning(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(WarningPrefix, logMsg, WarningLevel)
}

func Error(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithStackTrace(ErrorPrefix, logMsg, ErrorLevel)
}

func Debug(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(DebugPrefix, logMsg, DebugLevel)
}

func Trace(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(TracePrefix, logMsg, TraceLevel)
}

func Panic(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithStackTrace(PanicPrefix, logMsg, PanicLevel)
	panic(logMsg)
}

func Metric(name string, value interface{}, tags ...string) {
	tagStr := ""
	if len(tags) > 0 {
		tagStr = fmt.Sprintf(" [%s]", strings.Join(tags, ", "))
	}
	logMsg := fmt.Sprintf("%s: %v%s", name, value, tagStr)
	LogWithTimestamp(MetricPrefix, logMsg, InfoLevel)
}

// GetCallInfo returns current call information for external use
func GetCallInfo() CallInfo {
	return getCallInfo(2)
}

// GetStackTrace returns current stack trace for external use
func GetStackTrace() []StackFrame {
	return getStackTrace(2)
}
