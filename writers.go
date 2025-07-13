package pim

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// RotationConfig configures log file rotation
type RotationConfig struct {
	MaxSize         int64         `json:"max_size"`         // Max file size in bytes
	MaxAge          time.Duration `json:"max_age"`          // Max age of log files
	MaxFiles        int           `json:"max_files"`        // Max number of log files to keep
	Compress        bool          `json:"compress"`         // Whether to compress old log files
	RotateTime      time.Duration `json:"rotate_time"`      // Time-based rotation interval
	CleanupInterval time.Duration `json:"cleanup_interval"` // How often to run cleanup (default: 1 hour)
	VerboseCleanup  bool          `json:"verbose_cleanup"`  // Whether to log cleanup operations
}

// ConsoleWriter writes log entries to the console
type ConsoleWriter struct {
	config       LoggerConfig
	themeManager *ThemeManager
}

// NewConsoleWriter creates a new console writer
func NewConsoleWriter(config LoggerConfig) *ConsoleWriter {
	writer := &ConsoleWriter{
		config:       config,
		themeManager: NewThemeManager(),
	}

	// Initialize theme manager
	if config.CustomTheme != nil {
		writer.themeManager.currentTheme = config.CustomTheme
	} else if config.ThemeName != "" {
		writer.themeManager.SetTheme(config.ThemeName)
	}

	// Register custom format if provided
	if config.CustomFormat != "" {
		writer.themeManager.RegisterTemplate("custom", config.CustomFormat)
	}

	return writer
}

// Write implements LogWriter interface for console output
func (w *ConsoleWriter) Write(entry CoreLogEntry) error {
	if w.config.EnableJSON {
		return w.writeJSON(entry)
	}

	// Use theming if available
	if w.themeManager != nil && w.config.FormatName != "" {
		formatted := w.themeManager.Format(entry, w.config.FormatName)
		fmt.Println(formatted)
		return nil
	}

	return w.writeFormatted(entry)
}

// writeFormatted writes a formatted log entry to console
func (w *ConsoleWriter) writeFormatted(entry CoreLogEntry) error {
	var parts []string

	// Add prefix
	if entry.Prefix != "" {
		parts = append(parts, entry.Prefix)
	}

	// Add timestamp
	timestamp := entry.Timestamp.Format(w.config.TimestampFormat)
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))

	// Add file/line info
	if entry.File != "" {
		fileInfo := entry.File
		if entry.Function != "" {
			if w.config.ShowPackageName && entry.Package != "" {
				fileInfo += fmt.Sprintf(":%s.%s", entry.Package, entry.Function)
			} else {
				fileInfo += fmt.Sprintf(":%s", entry.Function)
			}
		}
		fileInfo += fmt.Sprintf(":L%d", entry.Line)
		parts = append(parts, fmt.Sprintf("[%s]", fileInfo))
	}

	// Add goroutine ID
	if entry.GoroutineID != "" {
		parts = append(parts, entry.GoroutineID)
	}

	// Add message
	parts = append(parts, entry.Message)

	// Add context if present
	if len(entry.Context) > 0 {
		contextStr := w.formatContext(entry.Context)
		parts = append(parts, fmt.Sprintf("{%s}", contextStr))
	}

	// Join and print
	logLine := strings.Join(parts, " ")
	fmt.Println(logLine)

	// Print stack trace if present
	if len(entry.StackTrace) > 0 {
		stackStr := w.formatStackTrace(entry.StackTrace)
		if stackStr != "" {
			fmt.Println(stackStr)
		}
	}

	return nil
}

// writeJSON writes a JSON log entry to console
func (w *ConsoleWriter) writeJSON(entry CoreLogEntry) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// formatContext formats context fields for console output
func (w *ConsoleWriter) formatContext(context map[string]interface{}) string {
	var pairs []string
	for k, v := range context {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(pairs, ", ")
}

// formatStackTrace formats stack trace for console output
func (w *ConsoleWriter) formatStackTrace(frames []StackFrame) string {
	if len(frames) == 0 {
		return ""
	}

	var lines []string
	for i, frame := range frames {
		indent := strings.Repeat("  ", i)
		parts := []string{frame.File}

		if w.config.ShowFunctionName && frame.Function != "" {
			if w.config.ShowPackageName && frame.Package != "" {
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

// Close implements LogWriter interface
func (w *ConsoleWriter) Close() error {
	// Console writer doesn't need to close anything
	return nil
}

// Flush implements LogWriter interface
func (w *ConsoleWriter) Flush() error {
	return nil
}

// FileWriter writes log entries to files with rotation
type FileWriter struct {
	file           *os.File
	config         LoggerConfig
	rotationConfig RotationConfig
	filePath       string
	fileSize       int64
	lastRotate     time.Time
	mu             sync.Mutex
}

// NewFileWriter creates a new file writer with rotation
func NewFileWriter(filename string, config LoggerConfig, rotationConfig RotationConfig) (*FileWriter, error) {
	writer := &FileWriter{
		config:         config,
		rotationConfig: rotationConfig,
		filePath:       filename,
		lastRotate:     time.Now(),
	}

	if err := writer.openFile(); err != nil {
		return nil, err
	}

	// Start cleanup goroutine if max age or max files is set
	if rotationConfig.MaxAge > 0 || rotationConfig.MaxFiles > 0 {
		go writer.cleanupOldFiles()
	}

	return writer, nil
}

// openFile opens the log file
func (w *FileWriter) openFile() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Close existing file if open
	if w.file != nil {
		w.file.Close()
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(w.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file
	file, err := os.OpenFile(w.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", w.filePath, err)
	}

	w.file = file

	// Get current file size
	if stat, err := file.Stat(); err == nil {
		w.fileSize = stat.Size()
	}

	return nil
}

// shouldRotate checks if the file should be rotated
func (w *FileWriter) shouldRotate() bool {
	// Size-based rotation
	if w.rotationConfig.MaxSize > 0 && w.fileSize >= w.rotationConfig.MaxSize {
		return true
	}

	// Time-based rotation
	if w.rotationConfig.RotateTime > 0 && time.Since(w.lastRotate) >= w.rotationConfig.RotateTime {
		return true
	}

	return false
}

// rotateFile rotates the current log file
func (w *FileWriter) rotateFile() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	// Close current file
	w.file.Close()

	// Generate rotated filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	ext := filepath.Ext(w.filePath)
	base := strings.TrimSuffix(w.filePath, ext)
	rotatedPath := fmt.Sprintf("%s.%s%s", base, timestamp, ext)

	// Rename current file to rotated name
	if err := os.Rename(w.filePath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Compress if enabled
	if w.rotationConfig.Compress {
		go w.compressFile(rotatedPath)
	}

	// Open new file
	if err := w.openFile(); err != nil {
		return err
	}

	w.lastRotate = time.Now()
	w.fileSize = 0

	return nil
}

// compressFile compresses a log file using gzip
func (w *FileWriter) compressFile(filePath string) {
	// Open the original file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file for compression %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	// Create the compressed file
	compressedPath := filePath + ".gz"
	compressedFile, err := os.Create(compressedPath)
	if err != nil {
		fmt.Printf("Failed to create compressed file %s: %v\n", compressedPath, err)
		return
	}
	defer compressedFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(compressedFile)
	defer gzipWriter.Close()

	// Copy data from original to compressed file
	_, err = io.Copy(gzipWriter, file)
	if err != nil {
		fmt.Printf("Failed to compress file %s: %v\n", filePath, err)
		// Clean up the partial compressed file
		os.Remove(compressedPath)
		return
	}

	// Remove the original file after successful compression
	if err := os.Remove(filePath); err != nil {
		fmt.Printf("Failed to remove original file %s after compression: %v\n", filePath, err)
	}
}

// cleanupOldFiles removes old log files based on age and count
func (w *FileWriter) cleanupOldFiles() {
	interval := w.rotationConfig.CleanupInterval
	if interval == 0 {
		interval = 1 * time.Hour // Default to 1 hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		w.cleanupFiles()
	}
}

// cleanupFiles performs the actual file cleanup
func (w *FileWriter) cleanupFiles() {
	dir := filepath.Dir(w.filePath)
	base := strings.TrimSuffix(filepath.Base(w.filePath), filepath.Ext(w.filePath))
	ext := filepath.Ext(w.filePath)

	// Find all log files (including compressed ones)
	patterns := []string{
		filepath.Join(dir, base+".*"+ext),       // Regular log files
		filepath.Join(dir, base+".*"+ext+".gz"), // Compressed log files
	}

	var allMatches []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		allMatches = append(allMatches, matches...)
	}

	var files []fileInfo
	for _, match := range allMatches {
		if stat, err := os.Stat(match); err == nil {
			files = append(files, fileInfo{
				path:    match,
				modTime: stat.ModTime(),
				size:    stat.Size(),
			})
		}
	}

	// Sort by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove files based on age
	if w.rotationConfig.MaxAge > 0 {
		cutoff := time.Now().Add(-w.rotationConfig.MaxAge)
		for _, file := range files {
			if file.modTime.Before(cutoff) {
				if err := os.Remove(file.path); err != nil {
					if w.rotationConfig.VerboseCleanup {
						fmt.Printf("Failed to remove old log file %s: %v\n", file.path, err)
					}
				} else if w.rotationConfig.VerboseCleanup {
					fmt.Printf("Removed old log file: %s\n", file.path)
				}
			}
		}
	}

	// Remove files based on count
	if w.rotationConfig.MaxFiles > 0 && len(files) > w.rotationConfig.MaxFiles {
		toRemove := len(files) - w.rotationConfig.MaxFiles
		for i := 0; i < toRemove; i++ {
			if err := os.Remove(files[i].path); err != nil {
				if w.rotationConfig.VerboseCleanup {
					fmt.Printf("Failed to remove excess log file %s: %v\n", files[i].path, err)
				}
			} else if w.rotationConfig.VerboseCleanup {
				fmt.Printf("Removed excess log file: %s\n", files[i].path)
			}
		}
	}
}

// fileInfo holds information about a log file
type fileInfo struct {
	path    string
	modTime time.Time
	size    int64
}

// Write implements LogWriter interface for file output with rotation
func (w *FileWriter) Write(entry CoreLogEntry) error {
	// Check if rotation is needed
	if w.shouldRotate() {
		if err := w.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return fmt.Errorf("log file is not open")
	}

	var data []byte
	var err error

	if w.config.EnableJSON {
		data, err = json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry: %w", err)
		}
		data = append(data, '\n')
	} else {
		data = []byte(w.formatLogEntry(entry) + "\n")
	}

	// Write data
	n, err := w.file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	// Update file size
	w.fileSize += int64(n)

	return nil
}

// formatLogEntry formats a log entry for text output
func (w *FileWriter) formatLogEntry(entry CoreLogEntry) string {
	var parts []string

	// Add timestamp
	timestamp := entry.Timestamp.Format(w.config.TimestampFormat)
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))

	// Add level
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.LevelString)))

	// Add service name
	if entry.ServiceName != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.ServiceName))
	}

	// Add file/line info
	if entry.File != "" {
		fileInfo := entry.File
		if entry.Function != "" {
			if w.config.ShowPackageName && entry.Package != "" {
				fileInfo += fmt.Sprintf(":%s.%s", entry.Package, entry.Function)
			} else {
				fileInfo += fmt.Sprintf(":%s", entry.Function)
			}
		}
		fileInfo += fmt.Sprintf(":L%d", entry.Line)
		parts = append(parts, fmt.Sprintf("[%s]", fileInfo))
	}

	// Add goroutine ID
	if entry.GoroutineID != "" {
		parts = append(parts, entry.GoroutineID)
	}

	// Add message
	parts = append(parts, entry.Message)

	// Add context if present
	if len(entry.Context) > 0 {
		contextStr := w.formatContext(entry.Context)
		parts = append(parts, fmt.Sprintf("{%s}", contextStr))
	}

	return strings.Join(parts, " ")
}

// formatContext formats context fields for file output
func (w *FileWriter) formatContext(context map[string]interface{}) string {
	var pairs []string
	for k, v := range context {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(pairs, ", ")
}

// formatStackTrace formats stack trace for file output
func (w *FileWriter) formatStackTrace(frames []StackFrame) string {
	if len(frames) == 0 {
		return ""
	}

	var lines []string
	for i, frame := range frames {
		indent := strings.Repeat("  ", i)
		parts := []string{frame.File}

		if w.config.ShowFunctionName && frame.Function != "" {
			if w.config.ShowPackageName && frame.Package != "" {
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

// Close implements LogWriter interface
func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Flush implements LogWriter interface
func (w *FileWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// MultiWriter writes to multiple writers
type MultiWriter struct {
	writers []LogWriter
}

// NewMultiWriter creates a new multi-writer
func NewMultiWriter(writers ...LogWriter) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Write implements LogWriter interface for multi-writer
func (w *MultiWriter) Write(entry CoreLogEntry) error {
	var errors []error

	for _, writer := range w.writers {
		if err := writer.Write(entry); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors writing to multiple writers: %v", errors)
	}

	return nil
}

// Close implements LogWriter interface
func (w *MultiWriter) Close() error {
	var errors []error

	for _, writer := range w.writers {
		if err := writer.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing multiple writers: %v", errors)
	}

	return nil
}

// Flush implements LogWriter interface
func (w *MultiWriter) Flush() error {
	var errors []error
	for _, writer := range w.writers {
		if err := writer.Flush(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors flushing multiple writers: %v", errors)
	}
	return nil
}

// StderrWriter writes log entries to stderr
type StderrWriter struct {
	config LoggerConfig
}

// NewStderrWriter creates a new stderr writer
func NewStderrWriter(config LoggerConfig) *StderrWriter {
	return &StderrWriter{
		config: config,
	}
}

// Write implements LogWriter interface for stderr output
func (w *StderrWriter) Write(entry CoreLogEntry) error {
	if w.config.EnableJSON {
		return w.writeJSON(entry)
	}
	return w.writeFormatted(entry)
}

// writeFormatted writes a formatted log entry to stderr
func (w *StderrWriter) writeFormatted(entry CoreLogEntry) error {
	var parts []string

	// Add prefix
	if entry.Prefix != "" {
		parts = append(parts, entry.Prefix)
	}

	// Add timestamp
	timestamp := entry.Timestamp.Format(w.config.TimestampFormat)
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))

	// Add file/line info
	if entry.File != "" {
		fileInfo := entry.File
		if entry.Function != "" {
			if w.config.ShowPackageName && entry.Package != "" {
				fileInfo += fmt.Sprintf(":%s.%s", entry.Package, entry.Function)
			} else {
				fileInfo += fmt.Sprintf(":%s", entry.Function)
			}
		}
		fileInfo += fmt.Sprintf(":L%d", entry.Line)
		parts = append(parts, fmt.Sprintf("[%s]", fileInfo))
	}

	// Add goroutine ID
	if entry.GoroutineID != "" {
		parts = append(parts, entry.GoroutineID)
	}

	// Add message
	parts = append(parts, entry.Message)

	// Add context if present
	if len(entry.Context) > 0 {
		contextStr := w.formatContext(entry.Context)
		parts = append(parts, fmt.Sprintf("{%s}", contextStr))
	}

	// Join and print to stderr
	logLine := strings.Join(parts, " ")
	fmt.Fprintln(os.Stderr, logLine)

	// Print stack trace if present
	if len(entry.StackTrace) > 0 {
		stackStr := w.formatStackTrace(entry.StackTrace)
		if stackStr != "" {
			fmt.Fprintln(os.Stderr, stackStr)
		}
	}

	return nil
}

// writeJSON writes a JSON log entry to stderr
func (w *StderrWriter) writeJSON(entry CoreLogEntry) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	fmt.Fprintln(os.Stderr, string(jsonData))
	return nil
}

// formatContext formats context fields for stderr output
func (w *StderrWriter) formatContext(context map[string]interface{}) string {
	var pairs []string
	for k, v := range context {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(pairs, ", ")
}

// formatStackTrace formats stack trace for stderr output
func (w *StderrWriter) formatStackTrace(frames []StackFrame) string {
	if len(frames) == 0 {
		return ""
	}

	var lines []string
	for i, frame := range frames {
		indent := strings.Repeat("  ", i)
		parts := []string{frame.File}

		if w.config.ShowFunctionName && frame.Function != "" {
			if w.config.ShowPackageName && frame.Package != "" {
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

// Close implements LogWriter interface
func (w *StderrWriter) Close() error {
	// Stderr writer doesn't need to close anything
	return nil
}

// Flush implements LogWriter interface
func (w *StderrWriter) Flush() error {
	return nil
}

// BufferWriter writes log entries to an in-memory buffer
type BufferWriter struct {
	config  LoggerConfig
	buffer  []CoreLogEntry
	mu      sync.RWMutex
	maxSize int // Maximum number of entries to keep
}

// NewBufferWriter creates a new buffer writer
func NewBufferWriter(config LoggerConfig, maxSize int) *BufferWriter {
	return &BufferWriter{
		config:  config,
		buffer:  make([]CoreLogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Write implements LogWriter interface for buffer output
func (w *BufferWriter) Write(entry CoreLogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Add entry to buffer
	w.buffer = append(w.buffer, entry)

	// Remove oldest entries if buffer is full
	if len(w.buffer) > w.maxSize {
		w.buffer = w.buffer[1:]
	}

	return nil
}

// GetBuffer returns a copy of the current buffer
func (w *BufferWriter) GetBuffer() []CoreLogEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]CoreLogEntry, len(w.buffer))
	copy(result, w.buffer)
	return result
}

// ClearBuffer clears the buffer
func (w *BufferWriter) ClearBuffer() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer = w.buffer[:0]
}

// GetBufferSize returns the current buffer size
func (w *BufferWriter) GetBufferSize() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.buffer)
}

// Close implements LogWriter interface
func (w *BufferWriter) Close() error {
	w.ClearBuffer()
	return nil
}

// Flush implements LogWriter interface
func (w *BufferWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	// No-op: buffer is in-memory, nothing to flush to disk
	return nil
}

// RemoteWriter writes log entries to a remote HTTP endpoint
type RemoteWriter struct {
	config     LoggerConfig
	client     *http.Client
	endpoint   string
	headers    map[string]string
	batchSize  int
	batchDelay time.Duration
	buffer     []CoreLogEntry
	mu         sync.Mutex
	stopCh     chan struct{}
}

// RemoteWriterConfig configures remote writer behavior
type RemoteWriterConfig struct {
	Endpoint      string            `json:"endpoint"`       // HTTP endpoint URL
	Headers       map[string]string `json:"headers"`        // Custom headers
	Timeout       time.Duration     `json:"timeout"`        // HTTP timeout
	BatchSize     int               `json:"batch_size"`     // Number of entries to batch
	BatchDelay    time.Duration     `json:"batch_delay"`    // Delay between batches
	RetryAttempts int               `json:"retry_attempts"` // Number of retry attempts
	RetryDelay    time.Duration     `json:"retry_delay"`    // Delay between retries
}

// NewRemoteWriter creates a new remote writer
func NewRemoteWriter(config LoggerConfig, remoteConfig RemoteWriterConfig) *RemoteWriter {
	if remoteConfig.Timeout == 0 {
		remoteConfig.Timeout = 30 * time.Second
	}
	if remoteConfig.BatchSize == 0 {
		remoteConfig.BatchSize = 100
	}
	if remoteConfig.BatchDelay == 0 {
		remoteConfig.BatchDelay = 5 * time.Second
	}
	if remoteConfig.RetryAttempts == 0 {
		remoteConfig.RetryAttempts = 3
	}
	if remoteConfig.RetryDelay == 0 {
		remoteConfig.RetryDelay = 1 * time.Second
	}

	writer := &RemoteWriter{
		config:     config,
		client:     &http.Client{Timeout: remoteConfig.Timeout},
		endpoint:   remoteConfig.Endpoint,
		headers:    remoteConfig.Headers,
		batchSize:  remoteConfig.BatchSize,
		batchDelay: remoteConfig.BatchDelay,
		buffer:     make([]CoreLogEntry, 0, remoteConfig.BatchSize),
		stopCh:     make(chan struct{}),
	}

	// Start background batch processor
	go writer.batchProcessor()

	return writer
}

// Write implements LogWriter interface for remote output
func (w *RemoteWriter) Write(entry CoreLogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer = append(w.buffer, entry)

	// Send immediately if buffer is full
	if len(w.buffer) >= w.batchSize {
		return w.sendBatch()
	}

	return nil
}

// batchProcessor runs in background to send batches periodically
func (w *RemoteWriter) batchProcessor() {
	ticker := time.NewTicker(w.batchDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mu.Lock()
			if len(w.buffer) > 0 {
				w.sendBatch()
			}
			w.mu.Unlock()
		case <-w.stopCh:
			// Send remaining entries on shutdown
			w.mu.Lock()
			if len(w.buffer) > 0 {
				w.sendBatch()
			}
			w.mu.Unlock()
			return
		}
	}
}

// sendBatch sends the current batch to the remote endpoint
func (w *RemoteWriter) sendBatch() error {
	if len(w.buffer) == 0 {
		return nil
	}

	// Prepare batch data
	var data []byte
	var err error

	if w.config.EnableJSON {
		data, err = json.Marshal(w.buffer)
	} else {
		// Convert to text format
		var lines []string
		for _, entry := range w.buffer {
			lines = append(lines, w.formatLogEntry(entry))
		}
		data = []byte(strings.Join(lines, "\n") + "\n")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", w.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if w.config.EnableJSON {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "text/plain")
	}

	for k, v := range w.headers {
		req.Header.Set(k, v)
	}

	// Send request with retries
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := w.client.Do(req)
		if err != nil {
			if attempt < 2 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return fmt.Errorf("failed to send batch after retries: %w", err)
		}

		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			break
		}

		if attempt < 2 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		return fmt.Errorf("remote endpoint returned status %d", resp.StatusCode)
	}

	// Clear buffer after successful send
	w.buffer = w.buffer[:0]
	return nil
}

// formatLogEntry formats a log entry for remote text output
func (w *RemoteWriter) formatLogEntry(entry CoreLogEntry) string {
	var parts []string

	// Add timestamp
	timestamp := entry.Timestamp.Format(w.config.TimestampFormat)
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))

	// Add level
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.LevelString)))

	// Add service name
	if entry.ServiceName != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.ServiceName))
	}

	// Add message
	parts = append(parts, entry.Message)

	// Add context if present
	if len(entry.Context) > 0 {
		contextStr := w.formatContext(entry.Context)
		parts = append(parts, fmt.Sprintf("{%s}", contextStr))
	}

	return strings.Join(parts, " ")
}

// formatContext formats context fields for remote output
func (w *RemoteWriter) formatContext(context map[string]interface{}) string {
	var pairs []string
	for k, v := range context {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(pairs, ", ")
}

// Close implements LogWriter interface
func (w *RemoteWriter) Close() error {
	close(w.stopCh)
	return nil
}

// Flush implements LogWriter interface
func (w *RemoteWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.buffer) > 0 {
		return w.sendBatch()
	}
	return nil
}

// SyslogWriter writes log entries to syslog (Unix systems only)
type SyslogWriter struct {
	config LoggerConfig
	tag    string
}

// NewSyslogWriter creates a new syslog writer
func NewSyslogWriter(config LoggerConfig, tag string) *SyslogWriter {
	return &SyslogWriter{
		config: config,
		tag:    tag,
	}
}

// Write implements LogWriter interface for syslog output
func (w *SyslogWriter) Write(entry CoreLogEntry) error {
	// Format message
	message := w.formatSyslogMessage(entry)

	// On Windows, just print to stderr with syslog format
	// On Unix systems, this would use the actual syslog
	fmt.Fprintf(os.Stderr, "[SYSLOG] %s: %s\n", w.tag, message)

	return nil
}

// formatSyslogMessage formats a log entry for syslog
func (w *SyslogWriter) formatSyslogMessage(entry CoreLogEntry) string {
	var parts []string

	// Add service name if available
	if entry.ServiceName != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.ServiceName))
	}

	// Add message
	parts = append(parts, entry.Message)

	// Add context if present
	if len(entry.Context) > 0 {
		contextStr := w.formatContext(entry.Context)
		parts = append(parts, fmt.Sprintf("{%s}", contextStr))
	}

	return strings.Join(parts, " ")
}

// formatContext formats context fields for syslog output
func (w *SyslogWriter) formatContext(context map[string]interface{}) string {
	var pairs []string
	for k, v := range context {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(pairs, ", ")
}

// Close implements LogWriter interface
func (w *SyslogWriter) Close() error {
	// Syslog writer doesn't need to close anything
	return nil
}

// Flush implements LogWriter interface
func (w *SyslogWriter) Flush() error {
	return nil
}

// NullWriter discards all log entries (useful for testing or conditional logging)
type NullWriter struct{}

// NewNullWriter creates a new null writer
func NewNullWriter() *NullWriter {
	return &NullWriter{}
}

// Write implements LogWriter interface for null output
func (w *NullWriter) Write(entry CoreLogEntry) error {
	// Do nothing - discard all entries
	return nil
}

// Close implements LogWriter interface
func (w *NullWriter) Close() error {
	// Null writer doesn't need to close anything
	return nil
}

// Flush implements LogWriter interface
func (w *NullWriter) Flush() error {
	return nil
}

// ConditionalWriter writes log entries based on a condition
type ConditionalWriter struct {
	writer    LogWriter
	condition func(CoreLogEntry) bool
}

// NewConditionalWriter creates a new conditional writer
func NewConditionalWriter(writer LogWriter, condition func(CoreLogEntry) bool) *ConditionalWriter {
	return &ConditionalWriter{
		writer:    writer,
		condition: condition,
	}
}

// Write implements LogWriter interface for conditional output
func (w *ConditionalWriter) Write(entry CoreLogEntry) error {
	if w.condition(entry) {
		return w.writer.Write(entry)
	}
	return nil
}

// Close implements LogWriter interface
func (w *ConditionalWriter) Close() error {
	return w.writer.Close()
}

// Flush implements LogWriter interface
func (w *ConditionalWriter) Flush() error {
	return w.writer.Flush()
}

// RateLimitedWriter limits the rate of log writes
type RateLimitedWriter struct {
	writer      LogWriter
	rateLimiter *time.Ticker
	lastWrite   time.Time
	minInterval time.Duration
}

// NewRateLimitedWriter creates a new rate-limited writer
func NewRateLimitedWriter(writer LogWriter, maxPerSecond int) *RateLimitedWriter {
	interval := time.Second / time.Duration(maxPerSecond)
	return &RateLimitedWriter{
		writer:      writer,
		rateLimiter: time.NewTicker(interval),
		minInterval: interval,
	}
}

// Write implements LogWriter interface for rate-limited output
func (w *RateLimitedWriter) Write(entry CoreLogEntry) error {
	// Check if enough time has passed since last write
	if time.Since(w.lastWrite) < w.minInterval {
		<-w.rateLimiter.C
	}

	w.lastWrite = time.Now()
	return w.writer.Write(entry)
}

// Close implements LogWriter interface
func (w *RateLimitedWriter) Close() error {
	w.rateLimiter.Stop()
	return w.writer.Close()
}

// Flush implements LogWriter interface
func (w *RateLimitedWriter) Flush() error {
	return w.writer.Flush()
}
