package main

import (
	"fmt"
	"time"

	"github.com/refactorroom/pim"
)

func main() {
	fmt.Println("=== Graceful Shutdown and Flush Demo ===\n")

	// Install global exit handler for all loggers
	pim.InstallExitHandler()

	config := pim.LoggerConfig{
		Level:         pim.InfoLevel,
		ServiceName:   "graceful-demo",
		EnableConsole: true,
		Async:         true,
		BufferSize:    10,
		FlushInterval: 5 * time.Second,
	}

	logger := pim.NewLoggerCore(config)
	defer logger.Close()

	fmt.Println("Logging messages...")
	for i := 1; i <= 5; i++ {
		logger.InfoKV("Message before panic", "sequence", i)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Simulating panic...")
	time.Sleep(1 * time.Second)
	panic("Simulated panic for graceful shutdown demo")

	// This will not be reached, but logs will be flushed
	logger.Info("This message should not appear")
}
