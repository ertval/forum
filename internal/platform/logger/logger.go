// Package logger provides structured logging for the forum application.
// It supports multiple log levels and outputs logs in JSON format for production
// and human-readable format for development.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
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
		// default to show HTTP request info and other essential fields
		AllowedFields: []string{"method", "path", "query", "status", "size", "duration_ms", "remote", "url", "response", "error", "errors", "proto"},
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

// formatHTTPRequest creates a compact, colorful one-line log for HTTP requests.
// Format: TS PROTO STATUS METHOD PATH?QUERY (SIZEb, DURms) [IP]
// Example: 18:33:58 🔒 ✓ 200 GET /board?my_posts=true (6.4kb, 1ms) [127.0.0.1]
func (l *Logger) formatHTTPRequest(ts string, level Level, data map[string]any) string {
	// Extract fields
	method := getStringField(data, "method", "???")
	path := getStringField(data, "path", "/")
	query := getStringField(data, "query", "")
	status := getIntField(data, "status", 0)
	size := getIntField(data, "size", 0)
	durationMs := getIntField(data, "duration_ms", 0)
	remote := getStringField(data, "remote", "")
	proto := getStringField(data, "proto", "http")

	// Build the full path with query
	fullPath := path
	if query != "" {
		fullPath = path + "?" + query
	}

	// Status indicator and color
	statusColor := colorForStatusCode(status)
	var statusIndicator string
	if status >= 200 && status < 300 {
		statusIndicator = "✓"
	} else if status >= 400 && status < 500 {
		statusIndicator = "⚠"
	} else if status >= 500 {
		statusIndicator = "✗"
	} else if status >= 300 && status < 400 {
		statusIndicator = "→"
	} else {
		statusIndicator = "?"
	}

	// Format size in human-readable format
	sizeStr := formatBytes(size)

	// Color method based on type
	methodColor := colorForMethod(method)

	// Extract just IP without port for cleaner output
	ip := remote
	if idx := strings.LastIndex(remote, ":"); idx > 0 {
		ip = remote[:idx]
	}

	// Build compact output
	// Format: TS PROTO INDICATOR STATUS METHOD PATH (SIZE, DUR) [IP]
	// Protocol indicator: 🔒 for HTTPS, 🔓 for HTTP
	var protoIndicator string
	var protoColor string
	if proto == "https" {
		protoIndicator = "🔒"
		protoColor = colorGreen
	} else {
		protoIndicator = "🔓"
		protoColor = colorYellow
	}
	protoPart := l.applyColor(protoIndicator, protoColor)

	statusPart := l.applyColor(fmt.Sprintf("%s %d", statusIndicator, status), statusColor)
	methodPart := l.applyColor(fmt.Sprintf("%-4s", method), methodColor)

	// Dim the metadata for less important info
	metaPart := l.applyColor(fmt.Sprintf("(%s, %dms)", sizeStr, durationMs), colorWhite)
	ipPart := l.applyColor(fmt.Sprintf("[%s]", ip), colorWhite)

	return fmt.Sprintf("%s %s %s %s %s %s %s", ts, protoPart, statusPart, methodPart, fullPath, metaPart, ipPart)
}

// colorForMethod returns a color based on HTTP method.
func colorForMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return colorGreen
	case "POST":
		return colorBlue
	case "PUT", "PATCH":
		return colorYellow
	case "DELETE":
		return colorRed
	default:
		return colorWhite
	}
}

// formatBytes converts bytes to human-readable format.
func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%db", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fkb", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1fmb", float64(bytes)/(1024*1024))
	}
}

// getStringField extracts a string field from a map with a default value.
func getStringField(data map[string]any, key, defaultVal string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// getIntField extracts an int field from a map with a default value.
func getIntField(data map[string]any, key string, defaultVal int) int {
	if v, ok := data[key]; ok {
		switch tv := v.(type) {
		case int:
			return tv
		case int64:
			return int(tv)
		case int32:
			return int(tv)
		case float64:
			return int(tv)
		}
	}
	return defaultVal
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
				ts = parsed.Format("15:04:05") // Compact time format (HH:MM:SS only)
			}
		}

		// Check if this is an HTTP request log for compact formatting
		if msg == "http.request" {
			out := l.formatHTTPRequest(ts, level, data)
			out += "\n"
			_, _ = l.output.Write([]byte(out))
			return
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
			// iterate fields in sorted order for stable output
			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				v := data[k]
				if useAllowed {
					if _, ok := allowed[k]; !ok {
						continue
					}
				} else if l.config != nil {
					// if not using AllowedFields, respect OmitFields
					skip := slices.Contains(l.config.OmitFields, k)
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

				// color error values red
				if valColor == "" && k == "error" {
					valColor = colorRed
				}

				if valColor != "" {
					out += fmt.Sprintf(" %s:%s", k, l.applyColor(valStr, valColor))
				} else {
					out += fmt.Sprintf(" %s:%v", k, v)
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
