// Package logger provides structured logging for the forum application.
// It supports multiple log levels and outputs logs in JSON format for production
// and human-readable format for development.
//
// Internally, this package delegates to Go's standard log/slog package,
// using custom slog.Handler implementations to maintain the forum's
// specific output formats.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
	"unicode"
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
// Internally delegates to a *slog.Logger with custom handlers.
type Logger struct {
	slogger *slog.Logger
	level   *slog.LevelVar
	output  io.Writer
	human   bool
	config  *Config
}

// TimePrecision controls how timestamps are rendered in human output.
type TimePrecision int

const (
	// TimePrecisionSeconds prints up to seconds (yyyy-mm-dd hh:mm:ss)
	TimePrecisionSeconds TimePrecision = iota
	// TimePrecisionNano preserves full RFC3339Nano formatting
	TimePrecisionNano
)

// ANSI color codes for terminal human output.
const (
	colorReset   = "\x1b[0m"
	colorRed     = "\x1b[31m"
	colorGreen   = "\x1b[32m"
	colorYellow  = "\x1b[33m"
	colorBlue    = "\x1b[34m"
	colorMagenta = "\x1b[35m"
	colorCyan    = "\x1b[36m"
	colorWhite   = "\x1b[37m"
)

// Config holds runtime options for human log formatting.
type Config struct {
	// TimePrecision controls how much time detail to include (seconds or nano).
	TimePrecision TimePrecision
	// OmitFields lists field keys to exclude from human output (e.g. "user_agent").
	OmitFields []string
	// AllowedFields, when non-empty, restricts human output to only these keys.
	// If empty, all fields are considered (except those in OmitFields).
	AllowedFields []string
	// MaxLineWidth limits the length of a single human-readable log line.
	// If <= 0 no truncation is applied.
	MaxLineWidth int
	// Colorize enables ANSI coloring in human output. Set to false to disable colors (e.g., when piping to files).
	Colorize bool
}

// defaultConfig returns the default human formatting config.
func defaultConfig() *Config {
	return &Config{
		TimePrecision: TimePrecisionSeconds,
		OmitFields:    []string{"user_agent"},
		AllowedFields: []string{"method", "path", "query", "status", "size", "duration_ms", "remote", "url", "response", "error", "errors", "proto"},
		MaxLineWidth:  200,
		Colorize:      true,
	}
}

// newLogger is the internal constructor that wires up the slog backend.
func newLogger(level Level, output io.Writer, human bool, cfg *Config) *Logger {
	levelVar := &slog.LevelVar{}
	levelVar.Set(levelToSlog(level))

	var handler slog.Handler
	if human {
		handler = newHumanHandler(output, cfg, levelVar)
	} else {
		handler = newJSONHandler(output, levelVar)
	}

	return &Logger{
		slogger: slog.New(handler),
		level:   levelVar,
		output:  output,
		human:   human,
		config:  cfg,
	}
}

// New creates a new logger with the specified level and output.
func New(level Level, output io.Writer) *Logger {
	// Decide whether to use human-readable output. Use human readable
	// when output is a terminal (stdout/stderr). We conservatively
	// treat os.Stdout and os.Stderr as terminals.
	human := false
	if output == os.Stdout || output == os.Stderr {
		human = true
	}
	return newLogger(level, output, human, defaultConfig())
}

// NewWithConfig creates a new Logger and accepts a formatting Config used for human output.
// If cfg is nil, defaults are applied. JSON output is unaffected by these config options.
func NewWithConfig(level Level, output io.Writer, cfg *Config) *Logger {
	l := New(level, output)
	if cfg == nil {
		return l
	}
	return newLogger(level, output, l.human, cfg)
}

// Debug logs a debug message with optional fields.
// Debug logs are typically used for detailed debugging information.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs an info message with optional fields.
// Info logs are used for general informational messages.
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a warning message with optional fields.
// Warn logs indicate something unexpected but not critical.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs an error message with optional fields.
// Error logs indicate a failure that should be investigated.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

// WithFields returns a new logger with additional fields.
// The returned logger will include these fields in all log messages.
func (l *Logger) WithFields(fields ...Field) *Logger {
	args := make([]any, len(fields))
	for i, f := range fields {
		args[i] = f.attr
	}
	return &Logger{
		slogger: l.slogger.With(args...),
		level:   l.level,
		output:  l.output,
		human:   l.human,
		config:  l.config,
	}
}

// log is the internal implementation that delegates to the slog backend.
func (l *Logger) log(level Level, msg string, fields ...Field) {
	slogLevel := levelToSlog(level)
	if !l.slogger.Enabled(context.Background(), slogLevel) {
		return
	}
	attrs := make([]slog.Attr, len(fields))
	for i, f := range fields {
		attrs[i] = f.attr
	}
	l.slogger.LogAttrs(context.Background(), slogLevel, msg, attrs...)
}

// Field represents a structured log field (key-value pair).
// Internally wraps a slog.Attr.
type Field struct {
	attr slog.Attr
}

// String creates a string field.
func String(key string, value string) Field {
	return Field{attr: slog.String(key, value)}
}

// Int creates an integer field.
func Int(key string, value int) Field {
	return Field{attr: slog.Int(key, value)}
}

// Error creates an error field.
// If err is nil, an empty string value is used to avoid a nil-pointer panic.
func Error(err error) Field {
	if err == nil {
		return Field{attr: slog.String("error", "")}
	}
	return Field{attr: slog.String("error", err.Error())}
}

// Any creates a field with any value type.
func Any(key string, value any) Field {
	return Field{attr: slog.Any(key, value)}
}

// Duration creates a duration field (in milliseconds).
func Duration(key string, value time.Duration) Field {
	return Field{attr: slog.Int64(key, value.Milliseconds())}
}

// levelToSlog converts a custom Level to a slog.Level.
func levelToSlog(l Level) slog.Level {
	switch l {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// slogToLevel converts a slog.Level back to a custom Level.
func slogToLevel(l slog.Level) Level {
	switch {
	case l < slog.LevelInfo:
		return DebugLevel
	case l < slog.LevelWarn:
		return InfoLevel
	case l < slog.LevelError:
		return WarnLevel
	default:
		return ErrorLevel
	}
}

// levelToString converts a Level to its string representation.
func levelToString(l Level) string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// truncateToWidth truncates the input string to at most width runes.
// If truncation occurs, an ellipsis (single rune) is appended to indicate truncation.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	rs := []rune(s)
	if len(rs) <= width {
		return s
	}
	// leave room for ellipsis
	if width <= 1 {
		return string(rs[:width])
	}
	truncated := string(rs[:width-1]) + "…"
	return truncated
}

func sanitizeFieldValuesForPlainText(data map[string]any) map[string]any {
	if len(data) == 0 {
		return data
	}

	sanitized := make(map[string]any, len(data))
	for key, value := range data {
		if strValue, ok := value.(string); ok {
			sanitized[key] = sanitizePlainText(strValue)
			continue
		}
		sanitized[key] = value
	}

	return sanitized
}

func sanitizePlainText(input string) string {
	if input == "" {
		return input
	}

	var builder strings.Builder
	builder.Grow(len(input))

	for _, r := range input {
		switch r {
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if unicode.IsControl(r) {
				builder.WriteRune(' ')
				continue
			}
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
