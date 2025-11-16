// Package logger provides structured logging for the forum application.
// It supports multiple log levels and outputs logs in JSON format for production
// and human-readable format for development.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
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
	mu     sync.Mutex
	// persistent fields included on every log entry
	fields []Field
	// human indicates whether to use human-readable output (true)
	// or JSON output (false).
	human bool
	// config controls human formatting options (only used when human==true)
	config *Config
}

// TimePrecision controls how timestamps are rendered in human output.
type TimePrecision int

const (
	// TimePrecisionSeconds prints up to seconds (yyyy-mm-ddThh:mm:ss)
	TimePrecisionSeconds TimePrecision = iota
	// TimePrecisionNano preserves full RFC3339Nano formatting
	TimePrecisionNano
)

// Config holds runtime options for human log formatting.
type Config struct {
	// TimePrecision controls how much time detail to include (seconds or nano).
	TimePrecision TimePrecision
	// OmitFields lists field keys to exclude from human output (e.g. "user_agent").
	OmitFields []string
	// MaxLineWidth limits the length of a single human-readable log line.
	// If <= 0 no truncation is applied.
	MaxLineWidth int
}

// New creates a new logger with the specified level and output.
// TODO: Implement logger initialization.
func New(level Level, output io.Writer) *Logger {
	// Decide whether to use human-readable output. Use human readable
	// when output is a terminal (stdout/stderr). We conservatively
	// treat os.Stdout and os.Stderr as terminals.
	human := false
	if output == os.Stdout || output == os.Stderr {
		human = true
	}

	// default config for human output: seconds precision, omit user_agent, 80 cols
	defCfg := &Config{
		TimePrecision: TimePrecisionSeconds,
		OmitFields:    []string{"user_agent"},
		MaxLineWidth:  80,
	}

	return &Logger{
		level:  level,
		output: output,
		fields: nil,
		human:  human,
		config: defCfg,
	}
}

// NewWithConfig creates a new Logger and accepts a formatting Config used for human output.
// If cfg is nil, defaults are applied. JSON output is unaffected by these config options.
func NewWithConfig(level Level, output io.Writer, cfg *Config) *Logger {
	l := New(level, output)
	if cfg == nil {
		return l
	}
	l.config = cfg
	return l
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
	// Return a new logger that includes the additional persistent fields.
	// Do a defensive copy of slices so callers can mutate after.
	newFields := make([]Field, 0, len(l.fields)+len(fields))
	newFields = append(newFields, l.fields...)
	newFields = append(newFields, fields...)

	return &Logger{
		level:  l.level,
		output: l.output,
		fields: newFields,
		human:  l.human,
		config: l.config,
	}
}

// internal helper: convert Level to string
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

// internal log implementation
func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	// Merge persistent fields and call-site fields into a map
	data := make(map[string]any)
	for _, f := range l.fields {
		// later fields override earlier ones
		data[f.Key] = f.Value
	}
	for _, f := range fields {
		data[f.Key] = f.Value
	}

	// base entry
	entry := map[string]any{
		"level": levelToString(level),
		"msg":   msg,
		"ts":    time.Now().Format(time.RFC3339Nano),
	}

	// attach fields
	if len(data) > 0 {
		entry["fields"] = data
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.human {
		// human readable: [LEVEL] ts msg key=val ...
		ts := entry["ts"].(string)
		// apply configured time precision if available
		if l.config != nil && l.config.TimePrecision == TimePrecisionSeconds {
			if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
				ts = parsed.Format("2006-01-02T15:04:05")
			}
		}

		// prepare fields, respecting omit list
		omit := map[string]struct{}{}
		if l.config != nil {
			for _, k := range l.config.OmitFields {
				omit[k] = struct{}{}
			}
		}

		out := fmt.Sprintf("[%s] %s %s", entry["level"], ts, entry["msg"])
		if len(data) > 0 {
			for k, v := range data {
				if _, ok := omit[k]; ok {
					continue
				}
				out += fmt.Sprintf(" %s=%v", k, v)
			}
		}
		// apply max line width truncation when requested
		if l.config != nil && l.config.MaxLineWidth > 0 {
			// ensure newline at end afterwards
			truncated := truncateToWidth(out, l.config.MaxLineWidth)
			out = truncated
		}
		out += "\n"
		_, _ = l.output.Write([]byte(out))
		return
	}

	// JSON output
	enc, err := json.Marshal(entry)
	if err != nil {
		// fallback to fmt if JSON encoding fails
		fallback := fmt.Sprintf("%s %s %v\n", levelToString(level), msg, data)
		_, _ = l.output.Write([]byte(fallback))
		return
	}
	enc = append(enc, '\n')
	_, _ = l.output.Write(enc)
}

// Field represents a structured log field (key-value pair).
type Field struct {
	Key   string
	Value any
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
func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field (in milliseconds).
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.Milliseconds()}
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
