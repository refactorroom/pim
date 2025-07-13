package pim

import (
	"testing"
	"time"
)

func TestThemeManager(t *testing.T) {
	tm := NewThemeManager()

	// Test setting themes
	themes := []string{"default", "dark", "light", "monokai", "minimal"}
	for _, themeName := range themes {
		err := tm.SetTheme(themeName)
		if err != nil {
			t.Errorf("Failed to set theme '%s': %v", themeName, err)
		}

		theme := tm.GetTheme()
		if theme == nil {
			t.Errorf("Theme is nil after setting '%s'", themeName)
		}

		if theme.Name != themeName {
			t.Errorf("Expected theme name '%s', got '%s'", themeName, theme.Name)
		}
	}

	// Test invalid theme
	err := tm.SetTheme("invalid-theme")
	if err == nil {
		t.Error("Expected error when setting invalid theme")
	}
}

func TestThemeManagerTemplates(t *testing.T) {
	tm := NewThemeManager()

	// Test registering template
	templateStr := `{{.Timestamp.Format "15:04"}} [{{.Level}}] {{.Message}}`
	err := tm.RegisterTemplate("test", templateStr)
	if err != nil {
		t.Errorf("Failed to register template: %v", err)
	}

	// Test formatting with template
	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		LevelString: "INFO",
		Message:     "Test message",
		ServiceName: "test-service",
	}

	formatted := tm.Format(entry, "test")
	if formatted == "" {
		t.Error("Formatted output is empty")
	}
}

func TestThemeManagerFormatters(t *testing.T) {
	tm := NewThemeManager()

	// Test registering formatter
	tm.RegisterFormatter("test", func(entry CoreLogEntry, theme *Theme) string {
		return "test-formatted: " + entry.Message
	})

	// Test formatting with custom formatter
	entry := CoreLogEntry{
		Timestamp:   time.Now(),
		Level:       InfoLevel,
		LevelString: "INFO",
		Message:     "Test message",
		ServiceName: "test-service",
	}

	formatted := tm.Format(entry, "test")
	expected := "test-formatted: Test message"
	if formatted != expected {
		t.Errorf("Expected '%s', got '%s'", expected, formatted)
	}
}

func TestBuiltinThemes(t *testing.T) {
	// Test that all built-in themes exist
	expectedThemes := []string{"default", "dark", "light", "monokai", "minimal"}

	for _, themeName := range expectedThemes {
		theme, exists := builtinThemes[themeName]
		if !exists {
			t.Errorf("Built-in theme '%s' not found", themeName)
		}

		if theme.Name != themeName {
			t.Errorf("Theme name mismatch: expected '%s', got '%s'", themeName, theme.Name)
		}

		if theme.Description == "" {
			t.Errorf("Theme '%s' has no description", themeName)
		}
	}
}

func TestThemeColors(t *testing.T) {
	tm := NewThemeManager()
	tm.SetTheme("default")
	theme := tm.GetTheme()

	// Test that colors are set
	if theme.Colors.Info == nil {
		t.Error("Info color is nil")
	}
	if theme.Colors.Error == nil {
		t.Error("Error color is nil")
	}
	if theme.Colors.Warning == nil {
		t.Error("Warning color is nil")
	}
}

func TestThemeIcons(t *testing.T) {
	tm := NewThemeManager()
	tm.SetTheme("default")
	theme := tm.GetTheme()

	// Test that icons are set
	if theme.Icons.Info == "" {
		t.Error("Info icon is empty")
	}
	if theme.Icons.Error == "" {
		t.Error("Error icon is empty")
	}
	if theme.Icons.Warning == "" {
		t.Error("Warning icon is empty")
	}
}

func TestGlobalThemeManager(t *testing.T) {
	// Test global theme manager functions
	err := SetGlobalTheme("dark")
	if err != nil {
		t.Errorf("Failed to set global theme: %v", err)
	}

	theme := GetGlobalTheme()
	if theme == nil {
		t.Error("Global theme is nil")
	}

	if theme.Name != "dark" {
		t.Errorf("Expected global theme 'dark', got '%s'", theme.Name)
	}
}
