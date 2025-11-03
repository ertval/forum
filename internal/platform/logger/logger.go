package logger
// Package logger provides structured logging for the application.
// It wraps the standard library log package with additional functionality
// for log levels, structured fields, and context-aware logging.
package logger

// Logger provides structured logging capabilities.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// logger is the default logger implementation.
type logger struct {
	// TODO: Add logger fields
}

// New creates a new logger instance.
func New(level string) Logger {
	// TODO: Implement logger
	return &logger{}
}

// String creates a string field for logging.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field for logging.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field for logging.
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}
