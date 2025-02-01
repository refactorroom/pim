package console

import (
	"fmt"
)

// Print formats using the default formats for its operands and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Print(args ...interface{}) (n int, err error) {
	msg := fmt.Sprint(args...)
	LogWithTimestamp(printPrefix, msg, InfoLevel)
	return fmt.Print(msg)
}

// Println formats using the default formats for its operands and writes to standard output.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Println(args ...interface{}) (n int, err error) {
	msg := fmt.Sprintln(args...)
	// Remove trailing newline as LogWithTimestamp adds one
	msg = msg[:len(msg)-1]
	LogWithTimestamp(printPrefix, msg, InfoLevel)
	return fmt.Println(args...)
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, args ...interface{}) (n int, err error) {
	msg := fmt.Sprintf(format, args...)
	LogWithTimestamp(printPrefix, msg, InfoLevel)
	return fmt.Print(msg)
}

// Scanf scans text read from standard input, storing successive space-separated
// values into successive arguments as determined by the format. It returns
// the number of items successfully scanned and any error encountered.
func Scanf(format string, args ...interface{}) (n int, err error) {
	n, err = fmt.Scanf(format, args...)
	if err != nil {
		LogWithTimestamp(ErrorPrefix, fmt.Sprintf("Scanf error: %v", err), ErrorLevel)
	} else {
		LogWithTimestamp(DebugPrefix, fmt.Sprintf("Scanf read %d items", n), DebugLevel)
	}
	return n, err
}

// Scanln is similar to Scanf, but stops scanning at a newline and
// after the final item there must be a newline or EOF.
func Scanln(args ...interface{}) (n int, err error) {
	n, err = fmt.Scanln(args...)
	if err != nil {
		LogWithTimestamp(ErrorPrefix, fmt.Sprintf("Scanln error: %v", err), ErrorLevel)
	} else {
		LogWithTimestamp(DebugPrefix, fmt.Sprintf("Scanln read %d items", n), DebugLevel)
	}
	return n, err
}

// printPrefix is the prefix used for print statements in the log
const printPrefix = "📄"

// You might want to add these to your existing prefix definitions:
// const (
//     printPrefix    = "📄"
// )
