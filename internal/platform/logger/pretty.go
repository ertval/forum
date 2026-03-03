// Pretty formatting helpers for human-readable log output.
// These functions handle HTTP-request colouring, emoji selection, and
// formatted human output. They are separated from the core Logger to
// keep logger.go focused on the Logger struct, Field constructors, and
// the central log() method.
package logger

import (
	"fmt"
	"strings"
)

// applyColor applies colorization depending on the config.
func applyColor(s, color string, cfg *Config) string {
	if cfg == nil {
		// default to color enabled for terminal output
		if color == "" {
			return s
		}
		return color + s + colorReset
	}
	if !cfg.Colorize || color == "" {
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

// colorForMessage returns a highlight colour for the log message based
// solely on the log level. This avoids false-positive highlighting caused
// by keyword matching (e.g. "no error found" being coloured red).
func colorForMessage(level Level) string {
	switch level {
	case ErrorLevel:
		return colorRed
	case WarnLevel:
		return colorYellow
	default:
		return ""
	}
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

// formatHTTPRequest creates a compact, colorful one-line log for HTTP requests.
// Format: TS PROTO STATUS METHOD PATH?QUERY (SIZEb, DURms) [IP]
// Example: 18:33:58 🔒 ✓ 200 GET /board?my_posts=true (6.4kb, 1ms) [127.0.0.1]
func formatHTTPRequest(ts string, data map[string]any, cfg *Config) string {
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
	protoPart := applyColor(protoIndicator, protoColor, cfg)

	statusPart := applyColor(fmt.Sprintf("%s %d", statusIndicator, status), statusColor, cfg)
	methodPart := applyColor(fmt.Sprintf("%-4s", method), methodColor, cfg)

	// Dim the metadata for less important info
	metaPart := applyColor(fmt.Sprintf("(%s, %dms)", sizeStr, durationMs), colorWhite, cfg)
	ipPart := applyColor(fmt.Sprintf("[%s]", ip), colorWhite, cfg)

	return fmt.Sprintf("%s %s %s %s %s %s %s", ts, protoPart, statusPart, methodPart, fullPath, metaPart, ipPart)
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
