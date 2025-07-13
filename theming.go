package pim

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
)

// Theme defines a complete color theme for log output
type Theme struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Colors      ThemeColors       `json:"colors"`
	Styles      ThemeStyles       `json:"styles"`
	Icons       ThemeIcons        `json:"icons"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// ThemeColors defines colors for different log elements
type ThemeColors struct {
	// Level-specific colors
	Panic   *color.Color `json:"-"`
	Error   *color.Color `json:"-"`
	Warning *color.Color `json:"-"`
	Info    *color.Color `json:"-"`
	Success *color.Color `json:"-"`
	Debug   *color.Color `json:"-"`
	Trace   *color.Color `json:"-"`
	Config  *color.Color `json:"-"`

	// Element colors
	Timestamp *color.Color `json:"-"`
	Service   *color.Color `json:"-"`
	File      *color.Color `json:"-"`
	Function  *color.Color `json:"-"`
	Package   *color.Color `json:"-"`
	Goroutine *color.Color `json:"-"`
	Message   *color.Color `json:"-"`
	Context   *color.Color `json:"-"`
	Key       *color.Color `json:"-"`
	Value     *color.Color `json:"-"`
	Bracket   *color.Color `json:"-"`
	Separator *color.Color `json:"-"`

	// Background colors
	Background *color.Color `json:"-"`
	Highlight  *color.Color `json:"-"`
}

// ThemeStyles defines text styles for different elements
type ThemeStyles struct {
	Timestamp string `json:"timestamp"` // bold, italic, underline
	Service   string `json:"service"`   // bold, italic, underline
	File      string `json:"file"`      // bold, italic, underline
	Function  string `json:"function"`  // bold, italic, underline
	Package   string `json:"package"`   // bold, italic, underline
	Message   string `json:"message"`   // bold, italic, underline
	Context   string `json:"context"`   // bold, italic, underline
}

// ThemeIcons defines icons for different log levels
type ThemeIcons struct {
	Panic   string `json:"panic"`
	Error   string `json:"error"`
	Warning string `json:"warning"`
	Info    string `json:"info"`
	Success string `json:"success"`
	Debug   string `json:"debug"`
	Trace   string `json:"trace"`
	Config  string `json:"config"`
}

// FormatTemplate defines a custom log format template
type FormatTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Template    string `json:"template"`
	Example     string `json:"example"`
}

// TemplateData provides data for template rendering
type TemplateData struct {
	Timestamp   time.Time
	Level       string
	LevelString string
	Message     string
	Service     string
	File        string
	Line        int
	Function    string
	Package     string
	GoroutineID string
	Context     map[string]interface{}
	Hostname    string
	PID         int
	TraceID     string
	SpanID      string
	UserID      string
	RequestID   string
	SessionID   string
}

// ThemeManager manages themes and formatting
type ThemeManager struct {
	currentTheme *Theme
	templates    map[string]*template.Template
	formatters   map[string]LogFormatter
}

// LogFormatter is a function that formats a log entry
type LogFormatter func(entry CoreLogEntry, theme *Theme) string

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		templates:  make(map[string]*template.Template),
		formatters: make(map[string]LogFormatter),
	}

	// Register built-in themes
	tm.registerBuiltinThemes()

	// Register built-in templates
	tm.registerBuiltinTemplates()

	// Register built-in formatters
	tm.registerBuiltinFormatters()

	// Set default theme
	tm.SetTheme("default")

	return tm
}

// SetTheme sets the current theme
func (tm *ThemeManager) SetTheme(name string) error {
	theme, exists := builtinThemes[name]
	if !exists {
		return fmt.Errorf("theme '%s' not found", name)
	}
	tm.currentTheme = &theme
	return nil
}

// GetTheme returns the current theme
func (tm *ThemeManager) GetTheme() *Theme {
	return tm.currentTheme
}

// RegisterTemplate registers a custom template
func (tm *ThemeManager) RegisterTemplate(name, templateStr string) error {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template '%s': %w", name, err)
	}
	tm.templates[name] = tmpl
	return nil
}

// RegisterFormatter registers a custom formatter
func (tm *ThemeManager) RegisterFormatter(name string, formatter LogFormatter) {
	tm.formatters[name] = formatter
}

// Format formats a log entry using the current theme and specified format
func (tm *ThemeManager) Format(entry CoreLogEntry, formatName string) string {
	// Check if we have a custom formatter
	if formatter, exists := tm.formatters[formatName]; exists {
		return formatter(entry, tm.currentTheme)
	}

	// Check if we have a template
	if tmpl, exists := tm.templates[formatName]; exists {
		data := tm.entryToTemplateData(entry)
		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Sprintf("template error: %v", err)
		}
		return buf.String()
	}

	// Use default formatter
	return tm.defaultFormatter(entry, tm.currentTheme)
}

// entryToTemplateData converts a CoreLogEntry to TemplateData
func (tm *ThemeManager) entryToTemplateData(entry CoreLogEntry) TemplateData {
	return TemplateData{
		Timestamp:   entry.Timestamp,
		Level:       entry.LevelString,
		LevelString: entry.LevelString,
		Message:     entry.Message,
		Service:     entry.ServiceName,
		File:        entry.File,
		Line:        entry.Line,
		Function:    entry.Function,
		Package:     entry.Package,
		GoroutineID: entry.GoroutineID,
		Context:     entry.Context,
		Hostname:    entry.Hostname,
		PID:         entry.PID,
		TraceID:     entry.TraceID,
		SpanID:      entry.SpanID,
		UserID:      entry.UserID,
		RequestID:   entry.RequestID,
		SessionID:   entry.SessionID,
	}
}

// defaultFormatter is the default log formatter
func (tm *ThemeManager) defaultFormatter(entry CoreLogEntry, theme *Theme) string {
	var parts []string

	// Add timestamp
	if theme.Colors.Timestamp != nil {
		parts = append(parts, theme.Colors.Timestamp.Sprintf("[%s]", entry.Timestamp.Format("2006-01-02 15:04:05")))
	} else {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Timestamp.Format("2006-01-02 15:04:05")))
	}

	// Add level with icon
	levelColor := tm.getLevelColor(entry.Level, theme)
	icon := tm.getLevelIcon(entry.Level, theme)
	levelStr := fmt.Sprintf("%s %s", icon, strings.ToUpper(entry.LevelString))
	if levelColor != nil {
		parts = append(parts, levelColor.Sprintf("%-10s", levelStr))
	} else {
		parts = append(parts, fmt.Sprintf("%-10s", levelStr))
	}

	// Add service name
	if entry.ServiceName != "" && theme.Colors.Service != nil {
		parts = append(parts, theme.Colors.Service.Sprintf("[%s]", entry.ServiceName))
	} else if entry.ServiceName != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.ServiceName))
	}

	// Add file/line info
	if entry.File != "" {
		fileInfo := entry.File
		if entry.Function != "" {
			if entry.Package != "" {
				fileInfo += fmt.Sprintf(":%s.%s", entry.Package, entry.Function)
			} else {
				fileInfo += fmt.Sprintf(":%s", entry.Function)
			}
		}
		fileInfo += fmt.Sprintf(":L%d", entry.Line)

		if theme.Colors.File != nil {
			parts = append(parts, theme.Colors.File.Sprintf("[%s]", fileInfo))
		} else {
			parts = append(parts, fmt.Sprintf("[%s]", fileInfo))
		}
	}

	// Add goroutine ID
	if entry.GoroutineID != "" && theme.Colors.Goroutine != nil {
		parts = append(parts, theme.Colors.Goroutine.Sprintf("%s", entry.GoroutineID))
	} else if entry.GoroutineID != "" {
		parts = append(parts, entry.GoroutineID)
	}

	// Add message
	if theme.Colors.Message != nil {
		parts = append(parts, theme.Colors.Message.Sprintf("%s", entry.Message))
	} else {
		parts = append(parts, entry.Message)
	}

	// Add context if present
	if len(entry.Context) > 0 {
		contextStr := tm.formatContext(entry.Context, theme)
		parts = append(parts, contextStr)
	}

	return strings.Join(parts, " ")
}

// getLevelColor returns the color for a log level
func (tm *ThemeManager) getLevelColor(level LogLevel, theme *Theme) *color.Color {
	switch level {
	case PanicLevel:
		return theme.Colors.Panic
	case ErrorLevel:
		return theme.Colors.Error
	case WarningLevel:
		return theme.Colors.Warning
	case InfoLevel:
		return theme.Colors.Info
	case DebugLevel:
		return theme.Colors.Debug
	case TraceLevel:
		return theme.Colors.Trace
	default:
		return theme.Colors.Info
	}
}

// getLevelIcon returns the icon for a log level
func (tm *ThemeManager) getLevelIcon(level LogLevel, theme *Theme) string {
	switch level {
	case PanicLevel:
		return theme.Icons.Panic
	case ErrorLevel:
		return theme.Icons.Error
	case WarningLevel:
		return theme.Icons.Warning
	case InfoLevel:
		return theme.Icons.Info
	case DebugLevel:
		return theme.Icons.Debug
	case TraceLevel:
		return theme.Icons.Trace
	default:
		return theme.Icons.Info
	}
}

// formatContext formats context fields with theme colors
func (tm *ThemeManager) formatContext(context map[string]interface{}, theme *Theme) string {
	if len(context) == 0 {
		return ""
	}

	var pairs []string
	for k, v := range context {
		keyStr := fmt.Sprintf("%s", k)
		valueStr := fmt.Sprintf("%v", v)

		if theme.Colors.Key != nil && theme.Colors.Value != nil {
			pairs = append(pairs, fmt.Sprintf("%s=%s",
				theme.Colors.Key.Sprintf(keyStr),
				theme.Colors.Value.Sprintf(valueStr)))
		} else {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}
	}

	contextStr := strings.Join(pairs, ", ")
	if theme.Colors.Bracket != nil {
		return theme.Colors.Bracket.Sprintf("{%s}", contextStr)
	}
	return fmt.Sprintf("{%s}", contextStr)
}

// registerBuiltinThemes registers built-in themes
func (tm *ThemeManager) registerBuiltinThemes() {
	// Themes are defined in the builtinThemes map below
}

// registerBuiltinTemplates registers built-in templates
func (tm *ThemeManager) registerBuiltinTemplates() {
	templates := map[string]string{
		"compact":    `{{.Timestamp.Format "15:04:05"}} [{{.Level}}] {{.Message}}`,
		"detailed":   `{{.Timestamp.Format "2006-01-02 15:04:05.000"}} [{{.Level}}] [{{.Service}}] {{.File}}:{{.Line}} {{.Message}}`,
		"json-like":  `{"timestamp":"{{.Timestamp.Format "2006-01-02T15:04:05.000Z07:00"}}","level":"{{.Level}}","message":"{{.Message}}","service":"{{.Service}}"}`,
		"minimal":    `{{.Level}}: {{.Message}}`,
		"structured": `{{.Timestamp.Format "15:04:05"}} [{{.Level}}] {{.Service}} | {{.Message}} | {{range $k,$v := .Context}}{{$k}}={{$v}} {{end}}`,
	}

	for name, tmpl := range templates {
		tm.RegisterTemplate(name, tmpl)
	}
}

// registerBuiltinFormatters registers built-in formatters
func (tm *ThemeManager) registerBuiltinFormatters() {
	// Compact formatter
	tm.RegisterFormatter("compact", func(entry CoreLogEntry, theme *Theme) string {
		levelColor := tm.getLevelColor(entry.Level, theme)
		icon := tm.getLevelIcon(entry.Level, theme)

		var parts []string
		parts = append(parts, fmt.Sprintf("%s %s", icon, strings.ToUpper(entry.LevelString)))
		parts = append(parts, entry.Message)

		if len(entry.Context) > 0 {
			parts = append(parts, tm.formatContext(entry.Context, theme))
		}

		result := strings.Join(parts, " ")
		if levelColor != nil {
			return levelColor.Sprintf(result)
		}
		return result
	})

	// Colorful formatter
	tm.RegisterFormatter("colorful", func(entry CoreLogEntry, theme *Theme) string {
		return tm.defaultFormatter(entry, theme)
	})

	// Plain formatter (no colors)
	tm.RegisterFormatter("plain", func(entry CoreLogEntry, theme *Theme) string {
		var parts []string
		parts = append(parts, fmt.Sprintf("[%s]", entry.Timestamp.Format("2006-01-02 15:04:05")))
		parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.LevelString)))

		if entry.ServiceName != "" {
			parts = append(parts, fmt.Sprintf("[%s]", entry.ServiceName))
		}

		parts = append(parts, entry.Message)

		if len(entry.Context) > 0 {
			contextStr := tm.formatContext(entry.Context, nil) // No colors
			parts = append(parts, contextStr)
		}

		return strings.Join(parts, " ")
	})
}

// Built-in themes
var builtinThemes = map[string]Theme{
	"default": {
		Name:        "default",
		Description: "Default theme with standard colors",
		Colors: ThemeColors{
			Panic:   HiRed.Add(color.Bold),
			Error:   HiRed,
			Warning: HiYellow,
			Info:    HiBlue,
			Success: HiGreen,
			Debug:   HiPurple,
			Trace:   HiCyan,
			Config:  HiPurple,

			Timestamp: Cyan,
			Service:   HiBlue,
			File:      Gray,
			Function:  HiWhite,
			Package:   Gray,
			Goroutine: HiBlack,
			Message:   HiWhite,
			Context:   Gray,
			Key:       HiBlue,
			Value:     HiGreen,
			Bracket:   Gray,
			Separator: Gray,
		},
		Icons: ThemeIcons{
			Panic:   "💥",
			Error:   "❌",
			Warning: "⚠️",
			Info:    "ℹ️",
			Success: "✅",
			Debug:   "🔍",
			Trace:   "📍",
			Config:  "⚙️",
		},
	},

	"dark": {
		Name:        "dark",
		Description: "Dark theme optimized for dark terminals",
		Colors: ThemeColors{
			Panic:   color.New(color.FgRed, color.Bold),
			Error:   color.New(color.FgRed),
			Warning: color.New(color.FgYellow),
			Info:    color.New(color.FgCyan),
			Success: color.New(color.FgGreen),
			Debug:   color.New(color.FgMagenta),
			Trace:   color.New(color.FgBlue),
			Config:  color.New(color.FgMagenta),

			Timestamp: color.New(color.FgHiCyan),
			Service:   color.New(color.FgHiBlue),
			File:      color.New(color.FgHiBlack),
			Function:  color.New(color.FgHiWhite),
			Package:   color.New(color.FgHiBlack),
			Goroutine: color.New(color.FgHiBlack),
			Message:   color.New(color.FgHiWhite),
			Context:   color.New(color.FgHiBlack),
			Key:       color.New(color.FgHiBlue),
			Value:     color.New(color.FgHiGreen),
			Bracket:   color.New(color.FgHiBlack),
			Separator: color.New(color.FgHiBlack),
		},
		Icons: ThemeIcons{
			Panic:   "💥",
			Error:   "❌",
			Warning: "⚠️",
			Info:    "ℹ️",
			Success: "✅",
			Debug:   "🔍",
			Trace:   "📍",
			Config:  "⚙️",
		},
	},

	"light": {
		Name:        "light",
		Description: "Light theme optimized for light terminals",
		Colors: ThemeColors{
			Panic:   color.New(color.FgRed, color.Bold),
			Error:   color.New(color.FgRed),
			Warning: color.New(color.FgYellow),
			Info:    color.New(color.FgBlue),
			Success: color.New(color.FgGreen),
			Debug:   color.New(color.FgMagenta),
			Trace:   color.New(color.FgCyan),
			Config:  color.New(color.FgMagenta),

			Timestamp: color.New(color.FgCyan),
			Service:   color.New(color.FgBlue),
			File:      color.New(color.FgBlack),
			Function:  color.New(color.FgBlack),
			Package:   color.New(color.FgBlack),
			Goroutine: color.New(color.FgBlack),
			Message:   color.New(color.FgBlack),
			Context:   color.New(color.FgBlack),
			Key:       color.New(color.FgBlue),
			Value:     color.New(color.FgGreen),
			Bracket:   color.New(color.FgBlack),
			Separator: color.New(color.FgBlack),
		},
		Icons: ThemeIcons{
			Panic:   "💥",
			Error:   "❌",
			Warning: "⚠️",
			Info:    "ℹ️",
			Success: "✅",
			Debug:   "🔍",
			Trace:   "📍",
			Config:  "⚙️",
		},
	},

	"monokai": {
		Name:        "monokai",
		Description: "Monokai-inspired theme with vibrant colors",
		Colors: ThemeColors{
			Panic:   color.New(color.FgRed, color.Bold),
			Error:   color.New(color.FgRed),
			Warning: color.New(color.FgYellow),
			Info:    color.New(color.FgCyan),
			Success: color.New(color.FgGreen),
			Debug:   color.New(color.FgMagenta),
			Trace:   color.New(color.FgBlue),
			Config:  color.New(color.FgMagenta),

			Timestamp: color.New(color.FgHiCyan),
			Service:   color.New(color.FgHiBlue),
			File:      color.New(color.FgHiBlack),
			Function:  color.New(color.FgHiWhite),
			Package:   color.New(color.FgHiBlack),
			Goroutine: color.New(color.FgHiBlack),
			Message:   color.New(color.FgHiWhite),
			Context:   color.New(color.FgHiBlack),
			Key:       color.New(color.FgHiBlue),
			Value:     color.New(color.FgHiGreen),
			Bracket:   color.New(color.FgHiBlack),
			Separator: color.New(color.FgHiBlack),
		},
		Icons: ThemeIcons{
			Panic:   "💥",
			Error:   "❌",
			Warning: "⚠️",
			Info:    "ℹ️",
			Success: "✅",
			Debug:   "🔍",
			Trace:   "📍",
			Config:  "⚙️",
		},
	},

	"minimal": {
		Name:        "minimal",
		Description: "Minimal theme with subtle colors",
		Colors: ThemeColors{
			Panic:   color.New(color.FgRed),
			Error:   color.New(color.FgRed),
			Warning: color.New(color.FgYellow),
			Info:    color.New(color.FgBlue),
			Success: color.New(color.FgGreen),
			Debug:   color.New(color.FgMagenta),
			Trace:   color.New(color.FgCyan),
			Config:  color.New(color.FgMagenta),

			Timestamp: color.New(color.FgHiBlack),
			Service:   color.New(color.FgHiBlack),
			File:      color.New(color.FgHiBlack),
			Function:  color.New(color.FgHiBlack),
			Package:   color.New(color.FgHiBlack),
			Goroutine: color.New(color.FgHiBlack),
			Message:   color.New(color.FgWhite),
			Context:   color.New(color.FgHiBlack),
			Key:       color.New(color.FgHiBlack),
			Value:     color.New(color.FgHiBlack),
			Bracket:   color.New(color.FgHiBlack),
			Separator: color.New(color.FgHiBlack),
		},
		Icons: ThemeIcons{
			Panic:   "!",
			Error:   "E",
			Warning: "W",
			Info:    "I",
			Success: "S",
			Debug:   "D",
			Trace:   "T",
			Config:  "C",
		},
	},
}

// Global theme manager instance
var globalThemeManager = NewThemeManager()

// SetGlobalTheme sets the global theme
func SetGlobalTheme(name string) error {
	return globalThemeManager.SetTheme(name)
}

// GetGlobalTheme returns the global theme
func GetGlobalTheme() *Theme {
	return globalThemeManager.GetTheme()
}

// RegisterGlobalTemplate registers a global template
func RegisterGlobalTemplate(name, templateStr string) error {
	return globalThemeManager.RegisterTemplate(name, templateStr)
}

// RegisterGlobalFormatter registers a global formatter
func RegisterGlobalFormatter(name string, formatter LogFormatter) {
	globalThemeManager.RegisterFormatter(name, formatter)
}

// FormatGlobal formats a log entry using the global theme manager
func FormatGlobal(entry CoreLogEntry, formatName string) string {
	return globalThemeManager.Format(entry, formatName)
}
