package main

import pim "github.com/refactorroom/pim"

// Example_basic demonstrates basic usage of pim logging.
func Example_basic() {
	pim.Info("Hello from pim!")
	pim.Success("Operation completed successfully!")
	pim.Warning("This is a warning message")
	pim.Error("This is an error message")
}
