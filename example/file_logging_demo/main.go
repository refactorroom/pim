package main

import (
	"fmt"
	"os"

	pim "github.com/refactorroom/pim"
)

const (
	fileLoggingStatusFormat = "File logging enabled: %v\n"
	logsDirectory           = "./logs"
)

func main() {
	fmt.Println("=== Testing File Logging Disabled by Default ===")
	pim.Info("This log will only appear in console")
	pim.Error("This error will only appear in console")

	fmt.Printf(fileLoggingStatusFormat, pim.GetFileLogging())

	fmt.Println("\n=== Enabling File Logging ===")

	// Method 1: Simple enable
	pim.SetFileLogging(true)
	pim.EnableFileLogging(true) // This enables it in the stamper config
	fmt.Printf(fileLoggingStatusFormat, pim.GetFileLogging())
	// Method 2: Enable with initialization
	err := pim.InitializeFileLogging(logsDirectory, "demo-app")
	if err != nil {
		fmt.Printf("Error initializing file logging: %v\n", err)
	} else {
		fmt.Println("File logging initialized successfully")
	}

	fmt.Println("\n=== Testing File Logging Enabled ===")
	pim.Info("This log will appear in both console and file")
	pim.Error("This error will appear in both console and file")

	fmt.Println("\n=== Disabling File Logging ===")
	pim.SetFileLogging(false)
	pim.EnableFileLogging(false)
	fmt.Printf(fileLoggingStatusFormat, pim.GetFileLogging())

	pim.Info("This log will only appear in console again")

	// Clean up
	pim.CloseLogFiles()

	// Check if log files were created
	if _, err := os.Stat(logsDirectory); err == nil {
		fmt.Println("Log directory was created successfully")
		// List files in logs directory
		files, err := os.ReadDir(logsDirectory)
		if err == nil {
			fmt.Printf("Log files created: %d\n", len(files))
			for _, file := range files {
				fmt.Printf("  - %s\n", file.Name())
			}
		}
	}
}
