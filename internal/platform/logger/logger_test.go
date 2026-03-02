package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	logger "forum/internal/platform/logger"
)

func unmarshalLastJSONLine(t *testing.T, b *bytes.Buffer) map[string]any {
	t.Helper()
	out := b.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		t.Fatalf("no log lines found")
	}
	last := lines[len(lines)-1]
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(last), &m); err != nil {
		t.Fatalf("failed to unmarshal JSON log: %v\nline: %s", err, last)
	}
	return m
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.InfoLevel, &buf)

	l.Info("hello", logger.String("user", "alice"), logger.Int("n", 5))

	m := unmarshalLastJSONLine(t, &buf)

	// level and msg
	if m["level"] != "INFO" {
		t.Fatalf("expected level INFO, got %v", m["level"])
	}
	if m["msg"] != "hello" {
		t.Fatalf("expected msg 'hello', got %v", m["msg"])
	}
	if _, ok := m["ts"]; !ok {
		t.Fatalf("expected ts field present")
	}

	// fields
	f, ok := m["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected fields map, got %T", m["fields"])
	}
	if f["user"] != "alice" {
		t.Fatalf("expected user 'alice', got %v", f["user"])
	}
	// JSON numbers decode as float64
	if n, ok := f["n"].(float64); !ok || int(n) != 5 {
		t.Fatalf("expected n 5, got %v (type %T)", f["n"], f["n"])
	}
}

func TestWithFieldsMergeOverride(t *testing.T) {
	var buf bytes.Buffer
	base := logger.New(logger.InfoLevel, &buf)

	// Add persistent fields
	gl := base.WithFields(logger.String("k", "v1"), logger.Int("n", 1))

	// Override k at call-site
	gl.Info("msg", logger.String("k", "v2"), logger.String("another", "val"))

	m := unmarshalLastJSONLine(t, &buf)

	f, ok := m["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected fields map, got %T", m["fields"])
	}

	if f["k"] != "v2" {
		t.Fatalf("expected k overridden to v2, got %v", f["k"])
	}
	// persistent int field should be present
	if n, ok := f["n"].(float64); !ok || int(n) != 1 {
		t.Fatalf("expected n 1, got %v", f["n"])
	}
	if f["another"] != "val" {
		t.Fatalf("expected another 'val', got %v", f["another"])
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.WarnLevel, &buf)

	l.Debug("debug msg")
	l.Info("info msg")
	l.Warn("warn msg")

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line logged, got %d: %v", len(lines), lines)
	}
	// ensure the single line is WARN
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if m["level"] != "WARN" {
		t.Fatalf("expected level WARN, got %v", m["level"])
	}
}

func TestConcurrentSafety(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.InfoLevel, &buf)

	const goroutines = 20
	const perGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				l.Info("concurrent", logger.Int("g", id), logger.Int("i", j))
			}
		}(i)
	}
	wg.Wait()

	// count non-empty lines
	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatalf("no output from concurrent logging")
	}
	lines := strings.Split(out, "\n")
	if got := len(lines); got != goroutines*perGoroutine {
		t.Fatalf("expected %d log lines, got %d", goroutines*perGoroutine, got)
	}
}

func TestJSONMarshalFallback(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.InfoLevel, &buf)

	// non-marshallable value (func) should trigger fallback
	l.Info("bad", logger.Any("bad", func() {}))

	out := buf.String()
	if !strings.Contains(out, "INFO bad") {
		t.Fatalf("expected fallback text containing 'INFO bad', got: %s", out)
	}
	// fallback uses fmt to print the map; ensure the 'bad' key is present in printed map
	if !strings.Contains(out, "bad:") && !strings.Contains(out, "map[bad") {
		t.Fatalf("expected fallback output to show key 'bad', got: %s", out)
	}
}

func TestErrorFieldAndTypes(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.InfoLevel, &buf)

	err := errors.New("boom")
	l.Error("oops", logger.Error(err))

	m := unmarshalLastJSONLine(t, &buf)

	f, ok := m["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected fields map, got %T", m["fields"])
	}
	if f["error"] != "boom" {
		t.Fatalf("expected error field 'boom', got %v", f["error"])
	}
}

func TestErrorFieldNilDoesNotPanic(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.InfoLevel, &buf)

	l.Error("noop", logger.Error(nil))

	m := unmarshalLastJSONLine(t, &buf)
	f, ok := m["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected fields map, got %T", m["fields"])
	}
	if f["error"] != "" {
		t.Fatalf("expected empty error field for nil error, got %v", f["error"])
	}
}

func TestTimestampIsRecent(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.InfoLevel, &buf)

	l.Info("tscheck")

	m := unmarshalLastJSONLine(t, &buf)
	tsStr, ok := m["ts"].(string)
	if !ok {
		t.Fatalf("ts not a string: %T", m["ts"])
	}
	parsed, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		t.Fatalf("failed to parse ts: %v", err)
	}
	if time.Since(parsed) > time.Minute {
		t.Fatalf("timestamp too old: %v", parsed)
	}
}

func TestDurationField(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.InfoLevel, &buf)

	// Log with duration field
	l.Info("operation completed", logger.Duration("elapsed_ms", 123*time.Millisecond))

	m := unmarshalLastJSONLine(t, &buf)
	f, ok := m["fields"].(map[string]any)
	if !ok {
		t.Fatalf("fields not a map")
	}

	elapsed, ok := f["elapsed_ms"].(float64)
	if !ok {
		t.Fatalf("elapsed_ms not a float64: %T", f["elapsed_ms"])
	}

	if elapsed != 123 {
		t.Errorf("elapsed_ms = %v, want 123", elapsed)
	}
}
