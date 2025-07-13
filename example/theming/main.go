package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/refactorroom/pim"
)

func main() {
	fmt.Println("=== Rich Formatting and Theming Demo ===\n")

	// Demo 1: Built-in themes
	demoBuiltinThemes()

	// Demo 2: Custom themes
	demoCustomThemes()

	// Demo 3: Custom templates
	demoCustomTemplates()

	// Demo 4: Custom formatters
	demoCustomFormatters()

	// Demo 5: Dynamic theme switching
	demoDynamicThemeSwitching()

	fmt.Println("\n=== Demo completed ===")
}

func demoBuiltinThemes() {
	fmt.Println("--- Built-in Themes ---")

	themes := []string{"default", "dark", "light", "monokai", "minimal"}

	for _, themeName := range themes {
		fmt.Printf("\nTheme: %s\n", themeName)
		fmt.Println(strings.Repeat("-", 50))

		// Create logger with specific theme
		config := pim.LoggerConfig{
			Level:            pim.InfoLevel,
			ServiceName:      "theme-demo",
			TimestampFormat:  "15:04:05",
			ShowFunctionName: true,
			ShowPackageName:  true,
			ThemeName:        themeName,
			FormatName:       "colorful",
			EnableConsole:    true,
		}

		logger := pim.NewLoggerCore(config)

		// Log different levels to show theme colors
		logger.Info("This is an info message")
		logger.Success("This is a success message")
		logger.Warning("This is a warning message")
		logger.Error("This is an error message")
		logger.Debug("This is a debug message")

		// Log with context to show theme formatting
		logger.InfoWithFields("User action completed", map[string]interface{}{
			"user_id":    "user123",
			"action":     "login",
			"duration":   "150ms",
			"ip_address": "192.168.1.100",
		})

		time.Sleep(1 * time.Second) // Brief pause between themes
	}
}

func demoCustomThemes() {
	fmt.Println("\n--- Custom Themes ---")

	// Create a custom theme
	customTheme := pim.Theme{
		Name:        "custom-rainbow",
		Description: "Rainbow theme with vibrant colors",
		Colors: pim.ThemeColors{
			Panic:   pim.HiRed.Add(color.Bold),
			Error:   pim.HiRed,
			Warning: pim.HiYellow,
			Info:    pim.HiBlue,
			Success: pim.HiGreen,
			Debug:   pim.HiPurple,
			Trace:   pim.HiCyan,
			Config:  pim.HiPurple,

			Timestamp: pim.HiCyan.Add(color.Bold),
			Service:   pim.HiBlue.Add(color.Underline),
			File:      pim.Gray,
			Function:  pim.HiWhite,
			Package:   pim.Gray,
			Goroutine: pim.HiBlack,
			Message:   pim.HiWhite,
			Context:   pim.Gray,
			Key:       pim.HiBlue,
			Value:     pim.HiGreen,
			Bracket:   pim.Gray,
			Separator: pim.Gray,
		},
		Icons: pim.ThemeIcons{
			Panic:   "🚨",
			Error:   "💥",
			Warning: "⚠️",
			Info:    "💡",
			Success: "🎉",
			Debug:   "🔧",
			Trace:   "🔍",
			Config:  "⚙️",
		},
	}

	// Create logger with custom theme
	config := pim.LoggerConfig{
		Level:           pim.InfoLevel,
		ServiceName:     "custom-theme-demo",
		TimestampFormat: "15:04:05",
		CustomTheme:     &customTheme,
		FormatName:      "colorful",
		EnableConsole:   true,
	}

	logger := pim.NewLoggerCore(config)

	fmt.Println("Custom Rainbow Theme:")
	fmt.Println(strings.Repeat("-", 50))

	logger.Info("Welcome to the rainbow theme!")
	logger.Success("Everything is working perfectly")
	logger.Warning("But be careful with the colors")
	logger.Error("Oops, something went wrong")
	logger.Debug("Let's debug this issue")
	logger.Config("Configuration updated")

	logger.InfoWithFields("Rainbow performance metrics", map[string]interface{}{
		"cpu_usage":           "45%",
		"memory_usage":        "60%",
		"response_time":       "120ms",
		"requests_per_second": 150,
	})
}

func demoCustomTemplates() {
	fmt.Println("\n--- Custom Templates ---")

	// Create logger
	config := pim.LoggerConfig{
		Level:           pim.InfoLevel,
		ServiceName:     "template-demo",
		TimestampFormat: "15:04:05",
		ThemeName:       "default",
		EnableConsole:   true,
	}

	logger := pim.NewLoggerCore(config)

	// Register custom templates
	templates := map[string]string{
		"emoji": `{{.Timestamp.Format "15:04"}} {{.Level}} {{.Message}} 🚀`,
		"table": `| {{.Timestamp.Format "15:04:05"}} | {{.Level}} | {{.Service}} | {{.Message}} |`,
		"chat":  `[{{.Timestamp.Format "15:04"}}] {{.Service}}: {{.Message}}`,
		"log":   `{{.Timestamp.Format "2006-01-02 15:04:05.000"}} [{{.Level}}] {{.Service}} - {{.Message}}`,
	}

	for name, template := range templates {
		logger.RegisterTemplate(name, template)
	}

	// Test each template
	templateNames := []string{"emoji", "table", "chat", "log"}

	for _, templateName := range templateNames {
		fmt.Printf("\nTemplate: %s\n", templateName)
		fmt.Println(strings.Repeat("-", 50))

		// Temporarily set the format
		logger.SetTheme("default") // Reset theme
		logger.RegisterTemplate("temp", templates[templateName])

		logger.Info("This is a test message")
		logger.Success("Operation completed successfully")
		logger.Warning("Please check the configuration")

		time.Sleep(500 * time.Millisecond)
	}
}

func demoCustomFormatters() {
	fmt.Println("\n--- Custom Formatters ---")

	// Create logger
	config := pim.LoggerConfig{
		Level:           pim.InfoLevel,
		ServiceName:     "formatter-demo",
		TimestampFormat: "15:04:05",
		ThemeName:       "default",
		EnableConsole:   true,
	}

	logger := pim.NewLoggerCore(config)

	// Register custom formatters
	logger.RegisterFormatter("banner", func(entry pim.CoreLogEntry, theme *pim.Theme) string {
		levelColor := getLevelColor(entry.Level, theme)
		icon := getLevelIcon(entry.Level, theme)

		banner := fmt.Sprintf("╔══════════════════════════════════════════════════════════╗")
		content := fmt.Sprintf("║ %s %s", icon, entry.Message)
		if len(content) > 58 {
			content = content[:55] + "..."
		}
		content = fmt.Sprintf("%-58s ║", content)
		footer := fmt.Sprintf("╚══════════════════════════════════════════════════════════╝")

		if levelColor != nil {
			return levelColor.Sprintf("%s\n%s\n%s", banner, content, footer)
		}
		return fmt.Sprintf("%s\n%s\n%s", banner, content, footer)
	})

	logger.RegisterFormatter("progress", func(entry pim.CoreLogEntry, theme *pim.Theme) string {
		levelColor := getLevelColor(entry.Level, theme)
		icon := getLevelIcon(entry.Level, theme)

		progress := "████████████████████████████████████████████████████████████"
		message := fmt.Sprintf("%s %s", icon, entry.Message)

		result := fmt.Sprintf("[%s] %s", progress, message)

		if levelColor != nil {
			return levelColor.Sprintf(result)
		}
		return result
	})

	logger.RegisterFormatter("minimalist", func(entry pim.CoreLogEntry, theme *pim.Theme) string {
		levelColor := getLevelColor(entry.Level, theme)
		icon := getLevelIcon(entry.Level, theme)

		result := fmt.Sprintf("%s %s", icon, entry.Message)

		if levelColor != nil {
			return levelColor.Sprintf(result)
		}
		return result
	})

	// Test custom formatters
	formatters := []string{"banner", "progress", "minimalist"}

	for _, formatterName := range formatters {
		fmt.Printf("\nFormatter: %s\n", formatterName)
		fmt.Println(strings.Repeat("-", 50))

		// Create a new logger with this formatter
		config.FormatName = formatterName
		tempLogger := pim.NewLoggerCore(config)

		tempLogger.Info("This is a test message")
		tempLogger.Success("Operation completed")
		tempLogger.Warning("Please be careful")

		time.Sleep(500 * time.Millisecond)
	}
}

func demoDynamicThemeSwitching() {
	fmt.Println("\n--- Dynamic Theme Switching ---")

	// Create logger
	config := pim.LoggerConfig{
		Level:           pim.InfoLevel,
		ServiceName:     "dynamic-demo",
		TimestampFormat: "15:04:05",
		ThemeName:       "default",
		FormatName:      "colorful",
		EnableConsole:   true,
	}

	logger := pim.NewLoggerCore(config)

	fmt.Println("Starting with default theme...")
	logger.Info("Application started")
	logger.Success("Initialization complete")

	// Switch to dark theme
	fmt.Println("\nSwitching to dark theme...")
	logger.SetTheme("dark")
	logger.Info("Theme switched to dark")
	logger.Warning("Dark theme activated")

	// Switch to light theme
	fmt.Println("\nSwitching to light theme...")
	logger.SetTheme("light")
	logger.Info("Theme switched to light")
	logger.Success("Light theme activated")

	// Switch to minimal theme
	fmt.Println("\nSwitching to minimal theme...")
	logger.SetTheme("minimal")
	logger.Info("Theme switched to minimal")
	logger.Debug("Minimal theme activated")

	// Switch back to default
	fmt.Println("\nSwitching back to default theme...")
	logger.SetTheme("default")
	logger.Info("Back to default theme")
	logger.Success("Theme switching demo completed")
}

// Helper functions for custom formatters
func getLevelColor(level pim.LogLevel, theme *pim.Theme) *color.Color {
	switch level {
	case pim.PanicLevel:
		return theme.Colors.Panic
	case pim.ErrorLevel:
		return theme.Colors.Error
	case pim.WarningLevel:
		return theme.Colors.Warning
	case pim.InfoLevel:
		return theme.Colors.Info
	case pim.DebugLevel:
		return theme.Colors.Debug
	case pim.TraceLevel:
		return theme.Colors.Trace
	default:
		return theme.Colors.Info
	}
}

func getLevelIcon(level pim.LogLevel, theme *pim.Theme) string {
	switch level {
	case pim.PanicLevel:
		return theme.Icons.Panic
	case pim.ErrorLevel:
		return theme.Icons.Error
	case pim.WarningLevel:
		return theme.Icons.Warning
	case pim.InfoLevel:
		return theme.Icons.Info
	case pim.DebugLevel:
		return theme.Icons.Debug
	case pim.TraceLevel:
		return theme.Icons.Trace
	default:
		return theme.Icons.Info
	}
}
