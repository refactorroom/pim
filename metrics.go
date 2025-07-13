package pim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogEntry represents a structured log entry
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

// WebhookConfig represents configuration for webhook delivery
type WebhookConfig struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
	RetryDelay time.Duration     `json:"retry_delay"`
	BatchSize  int               `json:"batch_size"`
	BatchDelay time.Duration     `json:"batch_delay"`
}

// StamperConfig represents the overall stamper configuration
type StamperConfig struct {
	Enabled     bool           `json:"enabled"`
	FileLogging bool           `json:"file_logging"`
	Webhook     *WebhookConfig `json:"webhook,omitempty"`
	Callback    LogCallback    `json:"-"`
}

// LogCallback is a function type for custom log processing
type LogCallback func(entry LogEntry, level LogLevel) error

var (
	logDir      string
	logFiles    = make(map[LogLevel]*os.File)
	logMu       sync.Mutex
	serviceName string

	// Stamper configuration
	stamperConfig = StamperConfig{
		Enabled:     true,
		FileLogging: true,
	}

	// Webhook delivery
	webhookClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Batch processing
	logBatch    = make([]LogEntry, 0)
	batchMu     sync.Mutex
	batchTicker *time.Ticker
	batchDone   chan bool
)

// InitializeFileLogging initializes file-based logging
func InitializeFileLogging(baseDir, service string) error {
	logDir = filepath.Join(baseDir, ".log/pim")
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

// ConfigureStamper sets the stamper configuration
func ConfigureStamper(config StamperConfig) {
	stamperConfig = config

	// Initialize webhook batch processing if webhook is configured
	if config.Webhook != nil && config.Webhook.URL != "" {
		initializeWebhookBatch()
	}
}

// EnableStamper enables or disables the stamper
func EnableStamper(enabled bool) {
	stamperConfig.Enabled = enabled
}

// EnableFileLogging enables or disables file logging
func EnableFileLogging(enabled bool) {
	stamperConfig.FileLogging = enabled
}

// SetWebhookConfig sets webhook configuration
func SetWebhookConfig(webhook *WebhookConfig) {
	stamperConfig.Webhook = webhook
	if webhook != nil && webhook.URL != "" {
		initializeWebhookBatch()
	}
}

// SetLogCallback sets a custom callback function for log processing
func SetLogCallback(callback LogCallback) {
	stamperConfig.Callback = callback
}

// initializeWebhookBatch initializes batch processing for webhook delivery
func initializeWebhookBatch() {
	if batchTicker != nil {
		batchTicker.Stop()
		close(batchDone)
	}

	if stamperConfig.Webhook == nil {
		return
	}

	batchTicker = time.NewTicker(stamperConfig.Webhook.BatchDelay)
	batchDone = make(chan bool)

	go func() {
		for {
			select {
			case <-batchTicker.C:
				flushWebhookBatch()
			case <-batchDone:
				return
			}
		}
	}()
}

// flushWebhookBatch sends batched logs to webhook
func flushWebhookBatch() {
	batchMu.Lock()
	if len(logBatch) == 0 {
		batchMu.Unlock()
		return
	}

	batch := make([]LogEntry, len(logBatch))
	copy(batch, logBatch)
	logBatch = logBatch[:0]
	batchMu.Unlock()

	if len(batch) > 0 {
		sendWebhookBatch(batch)
	}
}

// sendWebhookBatch sends a batch of logs to the webhook
func sendWebhookBatch(entries []LogEntry) {
	if stamperConfig.Webhook == nil || stamperConfig.Webhook.URL == "" {
		return
	}

	payload := map[string]interface{}{
		"service_name": serviceName,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"logs":         entries,
		"batch_size":   len(entries),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	// Retry logic
	for attempt := 0; attempt <= stamperConfig.Webhook.RetryCount; attempt++ {
		if err := sendWebhookRequest(jsonData); err == nil {
			return
		} else if attempt < stamperConfig.Webhook.RetryCount {
			time.Sleep(stamperConfig.Webhook.RetryDelay)
		}
	}
}

// sendWebhookRequest sends a single webhook request
func sendWebhookRequest(jsonData []byte) error {
	req, err := http.NewRequest(stamperConfig.Webhook.Method, stamperConfig.Webhook.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PIM-Logger/1.0")

	// Set custom headers
	for key, value := range stamperConfig.Webhook.Headers {
		req.Header.Set(key, value)
	}

	// Create client with timeout
	client := &http.Client{
		Timeout: stamperConfig.Webhook.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
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

// writeToFile writes log entry to file
func writeToFile(entry LogEntry, level LogLevel) error {
	if !stamperConfig.Enabled || !stamperConfig.FileLogging {
		return nil
	}

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

// writeToWebhook adds log entry to webhook batch
func writeToWebhook(entry LogEntry, level LogLevel) error {
	if !stamperConfig.Enabled || stamperConfig.Webhook == nil || stamperConfig.Webhook.URL == "" {
		return nil
	}

	entry.ServiceName = serviceName

	batchMu.Lock()
	logBatch = append(logBatch, entry)

	// Send immediately if batch is full
	if len(logBatch) >= stamperConfig.Webhook.BatchSize {
		batch := make([]LogEntry, len(logBatch))
		copy(batch, logBatch)
		logBatch = logBatch[:0]
		batchMu.Unlock()

		go sendWebhookBatch(batch)
	} else {
		batchMu.Unlock()
	}

	return nil
}

// writeToCallback calls the custom callback function
func writeToCallback(entry LogEntry, level LogLevel) error {
	if !stamperConfig.Enabled || stamperConfig.Callback == nil {
		return nil
	}

	entry.ServiceName = serviceName
	return stamperConfig.Callback(entry, level)
}

// writeLogEntry writes log entry to all configured outputs
func writeLogEntry(entry LogEntry, level LogLevel) error {
	var errors []error

	// Write to file
	if err := writeToFile(entry, level); err != nil {
		errors = append(errors, fmt.Errorf("file write error: %v", err))
	}

	// Write to webhook
	if err := writeToWebhook(entry, level); err != nil {
		errors = append(errors, fmt.Errorf("webhook write error: %v", err))
	}

	// Write to callback
	if err := writeToCallback(entry, level); err != nil {
		errors = append(errors, fmt.Errorf("callback write error: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("log write errors: %v", errors)
	}

	return nil
}

// CloseLogFiles closes all log files and flushes webhook batch
func CloseLogFiles() {
	// Flush webhook batch
	flushWebhookBatch()

	// Stop batch ticker
	if batchTicker != nil {
		batchTicker.Stop()
		close(batchDone)
	}

	// Close log files
	logMu.Lock()
	defer logMu.Unlock()

	for _, file := range logFiles {
		file.Close()
	}
}
