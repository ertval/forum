package logger

import (
	"bytes"
	"strings"
	"testing"
)

func newHumanTestLogger(buf *bytes.Buffer) *Logger {
	return newLogger(InfoLevel, buf, true, &Config{
		MaxLineWidth: 0,
		Colorize:     false,
	})
}

func TestHumanOutput_SanitizesMessageAndFields(t *testing.T) {
	var buf bytes.Buffer
	l := newHumanTestLogger(&buf)

	l.Info("line1\nline2\rline3\tend", String("ke\ny", "va\rl\nue\t"))

	out := buf.String()
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("expected exactly one newline terminator, got output: %q", out)
	}
	if !strings.Contains(out, "line1\\nline2\\rline3\\tend") {
		t.Fatalf("expected escaped message content, got: %q", out)
	}
	if !strings.Contains(out, "ke\\ny:va\\rl\\nue\\t") {
		t.Fatalf("expected escaped field key/value, got: %q", out)
	}
}

func TestHumanOutput_HTTPRequestSanitizesStringFields(t *testing.T) {
	var buf bytes.Buffer
	l := newHumanTestLogger(&buf)

	l.Info("http.request",
		String("method", "GET\nINJECT"),
		String("path", "/board\nX"),
		String("query", "a=1\r\nb=2"),
		Int("status", 200),
		Int("size", 10),
		Int("duration_ms", 1),
		String("remote", "127.0.0.1:1234"),
		String("proto", "http"),
	)

	out := buf.String()
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("expected exactly one newline terminator, got output: %q", out)
	}
	if !strings.Contains(out, "GET\\nINJECT") {
		t.Fatalf("expected escaped method, got: %q", out)
	}
	if !strings.Contains(out, "/board\\nX?a=1\\r\\nb=2") {
		t.Fatalf("expected escaped path/query, got: %q", out)
	}
}
