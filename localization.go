package pim

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Locale represents a language/region combination
type Locale struct {
	Language string // e.g., "en", "es", "fr"
	Region   string // e.g., "US", "ES", "FR" (optional)
}

// String returns the locale as a string (e.g., "en-US", "es")
func (l Locale) String() string {
	if l.Region != "" {
		return fmt.Sprintf("%s-%s", l.Language, l.Region)
	}
	return l.Language
}

// ParseLocale parses a locale string into a Locale struct
func ParseLocale(localeStr string) Locale {
	parts := strings.Split(localeStr, "-")
	if len(parts) >= 2 {
		return Locale{
			Language: strings.ToLower(parts[0]),
			Region:   strings.ToUpper(parts[1]),
		}
	}
	return Locale{Language: strings.ToLower(localeStr)}
}

// MessageCatalog holds translated messages for a specific locale
type MessageCatalog struct {
	Locale   Locale
	Messages map[string]string
	mu       sync.RWMutex
}

// NewMessageCatalog creates a new message catalog
func NewMessageCatalog(locale Locale) *MessageCatalog {
	return &MessageCatalog{
		Locale:   locale,
		Messages: make(map[string]string),
	}
}

// AddMessage adds a message to the catalog
func (mc *MessageCatalog) AddMessage(key, message string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.Messages[key] = message
}

// GetMessage retrieves a message from the catalog
func (mc *MessageCatalog) GetMessage(key string) (string, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	message, exists := mc.Messages[key]
	return message, exists
}

// LoadFromFile loads messages from a JSON file
func (mc *MessageCatalog) LoadFromFile(filePath string) error {
	// This would load from JSON, but for simplicity we'll use a simple format
	// In a real implementation, you'd use encoding/json
	return nil
}

// LocalizationManager manages multiple message catalogs and locale detection
type LocalizationManager struct {
	catalogs      map[string]*MessageCatalog
	defaultLocale Locale
	currentLocale Locale
	fallbackChain []Locale
	mu            sync.RWMutex
}

// NewLocalizationManager creates a new localization manager
func NewLocalizationManager(defaultLocale Locale) *LocalizationManager {
	lm := &LocalizationManager{
		catalogs:      make(map[string]*MessageCatalog),
		defaultLocale: defaultLocale,
		currentLocale: defaultLocale,
		fallbackChain: []Locale{defaultLocale},
	}

	// Initialize default catalog
	lm.catalogs[defaultLocale.String()] = NewMessageCatalog(defaultLocale)

	return lm
}

// SetCurrentLocale sets the current locale
func (lm *LocalizationManager) SetCurrentLocale(locale Locale) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.currentLocale = locale
	lm.updateFallbackChain()
}

// GetCurrentLocale returns the current locale
func (lm *LocalizationManager) GetCurrentLocale() Locale {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.currentLocale
}

// updateFallbackChain updates the fallback chain for the current locale
func (lm *LocalizationManager) updateFallbackChain() {
	chain := []Locale{lm.currentLocale}

	// Add language-only fallback if we have a region
	if lm.currentLocale.Region != "" {
		chain = append(chain, Locale{Language: lm.currentLocale.Language})
	}

	// Add default locale as final fallback
	if lm.currentLocale.String() != lm.defaultLocale.String() {
		chain = append(chain, lm.defaultLocale)
	}

	lm.fallbackChain = chain
}

// AddCatalog adds a message catalog for a locale
func (lm *LocalizationManager) AddCatalog(catalog *MessageCatalog) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.catalogs[catalog.Locale.String()] = catalog
}

// GetCatalog returns a catalog for a locale
func (lm *LocalizationManager) GetCatalog(locale Locale) *MessageCatalog {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.catalogs[locale.String()]
}

// Translate translates a message key to the current locale
func (lm *LocalizationManager) Translate(key string, args ...interface{}) string {
	lm.mu.RLock()
	fallbackChain := make([]Locale, len(lm.fallbackChain))
	copy(fallbackChain, lm.fallbackChain)
	lm.mu.RUnlock()

	// Try each locale in the fallback chain
	for _, locale := range fallbackChain {
		if catalog := lm.GetCatalog(locale); catalog != nil {
			if message, exists := catalog.GetMessage(key); exists {
				return lm.formatMessage(message, args...)
			}
		}
	}

	// Return the key if no translation found
	return key
}

// formatMessage formats a message with arguments
func (lm *LocalizationManager) formatMessage(message string, args ...interface{}) string {
	if len(args) == 0 {
		return message
	}

	// Simple placeholder replacement: {0}, {1}, etc.
	re := regexp.MustCompile(`\{(\d+)\}`)
	result := re.ReplaceAllStringFunc(message, func(match string) string {
		// Extract the index
		indexStr := match[1 : len(match)-1]
		index, err := strconv.Atoi(indexStr)
		if err != nil || index >= len(args) {
			return match
		}
		return fmt.Sprintf("%v", args[index])
	})

	return result
}

// DetectLocale detects the system locale
func DetectLocale() Locale {
	// Check environment variables
	if lang := os.Getenv("LANG"); lang != "" {
		parts := strings.Split(lang, ".")
		if len(parts) > 0 {
			return ParseLocale(parts[0])
		}
	}

	if lang := os.Getenv("LC_ALL"); lang != "" {
		parts := strings.Split(lang, ".")
		if len(parts) > 0 {
			return ParseLocale(parts[0])
		}
	}

	if lang := os.Getenv("LC_MESSAGES"); lang != "" {
		parts := strings.Split(lang, ".")
		if len(parts) > 0 {
			return ParseLocale(parts[0])
		}
	}

	// Default to English
	return Locale{Language: "en", Region: "US"}
}

// Built-in message catalogs for common languages
var (
	// English messages
	englishMessages = map[string]string{
		"app_started":          "Application started",
		"app_shutdown":         "Application shutting down",
		"user_login":           "User {0} logged in from {1}",
		"user_logout":          "User {0} logged out",
		"file_uploaded":        "File {0} uploaded successfully ({1} bytes)",
		"file_upload_failed":   "File upload failed: {0}",
		"db_connected":         "Database connected successfully",
		"db_connection_failed": "Database connection failed: {0}",
		"api_request":          "API request {0} {1} from {2}",
		"api_response":         "API response {0} {1} in {2}ms",
		"error_occurred":       "An error occurred: {0}",
		"warning_detected":     "Warning detected: {0}",
		"info_message":         "Information: {0}",
		"debug_info":           "Debug information: {0}",
		"config_loaded":        "Configuration loaded from {0}",
		"config_saved":         "Configuration saved to {0}",
		"backup_created":       "Backup created: {0}",
		"backup_failed":        "Backup failed: {0}",
		"email_sent":           "Email sent to {0}",
		"email_failed":         "Email failed: {0}",
		"task_completed":       "Task {0} completed successfully",
		"task_failed":          "Task {0} failed: {1}",
		"memory_usage":         "Memory usage: {0}MB",
		"cpu_usage":            "CPU usage: {0}%",
		"disk_usage":           "Disk usage: {0}%",
		"network_activity":     "Network activity: {0} bytes sent, {1} bytes received",
		"security_alert":       "Security alert: {0}",
		"performance_warning":  "Performance warning: {0}",
		"maintenance_mode":     "Maintenance mode {0}",
		"service_started":      "Service {0} started",
		"service_stopped":      "Service {0} stopped",
		"service_restarted":    "Service {0} restarted",
		"service_failed":       "Service {0} failed: {1}",
	}

	// Spanish messages
	spanishMessages = map[string]string{
		"app_started":          "Aplicación iniciada",
		"app_shutdown":         "Aplicación cerrando",
		"user_login":           "Usuario {0} conectado desde {1}",
		"user_logout":          "Usuario {0} desconectado",
		"file_uploaded":        "Archivo {0} subido exitosamente ({1} bytes)",
		"file_upload_failed":   "Error al subir archivo: {0}",
		"db_connected":         "Base de datos conectada exitosamente",
		"db_connection_failed": "Error de conexión a la base de datos: {0}",
		"api_request":          "Solicitud API {0} {1} desde {2}",
		"api_response":         "Respuesta API {0} {1} en {2}ms",
		"error_occurred":       "Ocurrió un error: {0}",
		"warning_detected":     "Advertencia detectada: {0}",
		"info_message":         "Información: {0}",
		"debug_info":           "Información de depuración: {0}",
		"config_loaded":        "Configuración cargada desde {0}",
		"config_saved":         "Configuración guardada en {0}",
		"backup_created":       "Respaldo creado: {0}",
		"backup_failed":        "Error al crear respaldo: {0}",
		"email_sent":           "Correo enviado a {0}",
		"email_failed":         "Error al enviar correo: {0}",
		"task_completed":       "Tarea {0} completada exitosamente",
		"task_failed":          "Tarea {0} falló: {1}",
		"memory_usage":         "Uso de memoria: {0}MB",
		"cpu_usage":            "Uso de CPU: {0}%",
		"disk_usage":           "Uso de disco: {0}%",
		"network_activity":     "Actividad de red: {0} bytes enviados, {1} bytes recibidos",
		"security_alert":       "Alerta de seguridad: {0}",
		"performance_warning":  "Advertencia de rendimiento: {0}",
		"maintenance_mode":     "Modo de mantenimiento {0}",
		"service_started":      "Servicio {0} iniciado",
		"service_stopped":      "Servicio {0} detenido",
		"service_restarted":    "Servicio {0} reiniciado",
		"service_failed":       "Servicio {0} falló: {1}",
	}

	// French messages
	frenchMessages = map[string]string{
		"app_started":          "Application démarrée",
		"app_shutdown":         "Application en cours d'arrêt",
		"user_login":           "Utilisateur {0} connecté depuis {1}",
		"user_logout":          "Utilisateur {0} déconnecté",
		"file_uploaded":        "Fichier {0} téléchargé avec succès ({1} octets)",
		"file_upload_failed":   "Échec du téléchargement de fichier: {0}",
		"db_connected":         "Base de données connectée avec succès",
		"db_connection_failed": "Échec de connexion à la base de données: {0}",
		"api_request":          "Requête API {0} {1} depuis {2}",
		"api_response":         "Réponse API {0} {1} en {2}ms",
		"error_occurred":       "Une erreur s'est produite: {0}",
		"warning_detected":     "Avertissement détecté: {0}",
		"info_message":         "Information: {0}",
		"debug_info":           "Informations de débogage: {0}",
		"config_loaded":        "Configuration chargée depuis {0}",
		"config_saved":         "Configuration sauvegardée dans {0}",
		"backup_created":       "Sauvegarde créée: {0}",
		"backup_failed":        "Échec de la sauvegarde: {0}",
		"email_sent":           "E-mail envoyé à {0}",
		"email_failed":         "Échec de l'envoi d'e-mail: {0}",
		"task_completed":       "Tâche {0} terminée avec succès",
		"task_failed":          "Tâche {0} échouée: {1}",
		"memory_usage":         "Utilisation de la mémoire: {0}MB",
		"cpu_usage":            "Utilisation du CPU: {0}%",
		"disk_usage":           "Utilisation du disque: {0}%",
		"network_activity":     "Activité réseau: {0} octets envoyés, {1} octets reçus",
		"security_alert":       "Alerte de sécurité: {0}",
		"performance_warning":  "Avertissement de performance: {0}",
		"maintenance_mode":     "Mode maintenance {0}",
		"service_started":      "Service {0} démarré",
		"service_stopped":      "Service {0} arrêté",
		"service_restarted":    "Service {0} redémarré",
		"service_failed":       "Service {0} échoué: {1}",
	}
)

// LoadBuiltinCatalogs loads built-in message catalogs
func LoadBuiltinCatalogs(lm *LocalizationManager) {
	// English
	enCatalog := NewMessageCatalog(Locale{Language: "en", Region: "US"})
	for key, message := range englishMessages {
		enCatalog.AddMessage(key, message)
	}
	lm.AddCatalog(enCatalog)

	// Spanish
	esCatalog := NewMessageCatalog(Locale{Language: "es"})
	for key, message := range spanishMessages {
		esCatalog.AddMessage(key, message)
	}
	lm.AddCatalog(esCatalog)

	// French
	frCatalog := NewMessageCatalog(Locale{Language: "fr"})
	for key, message := range frenchMessages {
		frCatalog.AddMessage(key, message)
	}
	lm.AddCatalog(frCatalog)
}

// LocalizedLogger extends LoggerCore with localization support
type LocalizedLogger struct {
	*LoggerCore
	localization *LocalizationManager
}

// NewLocalizedLogger creates a new localized logger
func NewLocalizedLogger(config LoggerConfig, defaultLocale Locale) *LocalizedLogger {
	core := NewLoggerCore(config)
	localization := NewLocalizationManager(defaultLocale)
	LoadBuiltinCatalogs(localization)

	return &LocalizedLogger{
		LoggerCore:   core,
		localization: localization,
	}
}

// SetLocale sets the current locale for the logger
func (l *LocalizedLogger) SetLocale(locale Locale) {
	l.localization.SetCurrentLocale(locale)
}

// GetLocale returns the current locale
func (l *LocalizedLogger) GetLocale() Locale {
	return l.localization.GetCurrentLocale()
}

// T translates and logs a message
func (l *LocalizedLogger) T(level LogLevel, key string, args ...interface{}) {
	message := l.localization.Translate(key, args...)
	l.Log(level, getPrefixForLevel(level), message)
}

// TInfo translates and logs an info message
func (l *LocalizedLogger) TInfo(key string, args ...interface{}) {
	l.T(InfoLevel, key, args...)
}

// TSuccess translates and logs a success message
func (l *LocalizedLogger) TSuccess(key string, args ...interface{}) {
	message := l.localization.Translate(key, args...)
	l.Log(InfoLevel, SuccessPrefix, message)
}

// TWarning translates and logs a warning message
func (l *LocalizedLogger) TWarning(key string, args ...interface{}) {
	l.T(WarningLevel, key, args...)
}

// TError translates and logs an error message
func (l *LocalizedLogger) TError(key string, args ...interface{}) {
	l.T(ErrorLevel, key, args...)
}

// TDebug translates and logs a debug message
func (l *LocalizedLogger) TDebug(key string, args ...interface{}) {
	l.T(DebugLevel, key, args...)
}

// TTrace translates and logs a trace message
func (l *LocalizedLogger) TTrace(key string, args ...interface{}) {
	l.T(TraceLevel, key, args...)
}

// TPanic translates and logs a panic message
func (l *LocalizedLogger) TPanic(key string, args ...interface{}) {
	message := l.localization.Translate(key, args...)
	l.Log(PanicLevel, PanicPrefix, message)
	panic(message)
}

// TWithContext translates and logs a message with context
func (l *LocalizedLogger) TWithContext(level LogLevel, key string, context map[string]interface{}, args ...interface{}) {
	message := l.localization.Translate(key, args...)
	l.LogWithContext(level, getPrefixForLevel(level), message, context)
}

// TInfoWithContext translates and logs an info message with context
func (l *LocalizedLogger) TInfoWithContext(key string, context map[string]interface{}, args ...interface{}) {
	l.TWithContext(InfoLevel, key, context, args...)
}

// TErrorWithContext translates and logs an error message with context
func (l *LocalizedLogger) TErrorWithContext(key string, context map[string]interface{}, args ...interface{}) {
	l.TWithContext(ErrorLevel, key, context, args...)
}

// AddCustomMessage adds a custom message to the current locale
func (l *LocalizedLogger) AddCustomMessage(key, message string) {
	locale := l.localization.GetCurrentLocale()
	if catalog := l.localization.GetCatalog(locale); catalog != nil {
		catalog.AddMessage(key, message)
	}
}

// LoadCustomCatalog loads a custom message catalog from a file
func (l *LocalizedLogger) LoadCustomCatalog(locale Locale, filePath string) error {
	catalog := NewMessageCatalog(locale)
	if err := catalog.LoadFromFile(filePath); err != nil {
		return err
	}
	l.localization.AddCatalog(catalog)
	return nil
}

// getPrefixForLevel returns the appropriate prefix for a log level
func getPrefixForLevel(level LogLevel) string {
	switch level {
	case PanicLevel:
		return PanicPrefix
	case ErrorLevel:
		return ErrorPrefix
	case WarningLevel:
		return WarningPrefix
	case InfoLevel:
		return InfoPrefix
	case DebugLevel:
		return DebugPrefix
	case TraceLevel:
		return TracePrefix
	default:
		return InfoPrefix
	}
}

// Convenience functions for global localized logging
var globalLocalizedLogger *LocalizedLogger

// InitLocalizedLogger initializes the global localized logger
func InitLocalizedLogger(config LoggerConfig, defaultLocale Locale) {
	globalLocalizedLogger = NewLocalizedLogger(config, defaultLocale)
}

// SetGlobalLocale sets the locale for the global localized logger
func SetGlobalLocale(locale Locale) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.SetLocale(locale)
	}
}

// TInfo translates and logs an info message using the global logger
func TInfo(key string, args ...interface{}) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.TInfo(key, args...)
	}
}

// TSuccess translates and logs a success message using the global logger
func TSuccess(key string, args ...interface{}) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.TSuccess(key, args...)
	}
}

// TWarning translates and logs a warning message using the global logger
func TWarning(key string, args ...interface{}) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.TWarning(key, args...)
	}
}

// TError translates and logs an error message using the global logger
func TError(key string, args ...interface{}) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.TError(key, args...)
	}
}

// TDebug translates and logs a debug message using the global logger
func TDebug(key string, args ...interface{}) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.TDebug(key, args...)
	}
}

// TTrace translates and logs a trace message using the global logger
func TTrace(key string, args ...interface{}) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.TTrace(key, args...)
	}
}

// TPanic translates and logs a panic message using the global logger
func TPanic(key string, args ...interface{}) {
	if globalLocalizedLogger != nil {
		globalLocalizedLogger.TPanic(key, args...)
	}
}
