package main

import (
	"fmt"
	"time"

	"github.com/refactorroom/pim"
)

func main() {
	fmt.Println("=== Localization/Internationalization Demo ===\n")

	// Initialize localized logger with English as default
	config := pim.LoggerConfig{
		Level:         pim.InfoLevel,
		ServiceName:   "localization-demo",
		EnableConsole: true,
		EnableColors:  true,
	}

	// Create localized logger with English as default
	logger := pim.NewLocalizedLogger(config, pim.Locale{Language: "en", Region: "US"})
	defer logger.Close()

	// Example 1: Basic localized logging
	fmt.Println("1. Basic Localized Logging (English):")
	logger.TInfo("app_started")
	logger.TSuccess("user_login", "john_doe", "192.168.1.100")
	logger.TWarning("performance_warning", "High memory usage detected")
	logger.TError("db_connection_failed", "Connection timeout")

	// Example 2: Switch to Spanish
	fmt.Println("\n2. Switching to Spanish:")
	logger.SetLocale(pim.Locale{Language: "es"})
	logger.TInfo("app_started")
	logger.TSuccess("user_login", "maria_garcia", "10.0.0.50")
	logger.TWarning("performance_warning", "Uso de memoria alto detectado")
	logger.TError("db_connection_failed", "Tiempo de conexión agotado")

	// Example 3: Switch to French
	fmt.Println("\n3. Switching to French:")
	logger.SetLocale(pim.Locale{Language: "fr"})
	logger.TInfo("app_started")
	logger.TSuccess("user_login", "pierre_dupont", "172.16.0.25")
	logger.TWarning("performance_warning", "Utilisation élevée de la mémoire détectée")
	logger.TError("db_connection_failed", "Délai de connexion dépassé")

	// Example 4: Back to English
	fmt.Println("\n4. Back to English:")
	logger.SetLocale(pim.Locale{Language: "en", Region: "US"})
	logger.TInfo("app_started")
	logger.TInfo("config_loaded", "/etc/app/config.json")
	logger.TInfo("service_started", "web-server")
	logger.TInfo("backup_created", "backup-2024-01-15.zip")

	// Example 5: Context-aware logging
	fmt.Println("\n5. Context-Aware Logging:")
	logger.TInfoWithContext("api_request", map[string]interface{}{
		"method":     "POST",
		"endpoint":   "/api/users",
		"user_agent": "Mozilla/5.0...",
		"ip":         "203.0.113.1",
	}, "POST", "/api/users", "203.0.113.1")

	logger.TErrorWithContext("task_failed", map[string]interface{}{
		"task_id":     "task_12345",
		"duration":    "2.5s",
		"retry_count": 3,
	}, "data_processing", "Invalid input format")

	// Example 6: Custom messages
	fmt.Println("\n6. Custom Messages:")
	logger.AddCustomMessage("custom_welcome", "Welcome to our application, {0}!")
	logger.AddCustomMessage("custom_goodbye", "Goodbye, {0}! Come back soon!")
	logger.AddCustomMessage("custom_processing", "Processing {0} items... ({1}% complete)")

	logger.TInfo("custom_welcome", "Alice")
	logger.TInfo("custom_processing", 1000, 75)
	logger.TInfo("custom_goodbye", "Alice")

	// Example 7: Different locales with custom messages
	fmt.Println("\n7. Custom Messages in Different Locales:")

	// Add Spanish custom messages
	logger.SetLocale(pim.Locale{Language: "es"})
	logger.AddCustomMessage("custom_welcome", "¡Bienvenido a nuestra aplicación, {0}!")
	logger.AddCustomMessage("custom_goodbye", "¡Adiós, {0}! ¡Vuelve pronto!")
	logger.AddCustomMessage("custom_processing", "Procesando {0} elementos... ({1}% completado)")

	logger.TInfo("custom_welcome", "María")
	logger.TInfo("custom_processing", 500, 60)
	logger.TInfo("custom_goodbye", "María")

	// Add French custom messages
	logger.SetLocale(pim.Locale{Language: "fr"})
	logger.AddCustomMessage("custom_welcome", "Bienvenue dans notre application, {0} !")
	logger.AddCustomMessage("custom_goodbye", "Au revoir, {0} ! Revenez bientôt !")
	logger.AddCustomMessage("custom_processing", "Traitement de {0} éléments... ({1}% terminé)")

	logger.TInfo("custom_welcome", "Pierre")
	logger.TInfo("custom_processing", 750, 90)
	logger.TInfo("custom_goodbye", "Pierre")

	// Example 8: System monitoring messages
	fmt.Println("\n8. System Monitoring Messages:")
	logger.SetLocale(pim.Locale{Language: "en", Region: "US"})

	logger.TInfo("memory_usage", 512)
	logger.TInfo("cpu_usage", 85)
	logger.TInfo("disk_usage", 67)
	logger.TInfo("network_activity", 1024000, 2048000)

	// Example 9: Security and maintenance messages
	fmt.Println("\n9. Security and Maintenance Messages:")
	logger.TWarning("security_alert", "Multiple failed login attempts detected")
	logger.TInfo("maintenance_mode", "enabled")
	logger.TInfo("service_restarted", "database")
	logger.TError("service_failed", "email-service", "SMTP server unreachable")

	// Example 10: File operations
	fmt.Println("\n10. File Operations:")
	logger.TInfo("file_uploaded", "document.pdf", 2048576)
	logger.TError("file_upload_failed", "File size exceeds limit")
	logger.TInfo("config_saved", "/home/user/.config/app.json")

	// Example 11: Email operations
	fmt.Println("\n11. Email Operations:")
	logger.TInfo("email_sent", "user@example.com")
	logger.TError("email_failed", "SMTP authentication failed")

	// Example 12: Task management
	fmt.Println("\n12. Task Management:")
	logger.TInfo("task_completed", "data_export")
	logger.TError("task_failed", "backup_job", "Insufficient disk space")

	// Example 13: Debug and trace messages
	fmt.Println("\n13. Debug and Trace Messages:")
	logger.SetLevel(pim.DebugLevel)
	logger.TDebug("debug_info", "Processing request ID: 12345")
	logger.TTrace("debug_info", "Entering function: processUserData")

	// Example 14: Locale detection and fallback
	fmt.Println("\n14. Locale Detection and Fallback:")

	// Test with unsupported locale (should fallback to English)
	logger.SetLocale(pim.Locale{Language: "de"}) // German not supported
	logger.TInfo("app_started")                  // Should show English message

	// Test with region-specific locale
	logger.SetLocale(pim.Locale{Language: "en", Region: "GB"}) // British English
	logger.TInfo("app_started")                                // Should fallback to US English

	// Example 15: Performance demonstration
	fmt.Println("\n15. Performance Demonstration:")
	logger.SetLocale(pim.Locale{Language: "en", Region: "US"})

	start := time.Now()
	for i := 0; i < 1000; i++ {
		logger.TInfo("info_message", fmt.Sprintf("Message %d", i))
	}
	duration := time.Since(start)

	fmt.Printf("\nLogged 1000 localized messages in %v\n", duration)

	// Example 16: Global localized logger
	fmt.Println("\n16. Global Localized Logger:")

	// Initialize global logger
	pim.InitLocalizedLogger(config, pim.Locale{Language: "en", Region: "US"})

	// Use global functions
	pim.TInfo("app_started")
	pim.TSuccess("user_login", "global_user", "127.0.0.1")
	pim.TWarning("performance_warning", "Global warning message")

	// Switch locale globally
	pim.SetGlobalLocale(pim.Locale{Language: "es"})
	pim.TInfo("app_started")
	pim.TSuccess("user_login", "usuario_global", "127.0.0.1")

	// Final summary
	fmt.Println("\n=== Localization Demo Summary ===")
	fmt.Printf("Current locale: %s\n", logger.GetLocale().String())
	fmt.Printf("Supported locales: en-US, es, fr\n")
	fmt.Printf("Fallback chain: %s\n", logger.GetLocale().String())

	fmt.Println("\nDemo completed successfully!")
	fmt.Println("All messages were properly localized and logged.")
}
