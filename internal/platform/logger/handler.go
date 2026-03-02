// OUTPUT ADAPTER - slog.Handler implementations for forum logger.
// This file provides custom slog.Handler implementations that maintain
// the forum's specific output formats (JSON with "fields" nesting and
// human-readable with colors).
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// JSON Handler
// ---------------------------------------------------------------------------

// jsonHandler outputs logs in JSON format compatible with the forum's
// original JSON output: {"level":"INFO","msg":"...","ts":"...","fields":{...}}
type jsonHandler struct {
	output io.Writer
	mu     *sync.Mutex
	attrs  []slog.Attr
	level  *slog.LevelVar
}

func newJSONHandler(output io.Writer, level *slog.LevelVar) *jsonHandler {
	return &jsonHandler{
		output: output,
		mu:     &sync.Mutex{},
		level:  level,
	}
}

func (h *jsonHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *jsonHandler) Handle(_ context.Context, r slog.Record) error {
	// Build fields map: pre-set attrs first, then record attrs (overrides).
	data := make(map[string]any)
	for _, a := range h.attrs {
		data[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		data[a.Key] = a.Value.Any()
		return true
	})

	entry := map[string]any{
		"level": r.Level.String(),
		"msg":   r.Message,
		"ts":    r.Time.Format(time.RFC3339Nano),
	}
	if len(data) > 0 {
		entry["fields"] = data
	}

	enc, err := json.Marshal(entry)
	if err != nil {
		// fallback to fmt if JSON encoding fails (e.g. func values)
		fallback := fmt.Sprintf("%s %s %v\n", r.Level.String(), r.Message, data)
		h.mu.Lock()
		_, _ = h.output.Write([]byte(fallback))
		h.mu.Unlock()
		return nil
	}
	enc = append(enc, '\n')
	h.mu.Lock()
	_, _ = h.output.Write(enc)
	h.mu.Unlock()
	return nil
}

func (h *jsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)
	return &jsonHandler{
		output: h.output,
		mu:     h.mu,
		attrs:  newAttrs,
		level:  h.level,
	}
}

func (h *jsonHandler) WithGroup(_ string) slog.Handler {
	return h // groups not used in our API
}

// ---------------------------------------------------------------------------
// Human-Readable Handler
// ---------------------------------------------------------------------------

// humanHandler outputs logs in a human-readable format with optional ANSI
// colors, field filtering, and compact HTTP-request formatting.
type humanHandler struct {
	output io.Writer
	mu     *sync.Mutex
	attrs  []slog.Attr
	config *Config
	level  *slog.LevelVar
}

func newHumanHandler(output io.Writer, config *Config, level *slog.LevelVar) *humanHandler {
	return &humanHandler{
		output: output,
		mu:     &sync.Mutex{},
		config: config,
		level:  level,
	}
}

func (h *humanHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *humanHandler) Handle(_ context.Context, r slog.Record) error {
	level := slogToLevel(r.Level)

	// Build data map: pre-set attrs first, then record attrs (overrides).
	data := make(map[string]any)
	for _, a := range h.attrs {
		data[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		data[a.Key] = a.Value.Any()
		return true
	})

	// Format timestamp
	ts := r.Time.Format(time.RFC3339Nano)
	if h.config != nil && h.config.TimePrecision == TimePrecisionSeconds {
		ts = r.Time.Format("15:04:05")
	}

	// HTTP request compact formatting
	if r.Message == "http.request" {
		out := formatHTTPRequest(ts, sanitizeFieldValuesForPlainText(data), h.config)
		out += "\n"
		h.mu.Lock()
		_, _ = h.output.Write([]byte(out))
		h.mu.Unlock()
		return nil
	}

	// Prepare field filtering
	allowed := map[string]struct{}{}
	useAllowed := false
	if h.config != nil && len(h.config.AllowedFields) > 0 {
		for _, k := range h.config.AllowedFields {
			allowed[k] = struct{}{}
		}
		allowed[""] = struct{}{} // Always allow empty keys for prefix-less fields
		useAllowed = true
	}

	// Colorize level label
	levelLabel := fmt.Sprintf("[%s]", levelToString(level))
	levelColored := applyColor(levelLabel, colorForLevel(level), h.config)

	// Colorize message when it's important (errors, warnings, etc.)
	msgStr := sanitizePlainText(r.Message)
	msgColor := colorForMessage(level)
	if msgColor != "" {
		msgStr = applyColor(msgStr, msgColor, h.config)
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
			} else if h.config != nil {
				// if not using AllowedFields, respect OmitFields
				if slices.Contains(h.config.OmitFields, k) {
					continue
				}
			}

			// prepare value string and optionally color status codes or URLs
			keyStr := sanitizePlainText(k)
			valStr := sanitizePlainText(fmt.Sprintf("%v", v))
			valColor := ""

			// color URL-like values for better visibility and clickability
			if vs, ok := v.(string); ok {
				lower := strings.ToLower(vs)
				if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
					valColor = colorBlue
				}
			}

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

			// color error values red
			if valColor == "" && k == "error" {
				valColor = colorRed
			}

			if k == "" {
				if valColor != "" {
					out += " " + applyColor(valStr, valColor, h.config)
				} else {
					out += " " + valStr
				}
			} else if valColor != "" {
				out += fmt.Sprintf(" %s:%s", keyStr, applyColor(valStr, valColor, h.config))
			} else {
				out += fmt.Sprintf(" %s:%s", keyStr, valStr)
			}
		}
	}

	// apply max line width truncation when requested
	if h.config != nil && h.config.MaxLineWidth > 0 {
		out = truncateToWidth(out, h.config.MaxLineWidth)
	}
	out += "\n"

	h.mu.Lock()
	_, _ = h.output.Write([]byte(out))
	h.mu.Unlock()
	return nil
}

func (h *humanHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)
	return &humanHandler{
		output: h.output,
		mu:     h.mu,
		attrs:  newAttrs,
		config: h.config,
		level:  h.level,
	}
}

func (h *humanHandler) WithGroup(_ string) slog.Handler {
	return h // groups not used in our API
}
