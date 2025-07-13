package main

import (
	"fmt"
	"time"

	"pim"
)

func main() {
	fmt.Println("=== Enhanced Caller Information Demo ===\n")

	// 1. Basic caller info with default configuration
	fmt.Println("1. Basic Caller Info (Default Configuration)")
	basicLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: pim.NewCallerInfoConfig(),
	})

	basicLogger.Info("This is a basic log message with caller info")
	basicLogger.Debug("Debug message with caller info")
	basicLogger.Error("Error message with caller info")
	fmt.Println()

	// 2. Minimal caller info (file:line only)
	fmt.Println("2. Minimal Caller Info (File:Line Only)")
	minimalConfig := pim.NewCallerInfoConfig()
	minimalConfig.ShowPackage = false
	minimalConfig.ShowFunction = false
	minimalConfig.ShowGoroutineID = false
	minimalConfig.Format = "{file}:{line}"

	minimalLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: minimalConfig,
	})

	minimalLogger.Info("Minimal caller info - just file and line")
	fmt.Println()

	// 3. Detailed caller info with full paths
	fmt.Println("3. Detailed Caller Info (Full Paths)")
	detailedConfig := pim.NewCallerInfoConfig()
	detailedConfig.ShowFullPath = true
	detailedConfig.IncludeRuntime = true
	detailedConfig.IncludeTest = true
	detailedConfig.IncludeVendor = true
	detailedConfig.Format = "{fullpath}:{line} {pkg:func} {goroutine}"

	detailedLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: detailedConfig,
	})

	detailedLogger.Info("Detailed caller info with full paths and runtime info")
	fmt.Println()

	// 4. Production-optimized caller info
	fmt.Println("4. Production-Optimized Caller Info")
	prodConfig := pim.NewCallerInfoConfig()
	prodConfig.ShowPackage = false
	prodConfig.ShowGoroutineID = false
	prodConfig.IncludeRuntime = false
	prodConfig.IncludeTest = false
	prodConfig.IncludeVendor = false
	prodConfig.CacheEnabled = true
	prodConfig.CacheSize = 5000
	prodConfig.Format = "{file}:{line} {function}"

	prodLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.InfoLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: prodConfig,
	})

	prodLogger.Info("Production-optimized caller info")
	fmt.Println()

	// 5. Custom format with placeholders
	fmt.Println("5. Custom Format with Placeholders")
	customConfig := pim.NewCallerInfoConfig()
	customConfig.Format = "[{file}:{line}] {pkg:func} (depth:{depth})"
	customConfig.Separator = " | "

	customLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: customConfig,
	})

	customLogger.Info("Custom formatted caller info")
	fmt.Println()

	// 6. Caller info with filtering
	fmt.Println("6. Caller Info with Filtering")
	filterConfig := pim.NewCallerInfoConfig()
	filterConfig.ExcludePatterns = []string{
		`^runtime\.`,
		`^reflect\.`,
		`^syscall\.`,
		`^internal/`,
		`^vendor/`,
	}
	filterConfig.PackageFilter = "main"
	filterConfig.Format = "{file}:{line} {function}"

	filterLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: filterConfig,
	})

	filterLogger.Info("Filtered caller info - excludes runtime and vendor packages")
	fmt.Println()

	// 7. Caller info with depth control
	fmt.Println("7. Caller Info with Depth Control")
	depthConfig := pim.NewCallerInfoConfig()
	depthConfig.CallDepth = 1
	depthConfig.MinCallDepth = 1
	depthConfig.MaxCallDepth = 5
	depthConfig.Format = "{file}:{line} {function} (depth:{depth})"

	depthLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: depthConfig,
	})

	// Call through a helper function to demonstrate depth
	callThroughHelper(depthLogger)
	fmt.Println()

	// 8. Enhanced stack traces
	fmt.Println("8. Enhanced Stack Traces")
	stackConfig := pim.NewCallerInfoConfig()
	stackConfig.StackDepth = 8
	stackConfig.Format = "{file}:{line} {function}"

	stackLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: stackConfig,
	})

	// Demonstrate stack trace
	stackLogger.Error("Error with enhanced stack trace")
	fmt.Println()

	// 9. Cache performance demonstration
	fmt.Println("9. Cache Performance Demonstration")
	cacheConfig := pim.NewCallerInfoConfig()
	cacheConfig.CacheEnabled = true
	cacheConfig.CacheSize = 100
	cacheConfig.Format = "{file}:{line} {function}"

	cacheLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    false, // Disable console to focus on cache
		CallerInfoConfig: cacheConfig,
	})

	// Log multiple messages to demonstrate caching
	start := time.Now()
	for i := 0; i < 100; i++ {
		cacheLogger.Info(fmt.Sprintf("Message %d", i))
	}
	duration := time.Since(start)

	stats := cacheLogger.GetCallerCacheStats()
	fmt.Printf("Logged 100 messages in %v\n", duration)
	fmt.Printf("Cache stats: %+v\n", stats)
	fmt.Println()

	// 10. Dynamic configuration changes
	fmt.Println("10. Dynamic Configuration Changes")
	dynamicLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: pim.NewCallerInfoConfig(),
	})

	dynamicLogger.Info("Original format")

	// Change configuration dynamically
	newConfig := pim.NewCallerInfoConfig()
	newConfig.Format = "🔍 {file}:{line} | {function}"
	dynamicLogger.SetCallerInfoConfig(newConfig)

	dynamicLogger.Info("New format after dynamic change")
	fmt.Println()

	// 11. Convenience constructors demonstration
	fmt.Println("11. Convenience Constructors")

	// Minimal caller info
	minimalLogger2 := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: pim.NewMinimalCallerInfo().GetConfig(),
	})
	minimalLogger2.Info("Using NewMinimalCallerInfo()")

	// Production caller info
	prodLogger2 := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: pim.NewProductionCallerInfo().GetConfig(),
	})
	prodLogger2.Info("Using NewProductionCallerInfo()")

	// Detailed caller info
	detailedLogger2 := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: pim.NewDetailedCallerInfo().GetConfig(),
	})
	detailedLogger2.Info("Using NewDetailedCallerInfo()")
	fmt.Println()

	// 12. Direct caller info access
	fmt.Println("12. Direct Caller Info Access")
	directLogger := pim.NewLoggerCore(pim.LoggerConfig{
		Level:            pim.DebugLevel,
		EnableConsole:    true,
		EnableColors:     true,
		CallerInfoConfig: pim.NewCallerInfoConfig(),
	})

	// Get caller info directly
	callerInfo := directLogger.GetCallerInfo(1)
	fmt.Printf("Direct caller info: %+v\n", callerInfo)

	// Get caller info at specific depth
	callerInfo2 := directLogger.GetCallerInfoAtDepth(2)
	fmt.Printf("Caller info at depth 2: %+v\n", callerInfo2)

	// Get enhanced stack trace
	stackTrace := directLogger.GetEnhancedStackTrace(1)
	fmt.Printf("Enhanced stack trace (%d frames):\n", len(stackTrace))
	for i, frame := range stackTrace {
		fmt.Printf("  %d: %s\n", i, directLogger.FormatCallerInfo(frame))
	}
	fmt.Println()

	fmt.Println("=== Enhanced Caller Information Demo Complete ===")
}

// Helper function to demonstrate call depth
func callThroughHelper(logger *pim.LoggerCore) {
	logger.Info("This message is called through a helper function")
}
