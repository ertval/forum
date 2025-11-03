// Package logger provides structured logging for the forum application.
// It supports multiple log levels and outputs logs in JSON format for production
// and human-readable format for development.
package logger

import (
	"io"
)

// Level represents the log level.
type Level int

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production.
	DebugLevel Level = iota
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual human review.
	WarnLevel
	// ErrorLevel logs are high-priority. Applications running smoothly shouldn't generate errors.
	ErrorLevel
)

// Logger represents a structured logger.
// It provides methods for logging at different levels with structured fields.
type Logger struct {
	level  Level
	output io.Writer
}

// New creates a new logger with the specified level and output.
// TODO: Implement logger initialization.
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		output: output,
	}
}

// Debug logs a debug message with optional fields.
// Debug logs are typically used for detailed debugging information.
func (l *Logger) Debug(msg string, fields ...Field) {
	// Implementation placeholder
}

// Info logs an info message with optional fields.
// Info logs are used for general informational messages.
func (l *Logger) Info(msg string, fields ...Field) {
	// Implementation placeholder
}

// Warn logs a warning message with optional fields.
// Warn logs indicate something unexpected but not critical.
func (l *Logger) Warn(msg string, fields ...Field) {
	// Implementation placeholder
}

// Error logs an error message with optional fields.
// Error logs indicate a failure that should be investigated.
func (l *Logger) Error(msg string, fields ...Field) {
	// Implementation placeholder
}

// WithFields returns a new logger with additional fields.
// The returned logger will include these fields in all log messages.
func (l *Logger) WithFields(fields ...Field) *Logger {
	// Implementation placeholder
	return l
}

// Field represents a structured log field (key-value pair).
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field.
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field.
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Any creates a field with any value type.
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
