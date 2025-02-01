package console

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type LogEntry struct {
	Timestamp     string            `json:"timestamp"`
	TraceID       string            `json:"trace_id"`
	SpanID        string            `json:"span_id"`
	ParentSpanID  string            `json:"parent_span_id,omitempty"`
	ServiceName   string            `json:"service_name"`
	OperationName string            `json:"operation_name"`
	Level         string            `json:"level"`
	Message       string            `json:"message"`
	File          string            `json:"file,omitempty"`
	GoroutineID   string            `json:"goroutine_id,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	Duration      int64             `json:"duration_ms,omitempty"`
	Error         bool              `json:"error"`
}

var (
	logDir      string
	logFiles    = make(map[LogLevel]*os.File)
	logMu       sync.Mutex
	serviceName string
)

func InitializeFileLogging(baseDir, service string) error {
	logDir = filepath.Join(baseDir, ".log/console")
	serviceName = service

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create jaeger log directory: %v", err)
	}

	levels := []LogLevel{PanicLevel, ErrorLevel, WarningLevel, InfoLevel, DebugLevel, TraceLevel}
	for _, level := range levels {
		if err := openJaegerLogFile(level); err != nil {
			return err
		}
	}
	return nil
}

func getLevelString(level LogLevel) string {
	switch level {
	case PanicLevel:
		return "panic"
	case ErrorLevel:
		return "error"
	case WarningLevel:
		return "warning"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	case TraceLevel:
		return "trace"
	default:
		return "unknown"
	}
}

func openJaegerLogFile(level LogLevel) error {
	fileName := fmt.Sprintf("%s.jaeger.json", getLevelString(level))
	filePath := filepath.Join(logDir, fileName)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open jaeger log file %s: %v", filePath, err)
	}

	logFiles[level] = file
	return nil
}
func writeToFile(entry LogEntry, level LogLevel) error {
	logMu.Lock()
	defer logMu.Unlock()

	file, ok := logFiles[level]
	if !ok {
		return fmt.Errorf("no jaeger log file for level: %v", level)
	}

	entry.ServiceName = serviceName

	// Format with indentation for readability
	jsonData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal jaeger entry: %v", err)
	}

	// Write the JSON object and comma
	if _, err := file.Write(append(jsonData, []byte(",\n")...)); err != nil {
		return fmt.Errorf("failed to write to jaeger log file: %v", err)
	}

	return nil
}
func CloseLogFiles() {
	logMu.Lock()
	defer logMu.Unlock()

	for _, file := range logFiles {
		file.Close()
	}
}
