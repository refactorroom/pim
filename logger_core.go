package pim

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// SamplingConfig allows per-level and per-message sampling
// If both probability and rate are set, probability is used.
type SamplingConfig struct {
	EnableSampling bool
	SampleRate     float64 // Probability-based (0.0-1.0)
	Rate           int     // Log every Nth event (if > 0)
}

// LoggerCore is the main logger instance with extensible features
type LoggerCore struct {
	mu              sync.RWMutex
	level           LogLevel
	writers         []LogWriter
	hooks           []LogHook
	hookManager     *HookManager // Enhanced hook manager
	config          LoggerConfig
	context         map[string]interface{}
	hostname        string
	pid             int
	serviceName     string
	rateCounters    map[LogLevel]int     // for rate-based sampling
	themeManager    *ThemeManager        // Theme manager for formatting
	callerFormatter *CallerInfoFormatter // Enhanced caller info formatter

	// Async logging fields
	asyncBuffer chan CoreLogEntry
	asyncWorker *asyncWorker
	asyncCtx    context.Context
	asyncCancel context.CancelFunc
	asyncWg     sync.WaitGroup
}

// asyncWorker handles background log processing
type asyncWorker struct {
	logger *LoggerCore
	ctx    context.Context
}

// newAsyncWorker creates a new async worker
func newAsyncWorker(logger *LoggerCore, ctx context.Context) *asyncWorker {
	return &asyncWorker{
		logger: logger,
		ctx:    ctx,
	}
}

// start starts the async worker goroutine
func (w *asyncWorker) start() {
	w.logger.asyncWg.Add(1)
	go w.run()
}

// run is the main worker loop
func (w *asyncWorker) run() {
	defer w.logger.asyncWg.Done()

	ticker := time.NewTicker(w.logger.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case entry, ok := <-w.logger.asyncBuffer:
			if !ok {
				// Channel closed, flush remaining entries
				w.flushRemaining()
				return
			}
			w.processEntry(entry)

		case <-ticker.C:
			// Periodic flush
			w.flushBuffer()

		case <-w.ctx.Done():
			// Context cancelled, flush and exit
			w.flushRemaining()
			return
		}
	}
}

// processEntry processes a single log entry
func (w *asyncWorker) processEntry(entry CoreLogEntry) {
	w.logger.writeToWriters(entry)
}

// flushBuffer flushes all entries currently in the buffer
func (w *asyncWorker) flushBuffer() {
	for {
		select {
		case entry, ok := <-w.logger.asyncBuffer:
			if !ok {
				return
			}
			w.processEntry(entry)
		default:
			return
		}
	}
}

// flushRemaining flushes all remaining entries in the buffer
func (w *asyncWorker) flushRemaining() {
	for {
		select {
		case entry, ok := <-w.logger.asyncBuffer:
			if !ok {
				return
			}
			w.processEntry(entry)
		default:
			return
		}
	}
}

// CoreLogEntry represents a complete log entry with all metadata
type CoreLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	LevelString string                 `json:"level_string"`
	Message     string                 `json:"message"`
	Prefix      string                 `json:"prefix"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Function    string                 `json:"function,omitempty"`
	Package     string                 `json:"package,omitempty"`
	GoroutineID string                 `json:"goroutine_id,omitempty"`
	StackTrace  []StackFrame           `json:"stack_trace,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	ServiceName string                 `json:"service_name,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Hostname    string                 `json:"hostname,omitempty"`
	PID         int                    `json:"pid,omitempty"`
}

// LogWriter defines the interface for log output destinations
type LogWriter interface {
	Write(entry CoreLogEntry) error
	Close() error
	Flush() error
}

// LogHook defines the interface for log hooks (pre-processing)
type LogHook interface {
	Process(entry CoreLogEntry) (CoreLogEntry, error)
}

// LogHookFunc is a function type that implements LogHook
type LogHookFunc func(entry CoreLogEntry) (CoreLogEntry, error)

// Process implements LogHook interface for LogHookFunc
func (f LogHookFunc) Process(entry CoreLogEntry) (CoreLogEntry, error) {
	return f(entry)
}

// LoggerConfig holds all logger configuration
type LoggerConfig struct {
	// Basic settings
	Level           LogLevel `json:"level"`
	ServiceName     string   `json:"service_name"`
	TimestampFormat string   `json:"timestamp_format"`

	// Caller information (legacy - for backward compatibility)
	ShowFileLine     bool `json:"show_file_line"`
	ShowFunctionName bool `json:"show_function_name"`
	ShowPackageName  bool `json:"show_package_name"`
	ShowGoroutineID  bool `json:"show_goroutine_id"`
	ShowFullPath     bool `json:"show_full_path"`
	StackDepth       int  `json:"stack_depth"`

	// Enhanced caller information
	CallerInfoConfig CallerInfoConfig `json:"caller_info_config"`

	// Output settings
	EnableColors  bool `json:"enable_colors"`
	EnableJSON    bool `json:"enable_json"`
	EnableConsole bool `json:"enable_console"`

	// Theming and formatting
	ThemeName    string `json:"theme_name"`    // Name of the theme to use
	FormatName   string `json:"format_name"`   // Name of the format to use
	CustomTheme  *Theme `json:"custom_theme"`  // Custom theme (overrides ThemeName)
	CustomFormat string `json:"custom_format"` // Custom format template

	// Performance settings
	Async         bool          `json:"async"`
	BufferSize    int           `json:"buffer_size"`
	FlushInterval time.Duration `json:"flush_interval"`

	// Sampling
	EnableSampling  bool                        `json:"enable_sampling"`
	SampleRate      float64                     `json:"sample_rate"`
	SamplingByLevel map[LogLevel]SamplingConfig `json:"sampling_by_level"`

	// Context propagation
	PropagateContext bool `json:"propagate_context"`
}

// DefaultLoggerConfig provides sensible defaults
var DefaultLoggerConfig = LoggerConfig{
	Level:            InfoLevel,
	ServiceName:      "pim-logger",
	TimestampFormat:  "2006-01-02 15:04:05.000 UTC",
	ShowFileLine:     true,
	ShowFunctionName: true,
	ShowPackageName:  true,
	ShowGoroutineID:  true,
	ShowFullPath:     false,
	StackDepth:       3,
	EnableColors:     true,
	EnableJSON:       false,
	EnableConsole:    true,
	ThemeName:        "default",
	FormatName:       "colorful",
	Async:            false,
	BufferSize:       1000,
	FlushInterval:    5 * time.Second,
	EnableSampling:   false,
	SampleRate:       1.0,
	PropagateContext: true,
}

// NewLoggerCore creates a new logger instance with the given configuration
func NewLoggerCore(config LoggerConfig) *LoggerCore {
	// Initialize caller info config with legacy settings if not provided
	if !config.CallerInfoConfig.Enabled && config.CallerInfoConfig.Format == "" {
		config.CallerInfoConfig = NewCallerInfoConfig()
		// Map legacy settings to new config
		config.CallerInfoConfig.ShowFile = config.ShowFileLine
		config.CallerInfoConfig.ShowLine = config.ShowFileLine
		config.CallerInfoConfig.ShowFunction = config.ShowFunctionName
		config.CallerInfoConfig.ShowPackage = config.ShowPackageName
		config.CallerInfoConfig.ShowGoroutineID = config.ShowGoroutineID
		config.CallerInfoConfig.ShowFullPath = config.ShowFullPath
		config.CallerInfoConfig.StackDepth = config.StackDepth
	}

	hostname, _ := os.Hostname()

	// Initialize caller formatter
	callerFormatter := NewCallerInfoFormatter(config.CallerInfoConfig)

	logger := &LoggerCore{
		level:           config.Level,
		writers:         make([]LogWriter, 0),
		hooks:           make([]LogHook, 0),
		hookManager:     NewHookManager(),
		config:          config,
		context:         make(map[string]interface{}),
		hostname:        hostname,
		pid:             os.Getpid(),
		serviceName:     config.ServiceName,
		rateCounters:    make(map[LogLevel]int),
		themeManager:    NewThemeManager(),
		callerFormatter: callerFormatter,
	}

	// Initialize theme manager
	if config.CustomTheme != nil {
		logger.themeManager.currentTheme = config.CustomTheme
	} else if config.ThemeName != "" {
		logger.themeManager.SetTheme(config.ThemeName)
	}

	// Register custom format if provided
	if config.CustomFormat != "" {
		logger.themeManager.RegisterTemplate("custom", config.CustomFormat)
	}

	// Initialize async logging if enabled
	if config.Async {
		logger.asyncCtx, logger.asyncCancel = context.WithCancel(context.Background())
		logger.asyncBuffer = make(chan CoreLogEntry, config.BufferSize)
		logger.asyncWorker = newAsyncWorker(logger, logger.asyncCtx)
		logger.asyncWorker.start()
	}

	// Add default console writer if enabled
	if config.EnableConsole {
		logger.AddWriter(NewConsoleWriter(config))
	}

	RegisterLoggerForShutdown(logger)

	return logger
}

// AddWriter adds a new log writer
func (l *LoggerCore) AddWriter(writer LogWriter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writers = append(l.writers, writer)
}

// RemoveWriter removes a log writer by index
func (l *LoggerCore) RemoveWriter(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index >= 0 && index < len(l.writers) {
		l.writers = append(l.writers[:index], l.writers[index+1:]...)
	}
}

// AddHook adds a new log hook
func (l *LoggerCore) AddHook(hook LogHook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// AddHookFunc adds a function as a log hook
func (l *LoggerCore) AddHookFunc(fn LogHookFunc) {
	l.AddHook(fn)
}

// RemoveHook removes a log hook by index
func (l *LoggerCore) RemoveHook(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index >= 0 && index < len(l.hooks) {
		l.hooks = append(l.hooks[:index], l.hooks[index+1:]...)
	}
}

// SetLevel sets the log level
func (l *LoggerCore) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *LoggerCore) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetContext sets a context value that will be included in all log entries
func (l *LoggerCore) SetContext(key string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.context[key] = value
}

// GetContext returns a context value
func (l *LoggerCore) GetContext(key string) interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.context[key]
}

// ClearContext removes all context values
func (l *LoggerCore) ClearContext() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.context = make(map[string]interface{})
}

// Log creates and writes a log entry
func (l *LoggerCore) Log(level LogLevel, prefix, message string, args ...interface{}) {
	if level > l.GetLevel() {
		return
	}

	// Apply sampling if enabled
	if !l.shouldSampleLevel(level) {
		return
	}

	// Format message with args
	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(message, args...)
	}

	// Create log entry
	entry := l.createLogEntry(level, prefix, formattedMessage)

	// Apply hooks
	entry = l.applyHooks(entry)

	// Check if entry was filtered out
	if entry.Message == "" && entry.Level == 0 {
		return // Entry was filtered, don't log
	}

	// Write to all writers (async or sync)
	if l.config.Async {
		select {
		case l.asyncBuffer <- entry:
			// Successfully queued
		default:
			// Buffer full, fall back to synchronous logging
			l.writeToWriters(entry)
		}
	} else {
		l.writeToWriters(entry)
	}
}

// LogWithContext creates and writes a log entry with additional context
func (l *LoggerCore) LogWithContext(level LogLevel, prefix, message string, context map[string]interface{}, args ...interface{}) {
	if level > l.GetLevel() {
		return
	}

	// Apply sampling if enabled
	if !l.shouldSampleLevel(level) {
		return
	}

	// Format message with args
	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(message, args...)
	}

	// Create log entry
	entry := l.createLogEntry(level, prefix, formattedMessage)

	// Add context
	if context != nil {
		if entry.Context == nil {
			entry.Context = make(map[string]interface{})
		}
		for k, v := range context {
			entry.Context[k] = v
		}
	}

	// Apply hooks
	entry = l.applyHooks(entry)

	// Check if entry was filtered out
	if entry.Message == "" && entry.Level == 0 {
		return // Entry was filtered, don't log
	}

	// Write to all writers (async or sync)
	if l.config.Async {
		select {
		case l.asyncBuffer <- entry:
			// Successfully queued
		default:
			// Buffer full, fall back to synchronous logging
			l.writeToWriters(entry)
		}
	} else {
		l.writeToWriters(entry)
	}
}

// LogWithStackTrace creates and writes a log entry with stack trace
func (l *LoggerCore) LogWithStackTrace(level LogLevel, prefix, message string, args ...interface{}) {
	if level > l.GetLevel() {
		return
	}

	// Apply sampling if enabled
	if !l.shouldSampleLevel(level) {
		return
	}

	// Format message with args
	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(message, args...)
	}

	// Create log entry with stack trace
	entry := l.createLogEntry(level, prefix, formattedMessage)

	// Get stack trace using enhanced formatter
	if l.callerFormatter != nil {
		callerFrames := l.callerFormatter.GetStackTrace(4) // Skip 4 frames
		// Convert CallerInfo to StackFrame for compatibility
		for _, frame := range callerFrames {
			entry.StackTrace = append(entry.StackTrace, StackFrame{
				File:     frame.File,
				Line:     frame.Line,
				Function: frame.Function,
				Package:  frame.Package,
			})
		}
	} else {
		// Fallback to legacy method
		entry.StackTrace = l.getStackTrace(4) // Skip 4 frames
	}

	// Apply hooks
	entry = l.applyHooks(entry)

	// Check if entry was filtered out
	if entry.Message == "" && entry.Level == 0 {
		return // Entry was filtered, don't log
	}

	// Write to all writers (async or sync)
	if l.config.Async {
		select {
		case l.asyncBuffer <- entry:
			// Successfully queued
		default:
			// Buffer full, fall back to synchronous logging
			l.writeToWriters(entry)
		}
	} else {
		l.writeToWriters(entry)
	}
}

// createLogEntry creates a new log entry with all metadata
func (l *LoggerCore) createLogEntry(level LogLevel, prefix, message string) CoreLogEntry {
	now := time.Now().UTC()

	entry := CoreLogEntry{
		Timestamp:   now,
		Level:       level,
		LevelString: l.getLevelString(level),
		Message:     message,
		Prefix:      prefix,
		ServiceName: l.serviceName,
		Hostname:    l.hostname,
		PID:         l.pid,
	}

	// Add caller information using enhanced formatter
	if l.callerFormatter != nil {
		callerInfo := l.callerFormatter.GetCallerInfo(3) // Skip 3 frames
		entry.File = callerInfo.File
		entry.Line = callerInfo.Line
		entry.Function = callerInfo.Function
		entry.Package = callerInfo.Package
		entry.GoroutineID = callerInfo.GoroutineID
	} else {
		// Fallback to legacy method
		if l.config.ShowFileLine {
			callInfo := l.getCallInfo(3) // Skip 3 frames
			entry.File = callInfo.File
			entry.Line = callInfo.Line
			entry.Function = callInfo.Function
			entry.Package = callInfo.Package
		}

		// Add goroutine ID
		if l.config.ShowGoroutineID {
			entry.GoroutineID = l.getGoroutineID()
		}
	}

	// Add global context
	if l.config.PropagateContext && len(l.context) > 0 {
		entry.Context = make(map[string]interface{})
		for k, v := range l.context {
			entry.Context[k] = v
		}
		// Propagate trace/span/request/session/correlation IDs if present
		if v, ok := l.context["trace_id"]; ok {
			if s, ok := v.(string); ok {
				entry.TraceID = s
			}
		}
		if v, ok := l.context["span_id"]; ok {
			if s, ok := v.(string); ok {
				entry.SpanID = s
			}
		}
		if v, ok := l.context["request_id"]; ok {
			if s, ok := v.(string); ok {
				entry.RequestID = s
			}
		}
		if v, ok := l.context["session_id"]; ok {
			if s, ok := v.(string); ok {
				entry.SessionID = s
			}
		}
		if v, ok := l.context["correlation_id"]; ok {
			if s, ok := v.(string); ok {
				// Prefer to set TraceID if not already set
				if entry.TraceID == "" {
					entry.TraceID = s
				}
				// Or set as RequestID if not already set
				if entry.RequestID == "" {
					entry.RequestID = s
				}
			}
		}
	}

	return entry
}

// applyHooks applies all registered hooks to the log entry
func (l *LoggerCore) applyHooks(entry CoreLogEntry) CoreLogEntry {
	// Apply legacy hooks first
	l.mu.RLock()
	hooks := make([]LogHook, len(l.hooks))
	copy(hooks, l.hooks)
	l.mu.RUnlock()

	for _, hook := range hooks {
		if modifiedEntry, err := hook.Process(entry); err == nil {
			entry = modifiedEntry
		}
	}

	// Apply enhanced hooks
	if l.hookManager != nil {
		if modifiedEntry, err := l.hookManager.ProcessHooks(entry); err == nil {
			entry = modifiedEntry
		} else if strings.Contains(err.Error(), "filtered by hook") {
			// Return filtered entry to indicate filtering
			return CoreLogEntry{
				Message: "", // Empty message indicates filtering
				Level:   0,  // Zero level indicates filtering
			}
		}
	}

	return entry
}

// writeToWriters writes the log entry to all registered writers
func (l *LoggerCore) writeToWriters(entry CoreLogEntry) {
	l.mu.RLock()
	writers := make([]LogWriter, len(l.writers))
	copy(writers, l.writers)
	l.mu.RUnlock()

	for _, writer := range writers {
		if err := writer.Write(entry); err != nil {
			// Log writer errors to stderr to avoid infinite loops
			fmt.Fprintf(os.Stderr, "Failed to write log entry: %v\n", err)
		}
	}
}

// shouldSample determines if this log entry should be sampled
func (l *LoggerCore) shouldSampleLevel(level LogLevel) bool {
	cfg, ok := l.config.SamplingByLevel[level]
	if !ok {
		cfg = SamplingConfig{
			EnableSampling: l.config.EnableSampling,
			SampleRate:     l.config.SampleRate,
			Rate:           0,
		}
	}
	if !cfg.EnableSampling {
		return true
	}
	if cfg.SampleRate > 0.0 && cfg.SampleRate < 1.0 {
		return time.Now().UnixNano()%100 < int64(cfg.SampleRate*100)
	}
	if cfg.Rate > 1 {
		// Use an atomic counter per level
		return l.incrementAndCheckRate(level, cfg.Rate)
	}
	return true
}

// incrementAndCheckRate increments a counter and returns true if this event should be logged
func (l *LoggerCore) incrementAndCheckRate(level LogLevel, rate int) bool {
	l.mu.Lock()
	if l.rateCounters == nil {
		l.rateCounters = make(map[LogLevel]int)
	}
	l.rateCounters[level]++
	count := l.rateCounters[level]
	l.mu.Unlock()
	return count%rate == 0
}

// getCallInfo retrieves detailed information about the calling function
func (l *LoggerCore) getCallInfo(skip int) CallInfo {
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
	if !l.config.ShowFullPath {
		file = filepath.Base(file)
	}

	return CallInfo{
		File:     file,
		Line:     line,
		Function: functionName,
		Package:  packageName,
	}
}

// getStackTrace returns a formatted stack trace
func (l *LoggerCore) getStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	for i := skip; i < skip+l.config.StackDepth; i++ {
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

		if !l.config.ShowFullPath {
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

// getGoroutineID returns the current goroutine ID
func (l *LoggerCore) getGoroutineID() string {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	id := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	return fmt.Sprintf("(goroutine %s)", id)
}

// getLevelString returns the string representation of a log level
func (l *LoggerCore) getLevelString(level LogLevel) string {
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

// Flush flushes all buffered log entries and writers
func (l *LoggerCore) Flush() {
	if l.config.Async && l.asyncWorker != nil {
		l.asyncCancel()
		l.asyncWg.Wait()
	}
	// Flush all writers
	l.mu.RLock()
	for _, writer := range l.writers {
		if err := writer.Flush(); err != nil {
			// Log the error but continue flushing other writers
			fmt.Fprintf(os.Stderr, "Error flushing writer: %v\n", err)
		}
	}
	l.mu.RUnlock()
}

// Close closes all writers and stops async logging
func (l *LoggerCore) Close() error {
	l.Flush()
	l.mu.Lock()
	defer l.mu.Unlock()

	var errors []error
	for _, writer := range l.writers {
		if err := writer.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing writers: %v", errors)
	}
	return nil
}

// --- Global graceful shutdown support ---
var (
	globalLoggersMu sync.Mutex
	globalLoggers   []*LoggerCore
)

// RegisterLoggerForShutdown registers a logger for global shutdown handling
func RegisterLoggerForShutdown(logger *LoggerCore) {
	globalLoggersMu.Lock()
	defer globalLoggersMu.Unlock()
	globalLoggers = append(globalLoggers, logger)
}

// InstallExitHandler installs a handler to flush/close all loggers on exit/panic/signals
func InstallExitHandler() {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		FlushAllLoggers()
		os.Exit(1)
	}()

	// Recover from panic and flush logs
	go func() {
		if r := recover(); r != nil {
			FlushAllLoggers()
			panic(r)
		}
	}()
}

// FlushAllLoggers flushes and closes all registered loggers
func FlushAllLoggers() {
	globalLoggersMu.Lock()
	defer globalLoggersMu.Unlock()
	for _, logger := range globalLoggers {
		logger.Flush()
		logger.Close()
	}
}

// Convenience methods for different log levels
func (l *LoggerCore) Trace(msg string, args ...interface{}) {
	l.Log(TraceLevel, TracePrefix, msg, args...)
}

func (l *LoggerCore) Debug(msg string, args ...interface{}) {
	l.Log(DebugLevel, DebugPrefix, msg, args...)
}

func (l *LoggerCore) Info(msg string, args ...interface{}) {
	l.Log(InfoLevel, InfoPrefix, msg, args...)
}

func (l *LoggerCore) Success(msg string, args ...interface{}) {
	l.Log(InfoLevel, SuccessPrefix, msg, args...)
}

func (l *LoggerCore) Init(msg string, args ...interface{}) {
	l.Log(InfoLevel, InitPrefix, msg, args...)
}

func (l *LoggerCore) Config(msg string, args ...interface{}) {
	l.Log(InfoLevel, ConfigPrefix, msg, args...)
}

func (l *LoggerCore) Warning(msg string, args ...interface{}) {
	l.Log(WarningLevel, WarningPrefix, msg, args...)
}

func (l *LoggerCore) Error(msg string, args ...interface{}) {
	l.LogWithStackTrace(ErrorLevel, ErrorPrefix, msg, args...)
}

func (l *LoggerCore) Panic(msg string, args ...interface{}) {
	l.LogWithStackTrace(PanicLevel, PanicPrefix, msg, args...)
	panic(fmt.Sprintf(msg, args...))
}

func (l *LoggerCore) Metric(name string, value interface{}, tags ...string) {
	tagStr := ""
	if len(tags) > 0 {
		tagStr = fmt.Sprintf(" [%s]", strings.Join(tags, ", "))
	}
	logMsg := fmt.Sprintf("%s: %v%s", name, value, tagStr)
	l.Log(InfoLevel, MetricPrefix, logMsg)
}

// WithContext returns a new logger with additional context
func (l *LoggerCore) WithContext(ctx map[string]interface{}) *LoggerCore {
	newLogger := &LoggerCore{
		level:        l.level,
		writers:      l.writers,
		hooks:        l.hooks,
		config:       l.config,
		context:      make(map[string]interface{}),
		hostname:     l.hostname,
		pid:          l.pid,
		serviceName:  l.serviceName,
		rateCounters: make(map[LogLevel]int),
	}

	// Copy existing context
	for k, v := range l.context {
		newLogger.context[k] = v
	}

	// Add new context
	for k, v := range ctx {
		newLogger.context[k] = v
	}

	return newLogger
}

// WithField returns a new logger with a single additional context field
func (l *LoggerCore) WithField(key string, value interface{}) *LoggerCore {
	return l.WithContext(map[string]interface{}{key: value})
}

// WithFields returns a new logger with multiple additional context fields
func (l *LoggerCore) WithFields(fields map[string]interface{}) *LoggerCore {
	return l.WithContext(fields)
}

// InfoWithFields logs an info message with structured fields
func (l *LoggerCore) InfoWithFields(msg string, fields map[string]interface{}) {
	l.LogWithContext(InfoLevel, InfoPrefix, msg, fields)
}

// DebugWithFields logs a debug message with structured fields
func (l *LoggerCore) DebugWithFields(msg string, fields map[string]interface{}) {
	l.LogWithContext(DebugLevel, DebugPrefix, msg, fields)
}

// WarningWithFields logs a warning message with structured fields
func (l *LoggerCore) WarningWithFields(msg string, fields map[string]interface{}) {
	l.LogWithContext(WarningLevel, WarningPrefix, msg, fields)
}

// ErrorWithFields logs an error message with structured fields
func (l *LoggerCore) ErrorWithFields(msg string, fields map[string]interface{}) {
	l.LogWithContext(ErrorLevel, ErrorPrefix, msg, fields)
}

// TraceWithFields logs a trace message with structured fields
func (l *LoggerCore) TraceWithFields(msg string, fields map[string]interface{}) {
	l.LogWithContext(TraceLevel, TracePrefix, msg, fields)
}

// SuccessWithFields logs a success message with structured fields
func (l *LoggerCore) SuccessWithFields(msg string, fields map[string]interface{}) {
	l.LogWithContext(InfoLevel, SuccessPrefix, msg, fields)
}

// InfoKV logs an info message with variadic key-value pairs
func (l *LoggerCore) InfoKV(msg string, kv ...interface{}) {
	l.LogWithContext(InfoLevel, InfoPrefix, msg, kvToMap(kv...))
}

// DebugKV logs a debug message with variadic key-value pairs
func (l *LoggerCore) DebugKV(msg string, kv ...interface{}) {
	l.LogWithContext(DebugLevel, DebugPrefix, msg, kvToMap(kv...))
}

// WarningKV logs a warning message with variadic key-value pairs
func (l *LoggerCore) WarningKV(msg string, kv ...interface{}) {
	l.LogWithContext(WarningLevel, WarningPrefix, msg, kvToMap(kv...))
}

// ErrorKV logs an error message with variadic key-value pairs
func (l *LoggerCore) ErrorKV(msg string, kv ...interface{}) {
	l.LogWithContext(ErrorLevel, ErrorPrefix, msg, kvToMap(kv...))
}

// TraceKV logs a trace message with variadic key-value pairs
func (l *LoggerCore) TraceKV(msg string, kv ...interface{}) {
	l.LogWithContext(TraceLevel, TracePrefix, msg, kvToMap(kv...))
}

// SuccessKV logs a success message with variadic key-value pairs
func (l *LoggerCore) SuccessKV(msg string, kv ...interface{}) {
	l.LogWithContext(InfoLevel, SuccessPrefix, msg, kvToMap(kv...))
}

// kvToMap converts variadic key-value pairs to a map[string]interface{}
func kvToMap(kv ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(kv)-1; i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		fields[key] = kv[i+1]
	}
	return fields
}

// SetLevelFromString sets the log level from a string (e.g., "info", "debug")
func (l *LoggerCore) SetLevelFromString(levelStr string) bool {
	levelStr = strings.ToLower(strings.TrimSpace(levelStr))
	var level LogLevel
	switch levelStr {
	case "panic":
		level = PanicLevel
	case "error":
		level = ErrorLevel
	case "warning", "warn":
		level = WarningLevel
	case "info":
		level = InfoLevel
	case "debug":
		level = DebugLevel
	case "trace":
		level = TraceLevel
	default:
		return false
	}
	l.SetLevel(level)
	return true
}

// SetLevelFromEnv sets the log level from an environment variable (e.g., PIM_LOG_LEVEL)
func (l *LoggerCore) SetLevelFromEnv(envVar string) bool {
	levelStr := os.Getenv(envVar)
	if levelStr == "" {
		return false
	}
	return l.SetLevelFromString(levelStr)
}

// WatchLevelFile watches a file for log level changes (e.g., for dynamic config reload)
func (l *LoggerCore) WatchLevelFile(filePath string, pollInterval time.Duration, stopCh <-chan struct{}) {
	go func() {
		var lastLevel string
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(pollInterval):
				data, err := os.ReadFile(filePath)
				if err != nil {
					continue
				}
				levelStr := strings.TrimSpace(string(data))
				if levelStr != "" && levelStr != lastLevel {
					if l.SetLevelFromString(levelStr) {
						lastLevel = levelStr
					}
				}
			}
		}
	}()
}

// SetSamplingByLevel sets the per-level sampling configuration
func (l *LoggerCore) SetSamplingByLevel(sampling map[LogLevel]SamplingConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.SamplingByLevel = sampling
}

// SetTheme sets the theme for the logger
func (l *LoggerCore) SetTheme(themeName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.themeManager.SetTheme(themeName)
}

// GetTheme returns the current theme
func (l *LoggerCore) GetTheme() *Theme {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.themeManager.GetTheme()
}

// SetCustomTheme sets a custom theme
func (l *LoggerCore) SetCustomTheme(theme *Theme) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.themeManager.currentTheme = theme
}

// RegisterTemplate registers a custom template
func (l *LoggerCore) RegisterTemplate(name, templateStr string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.themeManager.RegisterTemplate(name, templateStr)
}

// RegisterFormatter registers a custom formatter
func (l *LoggerCore) RegisterFormatter(name string, formatter LogFormatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.themeManager.RegisterFormatter(name, formatter)
}

// Format formats a log entry using the current theme and format
func (l *LoggerCore) Format(entry CoreLogEntry, formatName string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.themeManager.Format(entry, formatName)
}

// Enhanced caller information methods

// SetCallerInfoConfig updates the caller information configuration
func (l *LoggerCore) SetCallerInfoConfig(config CallerInfoConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.CallerInfoConfig = config
	if l.callerFormatter != nil {
		l.callerFormatter.SetConfig(config)
	}
}

// GetCallerInfoConfig returns the current caller information configuration
func (l *LoggerCore) GetCallerInfoConfig() CallerInfoConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.CallerInfoConfig
}

// GetCallerInfo retrieves caller information using the enhanced formatter
func (l *LoggerCore) GetCallerInfo(skip int) CallerInfo {
	if l.callerFormatter != nil {
		return l.callerFormatter.GetCallerInfo(skip)
	}
	// Fallback to legacy method
	callInfo := l.getCallInfo(skip)
	return CallerInfo{
		File:     callInfo.File,
		Line:     callInfo.Line,
		Function: callInfo.Function,
		Package:  callInfo.Package,
	}
}

// GetCallerInfoAtDepth retrieves caller information at a specific depth
func (l *LoggerCore) GetCallerInfoAtDepth(depth int) CallerInfo {
	if l.callerFormatter != nil {
		return l.callerFormatter.GetCallerInfoAtDepth(depth)
	}
	// Fallback to legacy method
	callInfo := l.getCallInfo(depth)
	return CallerInfo{
		File:     callInfo.File,
		Line:     callInfo.Line,
		Function: callInfo.Function,
		Package:  callInfo.Package,
	}
}

// GetEnhancedStackTrace retrieves a stack trace using the enhanced formatter
func (l *LoggerCore) GetEnhancedStackTrace(skip int) []CallerInfo {
	if l.callerFormatter != nil {
		return l.callerFormatter.GetStackTrace(skip)
	}
	// Fallback to legacy method
	legacyFrames := l.getStackTrace(skip)
	var frames []CallerInfo
	for _, frame := range legacyFrames {
		frames = append(frames, CallerInfo{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
			Package:  frame.Package,
		})
	}
	return frames
}

// FormatCallerInfo formats caller information according to configuration
func (l *LoggerCore) FormatCallerInfo(info CallerInfo) string {
	if l.callerFormatter != nil {
		return l.callerFormatter.Format(info)
	}
	// Fallback to simple format
	var parts []string
	if info.File != "" {
		parts = append(parts, info.File)
	}
	if info.Line > 0 {
		parts = append(parts, fmt.Sprintf("L%d", info.Line))
	}
	if info.Function != "" {
		parts = append(parts, info.Function)
	}
	return strings.Join(parts, ":")
}

// FormatStackTrace formats a stack trace using the enhanced formatter
func (l *LoggerCore) FormatStackTrace(frames []CallerInfo) string {
	if l.callerFormatter != nil {
		return l.callerFormatter.FormatStackTrace(frames)
	}
	// Fallback to simple format
	var lines []string
	for i, frame := range frames {
		indent := strings.Repeat("  ", i)
		formatted := l.FormatCallerInfo(frame)
		line := fmt.Sprintf("%s↳ %s", indent, formatted)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// ClearCallerCache clears the caller formatter cache
func (l *LoggerCore) ClearCallerCache() {
	if l.callerFormatter != nil {
		l.callerFormatter.ClearCache()
	}
}

// GetCallerCacheStats returns caller formatter cache statistics
func (l *LoggerCore) GetCallerCacheStats() map[string]interface{} {
	if l.callerFormatter != nil {
		return l.callerFormatter.GetCacheStats()
	}
	return map[string]interface{}{
		"enabled": false,
		"reason":  "caller formatter not initialized",
	}
}

// Enhanced hooks management methods

// AddEnhancedHook adds an enhanced hook to the logger
func (l *LoggerCore) AddEnhancedHook(hook EnhancedLogHook) {
	if l.hookManager != nil {
		l.hookManager.AddHook(hook)
	}
}

// RemoveEnhancedHook removes an enhanced hook by name
func (l *LoggerCore) RemoveEnhancedHook(name string) {
	if l.hookManager != nil {
		l.hookManager.RemoveHook(name)
	}
}

// GetEnhancedHook returns an enhanced hook by name
func (l *LoggerCore) GetEnhancedHook(name string) EnhancedLogHook {
	if l.hookManager != nil {
		return l.hookManager.GetHook(name)
	}
	return nil
}

// GetEnhancedHooksByType returns all enhanced hooks of a specific type
func (l *LoggerCore) GetEnhancedHooksByType(hookType HookType) []EnhancedLogHook {
	if l.hookManager != nil {
		return l.hookManager.GetHooksByType(hookType)
	}
	return []EnhancedLogHook{}
}

// SetHookManagerEnabled enables or disables the hook manager
func (l *LoggerCore) SetHookManagerEnabled(enabled bool) {
	if l.hookManager != nil {
		l.hookManager.SetEnabled(enabled)
	}
}

// IsHookManagerEnabled returns whether the hook manager is enabled
func (l *LoggerCore) IsHookManagerEnabled() bool {
	if l.hookManager != nil {
		return l.hookManager.IsEnabled()
	}
	return false
}

// GetHookCount returns the total number of hooks (legacy + enhanced)
func (l *LoggerCore) GetHookCount() int {
	count := len(l.hooks)
	if l.hookManager != nil {
		count += l.hookManager.GetHookCount()
	}
	return count
}

// GetMetricsHook returns the metrics hook if it exists
func (l *LoggerCore) GetMetricsHook() *MetricsHook {
	if l.hookManager != nil {
		hooks := l.hookManager.GetHooksByType(HookTypeMetrics)
		for _, hook := range hooks {
			if metricsHook, ok := hook.(*MetricsHook); ok {
				return metricsHook
			}
		}
	}
	return nil
}

// GetMetrics returns metrics from the metrics hook
func (l *LoggerCore) GetMetrics() map[string]interface{} {
	if metricsHook := l.GetMetricsHook(); metricsHook != nil {
		return metricsHook.GetMetrics()
	}
	return map[string]interface{}{
		"error": "no metrics hook found",
	}
}

// ResetMetrics resets metrics in the metrics hook
func (l *LoggerCore) ResetMetrics() {
	if metricsHook := l.GetMetricsHook(); metricsHook != nil {
		metricsHook.ResetMetrics()
	}
}

// Convenience methods for common hooks

// AddSensitiveDataRedactHook adds a hook to redact sensitive data
func (l *LoggerCore) AddSensitiveDataRedactHook() {
	l.AddEnhancedHook(NewSensitiveDataRedactHook())
}

// AddRequestIDEnrichHook adds a hook to enrich with request IDs
func (l *LoggerCore) AddRequestIDEnrichHook() {
	l.AddEnhancedHook(NewRequestIDEnrichHook())
}

// AddMetricsHook adds a hook to collect metrics
func (l *LoggerCore) AddMetricsHook() {
	l.AddEnhancedHook(NewMetricsHook(MetricsConfig{
		HookConfig: HookConfig{
			Type:        HookTypeMetrics,
			Name:        "metrics_collector",
			Description: "Collects metrics about log entries",
			Enabled:     true,
			Priority:    100,
		},
	}))
}

// AddDebugFilterHook adds a hook to filter debug messages
func (l *LoggerCore) AddDebugFilterHook() {
	l.AddEnhancedHook(NewDebugFilterHook())
}

// WithTrace returns a new logger with the given trace ID set in context
func (l *LoggerCore) WithTrace(traceID string) *LoggerCore {
	return l.WithContext(map[string]interface{}{"trace_id": traceID})
}

// WithSpan returns a new logger with the given span ID set in context
func (l *LoggerCore) WithSpan(spanID string) *LoggerCore {
	return l.WithContext(map[string]interface{}{"span_id": spanID})
}

// WithRequestID returns a new logger with the given request ID set in context
func (l *LoggerCore) WithRequestID(requestID string) *LoggerCore {
	return l.WithContext(map[string]interface{}{"request_id": requestID})
}

// WithCorrelationID returns a new logger with the given correlation ID set in context (alias for request/session/trace)
func (l *LoggerCore) WithCorrelationID(correlationID string) *LoggerCore {
	return l.WithContext(map[string]interface{}{"correlation_id": correlationID})
}

// WithContextFromContext extracts trace/span/request/session IDs from context.Context and returns a new logger with them set
func (l *LoggerCore) WithContextFromContext(ctx context.Context) *LoggerCore {
	fields := map[string]interface{}{}
	if v := ctx.Value("trace_id"); v != nil {
		fields["trace_id"] = v
	}
	if v := ctx.Value("span_id"); v != nil {
		fields["span_id"] = v
	}
	if v := ctx.Value("request_id"); v != nil {
		fields["request_id"] = v
	}
	if v := ctx.Value("session_id"); v != nil {
		fields["session_id"] = v
	}
	if v := ctx.Value("correlation_id"); v != nil {
		fields["correlation_id"] = v
	}
	if len(fields) == 0 {
		return l
	}
	return l.WithContext(fields)
}
