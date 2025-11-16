// Package logger provides structured logging for the forum application.
// It supports multiple log levels and outputs logs in JSON format for production
// and human-readable format for development.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
		// default to only show essential HTTP info in human output
		AllowedFields: []string{"url", "response", "status", "error", "errors"},
		MaxLineWidth:  200,
		Colorize:      true,
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

// colorize wraps a string with an ANSI color code when color is non-empty.
func colorize(s, color string) string {
	if color == "" {
		return s
	}
	return color + s + colorReset
}

// applyColor applies colorization depending on the logger config.
func (l *Logger) applyColor(s, color string) string {
	if l.config == nil {
		// default to color enabled for terminal output
		if color == "" {
			return s
		}
		return color + s + colorReset
	}
	if !l.config.Colorize || color == "" {
		return s
	}
	return color + s + colorReset
}

// colorForLevel returns the ANSI color for a log level.
func colorForLevel(l Level) string {
	switch l {
	case DebugLevel:
		return colorMagenta
	case InfoLevel:
		return colorGreen
	case WarnLevel:
		return colorYellow
	case ErrorLevel:
		return colorRed
	default:
		return colorWhite
	}
}

// colorForStatusCode returns a color based on HTTP status code ranges.
func colorForStatusCode(code int) string {
	switch {
	case code >= 200 && code < 300:
		return colorGreen
	case code >= 300 && code < 400:
		return colorCyan
	case code >= 400 && code < 500:
		return colorYellow
	case code >= 500:
		return colorRed
	default:
		return colorWhite
	}
}

// colorForMessage decides whether a message should be highlighted based on
// its content (e.g., server started/stopped) or the log level.
func colorForMessage(msg string, lvl Level) string {
	lower := strings.ToLower(msg)
	// Important positive messages
	if strings.Contains(lower, "started") || strings.Contains(lower, "listening") || strings.Contains(lower, "listening on") || strings.Contains(lower, "server started") {
		// server start / listening messages: use blue to highlight
		return colorBlue
	}
	// Stopping/shutdown
	if strings.Contains(lower, "stopped") || strings.Contains(lower, "shutdown") || strings.Contains(lower, "stopping") {
		return colorRed
	}
	// Errors and failures
	if lvl == ErrorLevel || strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
		return colorRed
	}
	return ""
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
		// human readable: [LEVEL] ts msg key=val ... with colors for level and status codes
		ts := entry["ts"].(string)
		// apply configured time precision if available
		if l.config != nil && l.config.TimePrecision == TimePrecisionSeconds {
			if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
				ts = parsed.Format("2006-01-02T15:04:05")
			}
		}

		// prepare fields: if AllowedFields is set, only include those keys.
		allowed := map[string]struct{}{}
		useAllowed := false
		if l.config != nil && len(l.config.AllowedFields) > 0 {
			for _, k := range l.config.AllowedFields {
				allowed[k] = struct{}{}
			}
			useAllowed = true
		}

		// colorize level label
		levelLabel := fmt.Sprintf("[%s]", entry["level"])
		levelColored := l.applyColor(levelLabel, colorForLevel(level))

		// colorize message when it's important (server start/stop, errors, etc.)
		msgStr := fmt.Sprintf("%s", entry["msg"])
		msgColor := colorForMessage(msgStr, level)
		if msgColor != "" {
			msgStr = l.applyColor(msgStr, msgColor)
		}

		out := fmt.Sprintf("%s %s %s", levelColored, ts, msgStr)
		if len(data) > 0 {
			for k, v := range data {
				if useAllowed {
					if _, ok := allowed[k]; !ok {
						continue
					}
				} else if l.config != nil {
					// if not using AllowedFields, respect OmitFields
					skip := false
					for _, om := range l.config.OmitFields {
						if om == k {
							skip = true
							break
						}
					}
					if skip {
						continue
					}
				}

				// prepare value string and optionally color status codes or URLs
				valStr := fmt.Sprintf("%v", v)
				valColor := ""
				if k == "status" {
					// try to detect integer status code from various types
					switch tv := v.(type) {
					case int:
						valColor = colorForStatusCode(tv)
					case int32:
						valColor = colorForStatusCode(int(tv))
					case int64:
						valColor = colorForStatusCode(int(tv))
					case float64:
						// if it's float but integer valued, treat as status
						ival := int(tv)
						if float64(ival) == tv {
							valColor = colorForStatusCode(ival)
						}
					case string:
						if code, err := strconv.Atoi(tv); err == nil {
							valColor = colorForStatusCode(code)
						}
					}
				}

				// color URL-like values for better visibility and clickability
				if valColor == "" {
					if k == "url" {
						valColor = colorBlue
					} else if vs, ok := v.(string); ok {
						lower := strings.ToLower(vs)
						if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
							valColor = colorBlue
						}
					}
				}

				if valColor != "" {
					out += fmt.Sprintf(" %s=%s", k, l.applyColor(valStr, valColor))
				} else {
					out += fmt.Sprintf(" %s=%v", k, v)
				}
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
