// Package clog provides an enhanced logging system with colored output, multiple log levels,
// and debug features for Go applications.
//
// Basic usage:
//
//	console.Info("Starting application...")
//	console.Success("Server started on port %d", 8080)
//	console.Error("Failed to connect: %v", err)
//
// Log Levels (from lowest to highest):
//   - PanicLevel: System is unusable, halts execution
//   - ErrorLevel: Error events that might still allow the application to continue running
//   - WarningLevel: Warning messages for potentially harmful situations
//   - InfoLevel: General informational messages about system operation
//   - DebugLevel: Detailed information for debugging
//   - TraceLevel: Extremely detailed debugging information
//
// Configuration:
//
//	clog.SetLogLevel(console.DebugLevel)
//	clog.SetShowFileLine(true)
//	clog.SetShowGoroutineID(true)
//
// Features:
//   - Colored output using emoji prefixes
//   - UTC timestamp with millisecond precision
//   - File and line number tracking
//   - Goroutine ID tracking
//   - Performance metrics logging
//   - Multiple log levels with filtering
//   - Panic handling with stack traces
package console

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func getFileInfo() string {
	if !showFileLine {
		return ""
	}
	_, file, line, ok := runtime.Caller(3) // Skip 3 frames to get to the caller
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func getGoroutineID() string {
	if !showGoroutineID {
		return ""
	}
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	id := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	return fmt.Sprintf("(goroutine %s)", id)
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

	if err := writeToFile(entry, level); err != nil {
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

// func Error(format string, args ...interface{}) {
// 	LogWithTimestamp(ErrorPrefix, fmt.Sprintf(format, args...), ErrorLevel)
// }

func Error(msg string, args ...interface{}) {
	logMsg := msg
	if len(args) > 0 {
		logMsg += ": " + fmt.Sprint(args...)
	}
	LogWithTimestamp(ErrorPrefix, logMsg, ErrorLevel)
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
	LogWithTimestamp(PanicPrefix, logMsg, PanicLevel)
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
