package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// DefaultLogPath is the default path for the log file
var DefaultLogPath = "./app.log"

// Logger handles file-based logging
type Logger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// DefaultLogger is the default logger instance
var DefaultLogger *Logger

// Init initializes the default logger with the given file path
// If filePath is empty, uses DefaultLogPath
func Init(filePath string) error {
	if filePath == "" {
		filePath = DefaultLogPath
	}

	logger := &Logger{
		filePath: filePath,
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logger.file = file
	DefaultLogger = logger

	return nil
}

// Log writes a formatted message with timestamp to the log file
func (l *Logger) Log(format string, args ...interface{}) {
	if l == nil || l.file == nil {
		// Fallback to stderr if logger not initialized
		fmt.Fprintf(os.Stderr, "[LOGGER ERROR] Logger not initialized: "+format+"\n", args...)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)

	_, err := l.file.WriteString(logLine)
	if err != nil {
		// Fallback to stderr if write fails
		fmt.Fprintf(os.Stderr, "[LOGGER ERROR] Failed to write to log file: %v\n", err)
		fmt.Fprintf(os.Stderr, logLine)
	}
}

// Close closes the log file
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	err := l.file.Close()
	l.file = nil
	return err
}

// Log is a convenience function that logs to the default logger
func Log(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Log(format, args...)
	} else {
		// Fallback to stderr if default logger not initialized
		fmt.Fprintf(os.Stderr, "[LOGGER ERROR] Default logger not initialized: "+format+"\n", args...)
	}
}

// Close closes the default logger
func Close() error {
	if DefaultLogger != nil {
		return DefaultLogger.Close()
	}
	return nil
}
