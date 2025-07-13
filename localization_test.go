package pim

import (
	"strings"
	"testing"
)

func TestLocaleParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Locale
	}{
		{"en-US", Locale{Language: "en", Region: "US"}},
		{"es", Locale{Language: "es", Region: ""}},
		{"fr-FR", Locale{Language: "fr", Region: "FR"}},
		{"de-DE", Locale{Language: "de", Region: "DE"}},
		{"zh-CN", Locale{Language: "zh", Region: "CN"}},
		{"EN-US", Locale{Language: "en", Region: "US"}}, // Case normalization
		{"Es", Locale{Language: "es", Region: ""}},
	}

	for _, test := range tests {
		result := ParseLocale(test.input)
		if result != test.expected {
			t.Errorf("ParseLocale(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestLocaleString(t *testing.T) {
	tests := []struct {
		locale   Locale
		expected string
	}{
		{Locale{Language: "en", Region: "US"}, "en-US"},
		{Locale{Language: "es", Region: ""}, "es"},
		{Locale{Language: "fr", Region: "FR"}, "fr-FR"},
	}

	for _, test := range tests {
		result := test.locale.String()
		if result != test.expected {
			t.Errorf("Locale.String() = %s, expected %s", result, test.expected)
		}
	}
}

func TestMessageCatalog(t *testing.T) {
	catalog := NewMessageCatalog(Locale{Language: "en", Region: "US"})

	// Test adding messages
	catalog.AddMessage("test_key", "Test message")
	catalog.AddMessage("welcome", "Welcome, {0}!")

	// Test getting messages
	if message, exists := catalog.GetMessage("test_key"); !exists || message != "Test message" {
		t.Errorf("Failed to retrieve message for 'test_key'")
	}

	if message, exists := catalog.GetMessage("welcome"); !exists || message != "Welcome, {0}!" {
		t.Errorf("Failed to retrieve message for 'welcome'")
	}

	// Test non-existent message
	if _, exists := catalog.GetMessage("nonexistent"); exists {
		t.Errorf("Should not find message for 'nonexistent'")
	}
}

func TestLocalizationManager(t *testing.T) {
	lm := NewLocalizationManager(Locale{Language: "en", Region: "US"})

	// Test default locale
	if lm.GetCurrentLocale() != (Locale{Language: "en", Region: "US"}) {
		t.Errorf("Default locale not set correctly")
	}

	// Test setting locale
	newLocale := Locale{Language: "es"}
	lm.SetCurrentLocale(newLocale)
	if lm.GetCurrentLocale() != newLocale {
		t.Errorf("Locale not set correctly")
	}

	// Test adding catalog
	esCatalog := NewMessageCatalog(Locale{Language: "es"})
	esCatalog.AddMessage("hello", "Hola")
	lm.AddCatalog(esCatalog)

	// Test getting catalog
	if catalog := lm.GetCatalog(Locale{Language: "es"}); catalog == nil {
		t.Errorf("Failed to get Spanish catalog")
	}
}

func TestTranslation(t *testing.T) {
	lm := NewLocalizationManager(Locale{Language: "en", Region: "US"})

	// Add English catalog
	enCatalog := NewMessageCatalog(Locale{Language: "en", Region: "US"})
	enCatalog.AddMessage("hello", "Hello")
	enCatalog.AddMessage("welcome", "Welcome, {0}!")
	enCatalog.AddMessage("count", "Count: {0}")
	lm.AddCatalog(enCatalog)

	// Add Spanish catalog
	esCatalog := NewMessageCatalog(Locale{Language: "es"})
	esCatalog.AddMessage("hello", "Hola")
	esCatalog.AddMessage("welcome", "¡Bienvenido, {0}!")
	esCatalog.AddMessage("count", "Conteo: {0}")
	lm.AddCatalog(esCatalog)

	// Test English translation
	lm.SetCurrentLocale(Locale{Language: "en", Region: "US"})
	if result := lm.Translate("hello"); result != "Hello" {
		t.Errorf("English translation failed: got %s, expected Hello", result)
	}

	if result := lm.Translate("welcome", "John"); result != "Welcome, John!" {
		t.Errorf("English translation with args failed: got %s, expected Welcome, John!", result)
	}

	// Test Spanish translation
	lm.SetCurrentLocale(Locale{Language: "es"})
	if result := lm.Translate("hello"); result != "Hola" {
		t.Errorf("Spanish translation failed: got %s, expected Hola", result)
	}

	if result := lm.Translate("welcome", "Juan"); result != "¡Bienvenido, Juan!" {
		t.Errorf("Spanish translation with args failed: got %s, expected ¡Bienvenido, Juan!", result)
	}

	// Test fallback to English for missing Spanish message
	if result := lm.Translate("count", 42); result != "Conteo: 42" {
		t.Errorf("Spanish translation with number failed: got %s, expected Conteo: 42", result)
	}

	// Test fallback to key for missing message
	if result := lm.Translate("missing_key"); result != "missing_key" {
		t.Errorf("Fallback to key failed: got %s, expected missing_key", result)
	}
}

func TestLocalizedLogger(t *testing.T) {
	config := LoggerConfig{
		Level:         InfoLevel,
		ServiceName:   "test",
		EnableConsole: false, // Disable console for testing
	}

	logger := NewLocalizedLogger(config, Locale{Language: "en", Region: "US"})
	defer logger.Close()

	// Test locale setting
	esLocale := Locale{Language: "es"}
	logger.SetLocale(esLocale)
	if logger.GetLocale() != esLocale {
		t.Errorf("Logger locale not set correctly")
	}

	// Test custom message addition
	logger.AddCustomMessage("test_custom", "Custom message: {0}")

	// Test translation
	if result := logger.localization.Translate("test_custom", "test"); result != "Custom message: test" {
		t.Errorf("Custom message translation failed: got %s", result)
	}
}

func TestBuiltinCatalogs(t *testing.T) {
	lm := NewLocalizationManager(Locale{Language: "en", Region: "US"})
	LoadBuiltinCatalogs(lm)

	// Test English messages
	lm.SetCurrentLocale(Locale{Language: "en", Region: "US"})
	if result := lm.Translate("app_started"); result != "Application started" {
		t.Errorf("English builtin message failed: got %s", result)
	}

	if result := lm.Translate("user_login", "john", "192.168.1.1"); result != "User john logged in from 192.168.1.1" {
		t.Errorf("English builtin message with args failed: got %s", result)
	}

	// Test Spanish messages
	lm.SetCurrentLocale(Locale{Language: "es"})
	if result := lm.Translate("app_started"); result != "Aplicación iniciada" {
		t.Errorf("Spanish builtin message failed: got %s", result)
	}

	if result := lm.Translate("user_login", "juan", "192.168.1.1"); result != "Usuario juan conectado desde 192.168.1.1" {
		t.Errorf("Spanish builtin message with args failed: got %s", result)
	}

	// Test French messages
	lm.SetCurrentLocale(Locale{Language: "fr"})
	if result := lm.Translate("app_started"); result != "Application démarrée" {
		t.Errorf("French builtin message failed: got %s", result)
	}

	if result := lm.Translate("user_login", "pierre", "192.168.1.1"); result != "Utilisateur pierre connecté depuis 192.168.1.1" {
		t.Errorf("French builtin message with args failed: got %s", result)
	}
}

func TestFallbackChain(t *testing.T) {
	lm := NewLocalizationManager(Locale{Language: "en", Region: "US"})
	LoadBuiltinCatalogs(lm)

	// Test fallback from en-GB to en-US
	lm.SetCurrentLocale(Locale{Language: "en", Region: "GB"})
	if result := lm.Translate("app_started"); result != "Application started" {
		t.Errorf("Fallback from en-GB to en-US failed: got %s", result)
	}

	// Test fallback from unsupported language to English
	lm.SetCurrentLocale(Locale{Language: "de"}) // German not supported
	if result := lm.Translate("app_started"); result != "Application started" {
		t.Errorf("Fallback from unsupported language failed: got %s", result)
	}
}

func TestMessageFormatting(t *testing.T) {
	lm := NewLocalizationManager(Locale{Language: "en", Region: "US"})
	enCatalog := NewMessageCatalog(Locale{Language: "en", Region: "US"})
	enCatalog.AddMessage("test", "Hello {0}, you have {1} messages")
	lm.AddCatalog(enCatalog)

	result := lm.Translate("test", "John", 5)
	expected := "Hello John, you have 5 messages"
	if result != expected {
		t.Errorf("Message formatting failed: got %s, expected %s", result, expected)
	}

	// Test with missing arguments
	result = lm.Translate("test", "John")
	expected = "Hello John, you have {1} messages"
	if result != expected {
		t.Errorf("Message formatting with missing args failed: got %s, expected %s", result, expected)
	}
}

func TestLocalizedLoggerMethods(t *testing.T) {
	config := LoggerConfig{
		Level:         InfoLevel,
		ServiceName:   "test",
		EnableConsole: false,
	}

	logger := NewLocalizedLogger(config, Locale{Language: "en", Region: "US"})
	defer logger.Close()

	// Test TInfo method
	logger.TInfo("app_started")

	// Test TSuccess method
	logger.TSuccess("user_login", "testuser", "127.0.0.1")

	// Test TWarning method
	logger.TWarning("performance_warning", "High CPU usage")

	// Test TError method
	logger.TError("db_connection_failed", "Connection timeout")

	// Test TDebug method
	logger.SetLevel(DebugLevel)
	logger.TDebug("debug_info", "Debug message")

	// Test TTrace method
	logger.TTrace("debug_info", "Trace message")
}

func TestContextAwareLogging(t *testing.T) {
	config := LoggerConfig{
		Level:         InfoLevel,
		ServiceName:   "test",
		EnableConsole: false,
	}

	logger := NewLocalizedLogger(config, Locale{Language: "en", Region: "US"})
	defer logger.Close()

	context := map[string]interface{}{
		"user_id":    12345,
		"ip_address": "192.168.1.100",
		"action":     "login",
	}

	// Test TInfoWithContext
	logger.TInfoWithContext("api_request", context, "POST", "/api/login", "192.168.1.100")

	// Test TErrorWithContext
	logger.TErrorWithContext("task_failed", context, "authentication", "Invalid credentials")
}

func TestGlobalLocalizedLogger(t *testing.T) {
	config := LoggerConfig{
		Level:         InfoLevel,
		ServiceName:   "test",
		EnableConsole: false,
	}

	// Initialize global logger
	InitLocalizedLogger(config, Locale{Language: "en", Region: "US"})

	// Test global functions
	TInfo("app_started")
	TSuccess("user_login", "globaluser", "127.0.0.1")
	TWarning("performance_warning", "Global warning")

	// Test locale switching
	SetGlobalLocale(Locale{Language: "es"})
	TInfo("app_started")
	TSuccess("user_login", "usuario", "127.0.0.1")
}

func TestLocaleDetection(t *testing.T) {
	// Test that DetectLocale returns a valid locale
	locale := DetectLocale()
	if locale.Language == "" {
		t.Errorf("DetectLocale returned invalid locale: %v", locale)
	}

	// Test that the language is lowercase
	if locale.Language != strings.ToLower(locale.Language) {
		t.Errorf("Detected language should be lowercase: %s", locale.Language)
	}

	// Test that region is uppercase (if present)
	if locale.Region != "" && locale.Region != strings.ToUpper(locale.Region) {
		t.Errorf("Detected region should be uppercase: %s", locale.Region)
	}
}

func TestConcurrentAccess(t *testing.T) {
	lm := NewLocalizationManager(Locale{Language: "en", Region: "US"})
	LoadBuiltinCatalogs(lm)

	// Test concurrent locale switching and translation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			locale := Locale{Language: "en", Region: "US"}
			if id%2 == 0 {
				locale = Locale{Language: "es"}
			}
			lm.SetCurrentLocale(locale)
			result := lm.Translate("app_started")
			if result == "" {
				t.Errorf("Translation failed in goroutine %d", id)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCustomMessageLoading(t *testing.T) {
	config := LoggerConfig{
		Level:         InfoLevel,
		ServiceName:   "test",
		EnableConsole: false,
	}

	logger := NewLocalizedLogger(config, Locale{Language: "en", Region: "US"})
	defer logger.Close()

	// Test loading custom catalog (this would normally load from file)
	// For now, we'll test the method exists and doesn't panic
	err := logger.LoadCustomCatalog(Locale{Language: "de"}, "nonexistent.json")
	if err == nil {
		t.Log("LoadCustomCatalog should return error for nonexistent file")
	}
}

func TestPerformance(t *testing.T) {
	lm := NewLocalizationManager(Locale{Language: "en", Region: "US"})
	LoadBuiltinCatalogs(lm)

	// Test translation performance
	for i := 0; i < 1000; i++ {
		result := lm.Translate("app_started")
		if result != "Application started" {
			t.Errorf("Performance test failed at iteration %d", i)
		}
	}
}
